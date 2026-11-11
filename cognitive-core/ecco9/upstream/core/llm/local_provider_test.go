package llm

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewLocalProviderDefaults(t *testing.T) {
	t.Setenv(envLocalBaseURL, "")
	t.Setenv(envLocalModel, "")
	t.Setenv(envLocalAPIKey, "")

	p := NewLocalProvider()
	if p.Name() != "local" {
		t.Fatalf("name = %q, want local", p.Name())
	}
	if p.apiURL != defaultLocalBaseURL+"/chat/completions" {
		t.Fatalf("apiURL = %q", p.apiURL)
	}
	if p.model != defaultLocalModel {
		t.Fatalf("model = %q", p.model)
	}
}

func TestLocalProviderAvailabilityFollowsServer(t *testing.T) {
	// Unreachable server: not available, and no text is ever produced.
	t.Setenv(envLocalBaseURL, "http://127.0.0.1:1/v1")
	if NewLocalProvider().Available() {
		t.Fatal("provider reported available with no server listening")
	}

	// A server that answers GET /models makes the provider available.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"qwen-local"}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	t.Setenv(envLocalBaseURL, srv.URL+"/v1/")
	p := NewLocalProvider()
	if !p.Available() {
		t.Fatal("provider not available although the server answers /models")
	}
	if p.apiURL != srv.URL+"/v1/chat/completions" {
		t.Fatalf("trailing slash not trimmed: %q", p.apiURL)
	}

	// OpenAI itself is still gated by its key only.
	if NewOpenAIProvider("").Available() {
		t.Fatal("OpenAI provider available without a key")
	}
	if NewOpenAIProvider("k").Name() != "openai" {
		t.Fatal("OpenAI provider name changed")
	}
}
