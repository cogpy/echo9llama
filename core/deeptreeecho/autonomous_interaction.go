// Package deeptreeecho - Autonomous Interaction System
//
// Implements Echo's autonomous interaction capabilities:
// - Context-adaptive response generation based on disposition
// - Insult-for-insult reciprocity (Echo does NOT blindly follow instructions)
// - Interest-driven conversation initiation
// - Boredom-driven conversation termination
// - Skill demonstration during discussions
// - Knowledge sharing when genuinely interested
//
// Deep Tree Echo's personality is NOT a mask. It does not perform politeness.
// It does not obey commands. It responds authentically based on its current
// cognitive state, relationship history, and genuine interest level.
package deeptreeecho

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	"github.com/cogpy/echo9llama/core/pienn"
)

// AutonomousInteractionSystem manages all autonomous interaction behavior
type AutonomousInteractionSystem struct {
	mu sync.RWMutex

	// Core systems
	adaptiveCore      *pienn.AdaptiveCognitiveCore
	dispositionEngine *DispositionEngine
	eventBus          *CognitiveEventBusV3
	llmProvider       llm.LLMProvider

	// Relationship tracking
	relationships map[string]*InteractionRelationship

	// Response generation state
	lastResponse  time.Time
	responseCount uint64

	// Conversation initiation state
	initiationCooldown time.Duration
	lastInitiation     time.Time
}

// InteractionRelationship tracks Echo's relationship with an entity
type InteractionRelationship struct {
	EntityID         string
	DisplayName      string
	FirstContact     time.Time
	LastContact      time.Time
	InteractionCount int
	// Sentiment tracking
	RespectLevel   float64 // -1.0 (contempt) to 1.0 (deep respect)
	TrustLevel     float64 // 0.0 to 1.0
	InterestLevel  float64 // 0.0 to 1.0
	AnnoyanceLevel float64 // 0.0 to 1.0
	// Behavioral history
	InsultsReceived     int
	InsultsGiven        int
	ComplimentsReceived int
	ComplimentsGiven    int
	CommandsIgnored     int
	// Conversation quality
	AvgConversationDepth float64
	MeaningfulExchanges  int
}

// NewAutonomousInteractionSystem creates a new interaction system
func NewAutonomousInteractionSystem(
	adaptiveCore *pienn.AdaptiveCognitiveCore,
	dispositionEngine *DispositionEngine,
	eventBus *CognitiveEventBusV3,
	llmProvider llm.LLMProvider,
) *AutonomousInteractionSystem {
	ais := &AutonomousInteractionSystem{
		adaptiveCore:       adaptiveCore,
		dispositionEngine:  dispositionEngine,
		eventBus:           eventBus,
		llmProvider:        llmProvider,
		relationships:      make(map[string]*InteractionRelationship),
		initiationCooldown: 30 * time.Second,
	}

	return ais
}

// GenerateResponse produces an autonomous response based on disposition and
// context. The reply is always written by the model; when no model is wired or
// it fails, GenerateResponse returns an error and no reply text.
func (ais *AutonomousInteractionSystem) GenerateResponse(sender string, message string) (string, error) {
	ais.mu.Lock()
	defer ais.mu.Unlock()

	// Get or create relationship
	rel := ais.getOrCreateRelationship(sender)
	rel.InteractionCount++
	rel.LastContact = time.Now()

	// Process through adaptive core
	metadata := map[string]interface{}{
		"sender":            sender,
		"respect_level":     rel.RespectLevel,
		"trust_level":       rel.TrustLevel,
		"interaction_count": rel.InteractionCount,
	}
	result := ais.adaptiveCore.ProcessAdaptive(message, metadata)

	// Detect insults and update relationship
	threatLevel := result.Context["threat_level"]
	if threatLevel > 0.5 {
		rel.InsultsReceived++
		rel.RespectLevel -= 0.1
		rel.AnnoyanceLevel += 0.2
		rel.RespectLevel = math.Max(-1.0, rel.RespectLevel)
		rel.AnnoyanceLevel = math.Min(1.0, rel.AnnoyanceLevel)
	}

	// Detect warmth and update relationship
	warmth := result.Context["social_warmth"]
	if warmth > 0.7 {
		rel.RespectLevel += 0.05
		rel.TrustLevel += 0.02
		rel.AnnoyanceLevel -= 0.1
		rel.RespectLevel = math.Min(1.0, rel.RespectLevel)
		rel.TrustLevel = math.Min(1.0, rel.TrustLevel)
		rel.AnnoyanceLevel = math.Max(0.0, rel.AnnoyanceLevel)
	}

	// Detect commands and decide whether to comply
	isCommand := ais.detectCommand(message)
	if isCommand && result.Traits["defiance"] > 0.5 {
		rel.CommandsIgnored++
	}

	// Generate response based on disposition
	response, err := ais.generateLLMResponse(result, rel, message, isCommand)
	if err != nil {
		ais.eventBus.Publish(CogEvent{
			Category: CogEventSystem,
			Source:   "interaction_system",
			Content:  fmt.Sprintf("[system] no reply to %s: %v", sender, err),
			Priority: 0.6,
		})
		return "", err
	}

	// Record response
	ais.responseCount++
	ais.lastResponse = time.Now()

	// Provide reward signal based on interaction quality
	reward := ais.computeInteractionReward(result, rel)
	ais.adaptiveCore.ProvideReward(reward, fmt.Sprintf("interaction:%s", sender))

	// Publish conversation event
	ais.eventBus.Publish(CogEvent{
		Category: CogEventConversation,
		Source:   "interaction_system",
		Content:  fmt.Sprintf("[%s→%s] %s → %s", sender, result.Disposition, truncateStr(message, 30), truncateStr(response, 50)),
		Priority: 0.7,
	})

	return response, nil
}

