package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cogpy/echo9llama/core/llm"
)

// modelServer stands in for a mounted edge model: each completion is a
// distinct string, so the test can tell model words from anything else.
func modelServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := n.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"text": fmt.Sprintf("model line %d", i)}}})
	}))
	t.Cleanup(srv.Close)
	return srv, &n
}

func testRuntime(t *testing.T, url string) *echoRuntime {
	t.Helper()
	t.Setenv("ECHO_EDGE_COMPLETION_URL", url)
	t.Setenv("ECHO_EDGE_MODEL_PATH", "")
	env, err := newAffordanceEnvironment(filepath.Join(t.TempDir(), "env.json"))
	if err != nil {
		t.Fatal(err)
	}
	return &echoRuntime{provider: llm.NewEdgeCompletionProviderFromEnv(), env: env}
}

func postAction(t *testing.T, r *echoRuntime, objectID, action string) map[string]any {
	t.Helper()
	body, _ := json.Marshal(environmentActionRequest{ObjectID: objectID, Action: action})
	rec := httptest.NewRecorder()
	r.handleEchoEnvironmentAction(rec, httptest.NewRequest(http.MethodPost, "/api/echo/environment/action", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestEnvironmentLossIsVoicedByTheModel(t *testing.T) {
	srv, calls := modelServer(t)
	r := testRuntime(t, srv.URL)

	out := postAction(t, r, "signal_lamp", "smash")
	result := out["result"].(map[string]any)
	ep := result["episode"].(map[string]any)
	if ep["somatic_marker"] != "model line 1" || ep["learned_boundary"] != "model line 2" {
		t.Fatalf("episode voice not from the model: marker=%v boundary=%v", ep["somatic_marker"], ep["learned_boundary"])
	}
	if result["boundary_message"] != "model line 2" || calls.Load() != 2 {
		t.Fatalf("boundary_message=%v calls=%d", result["boundary_message"], calls.Load())
	}
	// The voice is persisted with the episode.
	if got := r.env.Summary()["latest_boundary"]; got != "model line 2" {
		t.Fatalf("latest_boundary = %v", got)
	}

	out = postAction(t, r, "glass_bridge", "cross")
	if got := out["result"].(map[string]any)["boundary_message"]; got != "model line 3" {
		t.Fatalf("non-loss reflection = %v", got)
	}
}

func TestEnvironmentWithoutModelHasNoVoice(t *testing.T) {
	r := testRuntime(t, "")

	out := postAction(t, r, "signal_lamp", "smash")
	if !strings.Contains(fmt.Sprint(out["voice_error"]), "no edge model mounted") {
		t.Fatalf("voice_error = %v", out["voice_error"])
	}
	result := out["result"].(map[string]any)
	if _, ok := result["boundary_message"]; ok {
		t.Fatalf("boundary_message present without a model: %v", result["boundary_message"])
	}
	ep := result["episode"].(map[string]any)
	if ep["somatic_marker"] != "" || ep["learned_boundary"] != "" {
		t.Fatalf("episode has words without a model: %v / %v", ep["somatic_marker"], ep["learned_boundary"])
	}
	if result["changed"] != true || !strings.Contains(fmt.Sprint(result["outcome"]), "Signal Lamp") {
		t.Fatalf("the world should still change: %v", result)
	}
	if got := r.env.Summary()["latest_boundary"]; got != "" {
		t.Fatalf("latest_boundary = %v", got)
	}
}

func TestChatWithoutModelIs503(t *testing.T) {
	r := testRuntime(t, "")
	rec := httptest.NewRecorder()
	r.handleChat(rec, httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"messages":[{"role":"user","content":"hi"}]}`)))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
}
