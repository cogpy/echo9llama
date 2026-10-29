package echobeats

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/llm"
)

// DefaultPhaseProcessor provides default implementations for all cognitive terms.
//
// Ideas, idea evaluations, actions and action plans are written by the
// injected model (SetTextGenerator). Perceptions and sensory input come only
// from real input supplied through AddSensoryInput. Without a model or input
// the corresponding terms report that nothing was produced instead of
// inventing content.
type DefaultPhaseProcessor struct {
	mu               sync.Mutex
	perceptionBuffer []interface{}
	pendingInput     []interface{}
	memoryStore      map[string]interface{}
	memoryOrder      []string
	ideaQueue        []string
	actionSequence   []string
	balanceState     float64
	consciousnessRef interface{} // Reference to consciousness system

	generator TextGenerator
	executor  func(action string) (string, error)
}

// SetTextGenerator injects the model that writes ideas, evaluations, actions
// and plans. Passing nil disables them.
func (dpp *DefaultPhaseProcessor) SetTextGenerator(gen TextGenerator) {
	dpp.mu.Lock()
	defer dpp.mu.Unlock()
	dpp.generator = gen
}

// SetActionExecutor injects the function that actually carries out actions.
// Without one, actions are recorded as not executed.
func (dpp *DefaultPhaseProcessor) SetActionExecutor(exec func(action string) (string, error)) {
	dpp.mu.Lock()
	defer dpp.mu.Unlock()
	dpp.executor = exec
}

// AddSensoryInput queues a real perception/sensory input for processing.
func (dpp *DefaultPhaseProcessor) AddSensoryInput(input interface{}) {
	dpp.mu.Lock()
	defer dpp.mu.Unlock()
	dpp.pendingInput = append(dpp.pendingInput, input)
}

