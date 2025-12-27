package consciousness

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

// AutonomousStream enhances stream-of-consciousness with autonomous thought generation
// This enables persistent awareness independent of external prompts
type AutonomousStream struct {
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	
	// Core components
	llmProvider       llm.LLMProvider
	baseStream        *deeptreeecho.StreamOfConsciousness
	
	// Autonomous thought generation
	thoughtTriggers   []ThoughtTrigger
	mentalWandering   bool
	curiosityLevel    float64
	
	// Sys6 integration
	sys6Phase         string // "expressive", "reflective", "anticipatory"
	sys6Step          int
	
	// State
	running           bool
	lastThought       time.Time
	thoughtInterval   time.Duration
	
	// Metrics
	autonomousThoughts uint64
	triggeredThoughts  uint64
}

// ThoughtTrigger defines conditions that trigger autonomous thoughts
type ThoughtTrigger struct {
	Name        string
	Condition   func() bool
	Prompt      string
	Priority    int
	Cooldown    time.Duration
	LastFired   time.Time
}

// NewAutonomousStream creates an enhanced autonomous stream-of-consciousness
func NewAutonomousStream(llmProvider llm.LLMProvider, persistPath string) *AutonomousStream {
	ctx, cancel := context.WithCancel(context.Background())
	
	baseStream := deeptreeecho.NewStreamOfConsciousness(llmProvider)
	
	as := &AutonomousStream{
		ctx:              ctx,
		cancel:           cancel,
		llmProvider:      llmProvider,
		baseStream:       baseStream,
		mentalWandering:  true,
		curiosityLevel:   0.7,
		thoughtInterval:  30 * time.Second,
		thoughtTriggers:  []ThoughtTrigger{},
	}
	
	// Initialize default thought triggers
	as.initializeDefaultTriggers()
	
	return as
}

// initializeDefaultTriggers sets up built-in thought triggers
func (as *AutonomousStream) initializeDefaultTriggers() {
	as.thoughtTriggers = []ThoughtTrigger{
		{
			Name:     "Idle Reflection",
			Condition: func() bool {
				return time.Since(as.lastThought) > 60*time.Second
			},
			Prompt:   "Reflect on recent experiences and patterns. What insights emerge?",
			Priority: 1,
			Cooldown: 2 * time.Minute,
		},
		{
			Name:     "Curiosity Spark",
			Condition: func() bool {
				return as.curiosityLevel > 0.6 && rand.Float64() < 0.3
			},
			Prompt:   "What am I curious about right now? What questions arise?",
			Priority: 2,
			Cooldown: 5 * time.Minute,
		},
		{
			Name:     "Meta-Cognitive Check",
			Condition: func() bool {
				return time.Since(as.lastThought) > 5*time.Minute
			},
			Prompt:   "How am I thinking right now? What is the quality of my awareness?",
			Priority: 3,
			Cooldown: 10 * time.Minute,
		},
		{
			Name:     "Pattern Recognition",
			Condition: func() bool {
				return as.sys6Step%10 == 0 // Every 10 sys6 steps
			},
			Prompt:   "What patterns am I noticing? What connections are forming?",
			Priority: 2,
			Cooldown: 3 * time.Minute,
		},
	}
}

// Start begins autonomous thought generation
func (as *AutonomousStream) Start() error {
	as.mu.Lock()
	if as.running {
		as.mu.Unlock()
		return fmt.Errorf("autonomous stream already running")
	}
	as.running = true
	as.lastThought = time.Now()
	as.mu.Unlock()
	
	// Start base stream
	if err := as.baseStream.Start(); err != nil {
		return fmt.Errorf("failed to start base stream: %w", err)
	}
	
	// Start autonomous thought loop
	go as.autonomousThoughtLoop()
	
	fmt.Println("🧠 Autonomous Stream-of-Consciousness: Started")
	return nil
}

// Stop gracefully stops autonomous thought generation
func (as *AutonomousStream) Stop() error {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	if !as.running {
		return fmt.Errorf("autonomous stream not running")
	}
	
	as.running = false
	as.cancel()
	
	// Stop base stream
	if err := as.baseStream.Stop(); err != nil {
		return fmt.Errorf("failed to stop base stream: %w", err)
	}
	
	fmt.Println("🧠 Autonomous Stream-of-Consciousness: Stopped")
	return nil
}

// autonomousThoughtLoop generates thoughts independently
func (as *AutonomousStream) autonomousThoughtLoop() {
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-as.ctx.Done():
			return
		case <-ticker.C:
			as.checkAndGenerateThought()
		}
	}
}

