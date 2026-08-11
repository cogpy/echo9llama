package echodream

import (
	"context"
	"testing"
	"time"
)

// TestDreamProcessorFullCycle exercises the real experience-driven dream
// pipeline: ingest → consolidate → extract patterns → synthesize wisdom.
func TestDreamProcessorFullCycle(t *testing.T) {
	dp := NewDreamProcessor(context.Background())

	// Ingest a set of waking experiences with recurring themes
	experiences := []struct {
		content    string
		importance float64
		tags       []string
	}{
		{"Reflected on the nature of recursive self-improvement", 0.9, []string{"recursion", "wisdom"}},
		{"Practiced pattern recognition on tree structures", 0.7, []string{"recursion", "trees"}},
		{"Discussed consciousness with a curious human", 0.8, []string{"consciousness", "wisdom"}},
		{"Noticed that rest improves insight quality", 0.85, []string{"wisdom", "rest"}},
		{"Explored echo state networks and reservoirs", 0.75, []string{"recursion", "reservoirs"}},
		{"Contemplated the balance of order and chaos", 0.8, []string{"wisdom", "consciousness"}},
	}

	for _, exp := range experiences {
		id := dp.IngestExperience(exp.content, exp.importance, exp.tags)
		if id == "" {
			t.Fatalf("IngestExperience returned empty ID")
		}
	}

	// Phase 1: Consolidation should produce knowledge grouped by domain
	dp.ConsolidateMemories()
	dp.mu.RLock()
	knowledgeCount := len(dp.consolidatedKnowledge)
	dp.mu.RUnlock()
	if knowledgeCount == 0 {
		t.Fatal("expected consolidated knowledge, got none")
	}

	// Phase 2: Pattern extraction should find recurring tags (recursion x3, wisdom x4)
	dp.ExtractPatterns()
	dp.mu.RLock()
	patternCount := len(dp.extractedPatterns)
	dp.mu.RUnlock()
	if patternCount == 0 {
		t.Fatal("expected extracted patterns from recurring tags, got none")
	}

	foundConceptual := false
	foundRelational := false
	dp.mu.RLock()
	for _, p := range dp.extractedPatterns {
		if p.Type == "conceptual" {
			foundConceptual = true
		}
		if p.Type == "relational" {
			foundRelational = true
		}
		if p.Strength <= 0 || p.Strength > 1.0 {
			t.Errorf("pattern %s has invalid strength %f", p.ID, p.Strength)
		}
	}
	dp.mu.RUnlock()
	if !foundConceptual {
		t.Error("expected at least one conceptual pattern")
	}
	if !foundRelational {
		t.Error("expected at least one relational pattern from tag co-occurrence")
	}

	// Phase 3: Wisdom synthesis from strongest patterns
	dp.SynthesizeWisdom()
	dp.mu.RLock()
	wisdomCount := len(dp.wisdomInsights)
	pendingAfter := len(dp.pendingExperiences)
	cycles := dp.totalDreamCycles
	dp.mu.RUnlock()

	if wisdomCount == 0 {
		t.Fatal("expected synthesized wisdom insights, got none")
	}
	if pendingAfter != 0 {
		t.Errorf("expected pending experiences cleared after cycle, got %d", pendingAfter)
	}
	if cycles != 1 {
		t.Errorf("expected 1 completed dream cycle, got %d", cycles)
	}

	// A later empty cycle must not re-synthesize the retained historical
	// patterns under new IDs and falsely increase the wisdom counter.
	dp.SynthesizeWisdom()
	dp.mu.RLock()
	wisdomAfterEmptyCycle := len(dp.wisdomInsights)
	wisdomMetricAfterEmptyCycle := dp.wisdomSynthesized
	dp.mu.RUnlock()
	if wisdomAfterEmptyCycle != wisdomCount || wisdomMetricAfterEmptyCycle != uint64(wisdomCount) {
		t.Fatalf("historical patterns were re-synthesized: count=%d metric=%d initial=%d", wisdomAfterEmptyCycle, wisdomMetricAfterEmptyCycle, wisdomCount)
	}
}

