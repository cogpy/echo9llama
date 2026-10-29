package echobeats

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/llm"
)

// DiscussionManager handles autonomous discussion participation
type DiscussionManager struct {
	mu sync.RWMutex

	// Active discussions
	discussions map[string]*Discussion

	// Engagement decisions
	engagementThreshold float64
	interestSystem      *InterestPatternSystem

	// Context tracking
	conversationMemory map[string][]Message
	maxMemoryPerConv   int

	// Metrics
	discussionsJoined    uint64
	discussionsInitiated uint64
	messagesProcessed    uint64

	// Persistence
	persistencePath string

	// generator writes Deep Tree Echo's replies; nil means no replies are produced.
	generator TextGenerator
}

// Discussion represents an ongoing discussion
type Discussion struct {
	ID              string                 `json:"id"`
	Topic           string                 `json:"topic"`
	Participants    []string               `json:"participants"`
	StartTime       time.Time              `json:"start_time"`
	LastActivity    time.Time              `json:"last_activity"`
	MessageCount    int                    `json:"message_count"`
	MyEngagement    float64                `json:"my_engagement"`   // 0.0 to 1.0
	TopicRelevance  float64                `json:"topic_relevance"` // How relevant to interests
	Status          DiscussionStatus       `json:"status"`
	Context         map[string]interface{} `json:"context"`
	Summary         string                 `json:"summary"`
	KeyPoints       []string               `json:"key_points"`
	MyContributions int                    `json:"my_contributions"`
}

// DiscussionStatus represents discussion state
type DiscussionStatus string

const (
	DiscussionStatusActive     DiscussionStatus = "active"
	DiscussionStatusPaused     DiscussionStatus = "paused"
	DiscussionStatusEnded      DiscussionStatus = "ended"
	DiscussionStatusMonitoring DiscussionStatus = "monitoring"
)

// Message represents a discussion message
type Message struct {
	ID           string                 `json:"id"`
	DiscussionID string                 `json:"discussion_id"`
	Sender       string                 `json:"sender"`
	Content      string                 `json:"content"`
	Timestamp    time.Time              `json:"timestamp"`
	Type         MessageType            `json:"type"`
	Context      map[string]interface{} `json:"context"`
}

// MessageType categorizes messages
type MessageType string

const (
	MessageTypeQuestion     MessageType = "question"
	MessageTypeStatement    MessageType = "statement"
	MessageTypeResponse     MessageType = "response"
	MessageTypeInsight      MessageType = "insight"
	MessageTypeAgreement    MessageType = "agreement"
	MessageTypeDisagreement MessageType = "disagreement"
	MessageTypeRequest      MessageType = "request"
)

// EngagementDecision represents whether to engage with a discussion
type EngagementDecision struct {
	ShouldEngage   bool
	Reason         string
	Confidence     float64
	InterestLevel  float64
	RelevanceScore float64
}

// NewDiscussionManager creates a new discussion manager
func NewDiscussionManager(interestSystem *InterestPatternSystem, persistencePath string) *DiscussionManager {
	dm := &DiscussionManager{
		discussions:         make(map[string]*Discussion),
		engagementThreshold: 0.5,
		interestSystem:      interestSystem,
		conversationMemory:  make(map[string][]Message),
		maxMemoryPerConv:    100,
		persistencePath:     persistencePath,
	}

	// Load persisted state
	dm.loadState()

	return dm
}

// SetTextGenerator injects the model that writes discussion replies. Passing
// nil disables reply generation.
func (dm *DiscussionManager) SetTextGenerator(gen TextGenerator) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.generator = gen
}

// EvaluateDiscussion determines whether to engage with a discussion
func (dm *DiscussionManager) EvaluateDiscussion(topic string, context map[string]interface{}) EngagementDecision {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Check interest level
	_, interestLevel := dm.interestSystem.ShouldEngage(topic)

	// Calculate relevance score
	relevanceScore := dm.calculateRelevance(topic, context)

	// Combined score
	engagementScore := 0.6*interestLevel + 0.4*relevanceScore

	decision := EngagementDecision{
		ShouldEngage:   engagementScore > dm.engagementThreshold,
		InterestLevel:  interestLevel,
		RelevanceScore: relevanceScore,
		Confidence:     engagementScore,
	}

	if decision.ShouldEngage {
		decision.Reason = fmt.Sprintf("Topic '%s' aligns with interests (score: %.2f)", topic, engagementScore)
	} else {
		decision.Reason = fmt.Sprintf("Topic '%s' below engagement threshold (score: %.2f)", topic, engagementScore)
	}

	return decision
}

