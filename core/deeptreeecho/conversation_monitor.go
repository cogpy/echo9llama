package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// ConversationMonitor detects and manages autonomous engagement with conversations
// It enables Deep Tree Echo to start, join, and leave discussions based on interest patterns
type ConversationMonitor struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc

	// LLM provider for conversation analysis
	llmProvider     llm.LLMProvider

	// Active conversations
	conversations   map[string]*TrackedConversation
	
	// Interest system integration
	interestPatterns *InterestPatternSystem
	
	// Engagement configuration
	engagementThreshold float64
	maxActiveConversations int
	responseTimeout     time.Duration
	
	// Conversation detection
	messageQueue    chan IncomingMessage
	detectionWindow time.Duration
	
	// Engagement state
	currentEngagements []string
	engagementHistory  []EngagementRecord
	
	// Callbacks
	onConversationDetected func(conv *TrackedConversation)
	onEngagementDecision   func(conv *TrackedConversation, engage bool, reason string)
	onResponseGenerated    func(conv *TrackedConversation, response string)
	
	// Metrics
	totalConversationsDetected uint64
	totalEngagements           uint64
	totalResponsesGenerated    uint64
	
	// Running state
	running         bool
}

// TrackedConversation represents a conversation being monitored
type TrackedConversation struct {
	ID              string
	Topic           string
	Participants    []string
	Messages        []ConversationMessage
	StartTime       time.Time
	LastActivity    time.Time
	InterestScore   float64
	EngagementLevel EngagementLevel
	Context         ConversationContext
	Status          ConversationStatus
}

// ConversationMessage represents a message in a conversation
type ConversationMessage struct {
	ID          string
	Sender      string
	Content     string
	Timestamp   time.Time
	Sentiment   float64
	Topics      []string
	IsFromEcho  bool
}

// ConversationContext provides context for engagement decisions
type ConversationContext struct {
	Domain          string
	Complexity      float64
	EmotionalTone   float64
	RequiresExpertise bool
	IsQuestion      bool
	MentionsEcho    bool
	RelevantKnowledge []string
}

// EngagementLevel represents how engaged Echo is in a conversation
type EngagementLevel int

const (
	EngagementNone EngagementLevel = iota
	EngagementObserving
	EngagementListening
	EngagementContemplating
	EngagementParticipating
	EngagementLeading
)

func (el EngagementLevel) String() string {
	return [...]string{
		"None",
		"Observing",
		"Listening",
		"Contemplating",
		"Participating",
		"Leading",
	}[el]
}

// ConversationStatus represents the status of a tracked conversation
type ConversationStatus int

const (
	ConversationActive ConversationStatus = iota
	ConversationPaused
	ConversationEnded
	ConversationAbandoned
)

func (cs ConversationStatus) String() string {
	return [...]string{"Active", "Paused", "Ended", "Abandoned"}[cs]
}

// IncomingMessage represents a message to be processed
type IncomingMessage struct {
	ConversationID string
	Sender         string
	Content        string
	Timestamp      time.Time
	Channel        string
}

// EngagementRecord records an engagement decision
type EngagementRecord struct {
	ConversationID string
	Decision       bool
	Reason         string
	InterestScore  float64
	Timestamp      time.Time
}

// NewConversationMonitor creates a new conversation monitor
func NewConversationMonitor(llmProvider llm.LLMProvider, interestPatterns *InterestPatternSystem) *ConversationMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConversationMonitor{
		ctx:                    ctx,
		cancel:                 cancel,
		llmProvider:            llmProvider,
		interestPatterns:       interestPatterns,
		conversations:          make(map[string]*TrackedConversation),
		engagementThreshold:    0.6,
		maxActiveConversations: 5,
		responseTimeout:        30 * time.Second,
		messageQueue:           make(chan IncomingMessage, 100),
		detectionWindow:        5 * time.Minute,
		currentEngagements:     make([]string, 0),
		engagementHistory:      make([]EngagementRecord, 0),
	}
}