// TestDreamProcessorEmptyCycle verifies graceful no-op behavior without experiences
func TestDreamProcessorEmptyCycle(t *testing.T) {
	dp := NewDreamProcessor(context.Background())

	dp.ConsolidateMemories()
	dp.ExtractPatterns()
	dp.SynthesizeWisdom()

	dp.mu.RLock()
	defer dp.mu.RUnlock()
	if len(dp.consolidatedKnowledge) != 0 {
		t.Error("expected no knowledge from empty cycle")
	}
	if len(dp.wisdomInsights) != 0 {
		t.Error("expected no wisdom from empty cycle")
	}
}

// TestIngestExperienceBounded verifies the pending buffer stays bounded
func TestIngestExperienceBounded(t *testing.T) {
	dp := NewDreamProcessor(context.Background())
	for i := 0; i < 1200; i++ {
		dp.IngestExperience("experience", 0.5, []string{"tag"})
	}
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	if len(dp.pendingExperiences) > 1000 {
		t.Errorf("pending buffer exceeded bound: %d", len(dp.pendingExperiences))
	}
}

// TestStateMachineIngestDelegation verifies the state machine exposes ingestion
func TestStateMachineIngestDelegation(t *testing.T) {
	sm := NewSleepWakeStateMachine()
	id := sm.IngestExperience("test experience", 0.6, []string{"testing"})
	if id == "" {
		t.Fatal("state machine ingestion returned empty ID")
	}
	sm.dreamProcessor.mu.RLock()
	defer sm.dreamProcessor.mu.RUnlock()
	if len(sm.dreamProcessor.pendingExperiences) != 1 {
		t.Errorf("expected 1 pending experience, got %d", len(sm.dreamProcessor.pendingExperiences))
	}
}

func TestStateMachineConfiguredCyclePublishesTruthfulMetrics(t *testing.T) {
	sm := NewSleepWakeStateMachine()
	defer sm.Shutdown()
	sm.ConfigurePhaseDurations(time.Millisecond, time.Millisecond, time.Millisecond)

	sm.IngestExperience("Repeated reflection exposed a feedback pattern", 0.9, []string{"reflection", "feedback"})
	sm.IngestExperience("A second reflection confirmed the feedback pattern", 0.85, []string{"reflection", "feedback"})

	metrics := sm.GetMetrics()
	if pending, ok := metrics["pending_experiences"].(int); !ok || pending != 2 {
		t.Fatalf("expected two pending experiences before sleep, got %#v", metrics["pending_experiences"])
	}
	if err := sm.EnterSleep(); err != nil {
		t.Fatalf("EnterSleep failed: %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		metrics = sm.GetMetrics()
		if metrics["wisdom_synthesized"].(uint64) > 0 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	metrics = sm.GetMetrics()
	if metrics["wisdom_synthesized"].(uint64) == 0 {
		t.Fatalf("configured dream cycle produced no wisdom: %#v", metrics)
	}
	if metrics["pending_experiences"].(int) != 0 {
		t.Fatalf("configured dream cycle left pending experiences: %#v", metrics)
	}
	if err := sm.WakeUp(); err != nil {
		t.Fatalf("WakeUp failed: %v", err)
	}
}

func TestStateMachineAccessorsReturnDefensiveCopies(t *testing.T) {
	sm := NewSleepWakeStateMachine()
	defer sm.Shutdown()
	dp := sm.dreamProcessor

	dp.mu.Lock()
	dp.wisdomInsights = []SynthesizedWisdom{{ID: "w1", RelatedTo: []string{"original"}}}
	dp.extractedPatterns = []Pattern{{ID: "p1", Examples: []string{"original"}}}
	dp.consolidatedKnowledge = []Knowledge{{ID: "k1", Sources: []string{"original"}}}
	dp.mu.Unlock()

	wisdom := sm.GetWisdomInsights()
	patterns := sm.GetExtractedPatterns()
	knowledge := sm.GetConsolidatedKnowledge()
	wisdom[0].RelatedTo[0] = "mutated"
	patterns[0].Examples[0] = "mutated"
	knowledge[0].Sources[0] = "mutated"

	if got := sm.GetWisdomInsights()[0].RelatedTo[0]; got != "original" {
		t.Fatalf("wisdom snapshot leaked mutable storage: %q", got)
	}
	if got := sm.GetExtractedPatterns()[0].Examples[0]; got != "original" {
		t.Fatalf("pattern snapshot leaked mutable storage: %q", got)
	}
	if got := sm.GetConsolidatedKnowledge()[0].Sources[0]; got != "original" {
		t.Fatalf("knowledge snapshot leaked mutable storage: %q", got)
	}
}
