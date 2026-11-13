package lampstack

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// scriptedAPI answers Generate calls in order; nil entries return errFake.
type scriptedAPI struct {
	replies []*string
	calls   int
}

var errFake = errors.New("model unavailable")

func strp(s string) *string { return &s }

func (s *scriptedAPI) Generate(_ context.Context, _ *GenerationRequest) (*GenerationResponse, error) {
	i := s.calls
	s.calls++
	if i >= len(s.replies) || s.replies[i] == nil {
		return nil, errFake
	}
	return &GenerationResponse{Content: *s.replies[i]}, nil
}

func (s *scriptedAPI) Chat(context.Context, []Message) (*ChatResponse, error) { return nil, errFake }

func (s *scriptedAPI) Embed(context.Context, string) ([]float64, error) { return nil, errFake }
func (s *scriptedAPI) GetProviders() []ProviderInfo                     { return nil }
func (s *scriptedAPI) SetPrimaryProvider(string) error                  { return nil }
func (s *scriptedAPI) IsAvailable() bool                                { return true }

func TestReflectInsightsComeFromModel(t *testing.T) {
	api := &scriptedAPI{replies: []*string{strp("a reflection about patterns"), strp("- Patterns repeat\n- Memory shapes thought\n")}}
	e := NewEchoIntegration(DefaultConfig(), api, nil, nil)
	resp, err := e.Reflect(context.Background(), "ctx")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Reflection != "a reflection about patterns" {
		t.Fatalf("reflection = %q", resp.Reflection)
	}
	if strings.Join(resp.Insights, "|") != "Patterns repeat|Memory shapes thought" {
		t.Fatalf("insights = %q", resp.Insights)
	}
}

func TestReflectInsightFailureYieldsNoFixedInsights(t *testing.T) {
	api := &scriptedAPI{replies: []*string{strp("an insight about a key pattern"), nil}}
	e := NewEchoIntegration(DefaultConfig(), api, nil, nil)
	resp, err := e.Reflect(context.Background(), "ctx")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Insights) != 0 {
		t.Fatalf("expected no insights on model failure, got %q", resp.Insights)
	}

	// Reflection generation failure: no text at all.
	e = NewEchoIntegration(DefaultConfig(), &scriptedAPI{}, nil, nil)
	if resp, err := e.Reflect(context.Background(), "ctx"); err == nil || resp != nil {
		t.Fatalf("expected error and no response, got %v", resp)
	}
}