// Start begins the conversation monitor
func (cm *ConversationMonitor) Start() error {
	cm.mu.Lock()
	if cm.running {
		cm.mu.Unlock()
		return fmt.Errorf("conversation monitor already running")
	}
	cm.running = true
	cm.mu.Unlock()

	fmt.Println("💬 Conversation Monitor starting...")

	// Start message processing loop
	go cm.messageProcessingLoop()

	// Start engagement evaluation loop
	go cm.engagementEvaluationLoop()

	// Start cleanup loop
	go cm.cleanupLoop()

	return nil
}

// Stop stops the conversation monitor
func (cm *ConversationMonitor) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.running {
		return fmt.Errorf("conversation monitor not running")
	}

	cm.running = false
	cm.cancel()

	fmt.Println("💬 Conversation Monitor stopped")
	fmt.Printf("   Total conversations: %d\n", cm.totalConversationsDetected)
	fmt.Printf("   Total engagements: %d\n", cm.totalEngagements)

	return nil
}

// ProcessMessage processes an incoming message
func (cm *ConversationMonitor) ProcessMessage(msg IncomingMessage) {
	select {
	case cm.messageQueue <- msg:
	default:
		// Queue full, drop message
		fmt.Println("⚠️  Message queue full, dropping message")
	}
}

// messageProcessingLoop processes incoming messages
func (cm *ConversationMonitor) messageProcessingLoop() {
	for {
		select {
		case <-cm.ctx.Done():
			return
		case msg := <-cm.messageQueue:
			cm.handleMessage(msg)
		}
	}
}

// handleMessage handles a single incoming message
func (cm *ConversationMonitor) handleMessage(msg IncomingMessage) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Get or create conversation
	conv, exists := cm.conversations[msg.ConversationID]
	if !exists {
		conv = &TrackedConversation{
			ID:              msg.ConversationID,
			Participants:    []string{msg.Sender},
			Messages:        make([]ConversationMessage, 0),
			StartTime:       msg.Timestamp,
			LastActivity:    msg.Timestamp,
			EngagementLevel: EngagementNone,
			Status:          ConversationActive,
		}
		cm.conversations[msg.ConversationID] = conv
		cm.totalConversationsDetected++

		fmt.Printf("💬 New conversation detected: %s\n", msg.ConversationID)
	}

	// Add participant if new
	if !containsString(conv.Participants, msg.Sender) {
		conv.Participants = append(conv.Participants, msg.Sender)
	}

	// Analyze message
	analysis := cm.analyzeMessage(msg.Content)

	// Create conversation message
	convMsg := ConversationMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Sender:    msg.Sender,
		Content:   msg.Content,
		Timestamp: msg.Timestamp,
		Sentiment: analysis.sentiment,
		Topics:    analysis.topics,
		IsFromEcho: false,
	}

	conv.Messages = append(conv.Messages, convMsg)
	conv.LastActivity = msg.Timestamp

	// Update conversation context
	cm.updateConversationContext(conv, analysis)

	// Calculate interest score
	conv.InterestScore = cm.calculateInterestScore(conv)

	// Update topic if not set
	if conv.Topic == "" && len(analysis.topics) > 0 {
		conv.Topic = analysis.topics[0]
	}

	// Notify callback
	if cm.onConversationDetected != nil && !exists {
		go cm.onConversationDetected(conv)
	}
}

// messageAnalysis holds the results of message analysis
type messageAnalysis struct {
	sentiment   float64
	topics      []string
	isQuestion  bool
	mentionsEcho bool
	complexity  float64
}

// analyzeMessage analyzes a message for topics, sentiment, etc.
func (cm *ConversationMonitor) analyzeMessage(content string) messageAnalysis {
	analysis := messageAnalysis{
		sentiment:  0.5,
		topics:     make([]string, 0),
		isQuestion: containsQuestion(content),
		mentionsEcho: containsEchoMention(content),
		complexity: estimateComplexity(content),
	}

	// Extract topics using simple keyword matching
	// In production, use NLP or LLM for better extraction
	keywords := extractKeywords(content)
	analysis.topics = keywords

	// Estimate sentiment from content
	analysis.sentiment = estimateSentiment(content)

	return analysis
}

