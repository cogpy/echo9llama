package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DgraphClient manages connections to Dgraph for persistent hypergraph storage
type DgraphClient struct {
	mu             sync.RWMutex
	conn           *grpc.ClientConn
	client         *dgo.Dgraph
	ctx            context.Context
	cancel         context.CancelFunc
	endpoint       string
	connected      bool
	retryCount     int
	retryDelay     time.Duration
	connectTimeout time.Duration
}

// DgraphConfig holds configuration for Dgraph connection
type DgraphConfig struct {
	Endpoint       string
	RetryCount     int
	RetryDelay     time.Duration
	ConnectTimeout time.Duration
}

// DefaultDgraphConfig returns default configuration
func DefaultDgraphConfig() *DgraphConfig {
	endpoint := os.Getenv("DGRAPH_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9080"
	}
	return &DgraphConfig{
		Endpoint:       endpoint,
		RetryCount:     3,
		RetryDelay:     time.Second * 2,
		ConnectTimeout: time.Second * 5,
	}
}

// NewDgraphClient creates a new Dgraph client
func NewDgraphClient(config *DgraphConfig) (*DgraphClient, error) {
	if config == nil {
		config = DefaultDgraphConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	if config.RetryCount <= 0 {
		config.RetryCount = 1
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = time.Second
	}
	if config.ConnectTimeout <= 0 {
		config.ConnectTimeout = time.Second * 5
	}

	client := &DgraphClient{
		ctx:            ctx,
		cancel:         cancel,
		endpoint:       config.Endpoint,
		retryCount:     config.RetryCount,
		retryDelay:     config.RetryDelay,
		connectTimeout: config.ConnectTimeout,
	}

	if err := client.connect(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Dgraph: %w", err)
	}

	return client, nil
}

// connect establishes connection to Dgraph
func (dc *DgraphClient) connect() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	var lastErr error
	for i := 0; i < dc.retryCount; i++ {
		dialCtx, cancel := context.WithTimeout(dc.ctx, dc.connectTimeout)
		conn, err := grpc.DialContext(
			dialCtx,
			dc.endpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		cancel()
		if err != nil {
			lastErr = err
			if i < dc.retryCount-1 {
				time.Sleep(dc.retryDelay)
			}
			continue
		}

		dc.conn = conn
		dc.client = dgo.NewDgraphClient(api.NewDgraphClient(conn))
		dc.connected = true
		return nil
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", dc.retryCount, lastErr)
}

// Close closes the Dgraph connection
func (dc *DgraphClient) Close() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.cancel()
	if dc.conn != nil {
		return dc.conn.Close()
	}
	return nil
}

// IsConnected returns connection status
func (dc *DgraphClient) IsConnected() bool {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.connected
}

// SetSchema sets the Dgraph schema
func (dc *DgraphClient) SetSchema(schema string) error {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if !dc.connected {
		return fmt.Errorf("not connected to Dgraph")
	}

	op := &api.Operation{Schema: schema}
	return dc.client.Alter(dc.ctx, op)
}

// NewTransaction creates a new read-write transaction
func (dc *DgraphClient) NewTransaction() *dgo.Txn {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.client.NewTxn()
}

// NewReadOnlyTransaction creates a new read-only transaction
func (dc *DgraphClient) NewReadOnlyTransaction() *dgo.Txn {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.client.NewReadOnlyTxn()
}

func mutationRequiresExplicitCommit(mu *api.Mutation) bool {
	return mu != nil && !mu.CommitNow
}

// Mutate performs a mutation operation
func (dc *DgraphClient) Mutate(ctx context.Context, mu *api.Mutation) (*api.Response, error) {
	if mu == nil {
		return nil, fmt.Errorf("mutation is required")
	}

	txn := dc.NewTransaction()
	defer txn.Discard(ctx)

	resp, err := txn.Mutate(ctx, mu)
	if err != nil {
		return nil, err
	}

	// Dgraph finalizes the transaction inside Mutate when CommitNow is set.
	// Committing that transaction again returns ErrFinished and makes a
	// successful mutation appear to have failed.
	if !mutationRequiresExplicitCommit(mu) {
		return resp, nil
	}
	if err := txn.Commit(ctx); err != nil {
		return nil, err
	}

	return resp, nil
}

// Query performs a query operation
func (dc *DgraphClient) Query(ctx context.Context, query string, vars map[string]string) (*api.Response, error) {
	txn := dc.NewReadOnlyTransaction()
	defer txn.Discard(ctx)

	if vars != nil {
		return txn.QueryWithVars(ctx, query, vars)
	}
	return txn.Query(ctx, query)
}

// Upsert performs an upsert operation (query + mutation in single transaction)
func (dc *DgraphClient) Upsert(ctx context.Context, query string, mu *api.Mutation) (*api.Response, error) {
	txn := dc.NewTransaction()
	defer txn.Discard(ctx)

	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	return txn.Do(ctx, req)
}

// DropAll drops all data from Dgraph (use with caution)
func (dc *DgraphClient) DropAll(ctx context.Context) error {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if !dc.connected {
		return fmt.Errorf("not connected to Dgraph")
	}

	return dc.client.Alter(ctx, &api.Operation{DropAll: true})
}

// DropData drops all data but keeps schema
func (dc *DgraphClient) DropData(ctx context.Context) error {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if !dc.connected {
		return fmt.Errorf("not connected to Dgraph")
	}

	return dc.client.Alter(ctx, &api.Operation{DropOp: api.Operation_DATA})
}

// MarshalJSON helper for mutations
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalJSON helper for query results
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
