package echodream

import "testing"

func TestDreamExperiencePreservesSourceEvent(t *testing.T) {
	processor := NewDreamProcessor(nil)
	processor.IngestExperienceWithSource(
		"evaluation.recorded:event-42",
		"A verified local artifact matched its declared hash.",
		0.9,
		[]string{"action", "verified"},
	)
	processor.ConsolidateMemories()

	processor.mu.RLock()
	defer processor.mu.RUnlock()
	if len(processor.consolidatedKnowledge) != 1 {
		t.Fatalf("expected one consolidated knowledge item, got %d", len(processor.consolidatedKnowledge))
	}
	sources := processor.consolidatedKnowledge[0].Sources
	if len(sources) != 1 || sources[0] != "evaluation.recorded:event-42" {
		t.Fatalf("source provenance was lost: %#v", sources)
	}
}