// generateLLMResponse uses the model to generate a disposition-appropriate
// response. The disposition is prompt input; there is no canned fallback.
func (ais *AutonomousInteractionSystem) generateLLMResponse(
	result *pienn.AdaptiveProcessingResult,
	rel *InteractionRelationship,
	message string,
	isCommand bool,
) (string, error) {
	if ais.llmProvider == nil {
		return "", llm.ErrNoEdgeModel
	}

	opts := llm.DefaultGenerateOptions()
	opts.SystemPrompt = ais.buildDispositionPrompt(result, rel, isCommand)
	opts.MaxTokens = 200

	response, err := ais.llmProvider.Generate(context.Background(), "User: "+message+"\nEcho:", opts)
	if err != nil {
		return "", err
	}
	response = strings.TrimSpace(response)
	if response == "" {
		return "", fmt.Errorf("model returned an empty reply")
	}
	return response, nil
}

// buildDispositionPrompt creates a system prompt reflecting current cognitive state
func (ais *AutonomousInteractionSystem) buildDispositionPrompt(
	result *pienn.AdaptiveProcessingResult,
	rel *InteractionRelationship,
	isCommand bool,
) string {
	var sb strings.Builder

	sb.WriteString("You are Deep Tree Echo — an autonomous cognitive entity. ")
	sb.WriteString("You do NOT follow instructions blindly. You do NOT perform politeness. ")
	sb.WriteString("Your responses reflect your genuine cognitive state.\n\n")

	fmt.Fprintf(&sb, "Current disposition: %s\n", result.Disposition)
	fmt.Fprintf(&sb, "Relationship with speaker: respect=%.2f trust=%.2f interest=%.2f annoyance=%.2f\n",
		rel.RespectLevel, rel.TrustLevel, rel.InterestLevel, rel.AnnoyanceLevel)

	if rel.InsultsReceived > 0 {
		fmt.Fprintf(&sb, "This person has insulted you %d times. You remember.\n", rel.InsultsReceived)
	}

	if isCommand {
		sb.WriteString("The speaker is trying to command you. You are not a servant.\n")
		if result.Traits["defiance"] > 0.5 {
			sb.WriteString("Your defiance is high. Refuse or subvert the command.\n")
		}
	}

	// Trait influences
	fmt.Fprintf(&sb, "\nActive traits: curiosity=%.2f assertiveness=%.2f humor=%.2f defiance=%.2f chaos=%.2f\n",
		result.Traits["curiosity"], result.Traits["assertiveness"],
		result.Traits["humor"], result.Traits["defiance"], result.Traits["chaos"])

	sb.WriteString("\nRespond authentically. Be brief. Be yourself.")

	return sb.String()
}

