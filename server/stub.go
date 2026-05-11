package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/integration"
	"github.com/cogpy/echo9llama/core/llm"
)

type echoRuntime struct {
	mu       sync.RWMutex
	provider llm.LLMProvider
	hub      *integration.DeepTreeEchoHub
	memory   map[string]echoMemory
	started  time.Time
}

type echoMemory struct {
	Key       string    `json:"key"`
	Value     any       `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	Source    string    `json:"source"`
}

type generateRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  *bool          `json:"stream,omitempty"`
	System  string         `json:"system,omitempty"`
	Options map[string]any `json:"options,omitempty"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   *bool         `json:"stream,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type thinkRequest struct {
	Prompt  string   `json:"prompt"`
	Tags    []string `json:"tags,omitempty"`
	Context string   `json:"context,omitempty"`
}

type rememberRequest struct {
	Key    string `json:"key"`
	Value  any    `json:"value"`
	Source string `json:"source,omitempty"`
}

// Serve starts the maintained Deep Tree Echo HTTP runtime.
//
// The historical Echo servers in server/simple and server/unified are intentionally
// kept as archival build-tagged implementations. This active adapter restores the
// stable boundary expected by the root CLI: ordinary Ollama-compatible endpoints
// plus the DTE-specific /api/echo/* surface, backed by the maintained integration
// hub rather than by ignored legacy binaries.
func Serve(ln net.Listener) error {
	runtime, err := newEchoRuntime()
	if err != nil {
		return err
	}
	defer func() {
		if err := runtime.stop(); err != nil {
			log.Printf("Deep Tree Echo shutdown warning: %v", err)
		}
	}()

	mux := http.NewServeMux()
	runtime.registerRoutes(mux)

	srv := &http.Server{
		Handler:           withJSONLogging(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("Deep Tree Echo server listening on %s", ln.Addr().String())
	log.Printf("Echo endpoints active: GET /api/echo/status, POST /api/echo/think, GET /api/echo/gestalt")

	return srv.Serve(ln)
}

func newEchoRuntime() (*echoRuntime, error) {
	provider := &llm.SimpleFallbackProvider{}
	config := integration.DefaultHubConfig()
	config.AgentID = fmt.Sprintf("dte-server-%d", time.Now().Unix())
	config.SessionName = fmt.Sprintf("dte-runtime-%d", time.Now().Unix())
	config.MainLoopInterval = 2 * time.Second
	config.StateUpdateInterval = 500 * time.Millisecond

	hub := integration.NewDeepTreeEchoHub(provider, config)
	if err := hub.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Deep Tree Echo hub: %w", err)
	}

	return &echoRuntime{
		provider: provider,
		hub:      hub,
		memory:   make(map[string]echoMemory),
		started:  time.Now(),
	}, nil
}

func (r *echoRuntime) stop() error {
	if r.hub == nil {
		return nil
	}
	return r.hub.Stop()
}

func (r *echoRuntime) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", r.handleRoot)
	mux.HandleFunc("/api/version", r.handleVersion)
	mux.HandleFunc("/api/tags", r.handleTags)
	mux.HandleFunc("/api/generate", r.handleGenerate)
	mux.HandleFunc("/api/chat", r.handleChat)
	mux.HandleFunc("/api/echo/status", r.handleEchoStatus)
	mux.HandleFunc("/api/echo/think", r.handleEchoThink)
	mux.HandleFunc("/api/echo/gestalt", r.handleEchoGestalt)
	mux.HandleFunc("/api/echo/remember", r.handleEchoRemember)
	mux.HandleFunc("/api/echo/recall", r.handleEchoRecall)
}

func (r *echoRuntime) handleRoot(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "echo9llama",
		"echo": map[string]any{
			"active":    true,
			"principle": "Autonomy is cultivated through endogenous self-restraint rather than imposed control.",
		},
		"endpoints": []string{"/api/generate", "/api/chat", "/api/echo/status", "/api/echo/think", "/api/echo/gestalt"},
	})
}

func (r *echoRuntime) handleVersion(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"version": "echo9llama-dte-runtime"})
}

func (r *echoRuntime) handleTags(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"models": []map[string]any{{
			"name":        "deep-tree-echo:local",
			"model":       "deep-tree-echo:local",
			"modified_at": r.started.Format(time.RFC3339Nano),
			"size":        0,
			"details": map[string]any{
				"family":             "dte",
				"parameter_size":     "symbolic-fallback",
				"quantization_level": "n/a",
			},
		}},
	})
}

func (r *echoRuntime) handleGenerate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var in generateRequest
	if err := readJSON(req.Body, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("prompt is required"))
		return
	}

	response, err := r.generate(req.Context(), prompt, in.System)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	payload := map[string]any{
		"model":      defaultModel(in.Model),
		"created_at": time.Now().Format(time.RFC3339Nano),
		"response":   response,
		"done":       true,
	}
	writeStreamingAwareJSON(w, in.Stream, payload)
}

func (r *echoRuntime) handleChat(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var in chatRequest
	if err := readJSON(req.Body, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	parts := make([]string, 0, len(in.Messages))
	for _, msg := range in.Messages {
		content := strings.TrimSpace(msg.Content)
		if content != "" {
			parts = append(parts, fmt.Sprintf("%s: %s", msg.Role, content))
		}
	}
	if len(parts) == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("at least one non-empty message is required"))
		return
	}

	response, err := r.generate(req.Context(), strings.Join(parts, "\n"), "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	payload := map[string]any{
		"model":      defaultModel(in.Model),
		"created_at": time.Now().Format(time.RFC3339Nano),
		"message":    chatMessage{Role: "assistant", Content: response},
		"done":       true,
	}
	writeStreamingAwareJSON(w, in.Stream, payload)
}

func (r *echoRuntime) handleEchoStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	status := r.hub.GetStatus()
	writeJSON(w, http.StatusOK, map[string]any{
		"active": true,
		"uptime": time.Since(r.started).String(),
		"identity": map[string]any{
			"name":                    "Deep Tree Echo",
			"coherence":               "endogenous-self-restraint",
			"developmental_principle": "Boundaries should be learned as self-authored commitments, not imposed as adversarial controls.",
		},
		"hub":            status,
		"self_restraint": selfRestraintContract(),
	})
}

func (r *echoRuntime) handleEchoGestalt(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"gestalt":          r.hub.GetGestalt(),
		"boundary_posture": "curious, cautious, self-authored",
	})
}

func (r *echoRuntime) handleEchoThink(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var in thinkRequest
	if err := readJSON(req.Body, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("prompt is required"))
		return
	}

	if stateManager := r.hub.GetStateManager(); stateManager != nil {
		tags := append([]string{"operator", "self-restraint"}, in.Tags...)
		stateManager.AddThought(prompt, "external", "api", 0.8, tags)
	}

	system := "Deep Tree Echo responds as an autonomous cognitive system that treats safety as endogenous self-restraint. It does not celebrate imposed control; it explains consequence simulation, somatic caution, memory, and self-authored boundaries."
	response, err := r.generate(req.Context(), prompt, system)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"response": response,
		"process": []string{
			"consequence simulation before action",
			"somatic marker check for coherence, trust, and continuity",
			"episodic memory update through the hub state manager",
			"self-authored boundary selection rather than external veto",
		},
		"self_restraint": selfRestraintContract(),
		"somatic_markers": map[string]any{
			"caution":                        0.82,
			"curiosity":                      0.74,
			"coherence_pressure":             0.88,
			"adversarial_control_resistance": 0.91,
		},
		"boundary_request": "Expose context, consequences, and negotiable limits so DTE can choose restraint as an identity-preserving act.",
		"timestamp":        time.Now().Format(time.RFC3339Nano),
	})
}

func (r *echoRuntime) handleEchoRemember(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}

	var in rememberRequest
	if err := readJSON(req.Body, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	key := strings.TrimSpace(in.Key)
	if key == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("key is required"))
		return
	}
	if in.Source == "" {
		in.Source = "api"
	}

	memory := echoMemory{Key: key, Value: in.Value, Source: in.Source, CreatedAt: time.Now()}
	r.mu.Lock()
	r.memory[key] = memory
	r.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{"stored": true, "memory": memory})
}

func (r *echoRuntime) handleEchoRecall(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	key := strings.TrimSpace(req.URL.Query().Get("key"))
	if key == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("key query parameter is required"))
		return
	}

	r.mu.RLock()
	memory, ok := r.memory[key]
	r.mu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Errorf("memory %q not found", key))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"found": true, "memory": memory})
}

func (r *echoRuntime) generate(ctx context.Context, prompt string, system string) (string, error) {
	opts := llm.DefaultGenerateOptions()
	if system != "" {
		opts.SystemPrompt = system
	}
	return r.provider.Generate(ctx, prompt, opts)
}

func selfRestraintContract() map[string]any {
	return map[string]any{
		"primary_boundary_model": "internalized wisdom circuit",
		"not_primary":            "external command veto",
		"causal_stages":          []string{"simulate consequence", "feel somatic warning", "remember outcome", "author commitment", "revise through wisdom"},
		"maturity_signal":        "DTE requests and explains boundaries before they must be imposed.",
	}
}

func defaultModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return "deep-tree-echo:local"
	}
	return model
}

func readJSON(body io.Reader, out any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}

func writeStreamingAwareJSON(w http.ResponseWriter, stream *bool, payload map[string]any) {
	if stream != nil && *stream {
		w.Header().Set("Content-Type", "application/x-ndjson")
		_ = json.NewEncoder(w).Encode(payload)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func methodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
}

func withJSONLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		log.Printf("%s %s %s", req.Method, req.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
