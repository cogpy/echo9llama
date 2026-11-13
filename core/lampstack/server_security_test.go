package lampstack

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

func securityTestServer(t *testing.T, config *ServerConfig) (*Server, *echo.Echo) {
	t.Helper()
	stack, err := NewStack(nil)
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(stack, config)
	e := echo.New()
	server.configureMiddleware(e)
	server.registerRoutes(e)
	return server, e
}

func TestServerForwardedHeadersCannotSpoofRealIP(t *testing.T) {
	for _, peer := range []struct{ remote, ip string }{
		{"198.51.100.23:4321", "198.51.100.23"},
		{"[2001:db8::23]:4321", "2001:db8::23"},
		{"198.51.100.23", "198.51.100.23"},
	} {
		t.Run(peer.remote, func(t *testing.T) {
			_, e := securityTestServer(t, DefaultServerConfig())
			var realIP string
			e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c *echo.Context) error {
					realIP = c.RealIP()
					return next(c)
				}
			})
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			req.RemoteAddr = peer.remote
			req.Header.Set("X-Real-IP", "203.0.113.99")
			req.Header.Set("X-Forwarded-For", "203.0.113.99, 192.0.2.1")
			req.Header.Set("Forwarded", "for=203.0.113.99;proto=https;host=attacker.example")
			req.Header.Set("X-Forwarded-Proto", "https")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK || realIP != peer.ip {
				t.Fatalf("status=%d RealIP=%q, want %q", rec.Code, realIP, peer.ip)
			}
			var payload map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(payload, map[string]string{"status": "ok", "version": Version}) {
				t.Fatalf("health payload changed: %v", payload)
			}
			if rec.Header().Get(echo.HeaderXRequestID) == "" {
				t.Fatal("request ID middleware did not run")
			}
		})
	}
}

func TestServerExistingRoutesAndJSONErrors(t *testing.T) {
	_, e := securityTestServer(t, DefaultServerConfig())
	for _, tc := range []struct {
		method, path, body string
		status             int
		errorMessage       string
	}{
		{http.MethodGet, "/health", "", 200, ""},
		{http.MethodGet, "/info", "", 200, ""},
		{http.MethodGet, "/metrics", "", 200, ""},
		{http.MethodPost, "/api/generate", "{", 400, "Invalid request body"},
		{http.MethodPost, "/api/chat", "{", 400, "Invalid request body"},
		{http.MethodPost, "/api/echo/think", "{", 400, "Invalid request body"},
		{http.MethodPost, "/api/echo/reflect", "{", 400, "Invalid request body"},
		{http.MethodGet, "/api/echo/state", "", 500, "Echo engine not available"},
		{http.MethodPost, "/api/memory/store", "{", 400, "Invalid request body"},
		{http.MethodPost, "/api/memory/retrieve", "{", 400, "Invalid request body"},
		{http.MethodGet, "/api/memory/stats", "", 503, "Memory not available"},
		{http.MethodGet, "/api/snapshot", "", 503, "Persistence not available"},
		{http.MethodPost, "/api/snapshot", "", 503, "Persistence not available"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Fatalf("status=%d want=%d body=%s", rec.Code, tc.status, rec.Body.String())
			}
			if tc.errorMessage != "" {
				var payload map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(payload, map[string]any{"error": tc.errorMessage}) {
					t.Fatalf("error payload changed: %v", payload)
				}
			}
		})
	}
}

func TestServerCORSAllowlist(t *testing.T) {
	for _, tc := range []struct {
		name         string
		origins      []string
		origin, want string
	}{
		{"legacy wildcard", nil, "https://client.example", "*"},
		{"allowed", []string{"https://client.example"}, "https://client.example", "https://client.example"},
		{"rejected", []string{"https://client.example"}, "https://attacker.example", ""},
		{"empty denies", []string{}, "https://client.example", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultServerConfig()
			config.CORSAllowOrigins = tc.origins
			_, e := securityTestServer(t, config)
			for _, method := range []string{http.MethodGet, http.MethodOptions} {
				req := httptest.NewRequest(method, "/health", nil)
				req.Header.Set(echo.HeaderOrigin, tc.origin)
				if method == http.MethodOptions {
					req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
				}
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
				if got := rec.Header().Get(echo.HeaderAccessControlAllowOrigin); got != tc.want {
					t.Fatalf("%s CORS origin=%q want=%q", method, got, tc.want)
				}
			}
		})
	}
}

// Only test code supplies this APIProvider; production stack behavior is unchanged.
type securityGenerationAPI struct {
	APIProvider
	request     *GenerationRequest
	hasDeadline bool
}

func (a *securityGenerationAPI) Generate(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error) {
	a.request = request
	_, a.hasDeadline = ctx.Deadline()
	return &GenerationResponse{Content: "answer", Model: "test-model", TokensUsed: 3, FinishReason: "stop"}, nil
}

func TestServerGeneratePreservesBindingAndResponse(t *testing.T) {
	server, e := securityTestServer(t, DefaultServerConfig())
	api := &securityGenerationAPI{}
	server.stack.api = api
	server.stack.initialized = true
	req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(`{"prompt":"hello","max_tokens":12,"temperature":0.25,"use_cognition":false}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if api.request == nil || api.request.Prompt != "hello" || api.request.MaxTokens != 12 || api.request.Temperature != 0.25 || api.request.UseCognition {
		t.Fatalf("request binding changed: %+v", api.request)
	}
	if !api.hasDeadline {
		t.Fatal("request timeout middleware did not set a deadline")
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(payload, map[string]any{"content": "answer", "model": "test-model", "tokens_used": float64(3), "finish_reason": "stop", "latency": float64(0)}) {
		t.Fatalf("generation payload changed: %v", payload)
	}
}

func TestServerStartStopRestart(t *testing.T) {
	config := DefaultServerConfig()
	config.Address = "127.0.0.1:0"
	server, _ := securityTestServer(t, config)
	for range 2 {
		if err := server.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		if server.Echo() == nil || server.httpServer.Handler != server.Echo() {
			t.Fatal("HTTP server does not use the configured Echo handler")
		}
		if server.httpServer.ReadTimeout != config.ReadTimeout || server.httpServer.WriteTimeout != config.WriteTimeout {
			t.Fatal("HTTP server timeouts differ from configuration")
		}
		if err := server.Start(context.Background()); err == nil {
			t.Fatal("duplicate Start should fail")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err := server.Stop(ctx)
		cancel()
		if err != nil {
			t.Fatal(err)
		}
		if err := server.Stop(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestServerStartReportsBindFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	config := DefaultServerConfig()
	config.Address = listener.Addr().String()
	server, _ := securityTestServer(t, config)
	if err := server.Start(context.Background()); err == nil {
		t.Fatal("occupied address should fail Start")
	}
	if server.running {
		t.Fatal("failed bind must not leave server running")
	}
}