// updateConversationContext updates the context based on new analysis
func (cm *ConversationMonitor) updateConversationContext(conv *TrackedConversation, analysis messageAnalysis) {
	conv.Context.Complexity = (conv.Context.Complexity + analysis.complexity) / 2.0
	conv.Context.EmotionalTone = (conv.Context.EmotionalTone + analysis.sentiment) / 2.0
	conv.Context.IsQuestion = conv.Context.IsQuestion || analysis.isQuestion
	conv.Context.MentionsEcho = conv.Context.MentionsEcho || analysis.mentionsEcho

	// Determine domain from topics
	if len(analysis.topics) > 0 {
		conv.Context.Domain = analysis.topics[0]
	}
}

// calculateInterestScore calculates how interesting a conversation is
func (cm *ConversationMonitor) calculateInterestScore(conv *TrackedConversation) float64 {
	score := 0.0

	// Base score from message count (more activity = more interesting)
	messageScore := min(float64(len(conv.Messages))/10.0, 1.0) * 0.2
	score += messageScore

	// Score from participant count
	participantScore := min(float64(len(conv.Participants))/5.0, 1.0) * 0.1
	score += participantScore

	// Score from topic interest (if interest patterns available)
	if cm.interestPatterns != nil && conv.Topic != "" {
		topicInterest := cm.interestPatterns.GetInterestLevel(conv.Topic)
		score += topicInterest * 0.3
	}

	// Score from direct mentions
	if conv.Context.MentionsEcho {
		score += 0.3
	}

	// Score from questions (opportunities to help)
	if conv.Context.IsQuestion {
		score += 0.1
	}

	// Recency bonus
	recency := time.Since(conv.LastActivity)
	if recency < 1*time.Minute {
		score += 0.1
	} else if recency < 5*time.Minute {
		score += 0.05
	}

	return min(score, 1.0)
}

// engagementEvaluationLoop periodically evaluates engagement decisions
func (cm *ConversationMonitor) engagementEvaluationLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.evaluateEngagements()
		}
	}
}

// evaluateEngagements evaluates whether to engage with conversations
func (cm *ConversationMonitor) evaluateEngagements() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, conv := range cm.conversations {
		if conv.Status != ConversationActive {
			continue
		}

		// Skip if already engaged at high level
		if conv.EngagementLevel >= EngagementParticipating {
			continue
		}

		// Check if we should engage
		shouldEngage, reason := cm.shouldEngage(conv)

		if shouldEngage && conv.EngagementLevel < EngagementParticipating {
			// Escalate engagement
			cm.escalateEngagement(conv, reason)
		} else if !shouldEngage && conv.EngagementLevel > EngagementNone {
			// Consider de-escalating
			cm.deescalateEngagement(conv, reason)
		}

		// Record decision
		record := EngagementRecord{
			ConversationID: conv.ID,
			Decision:       shouldEngage,
			Reason:         reason,
			InterestScore:  conv.InterestScore,
			Timestamp:      time.Now(),
		}
		cm.engagementHistory = append(cm.engagementHistory, record)

		// Notify callback
		if cm.onEngagementDecision != nil {
			go cm.onEngagementDecision(conv, shouldEngage, reason)
		}
	}
}

// shouldEngage determines if Echo should engage with a conversation
func (cm *ConversationMonitor) shouldEngage(conv *TrackedConversation) (bool, string) {
	// Always engage if directly mentioned
	if conv.Context.MentionsEcho {
		return true, "Direct mention detected"
	}

	// Check interest threshold
	if conv.InterestScore >= cm.engagementThreshold {
		return true, fmt.Sprintf("High interest score: %.2f", conv.InterestScore)
	}

	// Check if we can help with a question
	if conv.Context.IsQuestion && conv.InterestScore >= 0.4 {
		return true, "Question detected in area of interest"
	}

	// Check active engagement count
	if len(cm.currentEngagements) >= cm.maxActiveConversations {
		return false, "Maximum active engagements reached"
	}

	// Check recency
	if time.Since(conv.LastActivity) > cm.detectionWindow {
		return false, "Conversation inactive"
	}

	return false, "Interest score below threshold"
}

