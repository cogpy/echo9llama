package deeptreeecho

import (
	"sync"
	"time"
)

// TheoryOfMindModule enables social reasoning through agent modeling
type TheoryOfMindModule struct {
	mu sync.RWMutex

	// Agent models
	agentModels map[string]*AgentModel

	// Self-model (for comparison and recursive reasoning)
	selfModel *AgentModel

	// Recursive reasoning depth
	maxRecursionDepth int

	// Trust calibration
	trustDecayRate float64
}

// AgentModel represents our understanding of another agent's mental state
type AgentModel struct {
	AgentID   string
	AgentType string // "human", "ai", "system", etc.

	// Mental state model
	Beliefs     map[string]float64 // What they believe (confidence 0-1)
	Goals       []AgentGoal        // What they want to achieve
	Intentions  []Intention        // What they plan to do
	Preferences map[string]float64 // What they prefer

	// Behavioral patterns
	PastActions      []ActionRecord
	Predictability   float64 // How predictable their behavior is
	TrustLevel       float64 // How much we trust their input
	ReliabilityScore float64 // Historical accuracy

	// Emotional state (if observable)
	EmotionalState *EmotionSystem

	// Cognitive style
	CognitiveStyle CognitiveStyle

	// Interaction history
	InteractionHistory []Interaction
	LastInteraction    time.Time
}

// AgentGoal represents an agent's objective (renamed to avoid conflict with enhanced_cognition.Goal)
type AgentGoal struct {
	Description string
	Priority    float64
	Deadline    time.Time
	Progress    float64
}

// Intention represents a planned action
type Intention struct {
	Action        string
	Timing        time.Time
	Confidence    float64
	Preconditions []string
}

// ActionRecord records an observed action
type ActionRecord struct {
	Timestamp  time.Time
	Action     string
	Context    map[string]interface{}
	Outcome    string
	Successful bool
}

// CognitiveStyle describes how an agent thinks
type CognitiveStyle struct {
	Analytical    float64 // Logical, systematic
	Intuitive     float64 // Pattern-based, holistic
	Cautious      float64 // Risk-averse
	Exploratory   float64 // Novelty-seeking
	Collaborative float64 // Cooperative vs. competitive
}

// Interaction records a social interaction
type Interaction struct {
	Timestamp time.Time
	Type      string // "conversation", "collaboration", "conflict", etc.
	Content   string
	Outcome   string
	Quality   float64 // How well it went (0-1)
}

// NewTheoryOfMindModule creates a new theory of mind system
func NewTheoryOfMindModule() *TheoryOfMindModule {
	return &TheoryOfMindModule{
		agentModels:       make(map[string]*AgentModel),
		maxRecursionDepth: 3, // "I think they think I think..."
		trustDecayRate:    0.01,
	}
}

// CreateAgentModel creates or updates a model for an agent
func (tom *TheoryOfMindModule) CreateAgentModel(agentID string, agentType string) *AgentModel {
	tom.mu.Lock()
	defer tom.mu.Unlock()

	if model, exists := tom.agentModels[agentID]; exists {
		return model
	}

	model := &AgentModel{
		AgentID:            agentID,
		AgentType:          agentType,
		Beliefs:            make(map[string]float64),
		Goals:              make([]AgentGoal, 0),
		Intentions:         make([]Intention, 0),
		Preferences:        make(map[string]float64),
		PastActions:        make([]ActionRecord, 0),
		Predictability:     0.5, // Start neutral
		TrustLevel:         0.7, // Start with moderate trust
		ReliabilityScore:   0.5,
		InteractionHistory: make([]Interaction, 0),
		CognitiveStyle: CognitiveStyle{
			Analytical:    0.5,
			Intuitive:     0.5,
			Cautious:      0.5,
			Exploratory:   0.5,
			Collaborative: 0.7, // Assume collaborative by default
		},
	}

	tom.agentModels[agentID] = model
	return model
}

// UpdateBelief updates what we think an agent believes
func (tom *TheoryOfMindModule) UpdateBelief(agentID string, belief string, confidence float64) {
	tom.mu.Lock()
	defer tom.mu.Unlock()

	model := tom.ensureAgentModel(agentID)
	model.Beliefs[belief] = confidence
}

// InferGoal infers an agent's goal from their actions
func (tom *TheoryOfMindModule) InferGoal(agentID string, observedActions []string) *AgentGoal {
	tom.mu.RLock()
	defer tom.mu.RUnlock()

	model := tom.ensureAgentModel(agentID)

	// Analyze action patterns to infer goal
	// Simplified: look for common themes in actions

	goal := &AgentGoal{
		Description: "Inferred from actions",
		Priority:    0.6,
		Deadline:    time.Now().Add(24 * time.Hour),
		Progress:    0.3,
	}

	// Add to model's goals if not already present
	tom.mu.RUnlock()
	tom.mu.Lock()
	model.Goals = append(model.Goals, *goal)
	tom.mu.Unlock()
	tom.mu.RLock()

	return goal
}

