package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/cogpy/echo9llama/api"
)

// captureStdout runs fn and returns what it printed.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	runErr := fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), runErr
}

func tasksCmd() *cobra.Command {
	c := &cobra.Command{}
	c.SetContext(context.Background()) // cobra's Execute sets this in the real CLI
	c.Flags().StringSlice("tasks", nil, "")
	c.Flags().Bool("sequential", false, "")
	c.Flags().String("keepalive", "", "")
	return c
}

func TestRunTasksPrintsServerOutput(t *testing.T) {
	var got api.OrchestrationRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orchestration/tasks" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		_ = json.NewEncoder(w).Encode(api.OrchestrationResponse{
			AgentID: got.AgentID,
			Tasks:   []api.OrchestrationTaskResult{{Type: "generate", Status: "completed", Output: "words from the model"}},
		})
	}))
	defer srv.Close()
	t.Setenv("OLLAMA_HOST", srv.URL)

	c := tasksCmd()
	_ = c.Flags().Set("tasks", "generate:say hi")
	out, err := captureStdout(t, func() error { return RunTasksHandler(c, []string{"agent-1"}) })
	if err != nil {
		t.Fatal(err)
	}
	if got.AgentID != "agent-1" || len(got.Tasks) != 1 || got.Tasks[0].Input != "say hi" {
		t.Fatalf("request = %+v", got)
	}
	if !strings.Contains(out, "words from the model") || strings.Contains(out, "placeholder") {
		t.Fatalf("output = %q", out)
	}
}

func TestRunTasksServerErrorIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"agent not found"}`, http.StatusInternalServerError)
	}))
	defer srv.Close()
	t.Setenv("OLLAMA_HOST", srv.URL)

	c := tasksCmd()
	_ = c.Flags().Set("tasks", "generate:say hi")
	if _, err := captureStdout(t, func() error { return RunTasksHandler(c, []string{"agent-1"}) }); err == nil || !strings.Contains(err.Error(), "agent not found") {
		t.Fatalf("err = %v", err)
	}
}
