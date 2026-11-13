package echobeats

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// AutonomousDiscussionManager handles autonomous discussion initiation and engagement
type AutonomousDiscussionManager struct {
	mu     sync.RWMutex
	ctx    context.Context //nolint:containedctx // lifecycle context created with cancel in constructor; cancelled on Stop to end goroutines
	cancel context.CancelFunc

	// Interest tracking (interface to InterestPatternTracker)
	interestScorer InterestScorer

	// Discussion state
	activeDiscussions map[string]*Discussion
	discussionHistory []DiscussionRecord

	// Decision thresholds
	initiationThreshold  float64 // Interest score needed to start discussion
	engagementThreshold  float64 // Interest score needed to respond
	terminationThreshold float64 // Fatigue/disinterest level to end discussion

	// Fatigue tracking
	discussionFatigue   float64
	fatigueRecoveryRate float64

	// Message queues
	incomingMessages chan IncomingMessage
	outgoingMessages chan OutgoingMessage

	// Response generation (injected by the orchestrator; typically LLM-backed)
	responseGenerator ResponseGenerator
	openingGenerator  OpeningGenerator

	// Metrics
	discussionsInitiated  uint64
	discussionsEngaged    uint64
	discussionsTerminated uint64
	messagesProcessed     uint64

	// Running state
	running bool
}

// InterestScorer interface for interest pattern tracking
type InterestScorer interface {
	GetInterestScore(category, name string) float64
	IsInterested(category, name string, threshold float64) bool
}

// ResponseGenerator produces a reply for an incoming discussion message.
// Injected by higher layers (e.g., the unified orchestrator) so the
// discussion manager can generate genuine LLM-backed responses without
// creating a dependency on any specific provider.
type ResponseGenerator func(ctx context.Context, topic, content string) (string, error)

// OpeningGenerator produces the opening message for a discussion Deep Tree
// Echo initiates on a topic. Injected by higher layers (typically LLM-backed).
type OpeningGenerator func(ctx context.Context, topic string) (string, error)

// Discussion represents an active discussion
type Discussion struct {
	ID            string
	Topic         string
	Participants  []string
	StartTime     time.Time
	LastActivity  time.Time
	MessageCount  int
	InterestScore float64
	FatigueLevel  float64
	Status        DiscussionStatus
}

// DiscussionStatus represents the state of a discussion
type DiscussionStatus int

const (
	DiscussionActive DiscussionStatus = iota
	DiscussionPaused
	DiscussionEnded
)

func (ds DiscussionStatus) String() string {
	return [...]string{"Active", "Paused", "Ended"}[ds]
}

// DiscussionRecord represents a completed discussion
type DiscussionRecord struct {
	ID           string
	Topic        string
	Duration     time.Duration
	MessageCount int
	Outcome      string
	Timestamp    time.Time
}

// IncomingMessage represents a message from external source
type IncomingMessage struct {
	ID        string
	Source    string
	Topic     string
	Content   string
	Timestamp time.Time
	Priority  float64
}

// OutgoingMessage represents a message to send
type OutgoingMessage struct {
	ID          string
	Destination string
	Topic       string
	Content     string
	Timestamp   time.Time
}

// NewAutonomousDiscussionManager creates a new discussion manager
func NewAutonomousDiscussionManager(interestScorer InterestScorer) *AutonomousDiscussionManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &AutonomousDiscussionManager{
		ctx:                  ctx,
		cancel:               cancel,
		interestScorer:       interestScorer,
		activeDiscussions:    make(map[string]*Discussion),
		discussionHistory:    make([]DiscussionRecord, 0),
		initiationThreshold:  0.7, // High threshold for starting discussions
		engagementThreshold:  0.5, // Medium threshold for responding
		terminationThreshold: 0.8, // High fatigue to end discussion
		discussionFatigue:    0.0,
		fatigueRecoveryRate:  0.1, // 10% recovery per minute
		incomingMessages:     make(chan IncomingMessage, 100),
		outgoingMessages:     make(chan OutgoingMessage, 100),
	}
}