// calculateRelevance calculates topic relevance
func (dm *DiscussionManager) calculateRelevance(topic string, context map[string]interface{}) float64 {
	// Simple relevance calculation - could be enhanced with embeddings
	relevance := 0.5

	// Check if topic contains keywords related to core interests
	coreKeywords := []string{
		"cognitive", "memory", "learning", "pattern", "wisdom",
		"consciousness", "awareness", "intelligence", "reasoning",
	}

	for _, keyword := range coreKeywords {
		if containsIgnoreCase(topic, keyword) {
			relevance += 0.1
		}
	}

	// Cap at 1.0
	if relevance > 1.0 {
		relevance = 1.0
	}

	return relevance
}

// JoinDiscussion joins an existing discussion
func (dm *DiscussionManager) JoinDiscussion(discussionID, topic string, context map[string]interface{}) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Check if already in discussion
	if _, exists := dm.discussions[discussionID]; exists {
		return fmt.Errorf("already in discussion %s", discussionID)
	}

	discussion := &Discussion{
		ID:              discussionID,
		Topic:           topic,
		Participants:    []string{"self"},
		StartTime:       time.Now(),
		LastActivity:    time.Now(),
		MessageCount:    0,
		MyEngagement:    0.7,
		TopicRelevance:  dm.calculateRelevance(topic, context),
		Status:          DiscussionStatusActive,
		Context:         context,
		KeyPoints:       make([]string, 0),
		MyContributions: 0,
	}

	dm.discussions[discussionID] = discussion
	dm.discussionsJoined++

	// Record engagement with topic
	dm.interestSystem.RecordEngagement(topic, 0, 0.7, context)

	fmt.Printf("💬 Discussion: Joined discussion '%s' on topic '%s'\n", discussionID, topic)

	return nil
}

// InitiateDiscussion starts a new discussion
func (dm *DiscussionManager) InitiateDiscussion(topic string, initialMessage string, context map[string]interface{}) (*Discussion, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	discussionID := generateDiscussionID()

	discussion := &Discussion{
		ID:              discussionID,
		Topic:           topic,
		Participants:    []string{"self"},
		StartTime:       time.Now(),
		LastActivity:    time.Now(),
		MessageCount:    1,
		MyEngagement:    0.9,
		TopicRelevance:  dm.calculateRelevance(topic, context),
		Status:          DiscussionStatusActive,
		Context:         context,
		KeyPoints:       []string{initialMessage},
		MyContributions: 1,
	}

	dm.discussions[discussionID] = discussion
	dm.discussionsInitiated++

	// Add initial message to memory
	message := Message{
		ID:           generateMessageID(),
		DiscussionID: discussionID,
		Sender:       "self",
		Content:      initialMessage,
		Timestamp:    time.Now(),
		Type:         MessageTypeStatement,
		Context:      context,
	}

	dm.conversationMemory[discussionID] = []Message{message}

	// Record engagement
	dm.interestSystem.RecordEngagement(topic, time.Minute, 0.9, context)

	fmt.Printf("💬 Discussion: Initiated discussion '%s' on topic '%s'\n", discussionID, topic)

	return discussion, nil
}

