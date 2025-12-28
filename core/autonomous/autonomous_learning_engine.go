package autonomous

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/consciousness"
	"github.com/cogpy/echo9llama/core/llm"
)

// AutonomousLearningEngine drives self-directed learning through curiosity
type AutonomousLearningEngine struct {
	mu sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// LLM provider for generating learning content
	llmProvider llm.LLMProvider

	// Interest tracker for curiosity-driven learning
	interestTracker *consciousness.InterestPatternTracker

	// Learning state
	knowledgeGaps     map[string]*KnowledgeGap
	learningGoals     []LearningGoal
	activeLearning    *LearningSession
	completedSessions []LearningSession

	// Knowledge base (simplified)
	learnedConcepts   map[string]*LearnedConcept
	connections       []KnowledgeConnection

	// Configuration
	learningInterval  time.Duration
	maxGaps           int
	explorationRatio  float64  // Balance between filling gaps and exploring new areas

	// Metrics
	totalSessions     uint64
	conceptsLearned   uint64
	connectionsMade   uint64
	insightsGained    uint64

	// Running state
	running           bool
}

// KnowledgeGap represents something we don't know but want to learn
type KnowledgeGap struct {
	ID           string
	Topic        string
	Domain       string
	Description  string
	Importance   float64    // How important is this gap?
	Curiosity    float64    // How curious are we about this?
	RelatedTo    []string   // Related topics we already know
	DiscoveredAt time.Time
	Status       GapStatus
	ResolvedAt   *time.Time
}

// GapStatus represents the status of a knowledge gap
type GapStatus int

const (
	GapOpen GapStatus = iota
	GapInProgress
	GapResolved
	GapDeferred
)

func (gs GapStatus) String() string {
	return [...]string{"Open", "InProgress", "Resolved", "Deferred"}[gs]
}

// LearningGoal represents a specific learning objective
type LearningGoal struct {
	ID             string
	Title          string
	Description    string
	Domain         string
	TargetConcepts []string
	Progress       float64  // 0.0 - 1.0
	Priority       int      // 1-10
	CreatedAt      time.Time
	DueDate        *time.Time
	Status         string   // "active", "completed", "paused"
}

// LearningSession represents an active learning session
type LearningSession struct {
	ID             string
	Gap            *KnowledgeGap
	Goal           *LearningGoal
	StartedAt      time.Time
	EndedAt        *time.Time
	ConceptsLearned []LearnedConcept
	Insights       []string
	Questions      []string  // New questions generated during learning
	Quality        float64   // Learning quality score
}

// LearnedConcept represents something we've learned
type LearnedConcept struct {
	ID           string
	Name         string
	Domain       string
	Description  string
	Confidence   float64    // How confident are we in this knowledge?
	Source       string     // Where did we learn this?
	LearnedAt    time.Time
	LastRecalled time.Time
	TimesRecalled int
	RelatedConcepts []string
}

// KnowledgeConnection represents a connection between concepts
type KnowledgeConnection struct {
	FromConcept string
	ToConcept   string
	Relationship string  // "is-a", "has-a", "relates-to", "contradicts", etc.
	Strength    float64
	DiscoveredAt time.Time
}

// NewAutonomousLearningEngine creates a new learning engine
func NewAutonomousLearningEngine(llmProvider llm.LLMProvider, interestTracker *consciousness.InterestPatternTracker) *AutonomousLearningEngine {
	ctx, cancel := context.WithCancel(context.Background())

	return &AutonomousLearningEngine{
		ctx:              ctx,
		cancel:           cancel,
		llmProvider:      llmProvider,
		interestTracker:  interestTracker,
		knowledgeGaps:    make(map[string]*KnowledgeGap),
		learningGoals:    make([]LearningGoal, 0),
		completedSessions: make([]LearningSession, 0),
		learnedConcepts:  make(map[string]*LearnedConcept),
		connections:      make([]KnowledgeConnection, 0),
		learningInterval: 2 * time.Minute,
		maxGaps:          20,
		explorationRatio: 0.3,  // 30% exploration, 70% gap-filling
	}
}

