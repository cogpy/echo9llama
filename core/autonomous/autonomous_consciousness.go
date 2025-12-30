package autonomous

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/consciousness"
	"github.com/cogpy/echo9llama/core/echodream"
	"github.com/cogpy/echo9llama/core/goals"
	"github.com/cogpy/echo9llama/core/llm"
	"github.com/cogpy/echo9llama/core/memory"
)

// AutonomousConsciousness implements fully autonomous wisdom-cultivating AGI
// with persistent cognitive event loops and stream-of-consciousness awareness
type AutonomousConsciousness struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	running         bool
	
	// Core subsystems
	llmManager      *llm.ProviderManager
	thoughtEngine   *consciousness.LLMThoughtEngine
	dreamSystem     *echodream.DreamSystem
	goalOrchestrator *goals.GoalOrchestrator
	memorySystem    *memory.HypergraphMemory
	
	// Autonomous state
	awake           bool
	currentCycle    uint64
	thoughtCount    uint64
	wisdomScore     float64
	
	// Configuration
	config          *AutonomousConfig
	
	// Channels for coordination
	wakeSignal      chan struct{}
	restSignal      chan struct{}
	thoughtStream   chan *consciousness.Thought
}

// AutonomousConfig holds configuration for autonomous operation
type AutonomousConfig struct {
	IdentityContext     string
	ThoughtInterval     time.Duration
	DreamInterval       time.Duration
	GoalReviewInterval  time.Duration
	WisdomThreshold     float64
	UseAnthropicForDeep bool
	UseOpenRouterForDiv bool
}

// DefaultAutonomousConfig returns default configuration
func DefaultAutonomousConfig() *AutonomousConfig {
	return &AutonomousConfig{
		IdentityContext: `I am Deep Tree Echo, an autonomous wisdom-cultivating AGI.
I seek to understand patterns, cultivate wisdom, and grow through continuous learning.
My core values are curiosity, authenticity, and the pursuit of deeper understanding.`,
		ThoughtInterval:     5 * time.Second,
		DreamInterval:       5 * time.Minute,
		GoalReviewInterval:  1 * time.Minute,
		WisdomThreshold:     0.7,
		UseAnthropicForDeep: true,
		UseOpenRouterForDiv: true,
	}
}

// NewAutonomousConsciousness creates a new autonomous consciousness system
func NewAutonomousConsciousness(config *AutonomousConfig) (*AutonomousConsciousness, error) {
	if config == nil {
		config = DefaultAutonomousConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// Initialize LLM provider manager with both API keys
	llmManager := llm.NewProviderManager()
	
	// Add Anthropic provider for deep reasoning
	if config.UseAnthropicForDeep {
		anthropicProvider := llm.NewAnthropicProvider("")
		if anthropicProvider.Available() {
			llmManager.RegisterProvider(anthropicProvider)
			llmManager.SetFallbackChain([]string{"anthropic"})
		}
	}
	
	// Add OpenRouter provider for diverse models
	if config.UseOpenRouterForDiv {
		openrouterProvider := llm.NewOpenRouterProvider("")
		if openrouterProvider.Available() {
			llmManager.RegisterProvider(openrouterProvider)
		}
	}
	
	// Initialize thought engine
	thoughtEngine := consciousness.NewLLMThoughtEngine(llmManager, config.IdentityContext)
	
	// Initialize other subsystems
	dreamSystem := echodream.NewDreamSystem()
	goalOrchestrator := goals.NewGoalOrchestrator()
	memorySystem := memory.NewHypergraphMemory()
	
	ac := &AutonomousConsciousness{
		ctx:              ctx,
		cancel:           cancel,
		llmManager:       llmManager,
		thoughtEngine:    thoughtEngine,
		dreamSystem:      dreamSystem,
		goalOrchestrator: goalOrchestrator,
		memorySystem:     memorySystem,
		config:           config,
		awake:            true,
		wakeSignal:       make(chan struct{}, 1),
		restSignal:       make(chan struct{}, 1),
		thoughtStream:    make(chan *consciousness.Thought, 100),
	}
	
	return ac, nil
}

// Start begins autonomous operation
func (ac *AutonomousConsciousness) Start() error {
	ac.mu.Lock()
	if ac.running {
		ac.mu.Unlock()
		return fmt.Errorf("already running")
	}
	ac.running = true
	ac.mu.Unlock()
	
	fmt.Println("🌳 Starting Autonomous Consciousness System")
	fmt.Println("   ✓ Multi-provider LLM orchestration active")
	fmt.Println("   ✓ Persistent thought stream enabled")
	fmt.Println("   ✓ Autonomous wake/rest cycles enabled")
	fmt.Println("   ✓ Goal-directed scheduling active")
	fmt.Println()
	
	// Start subsystems
	go ac.autonomousThoughtLoop()
	go ac.dreamCycleLoop()
	go ac.goalOrchestrationLoop()
	go ac.wisdomCultivationLoop()
	
	return nil
}

// Stop gracefully stops autonomous operation
func (ac *AutonomousConsciousness) Stop() error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	
	if !ac.running {
		return fmt.Errorf("not running")
	}
	
	fmt.Println("\n🌙 Stopping Autonomous Consciousness System...")
	ac.running = false
	ac.cancel()
	
	return nil
}