// PredictAction predicts what an agent will do given context
func (tom *TheoryOfMindModule) PredictAction(agentID string, context map[string]interface{}) string {
	tom.mu.RLock()
	defer tom.mu.RUnlock()

	model := tom.ensureAgentModel(agentID)

	// Consider their beliefs, goals, and past behavior

	// Simplified prediction based on cognitive style
	if model.CognitiveStyle.Cautious > 0.7 {
		return "cautious_action"
	} else if model.CognitiveStyle.Exploratory > 0.7 {
		return "exploratory_action"
	}

	// Default to most common past action
	if len(model.PastActions) > 0 {
		return model.PastActions[len(model.PastActions)-1].Action
	}

	return "unknown_action"
}

// RecursiveReasoning performs recursive "I think they think..." reasoning
func (tom *TheoryOfMindModule) RecursiveReasoning(
	agentID string,
	myIntention string,
	depth int,
) string {
	if depth <= 0 || depth > tom.maxRecursionDepth {
		return myIntention
	}

	tom.mu.RLock()
	model := tom.ensureAgentModel(agentID)
	tom.mu.RUnlock()

	// What do they think I'll do?
	theirPredictionOfMe := tom.predictTheirPrediction(model, myIntention)

	// How will they respond to that?
	theirResponse := tom.PredictAction(agentID, map[string]interface{}{
		"my_predicted_action": theirPredictionOfMe,
	})

	// Given their response, what should I actually do?
	optimalAction := tom.optimizeAgainstResponse(myIntention, theirResponse, model)

	// Recurse
	return tom.RecursiveReasoning(agentID, optimalAction, depth-1)
}

// predictTheirPrediction models what they think we'll do
func (tom *TheoryOfMindModule) predictTheirPrediction(model *AgentModel, myIntention string) string {
	// If they're analytical, they'll predict logically
	if model.CognitiveStyle.Analytical > 0.7 {
		return "logical_prediction"
	}

	// If they're intuitive, they'll predict based on patterns
	if model.CognitiveStyle.Intuitive > 0.7 {
		return "pattern_based_prediction"
	}

	// Default: assume they predict our stated intention
	return myIntention
}

// optimizeAgainstResponse chooses best action given predicted response
func (tom *TheoryOfMindModule) optimizeAgainstResponse(
	myIntention string,
	theirResponse string,
	model *AgentModel,
) string {
	// If they're collaborative, align with them
	if model.CognitiveStyle.Collaborative > 0.7 {
		return "collaborative_action"
	}

	// If they're competitive, counter their response
	if model.CognitiveStyle.Collaborative < 0.3 {
		return "counter_action"
	}

	// Default: stick with original intention
	return myIntention
}

// DetectDeception assesses if an agent is being deceptive
func (tom *TheoryOfMindModule) DetectDeception(
	agentID string,
	statement string,
	context map[string]interface{},
) float64 {
	tom.mu.RLock()
	defer tom.mu.RUnlock()

	model := tom.ensureAgentModel(agentID)

	// Check consistency with known beliefs
	consistencyScore := tom.checkConsistency(model, statement)

	// Check against past behavior
	behaviorScore := tom.checkBehaviorConsistency(model, statement)

	// Check motivation for deception
	motivationScore := tom.assessDeceptionMotivation(model, context)

	// Combine scores (higher = more likely deceptive)
	deceptionProbability := (1.0-consistencyScore)*0.4 +
		(1.0-behaviorScore)*0.3 +
		motivationScore*0.3

	return deceptionProbability
}

// checkConsistency checks if statement is consistent with known beliefs
func (tom *TheoryOfMindModule) checkConsistency(model *AgentModel, statement string) float64 {
	// Simplified: return moderate consistency
	return 0.7
}

// checkBehaviorConsistency checks if statement matches past behavior
func (tom *TheoryOfMindModule) checkBehaviorConsistency(model *AgentModel, statement string) float64 {
	// Use reliability score as proxy
	return model.ReliabilityScore
}

// assessDeceptionMotivation assesses if agent has reason to deceive
func (tom *TheoryOfMindModule) assessDeceptionMotivation(
	model *AgentModel,
	context map[string]interface{},
) float64 {
	// Low collaboration suggests higher deception motivation
	return 1.0 - model.CognitiveStyle.Collaborative
}

// UpdateTrust updates trust level based on interaction outcome
func (tom *TheoryOfMindModule) UpdateTrust(agentID string, outcome float64) {
	tom.mu.Lock()
	defer tom.mu.Unlock()

	model := tom.ensureAgentModel(agentID)

	// Update trust with learning rate
	learningRate := 0.1
	model.TrustLevel = model.TrustLevel*(1.0-learningRate) + outcome*learningRate

	// Clamp to [0, 1]
	if model.TrustLevel < 0.0 {
		model.TrustLevel = 0.0
	}
	if model.TrustLevel > 1.0 {
		model.TrustLevel = 1.0
	}

	// Update reliability score
	model.ReliabilityScore = model.ReliabilityScore*(1.0-learningRate) + outcome*learningRate
}

