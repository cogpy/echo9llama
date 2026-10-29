package opencog

import (
	"context"
	"testing"

	"github.com/EchoCog/echollama/core/llm"
)

func TestAtomSpace(t *testing.T) {
	as := NewAtomSpace()

	// Test atom creation
	atom1, err := as.AddAtom(ConceptNode, "cat", &TruthValue{
		Strength:   0.9,
		Confidence: 0.8,
		Count:      1.0,
	})
	if err != nil {
		t.Fatalf("Failed to add atom: %v", err)
	}

	atom2, err := as.AddAtom(ConceptNode, "animal", nil)
	if err != nil {
		t.Fatalf("Failed to add atom: %v", err)
	}

	// Test link creation
	link, err := as.AddLink(InheritanceLink, []string{atom1.ID, atom2.ID}, &TruthValue{
		Strength:   0.95,
		Confidence: 0.9,
		Count:      1.0,
	})
	if err != nil {
		t.Fatalf("Failed to add link: %v", err)
	}

	// Verify link
	if link.Type != InheritanceLink {
		t.Errorf("Expected InheritanceLink, got %v", link.Type)
	}

	// Test incoming links
	incoming := as.GetIncoming(atom1.ID)
	if len(incoming) != 1 {
		t.Errorf("Expected 1 incoming link, got %d", len(incoming))
	}

	t.Logf("AtomSpace test passed: %d atoms, %d links", as.AtomCount, as.LinkCount)
}

func TestPaunPSystem(t *testing.T) {
	pps := NewPaunPSystem()

	// Add objects to root membrane
	rootMembrane := pps.Membranes[pps.RootMembrane]
	rootMembrane.Objects["energy"] = 15
	rootMembrane.Objects["pattern"] = 5

	// Evolve
	for i := 0; i < 10; i++ {
		pps.Evolve()
	}

	// Check evolution
	if len(pps.Membranes) < 1 {
		t.Errorf("Expected at least 1 membrane")
	}

	t.Logf("P-System: %d membranes", len(pps.Membranes))
}

func TestRicciFlow(t *testing.T) {
	rfe := NewRicciFlowEngine(10)

	// Apply flow
	for i := 0; i < 10; i++ {
		rfe.Flow(0.01)
	}

	// Check flow time
	if rfe.FlowTime < 0.09 {
		t.Errorf("Expected flow time >= 0.09, got %f", rfe.FlowTime)
	}

	t.Logf("Ricci flow: time=%f, curvature=%f", rfe.FlowTime, rfe.ScalarCurvature)
}

func BenchmarkDTESN(b *testing.B) {
	dtesn := NewDTESN(128, 1024, 128)

	input := make([]float64, 128)
	for i := range input {
		input[i] = float64(i%10) * 0.1
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dtesn.Update(input)
	}
}

type fakeResponseGenerator struct {
	reply  string
	err    error
	prompt string
}

func (f *fakeResponseGenerator) Generate(_ context.Context, prompt string, _ llm.GenerateOptions) (string, error) {
	f.prompt = prompt
	return f.reply, f.err
}
