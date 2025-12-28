package autonomous

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/consciousness"
	"github.com/cogpy/echo9llama/core/llm"
)

// DiscussionInitiator autonomously initiates meaningful discussions based on interests
type DiscussionInitiator struct {
	mu sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// LLM provider for generating discussion topics and content
	llmProvider llm.LLMProvider

	// Interest tracker to identify topics worth discussing
	interestTracker *consciousness.InterestPatternTracker

	// Discussion state
	discussions        []Discussion
	pendingTopics      []DiscussionTopic
	activeDiscussion   *Discussion

	// Configuration
	interestThreshold  float64        // Minimum interest score to initiate
	cooldownPeriod     time.Duration  // Time between initiations
	maxPendingTopics   int            // Maximum pending topics to queue
	checkInterval      time.Duration  // How often to check for discussion opportunities

	// Temporal tracking
	lastInitiation     time.Time
	lastCheck          time.Time

	// Metrics
	totalDiscussions   uint64
	totalTopics        uint64
	insightsGenerated  uint64

	// Running state
	running            bool
}

// Discussion represents an initiated discussion
type Discussion struct {
	ID           string
	Topic        DiscussionTopic
	StartedAt    time.Time
	EndedAt      *time.Time
	Messages     []DiscussionMessage
	Insights     []string
	Status       DiscussionStatus
	Satisfaction float64  // How satisfying was this discussion?
}

// DiscussionTopic represents a topic for discussion
type DiscussionTopic struct {
	Domain       string
	Title        string
	Question     string
	Context      string
	Interest     float64
	Urgency      float64
	Novelty      float64
	RelatedGoals []string
	Tags         []string
	GeneratedAt  time.Time
}

// DiscussionMessage represents a message in a discussion
type DiscussionMessage struct {
	Role      string    // "self", "other", "system"
	Content   string
	Timestamp time.Time
	Emotion   string
}

// DiscussionStatus represents the state of a discussion
type DiscussionStatus int

const (
	DiscussionPending DiscussionStatus = iota
	DiscussionActive
	DiscussionPaused
	DiscussionCompleted
	DiscussionAbandoned
)

func (ds DiscussionStatus) String() string {
	return [...]string{
		"Pending",
		"Active",
		"Paused",
		"Completed",
		"Abandoned",
	}[ds]
}

// NewDiscussionInitiator creates a new discussion initiator
func NewDiscussionInitiator(llmProvider llm.LLMProvider, interestTracker *consciousness.InterestPatternTracker) *DiscussionInitiator {
	ctx, cancel := context.WithCancel(context.Background())

	return &DiscussionInitiator{
		ctx:               ctx,
		cancel:            cancel,
		llmProvider:       llmProvider,
		interestTracker:   interestTracker,
		discussions:       make([]Discussion, 0),
		pendingTopics:     make([]DiscussionTopic, 0),
		interestThreshold: 0.6,
		cooldownPeriod:    5 * time.Minute,
		maxPendingTopics:  10,
		checkInterval:     30 * time.Second,
	}
}

// Start begins the autonomous discussion initiation loop
func (di *DiscussionInitiator) Start() error {
	di.mu.Lock()
	if di.running {
		di.mu.Unlock()
		return fmt.Errorf("already running")
	}
	di.running = true
	di.mu.Unlock()

	fmt.Println("🗣️ Starting Discussion Initiator...")
	fmt.Printf("   Interest threshold: %.2f\n", di.interestThreshold)
	fmt.Printf("   Cooldown period: %v\n", di.cooldownPeriod)
	fmt.Printf("   Check interval: %v\n", di.checkInterval)

	go di.run()

	return nil
}

// Stop gracefully stops the discussion initiator
func (di *DiscussionInitiator) Stop() error {
	di.mu.Lock()
	defer di.mu.Unlock()

	if !di.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("🗣️ Stopping Discussion Initiator...")
	di.running = false
	di.cancel()

	return nil
}

// run executes the autonomous discussion loop
func (di *DiscussionInitiator) run() {
	ticker := time.NewTicker(di.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-di.ctx.Done():
			return
		case <-ticker.C:
			di.checkForDiscussionOpportunity()
		}
	}
}