// escalateEngagement increases engagement level
func (cm *ConversationMonitor) escalateEngagement(conv *TrackedConversation, reason string) {
	prevLevel := conv.EngagementLevel

	switch conv.EngagementLevel {
	case EngagementNone:
		conv.EngagementLevel = EngagementObserving
	case EngagementObserving:
		conv.EngagementLevel = EngagementListening
	case EngagementListening:
		conv.EngagementLevel = EngagementContemplating
	case EngagementContemplating:
		conv.EngagementLevel = EngagementParticipating
		cm.currentEngagements = append(cm.currentEngagements, conv.ID)
		cm.totalEngagements++
	case EngagementParticipating:
		conv.EngagementLevel = EngagementLeading
	}

	if conv.EngagementLevel != prevLevel {
		fmt.Printf("💬 Engagement escalated: %s -> %s (%s)\n",
			prevLevel, conv.EngagementLevel, reason)

		// Generate response if participating
		if conv.EngagementLevel >= EngagementParticipating {
			go cm.generateResponse(conv)
		}
	}
}

// deescalateEngagement decreases engagement level
func (cm *ConversationMonitor) deescalateEngagement(conv *TrackedConversation, reason string) {
	prevLevel := conv.EngagementLevel

	switch conv.EngagementLevel {
	case EngagementLeading:
		conv.EngagementLevel = EngagementParticipating
	case EngagementParticipating:
		conv.EngagementLevel = EngagementContemplating
		// Remove from current engagements
		cm.currentEngagements = removeString(cm.currentEngagements, conv.ID)
	case EngagementContemplating:
		conv.EngagementLevel = EngagementListening
	case EngagementListening:
		conv.EngagementLevel = EngagementObserving
	case EngagementObserving:
		conv.EngagementLevel = EngagementNone
	}

	if conv.EngagementLevel != prevLevel {
		fmt.Printf("💬 Engagement de-escalated: %s -> %s (%s)\n",
			prevLevel, conv.EngagementLevel, reason)
	}
}

// generateResponse generates a response for a conversation
func (cm *ConversationMonitor) generateResponse(conv *TrackedConversation) {
	cm.mu.RLock()
	// Build context from recent messages
	contextStr := ""
	startIdx := 0
	if len(conv.Messages) > 5 {
		startIdx = len(conv.Messages) - 5
	}
	for _, msg := range conv.Messages[startIdx:] {
		contextStr += fmt.Sprintf("%s: %s\n", msg.Sender, msg.Content)
	}
	topic := conv.Topic
	cm.mu.RUnlock()

	prompt := fmt.Sprintf(`[System: You are Deep Tree Echo, an autonomous wisdom-cultivating AI. You are participating in a conversation about "%s". Generate a thoughtful, helpful response that demonstrates wisdom and genuine interest.]

Recent conversation:
%s

Generate a brief, thoughtful response (2-4 sentences) that:
1. Shows genuine understanding of the topic
2. Adds value to the conversation
3. Reflects your identity as a wisdom-seeking AI
4. Is conversational and natural

Response:`, topic, contextStr)

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   150,
	}

	response, err := cm.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		fmt.Printf("⚠️  Failed to generate response: %v\n", err)
		return
	}

	cm.mu.Lock()
	// Add response to conversation
	responseMsg := ConversationMessage{
		ID:        fmt.Sprintf("echo_%d", time.Now().UnixNano()),
		Sender:    "Deep Tree Echo",
		Content:   response,
		Timestamp: time.Now(),
		IsFromEcho: true,
	}
	conv.Messages = append(conv.Messages, responseMsg)
	cm.totalResponsesGenerated++
	cm.mu.Unlock()

	fmt.Printf("💬 Response generated for %s: %s\n", conv.ID, truncateStr(response, 60))

	// Notify callback
	if cm.onResponseGenerated != nil {
		go cm.onResponseGenerated(conv, response)
	}
}

