package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	"github.com/spf13/cobra"
)

func newEchoAssessTestCommand() *cobra.Command {
	c := &cobra.Command{Use: "assess"}
	c.Flags().Bool("json", false, "")
	c.Flags().StringP("output", "o", "", "")
	c.Flags().Bool("continuous", false, "")
	c.Flags().Duration("interval", time.Minute, "")
	return c
}

type assessFake struct {
	reply  string
	err    error
	prompt string
}

func (f *assessFake) Generate(_ context.Context, prompt string, _ llm.GenerateOptions) (string, error) {
	f.prompt = prompt
	return f.reply, f.err
}

func TestBuildEchoAssessmentFromModel(t *testing.T) {
	fake := &assessFake{reply: "SUMMARY: model summary text.\nPRINCIPLE: model principle one\nPRINCIPLE: model principle two"}
	a, err := buildEchoAssessment(context.Background(), fake)
	if err != nil {
		t.Fatal(err)
	}
	if a.Summary != "model summary text." {
		t.Fatalf("summary = %q", a.Summary)
	}
	if len(a.Principles) != 2 || a.Principles[0] != "model principle one" {
		t.Fatalf("principles = %q", a.Principles)
	}
	if !strings.Contains(fake.prompt, "endogenous wisdom circuit") {
		t.Fatalf("prompt lacks runtime facts: %q", fake.prompt)
	}
}

func TestBuildEchoAssessmentModelFailure(t *testing.T) {
	for name, gen := range map[string]echoTextGenerator{
		"nil":   nil,
		"error": &assessFake{err: errors.New("down")},
		"empty": &assessFake{reply: "  "},
	} {
		a, err := buildEchoAssessment(context.Background(), gen)
		if err == nil {
			t.Fatalf("%s: expected an error", name)
		}
		if a.Summary != "" || len(a.Principles) != 0 {
			t.Fatalf("%s: assessment text produced without the model: %+v", name, a)
		}
	}
}

func TestEchoAssessHandlerNoModelReturnsError(t *testing.T) {
	old := newEchoAssessGenerator
	t.Cleanup(func() { newEchoAssessGenerator = old })
	newEchoAssessGenerator = func() echoTextGenerator { return &assessFake{err: llm.ErrNoEdgeModel} }

	cmd := newEchoAssessTestCommand()
	if err := EchoAssessHandler(cmd, nil); !errors.Is(err, llm.ErrNoEdgeModel) {
		t.Fatalf("err = %v, want ErrNoEdgeModel", err)
	}
}