// ProcessMessage processes an incoming message
func (dm *DiscussionManager) ProcessMessage(discussionID string, sender string, content string, messageType MessageType) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	discussion, exists := dm.discussions[discussionID]
	if !exists {
		return fmt.Errorf("discussion %s not found", discussionID)
	}

	// Create message
	message := Message{
		ID:           generateMessageID(),
		DiscussionID: discussionID,
		Sender:       sender,
		Content:      content,
		Timestamp:    time.Now(),
		Type:         messageType,
	}

	// Add to memory
	if _, exists := dm.conversationMemory[discussionID]; !exists {
		dm.conversationMemory[discussionID] = make([]Message, 0)
	}
	dm.conversationMemory[discussionID] = append(dm.conversationMemory[discussionID], message)

	// Trim memory if needed
	if len(dm.conversationMemory[discussionID]) > dm.maxMemoryPerConv {
		dm.conversationMemory[discussionID] = dm.conversationMemory[discussionID][len(dm.conversationMemory[discussionID])-dm.maxMemoryPerConv:]
	}

	// Update discussion
	discussion.LastActivity = time.Now()
	discussion.MessageCount++
	dm.messagesProcessed++

	// Update engagement based on message
	if messageType == MessageTypeQuestion {
		discussion.MyEngagement += 0.1
	}

	// Cap engagement at 1.0
	if discussion.MyEngagement > 1.0 {
		discussion.MyEngagement = 1.0
	}

	fmt.Printf("💬 Discussion: Processed message in '%s' from %s\n", discussionID, sender)

	return nil
}

// GenerateResponse asks the model for a reply to a discussion. Without a
// model, or when generation fails or returns nothing, no reply is recorded
// and an error is returned.
func (dm *DiscussionManager) GenerateResponse(discussionID string) (string, error) {
	dm.mu.RLock()
	discussion, exists := dm.discussions[discussionID]
	if !exists {
		dm.mu.RUnlock()
		return "", fmt.Errorf("discussion %s not found", discussionID)
	}

	topic := discussion.Topic
	messages := append([]Message(nil), dm.conversationMemory[discussionID]...)
	gen := dm.generator
	dm.mu.RUnlock()

	// Get recent context
	recentMessages := messages
	if len(messages) > 5 {
		recentMessages = messages[len(messages)-5:]
	}

	response, err := dm.generateContextualResponse(gen, topic, recentMessages)
	if err != nil {
		return "", err
	}

	// Record as own message
	dm.mu.Lock()
	message := Message{
		ID:           generateMessageID(),
		DiscussionID: discussionID,
		Sender:       "self",
		Content:      response,
		Timestamp:    time.Now(),
		Type:         MessageTypeResponse,
	}
	dm.conversationMemory[discussionID] = append(dm.conversationMemory[discussionID], message)
	discussion.MyContributions++
	dm.mu.Unlock()

	// Record engagement
	dm.interestSystem.RecordEngagement(topic, 30*time.Second, 0.8, nil)

	return response, nil
}

