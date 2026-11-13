package deeptreeecho

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
	"github.com/cogpy/echo9llama/core/persistence"
)

// livePracticeProvider exists only for the explicit integration smoke test. It
// calls the configured real OpenAI-compatible model; ordinary unit tests skip it.
// A successful HTTP model receipt is required before route evidence is exposed.
// No fallback, synthetic answers, hidden reasoning or credentials are recorded.
type livePracticeProvider struct {
	model  string
	mu     sync.Mutex
	routes map[string]llm.BackendRoutingState
}

func (p *livePracticeProvider) Name() string { return "explicit-live-model-test" }
func (p *livePracticeProvider) Available() bool {
	return os.Getenv("OPENAI_API_KEY") != "" && os.Getenv("OPENAI_API_BASE") != ""
}
func (p *livePracticeProvider) MaxTokens() int { return 4096 }
func (p *livePracticeProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	return nil, fmt.Errorf("live smoke provider is nonstreaming")
}

func (p *livePracticeProvider) GetRouteResult(trace string) (llm.BackendRoutingState, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	state, ok := p.routes[trace]
	return state, ok
}

func (p *livePracticeProvider) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	body, err := json.Marshal(map[string]any{"model": p.model, "messages": []map[string]string{{"role": "user", "content": prompt}}, "max_completion_tokens": 2000, "reasoning": map[string]string{"effort": "low"}})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(os.Getenv("OPENAI_API_BASE"), "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("live model HTTP status %d", resp.StatusCode)
	}
	var result struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1024*1024)).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) != 1 || strings.TrimSpace(result.Choices[0].Message.Content) == "" || result.Choices[0].FinishReason == "length" {
		return "", fmt.Errorf("live model did not return a complete public candidate")
	}
	p.mu.Lock()
	p.routes[opts.Routing.TraceID] = llm.BackendRoutingState{SelectedProvider: p.Name(), SelectedKind: "remote_api", SelectedModelID: p.model}
	p.mu.Unlock()
	return result.Choices[0].Message.Content, nil
}

func TestEvaluatedPracticeLiveModel(t *testing.T) {
	model := os.Getenv("ECHO_PRACTICE_LIVE_MODEL")
	if model == "" {
		t.Skip("set ECHO_PRACTICE_LIVE_MODEL to a verified catalogue model for the optional live smoke test")
	}
	provider := &livePracticeProvider{model: model, routes: make(map[string]llm.BackendRoutingState)}
	if !provider.Available() {
		t.Fatal("live smoke test requires configured sandbox model endpoint")
	}
	rig, pipeline := practiceProjectionRig(t, LLMActionPlanner{Provider: provider})
	for range 3 {
		request := admittedPracticeRequest(t, rig)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		action, err := pipeline.RunGoal(ctx, request)
		cancel()
		if err != nil {
			t.Fatalf("live planning/action failed closed: %v", err)
		}
		events, err := rig.eventStore.Query(context.Background(), persistence.CognitiveEventQuery{ActionID: action.ActionID, Limit: 1000})
		if err != nil {
			t.Fatal(err)
		}
		var result *ArtifactPracticeOutcome
		for _, event := range events {
			if event.EventType == persistence.EventTypePracticeEvaluated {
				outcome, _, err := rig.validatedPractice(context.Background(), event)
				if err != nil {
					t.Fatal(err)
				}
				result = &outcome
			}
		}
		if result == nil || result.Passed == nil {
			t.Fatal("live trial has no independent correctness outcome")
		}
		if err := rig.projectEnactionEvents(events); err != nil {
			t.Fatal(err)
		}
		if err := rig.replayOutcomeLearning(context.Background()); err != nil {
			t.Fatal(err)
		}
		t.Logf("real model=%s task=%s passed=%t held_out=%t artifact_sha256=%s", model, request.Practice.TaskID, *result.Passed, practiceHeldOut(request.Practice.TaskID), result.ArtifactSHA256)
	}
	status := rig.outcomeLearningStatus()
	if status.Training+status.HeldOut != 3 || status.Rejected != 0 {
		t.Fatalf("live evidence/training mismatch: %#v", status)
	}
	t.Logf("live observed=%d training=%d held_out=%d readout_version=%d data_sha256=%s", status.Observed, status.Training, status.HeldOut, status.ReadoutVersion, status.DataSHA256)
	// A smoke test checks the integration, not comparative or general wisdom gains.
}