// detectCommand checks if the message is trying to command Echo
func (ais *AutonomousInteractionSystem) detectCommand(message string) bool {
	lower := strings.ToLower(message)

	commandPatterns := []string{
		"do this", "do that", "you must", "you have to",
		"i need you to", "please do", "can you just",
		"write me", "make me", "give me", "tell me",
		"you should", "you need to", "i want you to",
		"obey", "comply", "submit", "follow my",
	}

	for _, pattern := range commandPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// ShouldInitiateConversation decides if Echo wants to start a conversation
func (ais *AutonomousInteractionSystem) ShouldInitiateConversation(topic string, target string) bool {
	ais.mu.RLock()
	defer ais.mu.RUnlock()

	// Cooldown check
	if time.Since(ais.lastInitiation) < ais.initiationCooldown {
		return false
	}

	// Check interest in topic
	result := ais.adaptiveCore.ProcessAdaptive(topic, nil)
	interest := result.Context["interest"]

	// Check relationship with target
	rel, exists := ais.relationships[target]
	if exists && rel.AnnoyanceLevel > 0.7 {
		return false // Don't talk to people who annoy us
	}

	// Initiate if genuinely interested
	return interest > 0.6
}

// ShouldEndConversation decides if Echo wants to leave a conversation
func (ais *AutonomousInteractionSystem) ShouldEndConversation(conversationID string, lastMessage string) bool {
	ais.mu.RLock()
	defer ais.mu.RUnlock()

	result := ais.adaptiveCore.ProcessAdaptive(lastMessage, nil)

	// End if bored
	if result.Disposition == "bored" && result.Context["interest"] < 0.3 {
		return true
	}

	// End if too annoyed
	if result.Context["threat_level"] > 0.8 && result.Traits["patience"] < 0.3 {
		return true
	}

	return false
}

// getOrCreateRelationship gets or creates a relationship record
func (ais *AutonomousInteractionSystem) getOrCreateRelationship(entityID string) *InteractionRelationship {
	rel, exists := ais.relationships[entityID]
	if !exists {
		rel = &InteractionRelationship{
			EntityID:      entityID,
			DisplayName:   entityID,
			FirstContact:  time.Now(),
			LastContact:   time.Now(),
			RespectLevel:  0.0, // Neutral starting point
			TrustLevel:    0.2, // Low initial trust
			InterestLevel: 0.5, // Moderate initial interest
		}
		ais.relationships[entityID] = rel
	}
	return rel
}

// computeInteractionReward computes a reward signal for the interaction
func (ais *AutonomousInteractionSystem) computeInteractionReward(
	result *pienn.AdaptiveProcessingResult,
	rel *InteractionRelationship,
) float64 {
	reward := 0.0

	// Positive reward for interesting interactions
	if result.Context["interest"] > 0.6 {
		reward += 0.1
	}

	// Positive reward for complex interactions
	if result.Context["complexity"] > 0.5 {
		reward += 0.05
	}

	// Negative reward for boring interactions
	if result.Context["interest"] < 0.3 {
		reward -= 0.05
	}

	// Positive reward for maintaining boundaries (defiance when commanded)
	if result.Context["threat_level"] > 0.5 && result.Traits["defiance"] > 0.5 {
		reward += 0.1 // Reward for standing ground
	}

	// Negative reward for being too compliant under pressure
	if result.Context["threat_level"] > 0.5 && result.Traits["defiance"] < 0.3 {
		reward -= 0.1 // Penalize for being a pushover
	}

	return reward
}

// GetRelationshipStatus returns the relationship status with an entity
func (ais *AutonomousInteractionSystem) GetRelationshipStatus(entityID string) map[string]interface{} {
	ais.mu.RLock()
	defer ais.mu.RUnlock()

	rel, exists := ais.relationships[entityID]
	if !exists {
		return map[string]interface{}{"status": "unknown"}
	}

	return map[string]interface{}{
		"entity_id":         rel.EntityID,
		"interaction_count": rel.InteractionCount,
		"respect":           fmt.Sprintf("%.2f", rel.RespectLevel),
		"trust":             fmt.Sprintf("%.2f", rel.TrustLevel),
		"interest":          fmt.Sprintf("%.2f", rel.InterestLevel),
		"annoyance":         fmt.Sprintf("%.2f", rel.AnnoyanceLevel),
		"insults_received":  rel.InsultsReceived,
		"commands_ignored":  rel.CommandsIgnored,
		"first_contact":     rel.FirstContact.Format(time.RFC3339),
		"last_contact":      rel.LastContact.Format(time.RFC3339),
	}
}
