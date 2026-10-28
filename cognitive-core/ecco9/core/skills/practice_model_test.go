package skills

import (
	"context"
	"testing"

	"github.com/EchoCog/echollama/core/llm"
)

type scriptedModel struct {
	replies []string
}

func (m *scriptedModel) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	r := m.replies[0]
	m.replies = m.replies[1:]
	return r, nil
}

func TestPracticeRequiresModel(t *testing.T) {
	sps := NewSkillPracticeSystem()
	skill := &Skill{ID: "s", Name: "Patterns"}
	scenario := &PracticeScenario{Prompt: "2, 4, 8?", Evaluation: EvaluationCriteria{Metrics: []string{"accuracy", "speed"}}}

	if _, _, err := sps.executePractice(skill, scenario); err == nil {
		t.Fatal("expected an error without a model")
	}

	sps.SetTextGenerator(&scriptedModel{replies: []string{"16, doubling", "accuracy=0.9"}})
	res, answer, err := sps.executePractice(skill, scenario)
	if err != nil {
		t.Fatal(err)
	}
	if answer != "16, doubling" || res.MetricScores["accuracy"] != 0.9 {
		t.Fatalf("answer=%q scores=%v", answer, res.MetricScores)
	}
}