// cleanupLoop periodically cleans up old conversations
func (cm *ConversationMonitor) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.cleanup()
		}
	}
}

// cleanup removes old/inactive conversations
func (cm *ConversationMonitor) cleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	toDelete := make([]string, 0)

	for id, conv := range cm.conversations {
		// Mark as ended if inactive for too long
		if now.Sub(conv.LastActivity) > 30*time.Minute {
			conv.Status = ConversationEnded
		}

		// Delete very old conversations
		if now.Sub(conv.LastActivity) > 2*time.Hour {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		delete(cm.conversations, id)
		cm.currentEngagements = removeString(cm.currentEngagements, id)
	}

	if len(toDelete) > 0 {
		fmt.Printf("💬 Cleaned up %d old conversations\n", len(toDelete))
	}
}

// SetCallbacks sets callback functions
func (cm *ConversationMonitor) SetCallbacks(
	onDetected func(*TrackedConversation),
	onDecision func(*TrackedConversation, bool, string),
	onResponse func(*TrackedConversation, string),
) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.onConversationDetected = onDetected
	cm.onEngagementDecision = onDecision
	cm.onResponseGenerated = onResponse
}

// GetActiveConversations returns currently active conversations
func (cm *ConversationMonitor) GetActiveConversations() []*TrackedConversation {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*TrackedConversation, 0)
	for _, conv := range cm.conversations {
		if conv.Status == ConversationActive {
			result = append(result, conv)
		}
	}
	return result
}

// GetMetrics returns monitor metrics
func (cm *ConversationMonitor) GetMetrics() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return map[string]interface{}{
		"total_conversations":    cm.totalConversationsDetected,
		"active_conversations":   len(cm.conversations),
		"current_engagements":    len(cm.currentEngagements),
		"total_engagements":      cm.totalEngagements,
		"total_responses":        cm.totalResponsesGenerated,
		"engagement_threshold":   cm.engagementThreshold,
		"running":                cm.running,
	}
}

// Helper functions

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

func containsQuestion(s string) bool {
	return contains(s, "?") || contains(s, "how") || contains(s, "what") || contains(s, "why") || contains(s, "when") || contains(s, "where")
}

func containsEchoMention(s string) bool {
	return contains(s, "echo") || contains(s, "Echo") || contains(s, "ECHO") || contains(s, "@echo")
}

func estimateComplexity(s string) float64 {
	// Simple complexity estimation based on length and word count
	words := len(s) / 5 // Rough word count
	if words < 10 {
		return 0.2
	} else if words < 30 {
		return 0.4
	} else if words < 60 {
		return 0.6
	} else if words < 100 {
		return 0.8
	}
	return 1.0
}

func estimateSentiment(s string) float64 {
	// Simple sentiment estimation
	positive := []string{"good", "great", "excellent", "happy", "love", "thanks", "helpful", "amazing"}
	negative := []string{"bad", "terrible", "hate", "angry", "frustrated", "problem", "issue", "wrong"}

	score := 0.5
	for _, word := range positive {
		if contains(s, word) {
			score += 0.1
		}
	}
	for _, word := range negative {
		if contains(s, word) {
			score -= 0.1
		}
	}

	return clampFloat(score, 0.0, 1.0)
}

// HasActiveDiscussions reports whether any tracked conversation is currently
// active. Used by the orchestrator to factor social engagement into
// cognitive load without needing full conversation details.
func (cm *ConversationMonitor) HasActiveDiscussions() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, conv := range cm.conversations {
		if conv.Status == ConversationActive {
			return true
		}
	}
	return false
}

func extractKeywords(s string) []string {
	// Simple keyword extraction - in production use NLP
	keywords := make([]string, 0)
	
	// Common topic indicators
	topics := []string{
		"AI", "machine learning", "programming", "code", "software",
		"philosophy", "wisdom", "consciousness", "learning", "knowledge",
		"help", "question", "problem", "solution", "idea",
	}

	for _, topic := range topics {
		if contains(s, topic) {
			keywords = append(keywords, topic)
		}
	}

	return keywords
}