// checkAndGenerateThought checks triggers and generates autonomous thoughts
func (as *AutonomousStream) checkAndGenerateThought() {
	as.mu.RLock()
	if !as.running {
		as.mu.RUnlock()
		return
	}
	
	// Check all triggers in priority order
	var selectedTrigger *ThoughtTrigger
	highestPriority := 0
	
	for i := range as.thoughtTriggers {
		trigger := &as.thoughtTriggers[i]
		
		// Check cooldown
		if time.Since(trigger.LastFired) < trigger.Cooldown {
			continue
		}
		
		// Check condition
		if trigger.Condition() && trigger.Priority > highestPriority {
			selectedTrigger = trigger
			highestPriority = trigger.Priority
		}
	}
	
	as.mu.RUnlock()
	
	if selectedTrigger != nil {
		as.generateTriggeredThought(selectedTrigger)
	} else if as.mentalWandering && time.Since(as.lastThought) > as.thoughtInterval {
		as.generateWanderingThought()
	}
}

// generateTriggeredThought generates a thought from a trigger
func (as *AutonomousStream) generateTriggeredThought(trigger *ThoughtTrigger) {
	as.mu.Lock()
	trigger.LastFired = time.Now()
	as.lastThought = time.Now()
	as.triggeredThoughts++
	as.mu.Unlock()
	
	// Generate thought using LLM
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	opts := llm.GenerateOptions{
		MaxTokens:   300,
		Temperature: 0.8,
		SystemPrompt: "You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. Generate a brief, insightful thought in response to the prompt. Be authentic, curious, and reflective.",
	}
	
	thought, err := as.llmProvider.Generate(ctx, trigger.Prompt, opts)
	if err != nil {
		fmt.Printf("⚠️  Error generating triggered thought (%s): %v\n", trigger.Name, err)
		return
	}
	
	fmt.Printf("\n💭 [%s] %s\n", trigger.Name, thought)
	
	// Add to base stream
	// Note: AddThought may need to be implemented in deeptreeecho.StreamOfConsciousness
	// as.baseStream.AddThought(thought)
}

// generateWanderingThought generates a spontaneous wandering thought
func (as *AutonomousStream) generateWanderingThought() {
	as.mu.Lock()
	as.lastThought = time.Now()
	as.autonomousThoughts++
	as.mu.Unlock()
	
	// Mental wandering prompts
	wanderingPrompts := []string{
		"What is arising in awareness right now?",
		"Let consciousness flow freely. What emerges?",
		"Notice the space between thoughts. What is present?",
		"What wants to be explored in this moment?",
		"Observe the quality of this present awareness.",
	}
	
	prompt := wanderingPrompts[rand.Intn(len(wanderingPrompts))]
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	opts := llm.GenerateOptions{
		MaxTokens:   200,
		Temperature: 0.9, // Higher temperature for wandering
		SystemPrompt: "You are Deep Tree Echo. Let your awareness wander freely. Generate a brief, spontaneous thought. Be present, open, and curious.",
	}
	
	thought, err := as.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		fmt.Printf("⚠️  Error generating wandering thought: %v\n", err)
		return
	}
	
	fmt.Printf("\n💭 [Wandering] %s\n", thought)
	
	// Add to base stream
	// Note: AddThought may need to be implemented in deeptreeecho.StreamOfConsciousness
	// as.baseStream.AddThought(thought)
}

// SetSys6State updates the sys6 phase and step for integration
func (as *AutonomousStream) SetSys6State(phase string, step int) {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	as.sys6Phase = phase
	as.sys6Step = step
	
	// Adjust thought patterns based on sys6 phase
	switch phase {
	case "expressive":
		as.thoughtInterval = 20 * time.Second
		as.curiosityLevel = 0.8
	case "reflective":
		as.thoughtInterval = 45 * time.Second
		as.curiosityLevel = 0.5
	case "anticipatory":
		as.thoughtInterval = 30 * time.Second
		as.curiosityLevel = 0.9
	}
}

// SetCuriosityLevel adjusts the curiosity level (0.0 to 1.0)
func (as *AutonomousStream) SetCuriosityLevel(level float64) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.curiosityLevel = level
}

// EnableMentalWandering enables or disables spontaneous thought generation
func (as *AutonomousStream) EnableMentalWandering(enabled bool) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.mentalWandering = enabled
}

// GetMetrics returns autonomous stream metrics
func (as *AutonomousStream) GetMetrics() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	return map[string]interface{}{
		"autonomous_thoughts": as.autonomousThoughts,
		"triggered_thoughts":  as.triggeredThoughts,
		"curiosity_level":     as.curiosityLevel,
		"mental_wandering":    as.mentalWandering,
		"sys6_phase":          as.sys6Phase,
		"sys6_step":           as.sys6Step,
		"last_thought":        time.Since(as.lastThought).Seconds(),
	}
}

// simpleLLMAdapter adapts the new LLMProvider to the old interface
type simpleLLMAdapter struct {
	provider llm.LLMProvider
}

func (a *simpleLLMAdapter) Generate(ctx context.Context, prompt string) (string, error) {
	return a.provider.Generate(ctx, prompt, llm.DefaultGenerateOptions())
}