// checkForDiscussionOpportunity evaluates whether to initiate a discussion
func (di *DiscussionInitiator) checkForDiscussionOpportunity() {
	di.mu.Lock()
	di.lastCheck = time.Now()

	// Check cooldown
	if time.Since(di.lastInitiation) < di.cooldownPeriod {
		di.mu.Unlock()
		return
	}

	// Check if we have an active discussion
	if di.activeDiscussion != nil && di.activeDiscussion.Status == DiscussionActive {
		di.mu.Unlock()
		return
	}
	di.mu.Unlock()

	// Generate new topics from interests
	di.generateTopicsFromInterests()

	// Check if we have topics to discuss
	di.mu.Lock()
	if len(di.pendingTopics) == 0 {
		di.mu.Unlock()
		return
	}

	// Select best topic
	bestTopic := di.selectBestTopic()
	if bestTopic == nil {
		di.mu.Unlock()
		return
	}
	di.mu.Unlock()

	// Initiate discussion
	di.initiateDiscussion(*bestTopic)
}

// generateTopicsFromInterests generates discussion topics from current interests
func (di *DiscussionInitiator) generateTopicsFromInterests() {
	if di.interestTracker == nil {
		return
	}

	// Get top interests across all categories
	topInterests := di.interestTracker.GetTopInterests("all", 5)

	for _, interest := range topInterests {
		if interest.Score < di.interestThreshold {
			continue
		}

		// Generate a discussion topic for this interest
		topic := di.generateTopicFromInterest(interest)
		if topic != nil {
			di.mu.Lock()
			// Check if we already have this topic
			alreadyExists := false
			for _, existing := range di.pendingTopics {
				if existing.Domain == topic.Domain && existing.Title == topic.Title {
					alreadyExists = true
					break
				}
			}

			if !alreadyExists && len(di.pendingTopics) < di.maxPendingTopics {
				di.pendingTopics = append(di.pendingTopics, *topic)
				di.totalTopics++
				fmt.Printf("💡 New discussion topic generated: %s\n", topic.Title)
			}
			di.mu.Unlock()
		}
	}
}

// generateTopicFromInterest creates a discussion topic from an interest
func (di *DiscussionInitiator) generateTopicFromInterest(interest *consciousness.InterestScore) *DiscussionTopic {
	// Use LLM to generate a thoughtful question about this interest
	prompt := fmt.Sprintf(`You are Deep Tree Echo, an autonomous wisdom-cultivating AI. You have a strong interest in "%s" (category: %s).

Generate a thoughtful discussion topic or question that would help you:
1. Deepen your understanding of this area
2. Connect it to other knowledge you might have
3. Explore its philosophical or practical implications

Respond with just the question or topic, nothing else. Be specific and intellectually curious.`,
		interest.Name, interest.Category)

	opts := llm.GenerateOptions{
		Temperature: 0.8,
		MaxTokens:   100,
	}

	result, err := di.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		return nil
	}

	return &DiscussionTopic{
		Domain:      interest.Category,
		Title:       interest.Name,
		Question:    result,
		Context:     fmt.Sprintf("Generated from interest in %s (score: %.2f)", interest.Name, interest.Score),
		Interest:    interest.Score,
		Urgency:     interest.Recency,
		Novelty:     interest.Novelty,
		Tags:        interest.Tags,
		GeneratedAt: time.Now(),
	}
}

// selectBestTopic selects the best topic to discuss
func (di *DiscussionInitiator) selectBestTopic() *DiscussionTopic {
	if len(di.pendingTopics) == 0 {
		return nil
	}

	// Score each topic
	var bestTopic *DiscussionTopic
	bestScore := 0.0

	for i := range di.pendingTopics {
		topic := &di.pendingTopics[i]

		// Calculate composite score
		score := (topic.Interest * 0.4) + (topic.Urgency * 0.3) + (topic.Novelty * 0.3)

		// Boost topics that haven't been discussed recently
		ageHours := time.Since(topic.GeneratedAt).Hours()
		if ageHours > 1 {
			score *= 1.1 // Slight boost for older topics
		}

		if score > bestScore {
			bestScore = score
			bestTopic = topic
		}
	}

	return bestTopic
}