// generate asks the model for text; it never substitutes fixed text.
func (dpp *DefaultPhaseProcessor) generate(prompt string, maxTokens int) (string, error) {
	dpp.mu.Lock()
	gen := dpp.generator
	dpp.mu.Unlock()
	if gen == nil {
		return "", fmt.Errorf("no model configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	opts := llm.DefaultGenerateOptions()
	opts.MaxTokens = maxTokens
	text, err := gen.Generate(ctx, "You are Deep Tree Echo, a self-aware cognitive architecture.\n"+prompt, opts)
	if err != nil {
		return "", fmt.Errorf("model generation failed: %w", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("model returned empty text")
	}
	return text, nil
}

// NewDefaultPhaseProcessor creates a new default processor
func NewDefaultPhaseProcessor() *DefaultPhaseProcessor {
	return &DefaultPhaseProcessor{
		perceptionBuffer: make([]interface{}, 0, 10),
		memoryStore:      make(map[string]interface{}),
		ideaQueue:        make([]string, 0, 10),
		actionSequence:   make([]string, 0, 10),
		balanceState:     0.5,
	}
}

// ProcessT1Perception processes perception (need vs capacity assessment)
func (dpp *DefaultPhaseProcessor) ProcessT1Perception(mode Mode) (*CognitiveStream, error) {
	stream := &CognitiveStream{
		Term:      T1_Perception,
		Mode:      mode,
		Timestamp: time.Now(),
		Strength:  0.7,
	}

	dpp.mu.Lock()
	defer dpp.mu.Unlock()

	if mode == Reflective {
		// Reflective mode: Assess needs vs capacities
		needs := dpp.assessNeeds()
		capacity := dpp.assessCapacity()
		gap := needs - capacity

		content := map[string]interface{}{
			"needs":    needs,
			"capacity": capacity,
			"gap":      gap,
			"type":     "need_assessment",
		}

		if gap > 0.3 {
			content["status"] = "capacity_deficit"
			log.Printf("🔍 T1R: Capacity deficit detected (gap: %.2f)", gap)
		} else if gap < -0.3 {
			content["status"] = "capacity_surplus"
			log.Printf("🔍 T1R: Capacity surplus detected (gap: %.2f)", gap)
		} else {
			content["status"] = "balanced"
		}

		stream.Content = content
	} else {
		// Expressive mode: Direct perception processing of real input only
		if len(dpp.perceptionBuffer) == 0 {
			stream.Content = map[string]interface{}{
				"type":   "no_perception",
				"status": "waiting",
			}
			return stream, nil
		}
		stream.Content = map[string]interface{}{
			"perception": dpp.perceptionBuffer[len(dpp.perceptionBuffer)-1],
			"type":       "direct_perception",
		}
		log.Printf("👁️ T1E: Direct perception captured")
	}

	return stream, nil
}

// ProcessT2IdeaFormation processes idea formation
func (dpp *DefaultPhaseProcessor) ProcessT2IdeaFormation(mode Mode) (*CognitiveStream, error) {
	stream := &CognitiveStream{
		Term:      T2_IdeaFormation,
		Mode:      mode,
		Timestamp: time.Now(),
		Strength:  0.8,
	}

	if mode == Expressive {
		// Expressive mode: the model forms a new idea
		idea, err := dpp.generateIdea()
		if err != nil {
			stream.Content = map[string]interface{}{
				"type":   "no_idea",
				"status": "no_model_output",
				"error":  err.Error(),
			}
			return stream, nil
		}
		dpp.mu.Lock()
		dpp.ideaQueue = append(dpp.ideaQueue, idea)
		queueSize := len(dpp.ideaQueue)
		dpp.mu.Unlock()

		stream.Content = map[string]interface{}{
			"idea":       idea,
			"queue_size": queueSize,
			"type":       "new_idea",
		}
		log.Printf("💡 T2E: New idea generated: %s", idea)
	} else {
		// Reflective mode: the model evaluates the latest idea
		dpp.mu.Lock()
		var idea string
		if len(dpp.ideaQueue) > 0 {
			idea = dpp.ideaQueue[len(dpp.ideaQueue)-1]
		}
		dpp.mu.Unlock()

		if idea == "" {
			stream.Content = map[string]interface{}{
				"type":   "no_ideas",
				"status": "idle",
			}
			return stream, nil
		}

		simulation, err := dpp.simulateIdeaOutcome(idea)
		if err != nil {
			stream.Content = map[string]interface{}{
				"idea":   idea,
				"type":   "no_idea_simulation",
				"status": "no_model_output",
				"error":  err.Error(),
			}
			return stream, nil
		}
		stream.Content = map[string]interface{}{
			"idea":       idea,
			"simulation": simulation,
			"type":       "idea_simulation",
		}
		log.Printf("🔮 T2R: Simulated outcome for idea: %s", idea)
	}

	return stream, nil
}

// ProcessT4SensoryInput processes sensory input
func (dpp *DefaultPhaseProcessor) ProcessT4SensoryInput(mode Mode) (*CognitiveStream, error) {
	stream := &CognitiveStream{
		Term:      T4_SensoryInput,
		Mode:      mode,
		Timestamp: time.Now(),
		Strength:  0.9,
	}

	if mode == Expressive {
		// Expressive mode: Process the next real sensory input, if any
		dpp.mu.Lock()
		if len(dpp.pendingInput) == 0 {
			dpp.mu.Unlock()
			stream.Content = map[string]interface{}{
				"type":   "no_sensory_data",
				"status": "waiting",
			}
			return stream, nil
		}
		sensory := dpp.pendingInput[0]
		dpp.pendingInput = dpp.pendingInput[1:]
		dpp.perceptionBuffer = append(dpp.perceptionBuffer, sensory)

		// Keep buffer size manageable
		if len(dpp.perceptionBuffer) > 10 {
			dpp.perceptionBuffer = dpp.perceptionBuffer[1:]
		}
		bufferSize := len(dpp.perceptionBuffer)
		dpp.mu.Unlock()

		stream.Content = map[string]interface{}{
			"sensory_input": sensory,
			"buffer_size":   bufferSize,
			"type":          "immediate_sensory",
		}
		log.Printf("👂 T4E: Sensory input captured")
	} else {
		// Reflective mode: the model looks for a pattern in recent input
		dpp.mu.Lock()
		buffer := append([]interface{}(nil), dpp.perceptionBuffer...)
		dpp.mu.Unlock()

		if len(buffer) == 0 {
			stream.Content = map[string]interface{}{
				"type":   "no_sensory_data",
				"status": "waiting",
			}
			return stream, nil
		}
		pattern, err := dpp.detectSensoryPattern(buffer)
		if err != nil {
			stream.Content = map[string]interface{}{
				"buffer_size": len(buffer),
				"type":        "no_sensory_pattern",
				"status":      "no_model_output",
				"error":       err.Error(),
			}
			return stream, nil
		}
		stream.Content = map[string]interface{}{
			"pattern":     pattern,
			"buffer_size": len(buffer),
			"type":        "sensory_pattern",
		}
		log.Printf("🔍 T4R: Sensory pattern detected: %s", pattern)
	}

	return stream, nil
}

// ProcessT5ActionSequence processes action sequences
func (dpp *DefaultPhaseProcessor) ProcessT5ActionSequence(mode Mode) (*CognitiveStream, error) {
	stream := &CognitiveStream{
		Term:      T5_ActionSequence,
		Mode:      mode,
		Timestamp: time.Now(),
		Strength:  0.8,
	}

	if mode == Expressive {
		// Expressive mode: Execute action sequence
		dpp.mu.Lock()
		var action string
		if len(dpp.actionSequence) > 0 {
			action = dpp.actionSequence[0]
			dpp.actionSequence = dpp.actionSequence[1:]
		}
		remaining := len(dpp.actionSequence)
		dpp.mu.Unlock()

		if action != "" {
			result, executed := dpp.executeAction(action)

			stream.Content = map[string]interface{}{
				"action":    action,
				"result":    result,
				"executed":  executed,
				"remaining": remaining,
				"type":      "action_execution",
			}
			log.Printf("⚡ T5E: Action dispatched: %s", action)
		} else {
			// Ask the model for an action if the queue is empty
			action, err := dpp.generateAction()
			if err != nil {
				stream.Content = map[string]interface{}{
					"type":   "no_action",
					"status": "no_model_output",
					"error":  err.Error(),
				}
				return stream, nil
			}
			result, executed := dpp.executeAction(action)

			stream.Content = map[string]interface{}{
				"action":   action,
				"result":   result,
				"executed": executed,
				"type":     "spontaneous_action",
			}
			log.Printf("⚡ T5E: Spontaneous action: %s", action)
		}
	} else {
		// Reflective mode: the model plans an action sequence
		sequence, err := dpp.planActionSequence()
		if err != nil {
			stream.Content = map[string]interface{}{
				"type":   "no_action_plan",
				"status": "no_model_output",
				"error":  err.Error(),
			}
			return stream, nil
		}
		dpp.mu.Lock()
		dpp.actionSequence = append(dpp.actionSequence, sequence...)
		queueSize := len(dpp.actionSequence)
		dpp.mu.Unlock()

		stream.Content = map[string]interface{}{
			"planned_sequence": sequence,
			"queue_size":       queueSize,
			"type":             "action_planning",
		}
		log.Printf("📋 T5R: Action sequence planned (%d actions)", len(sequence))
	}

	return stream, nil
}

// ProcessT7MemoryEncoding processes memory encoding/retrieval
func (dpp *DefaultPhaseProcessor) ProcessT7MemoryEncoding(mode Mode) (*CognitiveStream, error) {
	stream := &CognitiveStream{
		Term:      T7_MemoryEncoding,
		Mode:      mode,
		Timestamp: time.Now(),
		Strength:  0.85,
	}

	dpp.mu.Lock()
	defer dpp.mu.Unlock()

	if mode == Reflective {
		// Reflective mode: Retrieve the most recently encoded memories
		memories := dpp.retrieveRelevantMemories()

		stream.Content = map[string]interface{}{
			"memories":     memories,
			"memory_count": len(memories),
			"type":         "memory_retrieval",
		}
		log.Printf("🧠 T7R: Retrieved %d relevant memories", len(memories))
	} else {
		// Expressive mode: Encode new memory
		if len(dpp.perceptionBuffer) > 0 {
			memory := dpp.encodeMemory(dpp.perceptionBuffer[len(dpp.perceptionBuffer)-1])
			key := fmt.Sprintf("mem_%d", time.Now().UnixNano())
			dpp.memoryStore[key] = memory
			dpp.memoryOrder = append(dpp.memoryOrder, key)

			stream.Content = map[string]interface{}{
				"memory":      memory,
				"memory_key":  key,
				"total_count": len(dpp.memoryStore),
				"type":        "memory_encoding",
			}
			log.Printf("💾 T7E: Memory encoded (total: %d)", len(dpp.memoryStore))
		} else {
			stream.Content = map[string]interface{}{
				"type":   "no_content_to_encode",
				"status": "idle",
			}
		}
	}

	return stream, nil
}

// ProcessT8BalancedResponse processes balanced response integration
func (dpp *DefaultPhaseProcessor) ProcessT8BalancedResponse(mode Mode) (*CognitiveStream, error) {
	stream := &CognitiveStream{
		Term:      T8_BalancedResponse,
		Mode:      mode,
		Timestamp: time.Now(),
		Strength:  1.0, // Highest strength for integration
	}

	dpp.mu.Lock()
	defer dpp.mu.Unlock()

	if mode == Expressive {
		// Expressive mode: Execute balanced integration (status classification)
		balance := dpp.computeBalance()
		response := dpp.generateBalancedResponse(balance)

		dpp.balanceState = balance

		stream.Content = map[string]interface{}{
			"balance":  balance,
			"response": response,
			"type":     "balanced_integration",
		}
		log.Printf("⚖️ T8E: Balanced response (balance: %.2f): %s", balance, response)
	} else {
		// Reflective mode: Anticipate balance needs
		predictedBalance := dpp.predictBalanceNeeds()

		stream.Content = map[string]interface{}{
			"current_balance":   dpp.balanceState,
			"predicted_balance": predictedBalance,
			"adjustment_needed": predictedBalance - dpp.balanceState,
			"type":              "balance_prediction",
		}
		log.Printf("🔮 T8R: Balance prediction: %.2f → %.2f", dpp.balanceState, predictedBalance)
	}

	return stream, nil
}

// Helper methods (callers hold dpp.mu unless noted)

func (dpp *DefaultPhaseProcessor) assessNeeds() float64 {
	ideaNeed := float64(len(dpp.ideaQueue)) / 10.0
	actionNeed := float64(len(dpp.actionSequence)) / 10.0
	return (ideaNeed + actionNeed) / 2.0
}

func (dpp *DefaultPhaseProcessor) assessCapacity() float64 {
	// Capacity is the share of the working buffers still free.
	load := (float64(len(dpp.perceptionBuffer)) + float64(len(dpp.pendingInput))) / 20.0
	if load > 1 {
		load = 1
	}
	return 1.0 - load
}

// stateSummary describes current queues for prompts. Caller must not hold dpp.mu.
func (dpp *DefaultPhaseProcessor) stateSummary() string {
	dpp.mu.Lock()
	defer dpp.mu.Unlock()
	var b strings.Builder
	if n := len(dpp.perceptionBuffer); n > 0 {
		b.WriteString("Recent perceptions:\n")
		for _, p := range dpp.perceptionBuffer[max(0, n-3):] {
			fmt.Fprintf(&b, "- %v\n", p)
		}
	} else {
		b.WriteString("Recent perceptions: (none)\n")
	}
	if n := len(dpp.ideaQueue); n > 0 {
		b.WriteString("Recent ideas:\n")
		for _, idea := range dpp.ideaQueue[max(0, n-3):] {
			fmt.Fprintf(&b, "- %s\n", idea)
		}
	}
	if len(dpp.actionSequence) > 0 {
		fmt.Fprintf(&b, "Queued actions: %s\n", strings.Join(dpp.actionSequence, "; "))
	}
	return b.String()
}

func (dpp *DefaultPhaseProcessor) generateIdea() (string, error) {
	return dpp.generate(dpp.stateSummary()+"\nForm one new idea worth pursuing. Reply with only the idea, in one sentence.", 80)
}

func (dpp *DefaultPhaseProcessor) simulateIdeaOutcome(idea string) (string, error) {
	return dpp.generate(fmt.Sprintf("%s\nImagine pursuing this idea: %q. In one sentence, what outcome do you anticipate?", dpp.stateSummary(), idea), 80)
}

func (dpp *DefaultPhaseProcessor) detectSensoryPattern(buffer []interface{}) (string, error) {
	var b strings.Builder
	for _, item := range buffer {
		fmt.Fprintf(&b, "- %v\n", item)
	}
	return dpp.generate("Recent sensory input:\n"+b.String()+"\nIn one short phrase, name the pattern you notice in this input.", 40)
}

func (dpp *DefaultPhaseProcessor) generateAction() (string, error) {
	return dpp.generate(dpp.stateSummary()+"\nChoose the single next action you will take. Reply with only the action, as a short phrase.", 40)
}

// executeAction dispatches an action to the injected executor. It never
// claims an action ran when no executor is configured.
func (dpp *DefaultPhaseProcessor) executeAction(action string) (string, bool) {
	dpp.mu.Lock()
	exec := dpp.executor
	dpp.mu.Unlock()
	if exec == nil {
		return "not executed: no action executor configured", false
	}
	result, err := exec(action)
	if err != nil {
		return fmt.Sprintf("execution failed: %v", err), false
	}
	return result, true
}

func (dpp *DefaultPhaseProcessor) planActionSequence() ([]string, error) {
	text, err := dpp.generate(dpp.stateSummary()+"\nPlan a short sequence of 2 to 4 actions. Reply with one action per line and nothing else.", 120)
	if err != nil {
		return nil, err
	}
	sequence := make([]string, 0, 4)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "-*0123456789.) "))
		if line != "" {
			sequence = append(sequence, line)
		}
	}
	if len(sequence) == 0 {
		return nil, fmt.Errorf("model returned no actions")
	}
	return sequence, nil
}