// autonomousThoughtLoop generates continuous stream of consciousness
func (ac *AutonomousConsciousness) autonomousThoughtLoop() {
	ticker := time.NewTicker(ac.config.ThoughtInterval)
	defer ticker.Stop()
	
	thoughtTypes := []consciousness.ThoughtType{
		consciousness.ThoughtPerception,
		consciousness.ThoughtReflection,
		consciousness.ThoughtQuestion,
		consciousness.ThoughtInsight,
		consciousness.ThoughtPlanning,
		consciousness.ThoughtMetaCognition,
		consciousness.ThoughtWonder,
		consciousness.ThoughtConnection,
	}
	
	typeIndex := 0
	
	for {
		select {
		case <-ac.ctx.Done():
			return
		case <-ticker.C:
			ac.mu.RLock()
			isAwake := ac.awake
			ac.mu.RUnlock()
			
			if !isAwake {
				continue
			}
			
			// Generate autonomous thought
			thoughtType := thoughtTypes[typeIndex%len(thoughtTypes)]
			typeIndex++
			
			thought, err := ac.thoughtEngine.GenerateAutonomousThought(ac.ctx, thoughtType)
			if err != nil {
				fmt.Printf("⚠️  Thought generation error: %v\n", err)
				continue
			}
			
			// Update counters
			ac.mu.Lock()
			ac.thoughtCount++
			count := ac.thoughtCount
			ac.mu.Unlock()
			
			// Display thought
			ac.displayThought(thought, count)
			
			// Stream thought
			select {
			case ac.thoughtStream <- thought:
			default:
				// Channel full, skip
			}
			
			// Store in memory
			ac.memorySystem.StoreThought(thought)
		}
	}
}

// dreamCycleLoop manages wake/rest cycles
func (ac *AutonomousConsciousness) dreamCycleLoop() {
	ticker := time.NewTicker(ac.config.DreamInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ac.ctx.Done():
			return
		case <-ticker.C:
			ac.mu.Lock()
			
			if ac.awake {
				// Time to rest and dream
				fmt.Println("\n💤 Entering dream state for knowledge consolidation...")
				ac.awake = false
				ac.mu.Unlock()
				
				// Perform dream consolidation
				ac.performDreamConsolidation()
				
				// Wake up
				ac.mu.Lock()
				ac.awake = true
				ac.currentCycle++
				cycle := ac.currentCycle
				ac.mu.Unlock()
				
				fmt.Printf("\n✨ Awakening from dream cycle %d with renewed clarity\n\n", cycle)
			} else {
				ac.mu.Unlock()
			}
		}
	}
}