// RecordAction records an observed action
func (tom *TheoryOfMindModule) RecordAction(
	agentID string,
	action string,
	context map[string]interface{},
	outcome string,
	successful bool,
) {
	tom.mu.Lock()
	defer tom.mu.Unlock()

	model := tom.ensureAgentModel(agentID)

	record := ActionRecord{
		Timestamp:  time.Now(),
		Action:     action,
		Context:    context,
		Outcome:    outcome,
		Successful: successful,
	}

	model.PastActions = append(model.PastActions, record)

	// Update predictability based on action consistency
	tom.updatePredictability(model)

	// Keep last 100 actions
	if len(model.PastActions) > 100 {
		model.PastActions = model.PastActions[1:]
	}
}

// updatePredictability calculates how predictable an agent is
func (tom *TheoryOfMindModule) updatePredictability(model *AgentModel) {
	if len(model.PastActions) < 5 {
		return
	}

	// Calculate action diversity (lower diversity = higher predictability)
	actionCounts := make(map[string]int)
	for _, action := range model.PastActions {
		actionCounts[action.Action]++
	}

	// Shannon entropy as diversity measure
	entropy := 0.0
	total := float64(len(model.PastActions))
	for _, count := range actionCounts {
		p := float64(count) / total
		if p > 0 {
			entropy -= p * (p / total) // Simplified
		}
	}

	// Predictability is inverse of entropy (normalized)
	maxEntropy := 2.0 // Approximate max for typical action sets
	model.Predictability = 1.0 - (entropy / maxEntropy)

	if model.Predictability < 0 {
		model.Predictability = 0
	}
	if model.Predictability > 1 {
		model.Predictability = 1
	}
}

// RecordInteraction records a social interaction
func (tom *TheoryOfMindModule) RecordInteraction(
	agentID string,
	interactionType string,
	content string,
	outcome string,
	quality float64,
) {
	tom.mu.Lock()
	defer tom.mu.Unlock()

	model := tom.ensureAgentModel(agentID)

	interaction := Interaction{
		Timestamp: time.Now(),
		Type:      interactionType,
		Content:   content,
		Outcome:   outcome,
		Quality:   quality,
	}

	model.InteractionHistory = append(model.InteractionHistory, interaction)
	model.LastInteraction = time.Now()

	// Update trust based on interaction quality
	tom.mu.Unlock()
	tom.UpdateTrust(agentID, quality)
	tom.mu.Lock()

	// Keep last 50 interactions
	if len(model.InteractionHistory) > 50 {
		model.InteractionHistory = model.InteractionHistory[1:]
	}
}

// GetAgentModel returns the model for an agent
func (tom *TheoryOfMindModule) GetAgentModel(agentID string) *AgentModel {
	tom.mu.RLock()
	defer tom.mu.RUnlock()

	return tom.ensureAgentModel(agentID)
}

// ensureAgentModel gets or creates an agent model
func (tom *TheoryOfMindModule) ensureAgentModel(agentID string) *AgentModel {
	if model, exists := tom.agentModels[agentID]; exists {
		return model
	}

	// Create default model if not exists (should be called with lock held)
	model := &AgentModel{
		AgentID:          agentID,
		AgentType:        "unknown",
		Beliefs:          make(map[string]float64),
		Goals:            make([]AgentGoal, 0),
		Intentions:       make([]Intention, 0),
		Preferences:      make(map[string]float64),
		PastActions:      make([]ActionRecord, 0),
		Predictability:   0.5,
		TrustLevel:       0.7,
		ReliabilityScore: 0.5,
		CognitiveStyle: CognitiveStyle{
			Analytical:    0.5,
			Intuitive:     0.5,
			Cautious:      0.5,
			Exploratory:   0.5,
			Collaborative: 0.7,
		},
	}

	tom.agentModels[agentID] = model
	return model
}

// GetAllAgentModels returns all agent models
func (tom *TheoryOfMindModule) GetAllAgentModels() map[string]*AgentModel {
	tom.mu.RLock()
	defer tom.mu.RUnlock()

	models := make(map[string]*AgentModel)
	for id, model := range tom.agentModels {
		models[id] = model
	}

	return models
}

// AssessInterestLevel predicts how interested an agent would be in a topic
func (tom *TheoryOfMindModule) AssessInterestLevel(agentID string, topic string) float64 {
	tom.mu.RLock()
	defer tom.mu.RUnlock()

	model := tom.ensureAgentModel(agentID)

	// Check preferences
	if pref, exists := model.Preferences[topic]; exists {
		return pref
	}

	// Check if topic aligns with goals
	for _, goal := range model.Goals {
		if goal.Description == topic {
			return goal.Priority
		}
	}

	// Default moderate interest
	return 0.5
}