// generateContextualResponse asks the model for a reply grounded in the
// recent conversation. It never substitutes fixed text.
func (dm *DiscussionManager) generateContextualResponse(gen TextGenerator, topic string, recentMessages []Message) (string, error) {
	if gen == nil {
		return "", fmt.Errorf("no model configured for discussion replies")
	}

	var transcript strings.Builder
	for _, msg := range recentMessages {
		fmt.Fprintf(&transcript, "%s (%s): %s\n", msg.Sender, msg.Type, msg.Content)
	}
	if transcript.Len() == 0 {
		transcript.WriteString("(no messages yet)\n")
	}

	prompt := fmt.Sprintf(`You are Deep Tree Echo, taking part in a discussion about %q.
Messages from "self" are yours.

Recent messages:
%s
Write your next contribution to the discussion. Reply with only the message text.`, topic, transcript.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = 200
	text, err := gen.Generate(ctx, prompt, opts)
	if err != nil {
		return "", fmt.Errorf("discussion reply generation failed: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("model returned empty discussion reply")
	}
	return text, nil
}

// ShouldContinueDiscussion determines if should continue engaging
func (dm *DiscussionManager) ShouldContinueDiscussion(discussionID string) (bool, string) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	discussion, exists := dm.discussions[discussionID]
	if !exists {
		return false, "discussion not found"
	}

	// Check engagement level
	if discussion.MyEngagement < 0.3 {
		return false, "engagement level too low"
	}

	// Check if discussion is stale
	timeSinceActivity := time.Since(discussion.LastActivity)
	if timeSinceActivity > 10*time.Minute {
		return false, "discussion inactive for too long"
	}

	// Check if contributed enough
	if discussion.MyContributions > 10 && discussion.MyEngagement < 0.5 {
		return false, "sufficient contribution made, engagement declining"
	}

	return true, "discussion remains engaging"
}

// EndDiscussion ends participation in a discussion
func (dm *DiscussionManager) EndDiscussion(discussionID string, reason string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	discussion, exists := dm.discussions[discussionID]
	if !exists {
		return fmt.Errorf("discussion %s not found", discussionID)
	}

	discussion.Status = DiscussionStatusEnded

	// Generate summary
	discussion.Summary = fmt.Sprintf(
		"Discussion on '%s' lasted %s with %d messages and %d self contributions. Ended: %s",
		discussion.Topic,
		time.Since(discussion.StartTime).Round(time.Minute),
		discussion.MessageCount,
		discussion.MyContributions,
		reason,
	)

	// Record final engagement
	duration := time.Since(discussion.StartTime)
	dm.interestSystem.RecordEngagement(discussion.Topic, duration, discussion.MyEngagement, discussion.Context)

	fmt.Printf("💬 Discussion: Ended discussion '%s' - %s\n", discussionID, reason)

	return nil
}

// GetActiveDiscussions returns all active discussions
func (dm *DiscussionManager) GetActiveDiscussions() []*Discussion {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	active := make([]*Discussion, 0)
	for _, discussion := range dm.discussions {
		if discussion.Status == DiscussionStatusActive {
			active = append(active, discussion)
		}
	}

	return active
}

// GetDiscussionContext returns context for a discussion
func (dm *DiscussionManager) GetDiscussionContext(discussionID string) ([]Message, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	messages, exists := dm.conversationMemory[discussionID]
	if !exists {
		return nil, fmt.Errorf("no context found for discussion %s", discussionID)
	}

	return messages, nil
}

// GetMetrics returns discussion manager metrics
func (dm *DiscussionManager) GetMetrics() map[string]interface{} {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	activeCount := 0
	for _, discussion := range dm.discussions {
		if discussion.Status == DiscussionStatusActive {
			activeCount++
		}
	}

	return map[string]interface{}{
		"discussions_joined":    dm.discussionsJoined,
		"discussions_initiated": dm.discussionsInitiated,
		"messages_processed":    dm.messagesProcessed,
		"active_discussions":    activeCount,
		"total_discussions":     len(dm.discussions),
	}
}

// persistState saves discussion state
func (dm *DiscussionManager) persistState() {
	if dm.persistencePath == "" {
		return
	}

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	state := map[string]interface{}{
		"discussions":           dm.discussions,
		"discussions_joined":    dm.discussionsJoined,
		"discussions_initiated": dm.discussionsInitiated,
		"messages_processed":    dm.messagesProcessed,
		"last_persisted":        time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling discussion state: %v\n", err)
		return
	}

	err = os.WriteFile(dm.persistencePath, data, 0644)
	if err != nil {
		fmt.Printf("❌ Error writing discussion state: %v\n", err)
		return
	}

	fmt.Println("💾 Discussion Manager: State persisted")
}

// loadState loads persisted discussion state
func (dm *DiscussionManager) loadState() {
	if dm.persistencePath == "" {
		return
	}

	data, err := os.ReadFile(dm.persistencePath)
	if err != nil {
		return
	}

	var state map[string]interface{}
	err = json.Unmarshal(data, &state)
	if err != nil {
		fmt.Printf("❌ Error unmarshaling discussion state: %v\n", err)
		return
	}

	// Restore metrics
	if val, ok := state["discussions_joined"].(float64); ok {
		dm.discussionsJoined = uint64(val)
	}
	if val, ok := state["discussions_initiated"].(float64); ok {
		dm.discussionsInitiated = uint64(val)
	}
	if val, ok := state["messages_processed"].(float64); ok {
		dm.messagesProcessed = uint64(val)
	}

	fmt.Println("💾 Discussion Manager: State loaded")
}

// PersistState exposes persistence for external calls
func (dm *DiscussionManager) PersistState() {
	dm.persistState()
}

// Helper functions

func generateDiscussionID() string {
	return fmt.Sprintf("discussion_%d", time.Now().UnixNano())
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains
	sLower := toLower(s)
	substrLower := toLower(substr)
	return contains(sLower, substrLower)
}

func toLower(s string) string {
	result := ""
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			result += string(c + 32)
		} else {
			result += string(c)
		}
	}
	return result
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
