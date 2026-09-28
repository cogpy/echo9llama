package live2d

import (
	"testing"
	"time"
)

func TestEmotionalStateMapping(t *testing.T) {
	mapper := NewDefaultParameterMapper()

	// Test happy emotion
	happy := EmotionalState{
		Valence:    0.8,
		Arousal:    0.6,
		Dominance:  0.6,
		Curiosity:  0.4,
		Confidence: 0.7,
	}

	params := mapper.MapEmotionalState(happy)

	if len(params) == 0 {
		t.Error("Expected parameters from emotional state mapping")
	}

	// Verify smile parameters exist
	foundSmile := false
	for _, param := range params {
		if param.ID == StandardParameterNames.MouthSmile {
			foundSmile = true
			if param.Value <= 0 {
				t.Errorf("Expected positive smile value for happy emotion, got %f", param.Value)
			}
		}
	}

	if !foundSmile {
		t.Error("Expected MouthSmile parameter in result")
	}
}

func TestCognitiveStateMapping(t *testing.T) {
	mapper := NewDefaultParameterMapper()

	cognitive := CognitiveState{
		Awareness:      0.8,
		Attention:      0.7,
		CognitiveLoad:  0.4,
		Coherence:      0.9,
		EnergyLevel:    0.9,
		ProcessingMode: "creative",
	}

	params := mapper.MapCognitiveState(cognitive)

	if len(params) == 0 {
		t.Error("Expected parameters from cognitive state mapping")
	}
}

func TestCombinedStateMapping(t *testing.T) {
	mapper := NewDefaultParameterMapper()

	state := AvatarState{
		Emotional: EmotionPresets["happy"],
		Cognitive: CognitiveState{
			Awareness:      0.7,
			Attention:      0.6,
			CognitiveLoad:  0.3,
			Coherence:      0.8,
			EnergyLevel:    0.8,
			ProcessingMode: "dynamic",
		},
		Timestamp: time.Now(),
	}

	params := mapper.MapCombinedState(state)

	if len(params) == 0 {
		t.Error("Expected parameters from combined state mapping")
	}
}

func TestLive2DModel(t *testing.T) {
	model := NewLive2DModel("TestModel", "/test/model.json")

	if model.Name != "TestModel" {
		t.Errorf("Expected model name 'TestModel', got '%s'", model.Name)
	}

	if model.ModelPath != "/test/model.json" {
		t.Errorf("Expected model path '/test/model.json', got '%s'", model.ModelPath)
	}

	// Test state update
	newState := AvatarState{
		Emotional: EmotionPresets["excited"],
		Cognitive: CognitiveState{
			Awareness:      0.9,
			Attention:      0.8,
			CognitiveLoad:  0.5,
			Coherence:      0.7,
			EnergyLevel:    0.9,
			ProcessingMode: "dynamic",
		},
		Timestamp: time.Now(),
	}

	err := model.UpdateState(newState)
	if err != nil {
		t.Errorf("Failed to update state: %v", err)
	}

	currentState := model.GetCurrentState()
	if currentState.Emotional.Arousal != newState.Emotional.Arousal {
		t.Errorf("State not updated correctly")
	}
}

func TestAvatarManager(t *testing.T) {
	manager := NewAvatarManager("TestAvatar", "/test/model.json")

	// Test start/stop
	err := manager.Start()
	if err != nil {
		t.Errorf("Failed to start avatar manager: %v", err)
	}

	if !manager.IsRunning() {
		t.Error("Expected manager to be running")
	}

	// Test state update
	emotional := EmotionPresets["curious"]
	err = manager.UpdateEmotionalState(emotional)
	if err != nil {
		t.Errorf("Failed to update emotional state: %v", err)
	}

	cognitive := CognitiveState{
		Awareness:      0.8,
		Attention:      0.7,
		CognitiveLoad:  0.4,
		Coherence:      0.9,
		EnergyLevel:    0.8,
		ProcessingMode: "contemplative",
	}
	err = manager.UpdateCognitiveState(cognitive)
	if err != nil {
		t.Errorf("Failed to update cognitive state: %v", err)
	}

	// Wait a bit for updates to process
	time.Sleep(100 * time.Millisecond)

	// Stop manager
	err = manager.Stop()
	if err != nil {
		t.Errorf("Failed to stop avatar manager: %v", err)
	}

	if manager.IsRunning() {
		t.Error("Expected manager to be stopped")
	}
}