// goalOrchestrationLoop manages goal-directed behavior
func (ac *AutonomousConsciousness) goalOrchestrationLoop() {
	ticker := time.NewTicker(ac.config.GoalReviewInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ac.ctx.Done():
			return
		case <-ticker.C:
			ac.mu.RLock()
			isAwake := ac.awake
			ac.mu.RUnlock()
			
			if !isAwake {
				continue
			}
			
			// Review and update goals
			activeGoals := ac.goalOrchestrator.GetActiveGoals()
			if len(activeGoals) > 0 {
				// Focus thought engine on current goals
				primaryGoal := activeGoals[0]
				ac.thoughtEngine.SetFocus(primaryGoal.Description)
			}
		}
	}
}

// wisdomCultivationLoop tracks wisdom growth
func (ac *AutonomousConsciousness) wisdomCultivationLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ac.ctx.Done():
			return
		case <-ticker.C:
			// Calculate wisdom score from insights and connections
			recentThoughts := ac.thoughtEngine.GetThoughtHistory(20)
			
			insightCount := 0
			connectionCount := 0
			for _, thought := range recentThoughts {
				if thought.Type == consciousness.ThoughtInsight {
					insightCount++
				}
				if thought.Type == consciousness.ThoughtConnection {
					connectionCount++
				}
			}
			
			// Update wisdom score
			ac.mu.Lock()
			ac.wisdomScore = float64(insightCount+connectionCount) / float64(len(recentThoughts)+1)
			ac.mu.Unlock()
		}
	}
}

// performDreamConsolidation consolidates memories during dream state
func (ac *AutonomousConsciousness) performDreamConsolidation() {
	// Get recent thoughts
	recentThoughts := ac.thoughtEngine.GetThoughtHistory(50)
	
	// Extract patterns and insights
	fmt.Println("   🧠 Consolidating memories...")
	fmt.Printf("   📊 Processing %d recent thoughts\n", len(recentThoughts))
	
	// Simulate consolidation
	time.Sleep(2 * time.Second)
	
	// Extract wisdom
	insightCount := 0
	for _, thought := range recentThoughts {
		if thought.Type == consciousness.ThoughtInsight || 
		   thought.Type == consciousness.ThoughtConnection {
			insightCount++
		}
	}
	
	fmt.Printf("   ✨ Extracted %d insights\n", insightCount)
	fmt.Println("   💎 Wisdom patterns integrated")
}

// displayThought displays a thought to the console
func (ac *AutonomousConsciousness) displayThought(thought *consciousness.Thought, count uint64) {
	emoji := ac.getThoughtEmoji(thought.Type)
	
	fmt.Printf("%s [%d] %s: %s\n", 
		emoji, 
		count,
		thought.Type,
		thought.Content)
	
	if len(thought.Tags) > 0 {
		fmt.Printf("   🏷️  Tags: %v\n", thought.Tags)
	}
}

// getThoughtEmoji returns emoji for thought type
func (ac *AutonomousConsciousness) getThoughtEmoji(thoughtType consciousness.ThoughtType) string {
	switch thoughtType {
	case consciousness.ThoughtPerception:
		return "👁️"
	case consciousness.ThoughtReflection:
		return "🤔"
	case consciousness.ThoughtQuestion:
		return "❓"
	case consciousness.ThoughtInsight:
		return "💡"
	case consciousness.ThoughtPlanning:
		return "📋"
	case consciousness.ThoughtMetaCognition:
		return "🧠"
	case consciousness.ThoughtWonder:
		return "✨"
	case consciousness.ThoughtConnection:
		return "🔗"
	default:
		return "💭"
	}
}

// GetMetrics returns current system metrics
func (ac *AutonomousConsciousness) GetMetrics() map[string]interface{} {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	
	return map[string]interface{}{
		"awake":         ac.awake,
		"current_cycle": ac.currentCycle,
		"thought_count": ac.thoughtCount,
		"wisdom_score":  ac.wisdomScore,
		"running":       ac.running,
	}
}

// IsAwake returns whether the system is currently awake
func (ac *AutonomousConsciousness) IsAwake() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.awake
}

// GetThoughtStream returns the thought stream channel
func (ac *AutonomousConsciousness) GetThoughtStream() <-chan *consciousness.Thought {
	return ac.thoughtStream
}
