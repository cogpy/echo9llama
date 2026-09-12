package deeptreeecho

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestFeatherlessClientCreation(t *testing.T) {
	tests := []struct {
		name        string
		config      FeatherlessConfig
		envVars     map[string]string
		expectError bool
	}{
		{
			name: "API key from config",
			config: FeatherlessConfig{
				APIKey: "test-key",
			},
			envVars:     nil,
			expectError: false,
		},
		{
			name:   "API key from FEATHERLESS_API_KEY env",
			config: FeatherlessConfig{},
			envVars: map[string]string{
				"FEATHERLESS_API_KEY": "test-env-key",
			},
			expectError: false,
		},
		{
			name:   "API key from FEARLESS env (typo support)",
			config: FeatherlessConfig{},
			envVars: map[string]string{
				"FEARLESS": "test-fearless-key",
			},
			expectError: false,
		},
		{
			name:        "No API key provided",
			config:      FeatherlessConfig{},
			envVars:     nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("FEATHERLESS_API_KEY")
			os.Unsetenv("FEARLESS")

			// Set test environment variables
			if tt.envVars != nil {
				for k, v := range tt.envVars {
					os.Setenv(k, v)
				}
			}

			// Create client
			client, err := NewFeatherlessClient(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client == nil {
					t.Errorf("Expected client but got nil")
				}
			}

			// Cleanup
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestFeatherlessClientDefaults(t *testing.T) {
	config := FeatherlessConfig{
		APIKey: "test-key",
	}

	client, err := NewFeatherlessClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client.baseURL != "https://api.featherless.ai/v1" {
		t.Errorf("Expected default baseURL, got %s", client.baseURL)
	}

	if client.model != "meta-llama/Meta-Llama-3.1-8B-Instruct" {
		t.Errorf("Expected default model, got %s", client.model)
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected 30s timeout, got %v", client.httpClient.Timeout)
	}
}

func TestFeatherlessClientCustomConfig(t *testing.T) {
	config := FeatherlessConfig{
		APIKey:  "test-key",
		BaseURL: "https://custom.api.com/v1",
		Model:   "custom-model",
		Timeout: 60 * time.Second,
	}

	client, err := NewFeatherlessClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client.baseURL != config.BaseURL {
		t.Errorf("Expected custom baseURL %s, got %s", config.BaseURL, client.baseURL)
	}

	if client.model != config.Model {
		t.Errorf("Expected custom model %s, got %s", config.Model, client.model)
	}

	if client.httpClient.Timeout != config.Timeout {
		t.Errorf("Expected custom timeout %v, got %v", config.Timeout, client.httpClient.Timeout)
	}
}

func TestGenerateThoughtStructure(t *testing.T) {
	// This test verifies the method signatures are correct
	config := FeatherlessConfig{
		APIKey: "test-key",
	}

	client, err := NewFeatherlessClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify methods exist and have correct signatures
	ctx := context.Background()

	// Test GenerateThought (will fail with API error, but that's expected)
	_, err = client.GenerateThought(ctx, "test prompt", "test system")
	// We expect an error since we're using a fake key
	if err == nil {
		t.Log("Warning: No error with fake API key - API may not be validating")
	}

	// Test GenerateThoughtStream
	contentChan, errorChan := client.GenerateThoughtStream(ctx, "test prompt", "test system")
	if contentChan == nil || errorChan == nil {
		t.Error("GenerateThoughtStream should return non-nil channels")
	}
}

func TestChatCompletionMessages(t *testing.T) {
	messages := []FeatherlessChatMessage{
		{
			Role:    "system",
			Content: "You are a helpful assistant",
		},
		{
			Role:    "user",
			Content: "Hello",
		},
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("Expected system role, got %s", messages[0].Role)
	}

	if messages[1].Role != "user" {
		t.Errorf("Expected user role, got %s", messages[1].Role)
	}
}

// TestFEARLESS_KeyPriorityOrder verifies the correct precedence of API key sources
func TestFEARLESS_KeyPriorityOrder(t *testing.T) {
	// Setup: config > FEATHERLESS_API_KEY > FEARLESS

	t.Run("Config takes precedence over environment", func(t *testing.T) {
		os.Setenv("FEATHERLESS_API_KEY", "env-key")
		os.Setenv("FEARLESS", "fearless-key")
		defer os.Unsetenv("FEATHERLESS_API_KEY")
		defer os.Unsetenv("FEARLESS")

		config := FeatherlessConfig{
			APIKey: "config-key",
		}

		client, err := NewFeatherlessClient(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		if client.apiKey != "config-key" {
			t.Errorf("Expected config key to take precedence, got %s", client.apiKey)
		}
	})

	t.Run("FEATHERLESS_API_KEY takes precedence over FEARLESS", func(t *testing.T) {
		os.Setenv("FEATHERLESS_API_KEY", "featherless-key")
		os.Setenv("FEARLESS", "fearless-key")
		defer os.Unsetenv("FEATHERLESS_API_KEY")
		defer os.Unsetenv("FEARLESS")

		config := FeatherlessConfig{}

		client, err := NewFeatherlessClient(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		if client.apiKey != "featherless-key" {
			t.Errorf("Expected FEATHERLESS_API_KEY to take precedence, got %s", client.apiKey)
		}
	})

	t.Run("FEARLESS used when FEATHERLESS_API_KEY not set", func(t *testing.T) {
		os.Unsetenv("FEATHERLESS_API_KEY")
		os.Setenv("FEARLESS", "fearless-key")
		defer os.Unsetenv("FEARLESS")

		config := FeatherlessConfig{}

		client, err := NewFeatherlessClient(config)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		if client.apiKey != "fearless-key" {
			t.Errorf("Expected FEARLESS key to be used, got %s", client.apiKey)
		}
	})
}
