package echobeats

import (
	"context"
	"errors"
	"testing"

	"github.com/EchoCog/echollama/core/llm"
)

type fakeModel struct {
	reply string
	err   error
	calls int
}

func (f *fakeModel) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	f.calls++
	return f.reply, f.err
}

func TestDiscussionReplyComesFromModel(t *testing.T) {
	dm := NewDiscussionManager(NewInterestPatternSystem(""), "")
	if err := dm.JoinDiscussion("d1", "memory", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := dm.GenerateResponse("d1"); err == nil {
		t.Fatal("expected an error without a model")
	}
	if msgs, _ := dm.GetDiscussionContext("d1"); len(msgs) != 0 {
		t.Fatalf("no reply should be recorded without a model, got %d", len(msgs))
	}

	dm.SetTextGenerator(&fakeModel{err: errors.New("down")})
	if _, err := dm.GenerateResponse("d1"); err == nil {
		t.Fatal("expected an error when the model fails")
	}

	dm.SetTextGenerator(&fakeModel{reply: "  model reply  "})
	got, err := dm.GenerateResponse("d1")
	if err != nil || got != "model reply" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestPhaseProcessorProducesNothingWithoutModel(t *testing.T) {
	p := NewDefaultPhaseProcessor()
	stream, err := p.ProcessT2IdeaFormation(Expressive)
	if err != nil {
		t.Fatal(err)
	}
	content := stream.Content.(map[string]interface{})
	if _, ok := content["idea"]; ok {
		t.Fatalf("idea produced without a model: %v", content)
	}
	stream, _ = p.ProcessT5ActionSequence(Expressive)
	if _, ok := stream.Content.(map[string]interface{})["action"]; ok {
		t.Fatal("action produced without a model")
	}
	stream, _ = p.ProcessT4SensoryInput(Expressive)
	if stream.Content.(map[string]interface{})["type"] != "no_sensory_data" {
		t.Fatal("sensory input invented without real input")
	}

	p.SetTextGenerator(&fakeModel{reply: "an idea from the model"})
	stream, _ = p.ProcessT2IdeaFormation(Expressive)
	if got := stream.Content.(map[string]interface{})["idea"]; got != "an idea from the model" {
		t.Fatalf("got %v", got)
	}
	stream, _ = p.ProcessT5ActionSequence(Expressive)
	content = stream.Content.(map[string]interface{})
	if content["executed"] != false {
		t.Fatalf("action claimed executed without an executor: %v", content)
	}
}

func TestThreePhaseStepsUseModel(t *testing.T) {
	eb := NewEchoBeatsThreePhase()
	var thoughts []string
	eb.SetThoughtCallback(func(th string) { thoughts = append(thoughts, th) })

	eb.executeNextStep()
	if eb.stepHistory[0].Success || len(thoughts) != 0 {
		t.Fatal("step without a model must not succeed or emit a thought")
	}

	eb.SetTextGenerator(&fakeModel{reply: "focus from the model"})
	eb.executeNextStep()
	if !eb.stepHistory[1].Success || len(thoughts) != 1 || thoughts[0] != "focus from the model" {
		t.Fatalf("history=%+v thoughts=%v", eb.stepHistory[1], thoughts)
	}
}

func TestNarrativeHasNoFixedGoals(t *testing.T) {
	np := NewNarrativeProcess()
	stream, _ := np.ProcessT3E()
	if goals := stream.Content.(map[string]interface{})["goals"].([]map[string]interface{}); len(goals) != 0 {
		t.Fatalf("fixed goals without a model: %v", goals)
	}
	np.SetTextGenerator(&fakeModel{reply: "- learn\n- reflect"})
	stream, _ = np.ProcessT3E()
	if goals := stream.Content.(map[string]interface{})["goals"].([]map[string]interface{}); len(goals) != 2 || goals[0]["description"] != "learn" {
		t.Fatalf("got %v", goals)
	}
}

func TestParseScenarioEstimate(t *testing.T) {
	d, p, q := parseScenarioEstimate("A scenario.\nprobability=0.4 desirability=0.9")
	if d != "A scenario." || p != 0.4 || q != 0.9 {
		t.Fatalf("got %q %v %v", d, p, q)
	}
}