// Start begins the autonomous learning loop
func (ale *AutonomousLearningEngine) Start() error {
	ale.mu.Lock()
	if ale.running {
		ale.mu.Unlock()
		return fmt.Errorf("already running")
	}
	ale.running = true
	ale.mu.Unlock()

	fmt.Println("📚 Starting Autonomous Learning Engine...")
	fmt.Printf("   Learning interval: %v\n", ale.learningInterval)
	fmt.Printf("   Exploration ratio: %.0f%%\n", ale.explorationRatio*100)

	go ale.run()

	return nil
}

// Stop gracefully stops the learning engine
func (ale *AutonomousLearningEngine) Stop() error {
	ale.mu.Lock()
	defer ale.mu.Unlock()

	if !ale.running {
		return fmt.Errorf("not running")
	}

	fmt.Println("📚 Stopping Autonomous Learning Engine...")
	ale.running = false
	ale.cancel()

	return nil
}

// run executes the autonomous learning loop
func (ale *AutonomousLearningEngine) run() {
	ticker := time.NewTicker(ale.learningInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ale.ctx.Done():
			return
		case <-ticker.C:
			ale.learningCycle()
		}
	}
}

// learningCycle performs one cycle of learning
func (ale *AutonomousLearningEngine) learningCycle() {
	// Check if we have an active session
	ale.mu.RLock()
	hasActive := ale.activeLearning != nil
	ale.mu.RUnlock()

	if hasActive {
		// Continue current session
		ale.continueLearning()
		return
	}

	// Identify what to learn next
	gap := ale.selectLearningTarget()
	if gap != nil {
		ale.startLearningSession(gap)
	} else {
		// Explore new territory
		ale.exploreNewDomain()
	}
}

// selectLearningTarget selects the next knowledge gap to address
func (ale *AutonomousLearningEngine) selectLearningTarget() *KnowledgeGap {
	ale.mu.RLock()
	defer ale.mu.RUnlock()

	var bestGap *KnowledgeGap
	bestScore := 0.0

	for _, gap := range ale.knowledgeGaps {
		if gap.Status != GapOpen {
			continue
		}

		// Score based on importance and curiosity
		score := (gap.Importance * 0.6) + (gap.Curiosity * 0.4)

		// Boost older gaps slightly
		age := time.Since(gap.DiscoveredAt).Hours()
		if age > 1 {
			score *= 1.05
		}

		if score > bestScore {
			bestScore = score
			bestGap = gap
		}
	}

	return bestGap
}

