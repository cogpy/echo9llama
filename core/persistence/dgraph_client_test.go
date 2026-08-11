package persistence

import (
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDgraphConfig tests the configuration handling
func TestDgraphConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		// Clear env var to test default
		originalEndpoint := os.Getenv("DGRAPH_ENDPOINT")
		os.Unsetenv("DGRAPH_ENDPOINT")
		defer os.Setenv("DGRAPH_ENDPOINT", originalEndpoint)

		config := DefaultDgraphConfig()
		require.NotNil(t, config)
		assert.Equal(t, "localhost:9080", config.Endpoint)
		assert.Equal(t, 3, config.RetryCount)
		assert.Equal(t, time.Second*2, config.RetryDelay)
	})

	t.Run("ConfigFromEnv", func(t *testing.T) {
		os.Setenv("DGRAPH_ENDPOINT", "dgraph-alpha:9080")
		defer os.Unsetenv("DGRAPH_ENDPOINT")

		config := DefaultDgraphConfig()
		assert.Equal(t, "dgraph-alpha:9080", config.Endpoint)
	})

	t.Run("CustomConfig", func(t *testing.T) {
		config := &DgraphConfig{
			Endpoint:   "custom:9080",
			RetryCount: 5,
			RetryDelay: time.Second * 5,
		}
		assert.Equal(t, "custom:9080", config.Endpoint)
		assert.Equal(t, 5, config.RetryCount)
		assert.Equal(t, time.Second*5, config.RetryDelay)
	})
}

func TestMutationRequiresExplicitCommit(t *testing.T) {
	assert.False(t, mutationRequiresExplicitCommit(nil))
	assert.True(t, mutationRequiresExplicitCommit(&api.Mutation{}))
	assert.False(t, mutationRequiresExplicitCommit(&api.Mutation{CommitNow: true}))
}

// TestMarshalJSON tests the JSON marshaling helper
func TestMarshalJSON(t *testing.T) {
	t.Run("MarshalStruct", func(t *testing.T) {
		data := struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}{
			Name:  "test",
			Value: 42,
		}

		result, err := MarshalJSON(data)
		require.NoError(t, err)
		assert.Contains(t, string(result), "test")
		assert.Contains(t, string(result), "42")
	})

	t.Run("MarshalMap", func(t *testing.T) {
		data := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}

		result, err := MarshalJSON(data)
		require.NoError(t, err)
		assert.Contains(t, string(result), "key1")
		assert.Contains(t, string(result), "value1")
	})
}

// TestUnmarshalJSON tests the JSON unmarshaling helper
func TestUnmarshalJSON(t *testing.T) {
	t.Run("UnmarshalStruct", func(t *testing.T) {
		jsonData := []byte(`{"name":"test","value":42}`)
		var result struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		err := UnmarshalJSON(jsonData, &result)
		require.NoError(t, err)
		assert.Equal(t, "test", result.Name)
		assert.Equal(t, 42, result.Value)
	})

	t.Run("UnmarshalMap", func(t *testing.T) {
		jsonData := []byte(`{"key1":"value1","key2":123}`)
		var result map[string]interface{}

		err := UnmarshalJSON(jsonData, &result)
		require.NoError(t, err)
		assert.Equal(t, "value1", result["key1"])
	})

	t.Run("UnmarshalInvalidJSON", func(t *testing.T) {
		jsonData := []byte(`invalid json`)
		var result map[string]interface{}

		err := UnmarshalJSON(jsonData, &result)
		assert.Error(t, err)
	})
}

// TestDgraphClientConnectionFailure tests connection failure handling
// Note: These tests don't require a running Dgraph instance
func TestDgraphClientConnectionFailure(t *testing.T) {
	t.Run("ConnectionFailure", func(t *testing.T) {
		config := &DgraphConfig{
			Endpoint:       "nonexistent:9080",
			RetryCount:     1,
			RetryDelay:     time.Millisecond * 100,
			ConnectTimeout: time.Millisecond * 200,
		}

		client, err := NewDgraphClient(config)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

// MockDgraphClient provides a mock implementation for testing
type MockDgraphClient struct {
	connected bool
	schema    string
	nodes     map[string]interface{}
	edges     map[string]interface{}
}

// NewMockDgraphClient creates a new mock client
func NewMockDgraphClient() *MockDgraphClient {
	return &MockDgraphClient{
		connected: true,
		nodes:     make(map[string]interface{}),
		edges:     make(map[string]interface{}),
	}
}

// IsConnected returns mock connection status
func (m *MockDgraphClient) IsConnected() bool {
	return m.connected
}

// SetSchema sets the mock schema
func (m *MockDgraphClient) SetSchema(schema string) error {
	m.schema = schema
	return nil
}

// TestMockDgraphClient tests the mock client
func TestMockDgraphClient(t *testing.T) {
	t.Run("MockClientOperations", func(t *testing.T) {
		mock := NewMockDgraphClient()
		assert.True(t, mock.IsConnected())

		err := mock.SetSchema("test schema")
		require.NoError(t, err)
		assert.Equal(t, "test schema", mock.schema)
	})
}