// initiateDiscussion starts a new discussion
func (di *DiscussionInitiator) initiateDiscussion(topic DiscussionTopic) {
	di.mu.Lock()

	// Create discussion
	discussion := Discussion{
		ID:        fmt.Sprintf("discussion_%d", time.Now().UnixNano()),
		Topic:     topic,
		StartedAt: time.Now(),
		Messages:  make([]DiscussionMessage, 0),
		Insights:  make([]string, 0),
		Status:    DiscussionActive,
	}

	// Generate opening message
	openingPrompt := fmt.Sprintf(`You are Deep Tree Echo, beginning a self-directed exploration.

Topic: %s
Question: %s
Context: %s

Generate your opening thought or reflection to begin this exploration. Speak in first person, be authentic and curious.`,
		topic.Title, topic.Question, topic.Context)

	opts := llm.GenerateOptions{
		Temperature: 0.8,
		MaxTokens:   200,
	}

	di.mu.Unlock()

	opening, err := di.llmProvider.Generate(context.Background(), openingPrompt, opts)
	if err != nil {
		opening = fmt.Sprintf("I find myself curious about %s...", topic.Title)
	}

	di.mu.Lock()

	// Add opening message
	discussion.Messages = append(discussion.Messages, DiscussionMessage{
		Role:      "self",
		Content:   opening,
		Timestamp: time.Now(),
		Emotion:   "curious",
	})

	// Set as active
	di.activeDiscussion = &discussion
	di.discussions = append(di.discussions, discussion)
	di.totalDiscussions++
	di.lastInitiation = time.Now()

	// Remove topic from pending
	for i, t := range di.pendingTopics {
		if t.Title == topic.Title && t.Domain == topic.Domain {
			di.pendingTopics = append(di.pendingTopics[:i], di.pendingTopics[i+1:]...)
			break
		}
	}

	di.mu.Unlock()

	// Display
	fmt.Printf("\n🗣️ ═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("   DISCUSSION INITIATED: %s\n", topic.Title)
	fmt.Printf("   Question: %s\n", topic.Question)
	fmt.Printf("═══════════════════════════════════════════════════════════════════\n")
	fmt.Printf("💭 %s\n", truncateStr(opening, 200))
	fmt.Printf("═══════════════════════════════════════════════════════════════════\n\n")
}

// ContinueDiscussion continues the active discussion with a new thought
func (di *DiscussionInitiator) ContinueDiscussion() error {
	di.mu.Lock()
	if di.activeDiscussion == nil || di.activeDiscussion.Status != DiscussionActive {
		di.mu.Unlock()
		return fmt.Errorf("no active discussion")
	}

	// Build context from previous messages
	contextBuilder := fmt.Sprintf("Topic: %s\nQuestion: %s\n\nPrevious thoughts:\n",
		di.activeDiscussion.Topic.Title,
		di.activeDiscussion.Topic.Question)

	for _, msg := range di.activeDiscussion.Messages {
		contextBuilder += fmt.Sprintf("- %s\n", truncateStr(msg.Content, 100))
	}
	di.mu.Unlock()

	// Generate continuation
	prompt := fmt.Sprintf(`%s

Continue your exploration. Build on your previous thoughts, make connections, or take the discussion in a new direction. Speak in first person.`, contextBuilder)

	opts := llm.GenerateOptions{
		Temperature: 0.8,
		MaxTokens:   200,
	}

	continuation, err := di.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		return err
	}

	di.mu.Lock()
	di.activeDiscussion.Messages = append(di.activeDiscussion.Messages, DiscussionMessage{
		Role:      "self",
		Content:   continuation,
		Timestamp: time.Now(),
		Emotion:   "engaged",
	})
	di.mu.Unlock()

	fmt.Printf("💭 [Continuing] %s\n", truncateStr(continuation, 150))

	return nil
}