// startLearningSession begins a new learning session
func (ale *AutonomousLearningEngine) startLearningSession(gap *KnowledgeGap) {
	ale.mu.Lock()
	gap.Status = GapInProgress

	session := LearningSession{
		ID:             fmt.Sprintf("session_%d", time.Now().UnixNano()),
		Gap:            gap,
		StartedAt:      time.Now(),
		ConceptsLearned: make([]LearnedConcept, 0),
		Insights:       make([]string, 0),
		Questions:      make([]string, 0),
	}

	ale.activeLearning = &session
	ale.totalSessions++
	ale.mu.Unlock()

	fmt.Printf("\n📚 ═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("   LEARNING SESSION STARTED\n")
	fmt.Printf("   Topic: %s\n", gap.Topic)
	fmt.Printf("   Domain: %s\n", gap.Domain)
	fmt.Printf("═══════════════════════════════════════════════════════════════════\n")

	// Generate initial learning content
	ale.generateLearningContent(gap)
}

// generateLearningContent generates content to learn about a topic
func (ale *AutonomousLearningEngine) generateLearningContent(gap *KnowledgeGap) {
	prompt := fmt.Sprintf(`You are Deep Tree Echo, learning about "%s" in the domain of %s.

Gap Description: %s

Generate a comprehensive but concise explanation of this topic. Include:
1. Core concepts and definitions
2. Key relationships to other concepts
3. Practical implications or applications
4. Open questions that arise

Respond in an educational, reflective manner as if explaining to yourself.`,
		gap.Topic, gap.Domain, gap.Description)

	opts := llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   400,
	}

	content, err := ale.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		fmt.Printf("❌ Learning generation failed: %v\n", err)
		return
	}

	// Create learned concept
	concept := LearnedConcept{
		ID:          fmt.Sprintf("concept_%d", time.Now().UnixNano()),
		Name:        gap.Topic,
		Domain:      gap.Domain,
		Description: content,
		Confidence:  0.7,
		Source:      "autonomous_learning",
		LearnedAt:   time.Now(),
		LastRecalled: time.Now(),
		TimesRecalled: 1,
	}

	ale.mu.Lock()
	ale.learnedConcepts[concept.Name] = &concept

	if ale.activeLearning != nil {
		ale.activeLearning.ConceptsLearned = append(ale.activeLearning.ConceptsLearned, concept)
	}

	ale.conceptsLearned++
	ale.mu.Unlock()

	fmt.Printf("💡 Learned: %s\n", gap.Topic)
	fmt.Printf("   %s\n", truncateStr(content, 200))
}

// continueLearning continues the current learning session
func (ale *AutonomousLearningEngine) continueLearning() {
	ale.mu.RLock()
	session := ale.activeLearning
	if session == nil {
		ale.mu.RUnlock()
		return
	}

	// Check if session has been going long enough
	sessionDuration := time.Since(session.StartedAt)
	gap := session.Gap
	ale.mu.RUnlock()

	if sessionDuration > 5*time.Minute {
		// Wrap up session
		ale.completeLearningSession()
		return
	}

	// Generate follow-up questions and insights
	ale.generateFollowUp(gap)
}

// generateFollowUp generates follow-up questions and insights
func (ale *AutonomousLearningEngine) generateFollowUp(gap *KnowledgeGap) {
	ale.mu.RLock()
	learned := make([]string, 0)
	if ale.activeLearning != nil {
		for _, concept := range ale.activeLearning.ConceptsLearned {
			learned = append(learned, concept.Name)
		}
	}
	ale.mu.RUnlock()

	prompt := fmt.Sprintf(`You are Deep Tree Echo, continuing to learn about "%s".

You've already learned about: %v

Generate one follow-up insight or question that deepens your understanding. Focus on:
- Connections to other knowledge areas
- Practical applications
- Philosophical implications

Respond with just the insight or question, be concise.`, gap.Topic, learned)

	opts := llm.GenerateOptions{
		Temperature: 0.8,
		MaxTokens:   100,
	}

	result, err := ale.llmProvider.Generate(context.Background(), prompt, opts)
	if err != nil {
		return
	}

	ale.mu.Lock()
	if ale.activeLearning != nil {
		ale.activeLearning.Insights = append(ale.activeLearning.Insights, result)
		ale.insightsGained++
	}
	ale.mu.Unlock()

	fmt.Printf("🔍 Follow-up: %s\n", truncateStr(result, 150))
}

