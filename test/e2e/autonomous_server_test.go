package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ServerConfig holds test server configuration
type ServerConfig struct {
	BaseURL     string
	HealthPath  string
	APIPath     string
	Timeout     time.Duration
}

// DefaultServerConfig returns default test configuration
func DefaultServerConfig() *ServerConfig {
	baseURL := os.Getenv("ECHOSELF_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	return &ServerConfig{
		BaseURL:    baseURL,
		HealthPath: "/health",
		APIPath:    "/api/v1",
		Timeout:    time.Second * 30,
	}
}

// TestAutonomousServerE2E tests the autonomous server end-to-end
func TestAutonomousServerE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := DefaultServerConfig()

	t.Run("HealthCheck", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}
		
		resp, err := client.Get(config.BaseURL + config.HealthPath)
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("StatusEndpoint", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}
		
		resp, err := client.Get(config.BaseURL + config.APIPath + "/status")
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var status map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&status)
		require.NoError(t, err)

		assert.Contains(t, status, "running")
		assert.Contains(t, status, "uptime")
	})

	t.Run("ThoughtGeneration", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}

		requestBody := map[string]interface{}{
			"prompt": "What is the meaning of consciousness?",
			"mode":   "thought",
			"max_tokens": 100,
		}
		
		body, _ := json.Marshal(requestBody)
		
		resp, err := client.Post(
			config.BaseURL + config.APIPath + "/generate",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Logf("Response body: %s", string(bodyBytes))
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "response")
		assert.Contains(t, response, "processing_time")
	})

	t.Run("GoalCreation", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}

		requestBody := map[string]interface{}{
			"prompt": "Learn about quantum computing",
			"mode":   "goal",
		}
		
		body, _ := json.Marshal(requestBody)
		
		resp, err := client.Post(
			config.BaseURL + config.APIPath + "/generate",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "goals")
	})

	t.Run("ConversationMode", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}

		requestBody := map[string]interface{}{
			"prompt": "Hello, how are you?",
			"mode":   "conversation",
			"max_tokens": 50,
		}
		
		body, _ := json.Marshal(requestBody)
		
		resp, err := client.Post(
			config.BaseURL + config.APIPath + "/generate",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "response")
	})

	t.Run("DreamCycle", func(t *testing.T) {
		client := &http.Client{Timeout: time.Minute} // Longer timeout for dream cycle

		requestBody := map[string]interface{}{
			"prompt": "Consolidate recent memories",
			"mode":   "dream",
		}
		
		body, _ := json.Marshal(requestBody)
		
		resp, err := client.Post(
			config.BaseURL + config.APIPath + "/generate",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "memories")
	})
}

// TestCognitiveLoopE2E tests the cognitive loop end-to-end
func TestCognitiveLoopE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := DefaultServerConfig()

	t.Run("CycleExecution", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}

		resp, err := client.Post(
			config.BaseURL + config.APIPath + "/cognitive/cycle",
			"application/json",
			nil,
		)
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "cycle_count")
		assert.Contains(t, response, "coherence")
	})

	t.Run("MetricsEndpoint", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}

		resp, err := client.Get(config.BaseURL + config.APIPath + "/cognitive/metrics")
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var metrics map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&metrics)
		require.NoError(t, err)

		assert.Contains(t, metrics, "running")
		assert.Contains(t, metrics, "cycle_count")
	})
}

// TestMemorySystemE2E tests the memory system end-to-end
func TestMemorySystemE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := DefaultServerConfig()

	t.Run("AddMemory", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}

		requestBody := map[string]interface{}{
			"content": "Test memory content for E2E testing",
			"type":    "episodic",
			"metadata": map[string]interface{}{
				"source": "e2e_test",
				"timestamp": time.Now().Unix(),
			},
		}
		
		body, _ := json.Marshal(requestBody)
		
		resp, err := client.Post(
			config.BaseURL + config.APIPath + "/memory/add",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "id")
	})

	t.Run("SearchMemory", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}

		requestBody := map[string]interface{}{
			"query": "E2E testing",
			"limit": 10,
		}
		
		body, _ := json.Marshal(requestBody)
		
		resp, err := client.Post(
			config.BaseURL + config.APIPath + "/memory/search",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "results")
	})
}

// TestWebSocketE2E tests WebSocket connections
func TestWebSocketE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// WebSocket tests would go here
	// Requires gorilla/websocket or similar
	t.Run("StreamOfConsciousnessConnection", func(t *testing.T) {
		t.Skip("WebSocket test not implemented")
	})
}

// TestErrorHandlingE2E tests error handling
func TestErrorHandlingE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := DefaultServerConfig()

	t.Run("InvalidEndpoint", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}

		resp, err := client.Get(config.BaseURL + "/nonexistent")
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}

		resp, err := client.Post(
			config.BaseURL + config.APIPath + "/generate",
			"application/json",
			bytes.NewReader([]byte("invalid json")),
		)
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		client := &http.Client{Timeout: config.Timeout}

		requestBody := map[string]interface{}{
			// Missing "prompt" field
			"mode": "thought",
		}
		
		body, _ := json.Marshal(requestBody)
		
		resp, err := client.Post(
			config.BaseURL + config.APIPath + "/generate",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			t.Skipf("Server not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// BenchmarkE2E benchmarks E2E operations
func BenchmarkE2E(b *testing.B) {
	config := DefaultServerConfig()
	client := &http.Client{Timeout: config.Timeout}

	// Check if server is available
	resp, err := client.Get(config.BaseURL + config.HealthPath)
	if err != nil {
		b.Skipf("Server not available: %v", err)
	}
	resp.Body.Close()

	b.Run("HealthCheck", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, _ := client.Get(config.BaseURL + config.HealthPath)
			if resp != nil {
				resp.Body.Close()
			}
		}
	})

	b.Run("StatusEndpoint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, _ := client.Get(config.BaseURL + config.APIPath + "/status")
			if resp != nil {
				resp.Body.Close()
			}
		}
	})
}
