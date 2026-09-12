package memory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SupabaseClient provides HTTP-based access to Supabase
type SupabaseClient struct {
	url    string
	key    string
	client *http.Client
}

// NewSupabaseClient creates a new Supabase client
func NewSupabaseClient(url, key string) *SupabaseClient {
	return &SupabaseClient{
		url: url,
		key: key,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Insert inserts a record into a table
func (sc *SupabaseClient) Insert(table string, record interface{}) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/%s", sc.url, table)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", sc.key)
	req.Header.Set("Authorization", "Bearer "+sc.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := sc.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("insert failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Query queries records from a table
func (sc *SupabaseClient) Query(table string, filter map[string]interface{}, limit int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/rest/v1/%s", sc.url, table)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	for key, value := range filter {
		q.Add(key, fmt.Sprintf("eq.%v", value))
	}
	if limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", limit))
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("apikey", sc.key)
	req.Header.Set("Authorization", "Bearer "+sc.key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return results, nil
}

// Update updates records in a table
func (sc *SupabaseClient) Update(table string, filter map[string]interface{}, updates map[string]interface{}) error {
	data, err := json.Marshal(updates)
	if err != nil {
		return fmt.Errorf("failed to marshal updates: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/%s", sc.url, table)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters for filtering
	q := req.URL.Query()
	for key, value := range filter {
		q.Add(key, fmt.Sprintf("eq.%v", value))
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("apikey", sc.key)
	req.Header.Set("Authorization", "Bearer "+sc.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := sc.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Delete deletes records from a table
func (sc *SupabaseClient) Delete(table string, filter map[string]interface{}) error {
	url := fmt.Sprintf("%s/rest/v1/%s", sc.url, table)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters for filtering
	q := req.URL.Query()
	for key, value := range filter {
		q.Add(key, fmt.Sprintf("eq.%v", value))
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("apikey", sc.key)
	req.Header.Set("Authorization", "Bearer "+sc.key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := sc.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ExecuteSQL executes a raw SQL query via RPC
func (sc *SupabaseClient) ExecuteSQL(query string, params map[string]interface{}) ([]map[string]interface{}, error) {
	// This would require a custom Supabase function
	// For now, we'll use the REST API for standard operations
	return nil, fmt.Errorf("raw SQL execution not implemented")
}

// Implement actual database operations for PersistentMemory

// StoreNodeActual stores a node using actual Supabase client
func (pm *PersistentMemory) StoreNodeActual(node *MemoryNode) error {
	client := NewSupabaseClient(pm.supabaseURL, pm.supabaseKey)
	return client.Insert("memory_nodes", node)
}

// StoreEdgeActual stores an edge using actual Supabase client
func (pm *PersistentMemory) StoreEdgeActual(edge *MemoryEdge) error {
	client := NewSupabaseClient(pm.supabaseURL, pm.supabaseKey)
	return client.Insert("memory_edges", edge)
}

// StoreEpisodeActual stores an episode using actual Supabase client
func (pm *PersistentMemory) StoreEpisodeActual(episode *Episode) error {
	client := NewSupabaseClient(pm.supabaseURL, pm.supabaseKey)
	return client.Insert("episodes", episode)
}

// StoreIdentitySnapshotActual stores an identity snapshot using actual Supabase client
func (pm *PersistentMemory) StoreIdentitySnapshotActual(snapshot *IdentitySnapshot) error {
	client := NewSupabaseClient(pm.supabaseURL, pm.supabaseKey)
	return client.Insert("identity_snapshots", snapshot)
}

// StoreDreamJournalActual stores a dream journal using actual Supabase client
func (pm *PersistentMemory) StoreDreamJournalActual(journal *DreamJournal) error {
	client := NewSupabaseClient(pm.supabaseURL, pm.supabaseKey)
	return client.Insert("dream_journals", journal)
}

// QueryNodesActual queries nodes using actual Supabase client
func (pm *PersistentMemory) QueryNodesActual(nodeType NodeType, limit int) ([]*MemoryNode, error) {
	client := NewSupabaseClient(pm.supabaseURL, pm.supabaseKey)

	filter := make(map[string]interface{})
	if nodeType != "" {
		filter["type"] = string(nodeType)
	}

	results, err := client.Query("memory_nodes", filter, limit)
	if err != nil {
		return nil, err
	}

	var nodes []*MemoryNode
	for _, result := range results {
		node := &MemoryNode{}
		// Convert map to struct
		data, _ := json.Marshal(result)
		json.Unmarshal(data, node)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// GetLatestIdentitySnapshotActual retrieves the most recent identity snapshot
func (pm *PersistentMemory) GetLatestIdentitySnapshotActual() (*IdentitySnapshot, error) {
	client := NewSupabaseClient(pm.supabaseURL, pm.supabaseKey)

	results, err := client.Query("identity_snapshots", map[string]interface{}{}, 1)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no identity snapshots found")
	}

	snapshot := &IdentitySnapshot{}
	data, _ := json.Marshal(results[0])
	json.Unmarshal(data, snapshot)

	return snapshot, nil
}