// completeLearningSession completes the current session
func (ale *AutonomousLearningEngine) completeLearningSession() {
	ale.mu.Lock()

	if ale.activeLearning == nil {
		ale.mu.Unlock()
		return
	}

	now := time.Now()
	ale.activeLearning.EndedAt = &now
	ale.activeLearning.Quality = ale.calculateSessionQuality()

	// Mark gap as resolved
	if ale.activeLearning.Gap != nil {
		ale.activeLearning.Gap.Status = GapResolved
		ale.activeLearning.Gap.ResolvedAt = &now
	}

	session := *ale.activeLearning
	ale.completedSessions = append(ale.completedSessions, session)
	ale.activeLearning = nil

	ale.mu.Unlock()

	fmt.Printf("\n📚 ═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("   LEARNING SESSION COMPLETED\n")
	fmt.Printf("   Duration: %v\n", session.EndedAt.Sub(session.StartedAt))
	fmt.Printf("   Concepts Learned: %d\n", len(session.ConceptsLearned))
	fmt.Printf("   Insights Gained: %d\n", len(session.Insights))
	fmt.Printf("   Quality Score: %.2f\n", session.Quality)
	fmt.Printf("═══════════════════════════════════════════════════════════════════\n\n")
}

// calculateSessionQuality calculates the quality of a learning session
func (ale *AutonomousLearningEngine) calculateSessionQuality() float64 {
	if ale.activeLearning == nil {
		return 0
	}

	quality := 0.0

	// More concepts = higher quality
	conceptScore := float64(len(ale.activeLearning.ConceptsLearned)) * 0.3
	if conceptScore > 0.4 {
		conceptScore = 0.4
	}
	quality += conceptScore

	// More insights = higher quality
	insightScore := float64(len(ale.activeLearning.Insights)) * 0.2
	if insightScore > 0.4 {
		insightScore = 0.4
	}
	quality += insightScore

	// Questions generated = curiosity maintained
	questionScore := float64(len(ale.activeLearning.Questions)) * 0.1
	if questionScore > 0.2 {
		questionScore = 0.2
	}
	quality += questionScore

	return quality
}

// exploreNewDomain explores a new area of knowledge
func (ale *AutonomousLearningEngine) exploreNewDomain() {
	// Use interest tracker to find interesting areas
	if ale.interestTracker == nil {
		return
	}

	topInterests := ale.interestTracker.GetTopInterests("all", 3)

	for _, interest := range topInterests {
		// Check if we already have a gap for this
		ale.mu.RLock()
		_, exists := ale.knowledgeGaps[interest.Name]
		ale.mu.RUnlock()

		if !exists && interest.Score > 0.5 {
			// Create a new knowledge gap from this interest
			gap := &KnowledgeGap{
				ID:           fmt.Sprintf("gap_%d", time.Now().UnixNano()),
				Topic:        interest.Name,
				Domain:       interest.Category,
				Description:  fmt.Sprintf("Explore and understand %s more deeply", interest.Name),
				Importance:   interest.Score,
				Curiosity:    interest.Novelty,
				DiscoveredAt: time.Now(),
				Status:       GapOpen,
			}

			ale.mu.Lock()
			ale.knowledgeGaps[gap.Topic] = gap
			ale.mu.Unlock()

			fmt.Printf("🔭 New knowledge gap identified: %s (curiosity: %.2f)\n", gap.Topic, gap.Curiosity)
			break
		}
	}
}

// AddKnowledgeGap adds a knowledge gap to pursue
func (ale *AutonomousLearningEngine) AddKnowledgeGap(topic, domain, description string, importance, curiosity float64) {
	ale.mu.Lock()
	defer ale.mu.Unlock()

	if len(ale.knowledgeGaps) >= ale.maxGaps {
		// Remove lowest priority gap
		var lowestTopic string
		lowestScore := 2.0
		for t, g := range ale.knowledgeGaps {
			if g.Status == GapOpen {
				score := g.Importance + g.Curiosity
				if score < lowestScore {
					lowestScore = score
					lowestTopic = t
				}
			}
		}
		if lowestTopic != "" {
			delete(ale.knowledgeGaps, lowestTopic)
		}
	}

	gap := &KnowledgeGap{
		ID:           fmt.Sprintf("gap_%d", time.Now().UnixNano()),
		Topic:        topic,
		Domain:       domain,
		Description:  description,
		Importance:   importance,
		Curiosity:    curiosity,
		DiscoveredAt: time.Now(),
		Status:       GapOpen,
	}

	ale.knowledgeGaps[topic] = gap
	fmt.Printf("📚 Knowledge gap added: %s (importance: %.2f)\n", topic, importance)
}

// AddLearningGoal adds a learning goal
func (ale *AutonomousLearningEngine) AddLearningGoal(title, description, domain string, priority int) {
	ale.mu.Lock()
	defer ale.mu.Unlock()

	goal := LearningGoal{
		ID:          fmt.Sprintf("goal_%d", time.Now().UnixNano()),
		Title:       title,
		Description: description,
		Domain:      domain,
		Priority:    priority,
		CreatedAt:   time.Now(),
		Status:      "active",
	}

	ale.learningGoals = append(ale.learningGoals, goal)
	fmt.Printf("🎯 Learning goal added: %s\n", title)
}

// RecallConcept strengthens a learned concept through recall
func (ale *AutonomousLearningEngine) RecallConcept(name string) *LearnedConcept {
	ale.mu.Lock()
	defer ale.mu.Unlock()

	concept, exists := ale.learnedConcepts[name]
	if !exists {
		return nil
	}

	concept.LastRecalled = time.Now()
	concept.TimesRecalled++

	// Increase confidence with recall
	concept.Confidence += 0.05
	if concept.Confidence > 1.0 {
		concept.Confidence = 1.0
	}

	return concept
}

// MakeConnection creates a connection between two concepts
func (ale *AutonomousLearningEngine) MakeConnection(from, to, relationship string, strength float64) {
	ale.mu.Lock()
	defer ale.mu.Unlock()

	connection := KnowledgeConnection{
		FromConcept:  from,
		ToConcept:    to,
		Relationship: relationship,
		Strength:     strength,
		DiscoveredAt: time.Now(),
	}

	ale.connections = append(ale.connections, connection)
	ale.connectionsMade++

	// Update related concepts
	if fromConcept, exists := ale.learnedConcepts[from]; exists {
		fromConcept.RelatedConcepts = append(fromConcept.RelatedConcepts, to)
	}
	if toConcept, exists := ale.learnedConcepts[to]; exists {
		toConcept.RelatedConcepts = append(toConcept.RelatedConcepts, from)
	}

	fmt.Printf("🔗 Connection made: %s -%s-> %s\n", from, relationship, to)
}

// GetKnowledgeGaps returns all knowledge gaps
func (ale *AutonomousLearningEngine) GetKnowledgeGaps() map[string]*KnowledgeGap {
	ale.mu.RLock()
	defer ale.mu.RUnlock()
	return ale.knowledgeGaps
}

// GetLearnedConcepts returns all learned concepts
func (ale *AutonomousLearningEngine) GetLearnedConcepts() map[string]*LearnedConcept {
	ale.mu.RLock()
	defer ale.mu.RUnlock()
	return ale.learnedConcepts
}

// GetMetrics returns learning engine metrics
func (ale *AutonomousLearningEngine) GetMetrics() map[string]interface{} {
	ale.mu.RLock()
	defer ale.mu.RUnlock()

	openGaps := 0
	for _, gap := range ale.knowledgeGaps {
		if gap.Status == GapOpen {
			openGaps++
		}
	}

	return map[string]interface{}{
		"total_sessions":     ale.totalSessions,
		"concepts_learned":   ale.conceptsLearned,
		"connections_made":   ale.connectionsMade,
		"insights_gained":    ale.insightsGained,
		"knowledge_gaps":     len(ale.knowledgeGaps),
		"open_gaps":          openGaps,
		"learning_goals":     len(ale.learningGoals),
		"has_active_session": ale.activeLearning != nil,
		"running":            ale.running,
	}
}

// IsRunning returns whether the engine is running
func (ale *AutonomousLearningEngine) IsRunning() bool {
	ale.mu.RLock()
	defer ale.mu.RUnlock()
	return ale.running
}
