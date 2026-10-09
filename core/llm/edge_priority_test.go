package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestBackgroundViewYieldsToForeground(t *testing.T) {
	release := make(chan struct{})
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		<-release
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"text":"from the edge model"}]}`))
	}))
	defer srv.Close()
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()

	t.Setenv("ECHO_EDGE_COMPLETION_URL", srv.URL)
	fg := NewEdgeCompletionProviderFromEnv(nil)
	bg := fg.Background()

	done := make(chan string, 1)
	go func() {
		out, _ := fg.Generate(t.Context(), "how are you feeling?", DefaultGenerateOptions())
		done <- out
	}()
	for calls.Load() == 0 {
		time.Sleep(time.Millisecond)
	}

	// While the foreground call holds the slot, the background view must not
	// reach the model server: it answers from the fallback immediately.
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	out, err := bg.Generate(ctx, "autonomous thought", DefaultGenerateOptions())
	if err != nil || out == "" {
		t.Fatalf("background generate = %q, %v", out, err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("model server calls = %d, want 1 (background must not queue)", got)
	}
	if bg.Status().LastSource != "deterministic-fallback" {
		t.Fatalf("background source = %s", bg.Status().LastSource)
	}

	close(release)
	if got := <-done; got != "from the edge model" {
		t.Fatalf("foreground = %q", got)
	}

	// With the slot free, the background view uses the model too.
	if out, _ := bg.Generate(t.Context(), "autonomous thought", DefaultGenerateOptions()); out != "from the edge model" {
		t.Fatalf("background with free slot = %q", out)
	}
}

func TestEdgeStopsDefaultToTurnMarkers(t *testing.T) {
	if got := edgeStops(nil); len(got) == 0 || got[0] != "\nUser:" {
		t.Fatalf("default stops = %q", got)
	}
	if got := edgeStops([]string{"###"}); len(got) != 1 || got[0] != "###" {
		t.Fatalf("caller stops not honoured: %q", got)
	}
}

func TestBuildEdgePromptDoesNotDoubleWrapTranscripts(t *testing.T) {
	cases := []struct{ prompt, system, want string }{
		{"User: hi", "be Lucy", "be Lucy\n\nUser: hi\nEcho:"},
		{"User: hi", "", "User: hi\nEcho:"},
		{"plain thought", "be Lucy", "be Lucy\n\nUser: plain thought\nEcho:"},
		{"plain thought", "", "plain thought"},
	}
	for _, c := range cases {
		if got := buildEdgePrompt(c.prompt, c.system); got != c.want {
			t.Errorf("buildEdgePrompt(%q, %q) = %q, want %q", c.prompt, c.system, got, c.want)
		}
	}
}
