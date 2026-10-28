package orchestration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cogpy/echo9llama/api"
)

// newFakeModelClient returns an api.Client backed by a fake /api/generate
// endpoint that answers every prompt with reply. It records prompts.
func newFakeModelClient(t *testing.T, reply string, prompts *[]string) api.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			http.NotFound(w, r)
			return
		}
		var req api.GenerateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if prompts != nil {
			*prompts = append(*prompts, req.Prompt)
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		_ = json.NewEncoder(w).Encode(api.GenerateResponse{Model: req.Model, Response: reply, Done: true})
	}))
	t.Cleanup(srv.Close)
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	return *api.NewClient(u, srv.Client())
}

// newFailingModelClient returns an api.Client whose generate endpoint errors.
func newFailingModelClient(t *testing.T) api.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"model unavailable"}`, http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	u, _ := url.Parse(srv.URL)
	return *api.NewClient(u, srv.Client())
}

func TestReflectUsesModelOutput(t *testing.T) {
	var prompts []string
	engine := NewEngine(newFakeModelClient(t, "model reflection text", &prompts))
	agent := &Agent{Name: "r", Type: AgentTypeReflective, Models: []string{"m"}}
	if err := engine.CreateAgent(context.Background(), agent); err != nil {
		t.Fatal(err)
	}
	res, err := engine.ExecuteTask(context.Background(), &Task{ID: "t", Type: TaskTypeReflect, Input: "my progress"}, agent)
	if err != nil {
		t.Fatalf("reflect: %v", err)
	}
	if res.Output != "model reflection text" {
		t.Fatalf("output = %q, want model text", res.Output)
	}
	if len(prompts) != 1 || !strings.Contains(prompts[0], "my progress") {
		t.Fatalf("prompt did not carry input: %v", prompts)
	}
}

func TestReflectModelFailureYieldsNoText(t *testing.T) {
	for name, client := range map[string]api.Client{
		"no client": {},
		"failing":   newFailingModelClient(t),
		"empty":     newFakeModelClient(t, "   ", nil),
	} {
		t.Run(name, func(t *testing.T) {
			engine := NewEngine(client)
			agent := &Agent{Name: "r", Type: AgentTypeReflective, Models: []string{"m"}}
			_ = engine.CreateAgent(context.Background(), agent)
			res, err := engine.ExecuteTask(context.Background(), &Task{ID: "t", Type: TaskTypeReflect, Input: "x"}, agent)
			if err == nil || res != nil {
				t.Fatalf("expected error and no result, got %v, %v", res, err)
			}
			if _, ok := agent.State.Memory["reflection"]; ok {
				t.Fatal("no reflection should be recorded on model failure")
			}
		})
	}
	// No model configured on the agent at all.
	engine := NewEngine(newFakeModelClient(t, "x", nil))
	agent := &Agent{Name: "r"}
	_ = engine.CreateAgent(context.Background(), agent)
	if _, err := engine.ExecuteTask(context.Background(), &Task{Type: TaskTypeReflect}, agent); err == nil {
		t.Fatal("expected error with no model")
	}
}

func TestCustomTaskUsesModel(t *testing.T) {
	engine := NewEngine(newFakeModelClient(t, "custom answer", nil))
	agent := &Agent{Name: "a", Models: []string{"m"}}
	_ = engine.CreateAgent(context.Background(), agent)
	res, err := engine.ExecuteTask(context.Background(), &Task{Type: "custom", Input: "do it"}, agent)
	if err != nil || res.Output != "custom answer" {
		t.Fatalf("got %v, %v", res, err)
	}

	engine = NewEngine(api.Client{})
	_ = engine.CreateAgent(context.Background(), agent)
	if res, err := engine.ExecuteTask(context.Background(), &Task{Type: "custom", Input: "do it"}, agent); err == nil || res != nil {
		t.Fatalf("expected failure without model client, got %v", res)
	}
}

func TestCoordinationReflection(t *testing.T) {
	result := &CoordinatedWorkflowResult{Tasks: []CoordinatedTaskResult{{TaskID: "a", Success: true}}}
	coord := &Agent{Name: "c", Models: []string{"m"}}

	engine := NewEngine(newFakeModelClient(t, "coord reflection", nil))
	got, err := engine.performCoordinationReflection(context.Background(), coord, result)
	if err != nil || got != "coord reflection" {
		t.Fatalf("got %q, %v", got, err)
	}

	engine = NewEngine(newFailingModelClient(t))
	got, err = engine.performCoordinationReflection(context.Background(), coord, result)
	if err == nil || got != "" {
		t.Fatalf("expected error and empty text, got %q, %v", got, err)
	}
}

func TestFutureSelfGuidance(t *testing.T) {
	engine := NewEngine(api.Client{})
	if _, ok := engine.GetDeepTreeEchoDashboardData()["guidance_from_future_self"]; ok {
		t.Fatal("dashboard must not carry fixed guidance text")
	}
	if g, err := engine.GenerateFutureSelfGuidance(context.Background(), "m"); err == nil || g != "" {
		t.Fatalf("expected error without model client, got %q", g)
	}
	engine = NewEngine(newFakeModelClient(t, "guidance", nil))
	if g, err := engine.GenerateFutureSelfGuidance(context.Background(), "m"); err != nil || g != "guidance" {
		t.Fatalf("got %q, %v", g, err)
	}
}

func TestToolsDoNotFabricateResults(t *testing.T) {
	engine := NewEngine(api.Client{})
	RegisterDefaultTools(engine)
	RegisterDefaultPlugins(engine)
	agent := &Agent{Name: "a"}
	_ = engine.CreateAgent(context.Background(), agent)

	_, err := engine.ExecuteTask(context.Background(), &Task{Type: TaskTypeTool, Parameters: map[string]interface{}{
		"tool": map[string]interface{}{"name": "web_search", "parameters": map[string]interface{}{"query": "q"}},
	}}, agent)
	if !errors.Is(err, ErrNoSearchBackend) {
		t.Fatalf("web_search without backend: err = %v", err)
	}

	if _, err := engine.ExecuteTask(context.Background(), &Task{Type: TaskTypeTool, Parameters: map[string]interface{}{
		"tool": map[string]interface{}{"name": "missing"},
	}}, agent); err == nil {
		t.Fatal("missing tool must be an error")
	}

	if _, err := engine.ExecuteTask(context.Background(), &Task{Type: TaskTypePlugin, Input: "great", Parameters: map[string]interface{}{
		"plugin_name": "data_analysis", "type": "sentiment", "model": "m",
	}}, agent); err == nil {
		t.Fatal("sentiment without model client must fail")
	}

	engine2 := NewEngine(newFakeModelClient(t, "positive", nil))
	RegisterDefaultPlugins(engine2)
	_ = engine2.CreateAgent(context.Background(), agent)
	res, err := engine2.ExecuteTask(context.Background(), &Task{Type: TaskTypePlugin, Input: "great", Parameters: map[string]interface{}{
		"plugin_name": "data_analysis", "type": "sentiment", "model": "m",
	}}, agent)
	if err != nil || !strings.Contains(res.Output, "positive") {
		t.Fatalf("sentiment via model: %v, %v", res, err)
	}
}

func TestRepositoryIntrospectionReadsRealContent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	snap, err := NewRepositoryIntrospector(dir).AnalyzeRepository(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range snap.SalientFiles {
		if strings.Contains(f.Content, "Simulated content") {
			t.Fatalf("simulated content in %s", f.Path)
		}
		if f.Path == "main.go" && f.Content != "package main\n" {
			t.Fatalf("content = %q", f.Content)
		}
	}
}
