//go:build e2e

package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func autonomousBaseURL() string {
	if configured := strings.TrimRight(os.Getenv("ECHOSELF_BASE_URL"), "/"); configured != "" {
		return configured
	}
	return "http://127.0.0.1:8081"
}

func autonomousClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

func getJSON(t *testing.T, path string) (int, map[string]interface{}) {
	t.Helper()
	response, err := autonomousClient().Get(autonomousBaseURL() + path)
	require.NoError(t, err)
	defer response.Body.Close()

	var payload map[string]interface{}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&payload))
	return response.StatusCode, payload
}

func TestAutonomousServerHealthE2E(t *testing.T) {
	statusCode, health := getJSON(t, "/health")
	require.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "healthy", health["status"])
	assert.Equal(t, "Deep Tree Echo", health["identity"])
	assert.Equal(t, true, health["running"])
	assert.Equal(t, true, health["awake"])
	assert.Equal(t, true, health["provider_available"])
	assert.NotEmpty(t, health["wake_rest_state"])
}

func TestAutonomousServerStatusE2E(t *testing.T) {
	statusCode, status := getJSON(t, "/status")
	require.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "Deep Tree Echo", status["identity"])
	assert.Equal(t, true, status["running"])
	assert.Equal(t, true, status["autonomous"])
	assert.Equal(t, true, status["provider_available"])
	assert.NotEmpty(t, status["session_id"])
	assert.NotEmpty(t, status["uptime"])
	assert.NotContains(t, status, "state_directory")

	cycles, ok := status["cycles"].(float64)
	require.True(t, ok, "cycles should be a JSON number")
	assert.Greater(t, cycles, float64(0), "the prompt-independent main loop should advance")
}

func TestAutonomousServerMetricsE2E(t *testing.T) {
	response, err := autonomousClient().Get(autonomousBaseURL() + "/metrics")
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Contains(t, response.Header.Get("Content-Type"), "text/plain")

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	metrics := string(body)
	for _, name := range []string{
		"echo_running",
		"echo_awake",
		"echo_cycles_total",
		"echo_thoughts_total",
		"echo_goals_total",
		"echo_wisdom_total",
		"echo_dream_pending_experiences",
		"echo_experience_ledger_size",
	} {
		assert.Contains(t, metrics, name)
	}
}

func TestAutonomousServerHelpAndUnknownRouteE2E(t *testing.T) {
	response, err := autonomousClient().Get(autonomousBaseURL() + "/")
	require.NoError(t, err)
	body, readErr := io.ReadAll(response.Body)
	response.Body.Close()
	require.NoError(t, readErr)
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Contains(t, string(body), "GET /health")
	assert.Contains(t, string(body), "GET /status")
	assert.Contains(t, string(body), "GET /metrics")

	response, err = autonomousClient().Get(autonomousBaseURL() + "/nonexistent")
	require.NoError(t, err)
	response.Body.Close()
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}