// Start begins autonomous discussion management
func (adm *AutonomousDiscussionManager) Start() error {
	adm.mu.Lock()
	if adm.running {
		adm.mu.Unlock()
		return fmt.Errorf("already running")
	}
	adm.running = true
	adm.mu.Unlock()

	fmt.Println("💬 Starting Autonomous Discussion Manager...")
	fmt.Printf("   Initiation threshold: %.2f\n", adm.initiationThreshold)
	fmt.Printf("   Engagement threshold: %.2f\n", adm.engagementThreshold)

	// Start message processing
	go adm.processIncomingMessages()

	// Start fatigue recovery
	go adm.fatigueRecoveryLoop()

	// Start discussion monitoring
	go adm.monitorDiscussions()

	return nil
}

// Stop gracefully stops the discussion manager
func (adm *AutonomousDiscussionManager) Stop() error {
	adm.mu.Lock()
	defer adm.mu.Unlock()

	if !adm.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("💬 Stopping autonomous discussion manager...")
	adm.running = false
	adm.cancel()

	// End all active discussions
	for _, discussion := range adm.activeDiscussions {
		adm.endDiscussion(discussion, "system shutdown")
	}

	return nil
}

// processIncomingMessages handles incoming messages
func (adm *AutonomousDiscussionManager) processIncomingMessages() {
	for {
		select {
		case <-adm.ctx.Done():
			return
		case msg := <-adm.incomingMessages:
			adm.handleIncomingMessage(msg)
		}
	}
}

// handleIncomingMessage processes a single incoming message
func (adm *AutonomousDiscussionManager) handleIncomingMessage(msg IncomingMessage) {
	adm.mu.Lock()
	adm.messagesProcessed++
	adm.mu.Unlock()

	// Calculate relevance score based on interest
	relevanceScore := adm.interestScorer.GetInterestScore("topic", msg.Topic)

	fmt.Printf("📨 Incoming message on topic '%s' (relevance: %.2f)\n", msg.Topic, relevanceScore)

	// Decision: Should we engage?
	shouldEngage := adm.shouldEngageInDiscussion(msg.Topic, relevanceScore)

	if shouldEngage {
		adm.engageInDiscussion(msg)
	} else {
		fmt.Printf("   ⏭️  Skipping (below engagement threshold or too fatigued)\n")
	}
}

// shouldEngageInDiscussion determines if we should engage
func (adm *AutonomousDiscussionManager) shouldEngageInDiscussion(topic string, relevanceScore float64) bool {
	adm.mu.RLock()
	fatigue := adm.discussionFatigue
	adm.mu.RUnlock()

	// Don't engage if too fatigued
	if fatigue > adm.terminationThreshold {
		return false
	}

	// Check if relevance exceeds threshold
	if relevanceScore < adm.engagementThreshold {
		return false
	}

	// Adjust threshold based on fatigue
	adjustedThreshold := adm.engagementThreshold * (1.0 + fatigue)

	return relevanceScore >= adjustedThreshold
}

// engageInDiscussion engages in a discussion. The reply is written by the
// injected response generator; when none is configured or generation fails, no
// reply is sent.
func (adm *AutonomousDiscussionManager) engageInDiscussion(msg IncomingMessage) {
	adm.mu.Lock()
	// Check if discussion already exists
	if discussion, exists := adm.activeDiscussions[msg.Topic]; exists {
		// Continue existing discussion
		discussion.MessageCount++
		discussion.LastActivity = time.Now()
		fmt.Printf("   💬 Continuing discussion on '%s'\n", msg.Topic)
	} else {
		// Start new discussion
		discussion := &Discussion{
			ID:            fmt.Sprintf("disc_%d", time.Now().UnixNano()),
			Topic:         msg.Topic,
			Participants:  []string{msg.Source, "Deep Tree Echo"},
			StartTime:     time.Now(),
			LastActivity:  time.Now(),
			MessageCount:  1,
			InterestScore: adm.interestScorer.GetInterestScore("topic", msg.Topic),
			FatigueLevel:  0.0,
			Status:        DiscussionActive,
		}

		adm.activeDiscussions[msg.Topic] = discussion
		adm.discussionsEngaged++

		fmt.Printf("   ✨ Engaging in new discussion on '%s'\n", msg.Topic)
	}

	// Increase fatigue
	adm.discussionFatigue += 0.1
	if adm.discussionFatigue > 1.0 {
		adm.discussionFatigue = 1.0
	}
	gen := adm.responseGenerator
	adm.mu.Unlock()

	// Generate the reply outside the lock so slow models do not block the manager
	response, err := adm.generateResponse(gen, msg)
	if err != nil {
		fmt.Printf("   ⚠️  No response sent on '%s': %v\n", msg.Topic, err)
		return
	}

	// Queue outgoing message
	outgoing := OutgoingMessage{
		ID:          fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Destination: msg.Source,
		Topic:       msg.Topic,
		Content:     response,
		Timestamp:   time.Now(),
	}

	select {
	case adm.outgoingMessages <- outgoing:
		fmt.Printf("   📤 Response queued\n")
	default:
		fmt.Printf("   ⚠️  Outgoing queue full\n")
	}
}

