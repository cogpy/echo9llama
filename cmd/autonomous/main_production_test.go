package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

type productionMockProvider struct{}

func (productionMockProvider) Generate(context.Context, string, llm.GenerateOptions) (string, error) {
	return "reflective mock response", nil
}

func (productionMockProvider) StreamGenerate(context.Context, string, llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	stream := make(chan llm.StreamChunk, 1)
	stream <- llm.StreamChunk{Content: "reflective mock response", Done: true}
	close(stream)
	return stream, nil
}

func (productionMockProvider) Name() string    { return "production-mock" }
func (productionMockProvider) Available() bool { return true }
func (productionMockProvider) MaxTokens() int  { return 4096 }

func TestProductionConstructorReturnsUnifiedRuntime(t *testing.T) {
	t.Setenv("ECHO_STATE_DIRECTORY", t.TempDir())
	t.Setenv("ECHO_SESSION_NAME", "production-constructor-test")

	var orchestrator *deeptreeecho.UnifiedAutonomousOrchestrator = newProductionOrchestrator(productionMockProvider{})
	if orchestrator == nil {
		t.Fatal("production constructor returned nil")
	}
	status := orchestrator.GetStatus()
	if status.SessionID != "production-constructor-test" {
		t.Fatalf("expected configured session, got %q", status.SessionID)
	}
	if status.Provider != "production-mock" || !status.ProviderAvailable {
		t.Fatalf("unexpected provider status: %#v", status)
	}
}

func TestProductionEnvironmentConfiguration(t *testing.T) {
	t.Setenv("ECHO_SESSION_NAME", "env-session")
	t.Setenv("ECHO_IDENTITY", "A persistent test identity")
	t.Setenv("ECHO_PERSONA", "magnetic confidence, playful wit, scientific brilliance, bounded by wisdom")
	t.Setenv("ECHO_STATE_DIRECTORY", t.TempDir())
	t.Setenv("ECHO_MAIN_LOOP_INTERVAL", "125ms")
	t.Setenv("ECHO_THOUGHT_INTERVAL", "250ms")
	t.Setenv("ECHO_WAKE_DURATION", "2s")
	t.Setenv("ECHO_REST_DURATION", "750ms")
	t.Setenv("ECHO_DREAM_LIGHT_DURATION", "10ms")
	t.Setenv("ECHO_DREAM_DEEP_DURATION", "20ms")
	t.Setenv("ECHO_DREAM_REM_DURATION", "30ms")
	t.Setenv("ECHO_STATE_SYNC_INTERVAL", "invalid")

	config := loadOrchestratorConfigFromEnvironment()
	if config.SessionName != "env-session" || config.IdentityContext != "A persistent test identity" {
		t.Fatalf("string environment configuration was not applied: %#v", config)
	}
	if !strings.Contains(config.PersonaContext, "playful wit") || !strings.Contains(config.PersonaContext, "bounded by wisdom") {
		t.Fatalf("persona configuration was not preserved: %q", config.PersonaContext)
	}
	if config.MainLoopInterval != 125*time.Millisecond || config.ThoughtInterval != 250*time.Millisecond {
		t.Fatalf("cognitive intervals were not applied: main=%s thought=%s", config.MainLoopInterval, config.ThoughtInterval)
	}
	if config.WakeDuration != 2*time.Second || config.RestDuration != 750*time.Millisecond {
		t.Fatalf("wake/rest durations were not applied: wake=%s rest=%s", config.WakeDuration, config.RestDuration)
	}
	if config.DreamLightDuration != 10*time.Millisecond || config.DreamDeepDuration != 20*time.Millisecond || config.DreamREMDuration != 30*time.Millisecond {
		t.Fatalf("dream durations were not applied: %#v", config)
	}
	if config.StateSyncInterval <= 0 {
		t.Fatal("invalid duration should leave a positive default state-sync interval")
	}
}

func TestProductionHTTPServerDefaultsToLoopback(t *testing.T) {
	t.Setenv("PORT", "19090")
	t.Setenv("ECHO_HTTP_ADDR", "")
	if got := newProductionHTTPServer(nil).Addr; got != "127.0.0.1:19090" {
		t.Fatalf("expected secure loopback default, got %q", got)
	}

	t.Setenv("ECHO_HTTP_ADDR", "0.0.0.0")
	if got := newProductionHTTPServer(nil).Addr; got != "0.0.0.0:19090" {
		t.Fatalf("expected explicit deployment bind override, got %q", got)
	}
}

func TestProductionHealthAndStatusReflectLifecycle(t *testing.T) {
	config := deeptreeecho.DefaultOrchestratorConfig()
	config.EnablePersistence = false
	config.EnableSkillLearning = false
	config.EnableDiscussionMonitoring = false
	config.MainLoopInterval = time.Hour
	config.ThoughtInterval = time.Hour
	config.GoalReviewInterval = time.Hour
	config.WisdomSynthesisInterval = time.Hour
	config.StateSyncInterval = time.Hour
	orchestrator := deeptreeecho.NewUnifiedAutonomousOrchestrator(productionMockProvider{}, config)
	handler := newProductionHandler(orchestrator)

	resting := httptest.NewRecorder()
	handler.ServeHTTP(resting, httptest.NewRequest(http.MethodGet, "/health", nil))
	if resting.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected resting health to return 503, got %d: %s", resting.Code, resting.Body.String())
	}

	if err := orchestrator.Awaken(); err != nil {
		t.Fatalf("Awaken failed: %v", err)
	}
	t.Cleanup(func() {
		if orchestrator.GetStatus().Running {
			_ = orchestrator.Sleep()
		}
	})

	healthy := httptest.NewRecorder()
	handler.ServeHTTP(healthy, httptest.NewRequest(http.MethodGet, "/health", nil))
	if healthy.Code != http.StatusOK {
		t.Fatalf("expected awake health to return 200, got %d: %s", healthy.Code, healthy.Body.String())
	}

	var healthPayload map[string]interface{}
	if err := json.Unmarshal(healthy.Body.Bytes(), &healthPayload); err != nil {
		t.Fatalf("invalid health JSON: %v", err)
	}
	if healthPayload["running"] != true || healthPayload["awake"] != true {
		t.Fatalf("health payload did not reflect lifecycle: %#v", healthPayload)
	}

	statusResponse := httptest.NewRecorder()
	handler.ServeHTTP(statusResponse, httptest.NewRequest(http.MethodGet, "/status", nil))
	if statusResponse.Code != http.StatusOK {
		t.Fatalf("status endpoint failed: %d", statusResponse.Code)
	}
	var statusPayload map[string]interface{}
	if err := json.Unmarshal(statusResponse.Body.Bytes(), &statusPayload); err != nil {
		t.Fatalf("invalid status JSON: %v", err)
	}
	if statusPayload["provider"] != "production-mock" {
		t.Fatalf("status endpoint hid provider truth: %#v", statusPayload)
	}
	if statusPayload["iteration"] != productionIteration {
		t.Fatalf("status endpoint reported wrong iteration: %#v", statusPayload)
	}
	if _, exposed := statusPayload["state_directory"]; exposed {
		t.Fatalf("status endpoint exposed private state path: %#v", statusPayload)
	}
}
