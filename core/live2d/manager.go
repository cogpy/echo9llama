package live2d

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// AvatarManager manages the Live2D avatar and its connection to Echo9
type AvatarManager struct {
	mu          sync.RWMutex
	model       *Live2DModel
	mapper      ParameterMapper
	updateChan  chan AvatarState
	subscribers []chan ParameterUpdate
	ctx         context.Context
	cancel      context.CancelFunc
	running     bool
}

// NewAvatarManager creates a new avatar manager
func NewAvatarManager(modelName, modelPath string) *AvatarManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &AvatarManager{
		model:       NewLive2DModel(modelName, modelPath),
		mapper:      NewDefaultParameterMapper(),
		updateChan:  make(chan AvatarState, 100),
		subscribers: make([]chan ParameterUpdate, 0),
		ctx:         ctx,
		cancel:      cancel,
		running:     false,
	}
}

// Start begins processing avatar state updates
func (am *AvatarManager) Start() error {
	am.mu.Lock()
	if am.running {
		am.mu.Unlock()
		return fmt.Errorf("avatar manager already running")
	}
	am.running = true
	am.mu.Unlock()

	// Start update loop
	go am.updateLoop()

	return nil
}

// Stop stops the avatar manager
func (am *AvatarManager) Stop() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if !am.running {
		return fmt.Errorf("avatar manager not running")
	}

	am.cancel()
	am.running = false
	close(am.updateChan)

	// Close all subscriber channels
	for _, ch := range am.subscribers {
		close(ch)
	}
	am.subscribers = nil

	return nil
}

// UpdateEmotionalState updates the avatar's emotional state
func (am *AvatarManager) UpdateEmotionalState(emotional EmotionalState) error {
	am.mu.RLock()
	currentState := am.model.GetCurrentState()
	am.mu.RUnlock()

	// Create new state with updated emotional component
	newState := AvatarState{
		Emotional: emotional,
		Cognitive: currentState.Cognitive,
		Timestamp: time.Now(),
	}

	select {
	case am.updateChan <- newState:
		return nil
	case <-am.ctx.Done():
		return fmt.Errorf("avatar manager stopped")
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("timeout updating emotional state")
	}
}

// UpdateCognitiveState updates the avatar's cognitive state
func (am *AvatarManager) UpdateCognitiveState(cognitive CognitiveState) error {
	am.mu.RLock()
	currentState := am.model.GetCurrentState()
	am.mu.RUnlock()

	// Create new state with updated cognitive component
	newState := AvatarState{
		Emotional: currentState.Emotional,
		Cognitive: cognitive,
		Timestamp: time.Now(),
	}

	select {
	case am.updateChan <- newState:
		return nil
	case <-am.ctx.Done():
		return fmt.Errorf("avatar manager stopped")
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("timeout updating cognitive state")
	}
}

// UpdateFullState updates both emotional and cognitive states
func (am *AvatarManager) UpdateFullState(state AvatarState) error {
	state.Timestamp = time.Now()

	select {
	case am.updateChan <- state:
		return nil
	case <-am.ctx.Done():
		return fmt.Errorf("avatar manager stopped")
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("timeout updating full state")
	}
}

// Subscribe returns a channel that receives parameter updates
func (am *AvatarManager) Subscribe() (<-chan ParameterUpdate, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if !am.running {
		return nil, fmt.Errorf("avatar manager not running")
	}

	ch := make(chan ParameterUpdate, 10)
	am.subscribers = append(am.subscribers, ch)

	return ch, nil
}

// GetCurrentState returns the current avatar state
func (am *AvatarManager) GetCurrentState() AvatarState {
	am.mu.RLock()
	defer am.mu.RUnlock()

	return am.model.GetCurrentState()
}

// GetCurrentParameters returns current parameter values as JSON
func (am *AvatarManager) GetCurrentParameters() ([]byte, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	params := am.model.GetCurrentParameters(am.mapper)
	return json.Marshal(params)
}

// updateLoop continuously processes state updates and publishes parameter changes
func (am *AvatarManager) updateLoop() {
	ticker := time.NewTicker(am.model.UpdateRate)
	defer ticker.Stop()

	var lastState *AvatarState

	for {
		select {
		case <-am.ctx.Done():
			return

		case newState, ok := <-am.updateChan:
			if !ok {
				return
			}

			// Update model state
			if err := am.model.UpdateState(newState); err != nil {
				// Log error but continue
				continue
			}

			lastState = &newState

		case <-ticker.C:
			// Periodic parameter update even without state changes
			// This allows for smooth animations like breathing
			if lastState != nil {
				am.publishParameters()
			}
		}
	}
}

// publishParameters publishes current parameters to all subscribers
func (am *AvatarManager) publishParameters() {
	am.mu.RLock()
	params := am.model.GetCurrentParameters(am.mapper)
	subscribers := am.subscribers
	am.mu.RUnlock()

	update := ParameterUpdate{
		Timestamp:  time.Now(),
		Parameters: params,
	}

	// Send to all subscribers (non-blocking)
	for _, ch := range subscribers {
		select {
		case ch <- update:
			// Sent successfully
		default:
			// Channel full, skip this update
		}
	}
}

// SetEmotionPreset sets the emotional state to a preset
func (am *AvatarManager) SetEmotionPreset(presetName string) error {
	preset, ok := EmotionPresets[presetName]
	if !ok {
		return fmt.Errorf("unknown emotion preset: %s", presetName)
	}

	return am.UpdateEmotionalState(preset)
}

// BlendEmotions blends two emotional states with a weight (0.0 to 1.0)
func BlendEmotions(emotion1, emotion2 EmotionalState, weight float64) EmotionalState {
	weight = clamp(weight, 0.0, 1.0)
	invWeight := 1.0 - weight

	return EmotionalState{
		Valence:    emotion1.Valence*invWeight + emotion2.Valence*weight,
		Arousal:    emotion1.Arousal*invWeight + emotion2.Arousal*weight,
		Dominance:  emotion1.Dominance*invWeight + emotion2.Dominance*weight,
		Curiosity:  emotion1.Curiosity*invWeight + emotion2.Curiosity*weight,
		Confidence: emotion1.Confidence*invWeight + emotion2.Confidence*weight,
	}
}

// AnimateEmotionTransition smoothly transitions between two emotions over a duration
func (am *AvatarManager) AnimateEmotionTransition(from, to EmotionalState, duration time.Duration) error {
	steps := int(duration.Milliseconds() / am.model.UpdateRate.Milliseconds())
	if steps < 1 {
		steps = 1
	}

	go func() {
		for i := 0; i <= steps; i++ {
			weight := float64(i) / float64(steps)
			blended := BlendEmotions(from, to, weight)

			if err := am.UpdateEmotionalState(blended); err != nil {
				return
			}

			time.Sleep(am.model.UpdateRate)
		}
	}()

	return nil
}

// GetModelInfo returns information about the current model
func (am *AvatarManager) GetModelInfo() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()

	return map[string]interface{}{
		"name":        am.model.Name,
		"model_path":  am.model.ModelPath,
		"update_rate": am.model.UpdateRate.String(),
		"running":     am.running,
		"subscribers": len(am.subscribers),
	}
}

// IsRunning returns whether the avatar manager is running
func (am *AvatarManager) IsRunning() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.running
}