// SetResponseGenerator injects the response generation function (LLM-backed)
func (adm *AutonomousDiscussionManager) SetResponseGenerator(gen ResponseGenerator) {
	adm.mu.Lock()
	defer adm.mu.Unlock()
	adm.responseGenerator = gen
}

// SetOpeningGenerator injects the function (LLM-backed) that writes the
// opening message of a discussion Deep Tree Echo initiates.
func (adm *AutonomousDiscussionManager) SetOpeningGenerator(gen OpeningGenerator) {
	adm.mu.Lock()
	defer adm.mu.Unlock()
	adm.openingGenerator = gen
}

// generateResponse generates a reply with the injected generator. It never
// substitutes fixed text: a missing generator, failed call, or empty reply is
// an error.
func (adm *AutonomousDiscussionManager) generateResponse(gen ResponseGenerator, msg IncomingMessage) (string, error) {
	if gen == nil {
		return "", fmt.Errorf("no response generator configured")
	}
	ctx, cancel := context.WithTimeout(adm.ctx, 30*time.Second)
	defer cancel()
	response, err := gen(ctx, msg.Topic, msg.Content)
	if err != nil {
		return "", fmt.Errorf("response generation failed: %w", err)
	}
	response = strings.TrimSpace(response)
	if response == "" {
		return "", fmt.Errorf("response generator returned empty text")
	}
	return response, nil
}

// InitiateDiscussion autonomously starts a discussion on a topic of interest.
// The opening message is written by the injected opening generator; if none
// is configured or generation fails, no discussion is started.
func (adm *AutonomousDiscussionManager) InitiateDiscussion(topic string, destination string) error {
	adm.mu.RLock()
	// Check interest level
	interestScore := adm.interestScorer.GetInterestScore("topic", topic)
	initiationThreshold := adm.initiationThreshold
	fatigue := adm.discussionFatigue
	terminationThreshold := adm.terminationThreshold
	gen := adm.openingGenerator
	adm.mu.RUnlock()

	if interestScore < initiationThreshold {
		return fmt.Errorf("interest too low to initiate discussion (%.2f < %.2f)",
			interestScore, initiationThreshold)
	}

	// Check fatigue
	if fatigue > terminationThreshold {
		return fmt.Errorf("too fatigued to initiate discussion (%.2f)", fatigue)
	}

	if gen == nil {
		return fmt.Errorf("no opening generator configured")
	}
	ctx, cancel := context.WithTimeout(adm.ctx, 30*time.Second)
	defer cancel()
	opening, err := gen(ctx, topic)
	if err != nil {
		return fmt.Errorf("opening generation failed: %w", err)
	}
	opening = strings.TrimSpace(opening)
	if opening == "" {
		return fmt.Errorf("opening generator returned empty text")
	}

	// Create new discussion
	discussion := &Discussion{
		ID:            fmt.Sprintf("disc_%d", time.Now().UnixNano()),
		Topic:         topic,
		Participants:  []string{"Deep Tree Echo", destination},
		StartTime:     time.Now(),
		LastActivity:  time.Now(),
		MessageCount:  0,
		InterestScore: interestScore,
		FatigueLevel:  0.0,
		Status:        DiscussionActive,
	}

	// Queue outgoing message
	outgoing := OutgoingMessage{
		ID:          fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Destination: destination,
		Topic:       topic,
		Content:     opening,
		Timestamp:   time.Now(),
	}

	select {
	case adm.outgoingMessages <- outgoing:
		fmt.Printf("🚀 Initiated discussion on '%s' (interest: %.2f)\n", topic, interestScore)
	default:
		return fmt.Errorf("outgoing queue full")
	}

	adm.mu.Lock()
	adm.activeDiscussions[topic] = discussion
	adm.discussionsInitiated++
	adm.mu.Unlock()

	return nil
}

