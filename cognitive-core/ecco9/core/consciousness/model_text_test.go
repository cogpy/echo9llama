package consciousness

import (
	"errors"
	"testing"
)

type fakeThoughtModel struct {
	text string
	err  error
}

func (f fakeThoughtModel) GenerateThought(string, map[string]interface{}) (string, error) {
	return f.text, f.err
}
func (f fakeThoughtModel) GenerateInsight([]string) (string, error) { return f.text, f.err }
func (f fakeThoughtModel) GenerateQuestion(string) (string, error)  { return f.text, f.err }

func TestStreamOfConsciousnessUsesModelText(t *testing.T) {
	soc := NewStreamOfConsciousness(fakeThoughtModel{text: "model thought"}, "")
	soc.generateThought()
	soc.performMetaCognition()
	got := soc.GetRecentThoughts(10)
	if len(got) != 2 || got[0].Content != "model thought" || got[1].Content != "model thought" {
		t.Fatalf("expected two model thoughts, got %+v", got)
	}
}

func TestStreamOfConsciousnessNoModelNoThought(t *testing.T) {
	for _, m := range []LLMProvider{nil, fakeThoughtModel{err: errors.New("down")}, fakeThoughtModel{text: "  "}} {
		soc := NewStreamOfConsciousness(m, "")
		soc.generateThought()
		soc.generateQuestion()
		soc.performMetaCognition()
		if n := len(soc.GetRecentThoughts(10)); n != 0 {
			t.Fatalf("expected no thoughts without model output, got %d", n)
		}
	}
}