func TestEmotionPresets(t *testing.T) {
	presets := []string{"neutral", "happy", "sad", "curious", "confident", "contemplative", "excited"}

	for _, preset := range presets {
		if _, ok := EmotionPresets[preset]; !ok {
			t.Errorf("Expected preset '%s' to exist", preset)
		}
	}
}

func TestBlendEmotions(t *testing.T) {
	emotion1 := EmotionPresets["happy"]
	emotion2 := EmotionPresets["sad"]

	// Test 50/50 blend
	blended := BlendEmotions(emotion1, emotion2, 0.5)

	expectedValence := (emotion1.Valence + emotion2.Valence) / 2.0
	if blended.Valence != expectedValence {
		t.Errorf("Expected blended valence %f, got %f", expectedValence, blended.Valence)
	}

	// Test full weight to emotion2
	blended = BlendEmotions(emotion1, emotion2, 1.0)
	if blended.Valence != emotion2.Valence {
		t.Errorf("Expected emotion2 valence %f, got %f", emotion2.Valence, blended.Valence)
	}

	// Test no weight (all emotion1)
	blended = BlendEmotions(emotion1, emotion2, 0.0)
	if blended.Valence != emotion1.Valence {
		t.Errorf("Expected emotion1 valence %f, got %f", emotion1.Valence, blended.Valence)
	}
}

func TestEchoStateBridge(t *testing.T) {
	manager := NewAvatarManager("TestAvatar", "/test/model.json")
	manager.Start()
	defer manager.Stop()

	bridge := NewEchoStateBridge(manager)

	// Test Echo emotion update
	echoEmotion := map[string]float64{
		"joy":        0.7,
		"curiosity":  0.8,
		"confidence": 0.6,
	}

	err := bridge.UpdateFromEchoEmotion(echoEmotion)
	if err != nil {
		t.Errorf("Failed to update from Echo emotion: %v", err)
	}

	// Test Echo cognitive update
	echoCognitive := map[string]interface{}{
		"awareness":       0.8,
		"attention":       0.7,
		"cognitive_load":  0.4,
		"energy_level":    0.9,
		"processing_mode": "creative",
	}

	err = bridge.UpdateFromEchoCognitive(echoCognitive)
	if err != nil {
		t.Errorf("Failed to update from Echo cognitive: %v", err)
	}

	// Wait for updates to process
	time.Sleep(100 * time.Millisecond)

	// Verify state was updated
	state := manager.GetCurrentState()
	if state.Cognitive.ProcessingMode != "creative" {
		t.Errorf("Expected processing mode 'creative', got '%s'", state.Cognitive.ProcessingMode)
	}
}

func TestParameterSmoothing(t *testing.T) {
	mapper := NewDefaultParameterMapper()
	mapper.SetSmoothingFactor(0.5)

	state1 := AvatarState{
		Emotional: EmotionPresets["happy"],
		Cognitive: CognitiveState{
			Awareness:      0.5,
			Attention:      0.5,
			CognitiveLoad:  0.3,
			Coherence:      0.7,
			EnergyLevel:    0.7,
			ProcessingMode: "contemplative",
		},
		Timestamp: time.Now(),
	}

	// First call - no smoothing (no previous state)
	params1 := mapper.MapCombinedState(state1)

	// Second call - should apply smoothing
	state2 := state1
	state2.Emotional = EmotionPresets["sad"]
	params2 := mapper.MapCombinedState(state2)

	if len(params1) == 0 || len(params2) == 0 {
		t.Error("Expected parameters from both calls")
	}
}

func TestStandardParameterNames(t *testing.T) {
	expectedParams := []string{
		StandardParameterNames.EyeOpenLeft,
		StandardParameterNames.EyeOpenRight,
		StandardParameterNames.MouthSmile,
		StandardParameterNames.EyeBallX,
		StandardParameterNames.AngleX,
		StandardParameterNames.Breathing,
	}

	for _, param := range expectedParams {
		if param == "" {
			t.Error("Expected non-empty parameter name")
		}
	}
}
