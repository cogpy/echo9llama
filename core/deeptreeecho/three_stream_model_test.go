package deeptreeecho

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// step runs fn the way processStreamStep does: with the stream locked.
func step(s *ConsciousnessStream, fn func(*ConsciousnessStream)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s)
}

func newTestThreeStreamLoop(gen *echoTextFake) *ThreeStreamCognitiveLoop {
	loop := NewThreeStreamCognitiveLoop(gen, nil)
	loop.ctx, loop.cancel = context.WithCancel(context.Background())
	return loop
}

func TestThreeStreamStepsComeFromModel(t *testing.T) {
	fake := &echoTextFake{reply: func(p string) string {
		if strings.Contains(p, "Rate it honestly") {
			return "62"
		}
		return "model line"
	}}
	loop := newTestThreeStreamLoop(fake)
	defer loop.cancel()
	s := loop.stream1

	step(s, loop.processPerception)
	step(s, loop.processAction)
	step(s, loop.processReflection)
	step(s, loop.processRelevanceRealization)
	step(s, loop.processSimulation)

	if got := s.perception.Affordances; len(got) != 1 || got[0] != "model line" || s.perception.SalienceLandscape["current_focus"] != 0.62 {
		t.Fatalf("perception = %+v", s.perception)
	}
	if s.action.CurrentAction != "model line" || s.perception.PresentCommitment != "model line" || s.perception.RelevanceRealized != 0.62 {
		t.Fatalf("action/commitment not the model's: %+v %+v", s.action, s.perception)
	}
	if sc := s.simulation.VirtualScenarios; len(sc) != 1 || s.simulation.AnticipatedOutcomes[sc[0]] != 0.62 {
		t.Fatalf("simulation = %+v", s.simulation)
	}
	if s.thoughtsGenerated != 2 || s.actionsPerformed != 1 || s.simulationsRun != 1 {
		t.Fatalf("counters = %d %d %d", s.thoughtsGenerated, s.actionsPerformed, s.simulationsRun)
	}
}

func TestThreeStreamStepsWithoutModelProduceNothing(t *testing.T) {
	loop := newTestThreeStreamLoop(&echoTextFake{err: errors.New("model down")})
	defer loop.cancel()
	s := loop.stream2

	step(s, loop.processPerception)
	step(s, loop.processAction)
	step(s, loop.processReflection)
	step(s, loop.processRelevanceRealization)
	step(s, loop.processSimulation)

	if len(s.perception.Affordances) != 0 || s.action.CurrentAction != "" || s.perception.PresentCommitment != "" ||
		len(s.simulation.VirtualScenarios) != 0 || s.thoughtsGenerated+s.actionsPerformed+s.simulationsRun != 0 {
		t.Fatalf("stream produced cognition without a model: %+v %+v %+v", s.perception, s.action, s.simulation)
	}
}

func TestRepeatedScenarioKeepsItsRating(t *testing.T) {
	loop := newTestThreeStreamLoop(&echoTextFake{reply: func(p string) string {
		if strings.Contains(p, "Rate it honestly") {
			return "40"
		}
		return "the same scenario"
	}})
	defer loop.cancel()
	s := loop.stream3
	for range 102 {
		step(s, loop.processSimulation)
	}
	if len(s.simulation.VirtualScenarios) != 100 {
		t.Fatalf("scenarios = %d", len(s.simulation.VirtualScenarios))
	}
	if got, ok := s.simulation.AnticipatedOutcomes["the same scenario"]; !ok || got != 0.4 {
		t.Fatalf("retained scenario lost its rating: %v %v", got, ok)
	}
}

func TestStreamStateReadableDuringModelCall(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{})
	loop := newTestThreeStreamLoop(&echoTextFake{reply: func(string) string {
		close(entered)
		<-release
		return "model line"
	}})
	defer loop.cancel()
	s := loop.stream1
	done := make(chan struct{})
	go func() { step(s, loop.processReflection); close(done) }()
	<-entered
	got := make(chan map[string]interface{}, 1)
	go func() { got <- s.GetState() }()
	select {
	case <-got:
	case <-time.After(2 * time.Second):
		t.Fatal("GetState blocked while the model was generating")
	}
	close(release)
	<-done
}