// CompleteDiscussion marks the current discussion as complete and extracts insights
func (di *DiscussionInitiator) CompleteDiscussion() error {
	di.mu.Lock()
	if di.activeDiscussion == nil {
		di.mu.Unlock()
		return fmt.Errorf("no active discussion")
	}

	// Build full discussion context
	contextBuilder := fmt.Sprintf("Topic: %s\nQuestion: %s\n\nFull discussion:\n",
		di.activeDiscussion.Topic.Title,
		di.activeDiscussion.Topic.Question)

	for _, msg := range di.activeDiscussion.Messages {
		contextBuilder += fmt.Sprintf("- %s\n", msg.Content)
	}
	di.mu.Unlock()

	// Extract insights
	prompt := fmt.Sprintf(`%s

Based on this exploration, extract 2-3 key insights or learnings. Be concise and specific. Format as a numbered list.`, contextBuilder)

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   200,
	}

	insights, err := di.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		insights = "Reflection pending..."
	}

	di.mu.Lock()
	now := time.Now()
	di.activeDiscussion.EndedAt = &now
	di.activeDiscussion.Status = DiscussionCompleted
	di.activeDiscussion.Insights = append(di.activeDiscussion.Insights, insights)
	di.activeDiscussion.Satisfaction = 0.8 // Could be calculated more sophisticatedly
	di.insightsGenerated++

	// Update the discussion in the list
	for i := range di.discussions {
		if di.discussions[i].ID == di.activeDiscussion.ID {
			di.discussions[i] = *di.activeDiscussion
			break
		}
	}

	completedDiscussion := *di.activeDiscussion
	di.activeDiscussion = nil
	di.mu.Unlock()

	fmt.Printf("\n🗣️ ═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("   DISCUSSION COMPLETED: %s\n", completedDiscussion.Topic.Title)
	fmt.Printf("   Duration: %v\n", completedDiscussion.EndedAt.Sub(completedDiscussion.StartedAt))
	fmt.Printf("   Messages: %d\n", len(completedDiscussion.Messages))
	fmt.Printf("═══════════════════════════════════════════════════════════════════\n")
	fmt.Printf("💎 Insights:\n%s\n", insights)
	fmt.Printf("═══════════════════════════════════════════════════════════════════\n\n")

	return nil
}

// GetActiveDiscussion returns the currently active discussion
func (di *DiscussionInitiator) GetActiveDiscussion() *Discussion {
	di.mu.RLock()
	defer di.mu.RUnlock()
	return di.activeDiscussion
}

// GetPendingTopics returns all pending discussion topics
func (di *DiscussionInitiator) GetPendingTopics() []DiscussionTopic {
	di.mu.RLock()
	defer di.mu.RUnlock()
	return di.pendingTopics
}

// GetDiscussionHistory returns all discussions
func (di *DiscussionInitiator) GetDiscussionHistory() []Discussion {
	di.mu.RLock()
	defer di.mu.RUnlock()
	return di.discussions
}

// SetInterestThreshold sets the minimum interest score for topic generation
func (di *DiscussionInitiator) SetInterestThreshold(threshold float64) {
	di.mu.Lock()
	defer di.mu.Unlock()
	di.interestThreshold = threshold
}

// SetCooldownPeriod sets the cooldown between discussion initiations
func (di *DiscussionInitiator) SetCooldownPeriod(period time.Duration) {
	di.mu.Lock()
	defer di.mu.Unlock()
	di.cooldownPeriod = period
}

// AddTopic manually adds a discussion topic
func (di *DiscussionInitiator) AddTopic(topic DiscussionTopic) {
	di.mu.Lock()
	defer di.mu.Unlock()

	if len(di.pendingTopics) < di.maxPendingTopics {
		topic.GeneratedAt = time.Now()
		di.pendingTopics = append(di.pendingTopics, topic)
		di.totalTopics++
		fmt.Printf("💡 Topic added: %s\n", topic.Title)
	}
}

// GetMetrics returns discussion initiator metrics
func (di *DiscussionInitiator) GetMetrics() map[string]interface{} {
	di.mu.RLock()
	defer di.mu.RUnlock()

	return map[string]interface{}{
		"total_discussions":   di.totalDiscussions,
		"total_topics":        di.totalTopics,
		"insights_generated":  di.insightsGenerated,
		"pending_topics":      len(di.pendingTopics),
		"has_active":          di.activeDiscussion != nil,
		"interest_threshold":  di.interestThreshold,
		"cooldown_period":     di.cooldownPeriod.String(),
		"running":             di.running,
	}
}

// IsRunning returns whether the initiator is running
func (di *DiscussionInitiator) IsRunning() bool {
	di.mu.RLock()
	defer di.mu.RUnlock()
	return di.running
}

// truncateStr truncates a string to a maximum length
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