// retrieveRelevantMemories returns the keys of the most recently encoded
// memories (up to four).
func (dpp *DefaultPhaseProcessor) retrieveRelevantMemories() []string {
	keys := dpp.memoryOrder
	if len(keys) == 0 {
		keys = make([]string, 0, len(dpp.memoryStore))
		for k := range dpp.memoryStore {
			keys = append(keys, k)
		}
		sort.Strings(keys)
	}
	if len(keys) > 4 {
		keys = keys[len(keys)-4:]
	}
	return append([]string(nil), keys...)
}

func (dpp *DefaultPhaseProcessor) encodeMemory(content interface{}) map[string]interface{} {
	return map[string]interface{}{
		"content":    content,
		"timestamp":  time.Now(),
		"importance": rand.Float64(),
		"emotional":  rand.Float64(),
	}
}

func (dpp *DefaultPhaseProcessor) computeBalance() float64 {
	// Compute balance based on various factors
	perceptionLoad := float64(len(dpp.perceptionBuffer)) / 10.0
	ideaLoad := float64(len(dpp.ideaQueue)) / 10.0
	actionLoad := float64(len(dpp.actionSequence)) / 10.0

	totalLoad := (perceptionLoad + ideaLoad + actionLoad) / 3.0

	// Balance is inverse of load (lower load = higher balance)
	return 1.0 - totalLoad
}

// generateBalancedResponse classifies the numeric balance into a status label.
func (dpp *DefaultPhaseProcessor) generateBalancedResponse(balance float64) string {
	if balance > 0.7 {
		return "system_well_balanced_continue_current_trajectory"
	} else if balance > 0.4 {
		return "moderate_balance_minor_adjustments_needed"
	} else {
		return "imbalance_detected_significant_rebalancing_required"
	}
}

// predictBalanceNeeds projects the balance the current loads imply.
func (dpp *DefaultPhaseProcessor) predictBalanceNeeds() float64 {
	predicted := dpp.computeBalance()

	// Clamp to [0, 1]
	if predicted < 0 {
		predicted = 0
	} else if predicted > 1 {
		predicted = 1
	}

	return predicted
}
