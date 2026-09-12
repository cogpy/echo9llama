package main

import (
	"context"
	"fmt"
	"time"

	"github.com/EchoCog/echollama/core/relevance"
)

func main() {
	fmt.Println("🌊 Relevance Realization Ennead - Demo")
	fmt.Println("=======================================\n")

	// Create the relevance realization engine
	ctx := context.Background()
	engine := relevance.NewEngine(ctx)

	// Start continuous optimization
	err := engine.Start()
	if err != nil {
		fmt.Printf("Error starting engine: %v\n", err)
		return
	}
	defer engine.Stop()

	// Let it run for a moment to initialize
	time.Sleep(500 * time.Millisecond)

	// Display initial state
	displayState(engine, "Initial State")

	// Simulate learning experiences
	fmt.Println("\n📚 Learning from Experiences...")
	fmt.Println("------------------------------")

	experiences := []struct {
		description string
		input       string
		output      string
		feedback    float64
	}{
		{
			description: "Understanding a scientific concept",
			input:       "How does gravity work?",
			output:      "Explained gravitational force",
			feedback:    0.9,
		},
		{
			description: "Making an ethical decision",
			input:       "Should I help this person?",
			output:      "Chose to help based on compassion",
			feedback:    0.85,
		},
		{
			description: "Practicing a skill",
			input:       "Writing code",
			output:      "Completed coding task",
			feedback:    0.7,
		},
		{
			description: "Developing self-understanding",
			input:       "Who am I becoming?",
			output:      "Reflected on personal growth",
			feedback:    0.95,
		},
	}

	for i, exp := range experiences {
		fmt.Printf("\n%d. %s\n", i+1, exp.description)

		// Create experience
		experience := &relevance.Experience{
			Input:     exp.input,
			Output:    exp.output,
			Feedback:  exp.feedback,
			Context:   make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		// Update engine
		engine.UpdateFromExperience(experience)

		// Realize relevance
		rr := engine.RealizeRelevance(exp.input)
		fmt.Printf("   Relevance Score: %.3f\n", rr.RelevanceScore)

		// Short pause between experiences
		time.Sleep(200 * time.Millisecond)
	}

	// Let the system integrate
	fmt.Println("\n🔄 Integrating knowledge across triads...")
	time.Sleep(2 * time.Second)

	// Display final state
	displayState(engine, "Final State After Learning")

	// Display metrics
	displayMetrics(engine)

	// Test relevance realization on new inputs
	fmt.Println("\n🎯 Testing Relevance Realization")
	fmt.Println("--------------------------------")

	testInputs := []string{
		"What is the meaning of life?",
		"How can I become more skilled?",
		"What is the right thing to do?",
		"How does the universe work?",
	}

	for _, input := range testInputs {
		rr := engine.RealizeRelevance(input)
		fmt.Printf("\nInput: %s\n", input)
		fmt.Printf("  Relevance: %.3f\n", rr.RelevanceScore)
		fmt.Printf("  Knowing:       %.3f\n", rr.KnowingAnalysis.OverallScore)
		fmt.Printf("  Understanding: %.3f\n", rr.UnderstandingAnalysis.OverallScore)
		fmt.Printf("  Wisdom:        %.3f\n", rr.WisdomAnalysis.OverallScore)
	}

	fmt.Println("\n✨ Demo Complete!")
}

func displayState(engine *relevance.Engine, title string) {
	fmt.Printf("\n📊 %s\n", title)
	fmt.Println(repeat("=", len(title)+3))

	state := engine.GetState()

	// Triad I: Ways of Knowing
	fmt.Println("\n🌊 Triad I: Ways of Knowing (Epistemological)")
	fmt.Printf("  Propositional (knowing-that):  %.3f\n", state.PropositionalKnowledge)
	fmt.Printf("  Procedural (knowing-how):      %.3f\n", state.ProceduralKnowledge)
	fmt.Printf("  Perspectival (knowing-as):     %.3f\n", state.PerspectivalKnowledge)
	fmt.Printf("  Participatory (knowing-by):    %.3f\n", state.ParticipatoryKnowledge)

	// Triad II: Orders of Understanding
	fmt.Println("\n🧠 Triad II: Orders of Understanding (Ontological)")
	fmt.Printf("  Nomological (how things work): %.3f\n", state.NomologicalUnderstanding)
	fmt.Printf("  Normative (what matters):      %.3f\n", state.NormativeUnderstanding)
	fmt.Printf("  Narrative (how things develop):%.3f\n", state.NarrativeUnderstanding)

	// Triad III: Practices of Wisdom
	fmt.Println("\n✨ Triad III: Practices of Wisdom (Axiological)")
	fmt.Printf("  Morality (virtue):             %.3f\n", state.MoralDevelopment)
	fmt.Printf("  Meaning (purpose):             %.3f\n", state.MeaningRealization)
	fmt.Printf("  Mastery (excellence):          %.3f\n", state.MasteryAchievement)

	// Integration
	fmt.Println("\n🌟 Integration Metrics")
	fmt.Printf("  Overall Coherence:             %.3f\n", state.OverallCoherence)
	fmt.Printf("  Relevance Optimization:        %.3f\n", state.RelevanceOptimization)
}

func displayMetrics(engine *relevance.Engine) {
	fmt.Println("\n📈 Performance Metrics")
	fmt.Println("----------------------")

	metrics := engine.GetMetrics()

	fmt.Printf("  Total Optimization Cycles:     %d\n", metrics.TotalCycles)
	fmt.Printf("  Cross-Triad Integrations:      %d\n", metrics.CrossTriadIntegrations)
	fmt.Printf("  Sophrosyne Optimizations:      %d\n", metrics.SophrosyneOptimizations)
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