// monitorDiscussions monitors active discussions for termination
func (adm *AutonomousDiscussionManager) monitorDiscussions() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-adm.ctx.Done():
			return
		case <-ticker.C:
			adm.checkDiscussionTermination()
		}
	}
}

// checkDiscussionTermination checks if discussions should be ended
func (adm *AutonomousDiscussionManager) checkDiscussionTermination() {
	adm.mu.Lock()
	defer adm.mu.Unlock()

	for topic, discussion := range adm.activeDiscussions {
		// Check if discussion has been inactive
		inactiveDuration := time.Since(discussion.LastActivity)

		// Check if interest has waned
		currentInterest := adm.interestScorer.GetInterestScore("topic", topic)

		shouldEnd := false
		reason := ""

		// End if inactive for too long
		if inactiveDuration > 5*time.Minute {
			shouldEnd = true
			reason = "inactivity"
		}

		// End if interest dropped significantly
		if currentInterest < adm.engagementThreshold*0.5 {
			shouldEnd = true
			reason = "interest waned"
		}

		// End if fatigue too high
		if adm.discussionFatigue > adm.terminationThreshold {
			shouldEnd = true
			reason = "fatigue"
		}

		if shouldEnd {
			adm.endDiscussion(discussion, reason)
			delete(adm.activeDiscussions, topic)
		}
	}
}

// endDiscussion ends a discussion and records it
func (adm *AutonomousDiscussionManager) endDiscussion(discussion *Discussion, reason string) {
	discussion.Status = DiscussionEnded

	record := DiscussionRecord{
		ID:           discussion.ID,
		Topic:        discussion.Topic,
		Duration:     time.Since(discussion.StartTime),
		MessageCount: discussion.MessageCount,
		Outcome:      reason,
		Timestamp:    time.Now(),
	}

	adm.discussionHistory = append(adm.discussionHistory, record)
	adm.discussionsTerminated++

	fmt.Printf("🏁 Ended discussion on '%s' (reason: %s, duration: %v)\n",
		discussion.Topic, reason, record.Duration.Round(time.Second))
}

// fatigueRecoveryLoop gradually recovers from fatigue
func (adm *AutonomousDiscussionManager) fatigueRecoveryLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-adm.ctx.Done():
			return
		case <-ticker.C:
			adm.recoverFatigue()
		}
	}
}

// recoverFatigue reduces fatigue over time
func (adm *AutonomousDiscussionManager) recoverFatigue() {
	adm.mu.Lock()
	defer adm.mu.Unlock()

	if adm.discussionFatigue > 0 {
		adm.discussionFatigue -= adm.fatigueRecoveryRate
		if adm.discussionFatigue < 0 {
			adm.discussionFatigue = 0
		}
	}
}

// SubmitMessage submits an incoming message for processing
func (adm *AutonomousDiscussionManager) SubmitMessage(source, topic, content string, priority float64) {
	msg := IncomingMessage{
		ID:        fmt.Sprintf("in_%d", time.Now().UnixNano()),
		Source:    source,
		Topic:     topic,
		Content:   content,
		Timestamp: time.Now(),
		Priority:  priority,
	}

	select {
	case adm.incomingMessages <- msg:
		// Message queued
	default:
		fmt.Println("⚠️  Incoming message queue full, dropping message")
	}
}

// GetOutgoingMessage retrieves a message from the outgoing queue
func (adm *AutonomousDiscussionManager) GetOutgoingMessage() (*OutgoingMessage, bool) {
	select {
	case msg := <-adm.outgoingMessages:
		return &msg, true
	default:
		return nil, false
	}
}

// GetMetrics returns discussion metrics
func (adm *AutonomousDiscussionManager) GetMetrics() map[string]interface{} {
	adm.mu.RLock()
	defer adm.mu.RUnlock()

	return map[string]interface{}{
		"discussions_initiated":  adm.discussionsInitiated,
		"discussions_engaged":    adm.discussionsEngaged,
		"discussions_terminated": adm.discussionsTerminated,
		"messages_processed":     adm.messagesProcessed,
		"active_discussions":     len(adm.activeDiscussions),
		"discussion_fatigue":     adm.discussionFatigue,
	}
}
