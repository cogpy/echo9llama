//go:build cgo

package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func newTestCognitiveEventStore(t *testing.T) (*CognitiveEventStore, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "private", "events.db")
	store, err := NewCognitiveEventStore(path)
	if err != nil {
		t.Fatalf("NewCognitiveEventStore(%q): %v", path, err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close(): %v", err)
		}
	})
	return store, path
}

func testCognitiveEvent(id string, occurredAt time.Time) CognitiveEvent {
	return CognitiveEvent{
		SchemaVersion: CognitiveEventSchemaVersion,
		EventID:       id,
		EventType:     EventTypeActionCompleted,
		OccurredAt:    occurredAt.UTC(),
		IdentityID:    "identity-test",
		SessionID:     "session-test",
		CorrelationID: "correlation-test",
		GoalID:        "goal-test",
		ActionID:      "action-test",
		EvidenceClass: EvidenceClassOperational,
		PayloadJSON:   json.RawMessage(`{"result":"ok","value":1}`),
		Provider:      "provider-test",
		BackendKind:   "backend-test",
		ModelID:       "model-test",
		RouteReason:   "policy-match",
		PolicyVersion: "policy-v1",
	}
}

func TestCognitiveEventStoreMigrationIsIdempotent(t *testing.T) {
	store, path := newTestCognitiveEventStore(t)
	if err := store.migrate(); err != nil {
		t.Fatalf("first explicit migrate(): %v", err)
	}
	if err := store.migrate(); err != nil {
		t.Fatalf("second explicit migrate(): %v", err)
	}
	var migrations int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM cognitive_event_schema_migrations WHERE version = ?`, cognitiveEventSchemaMigrationVersion).Scan(&migrations); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if migrations != 1 {
		t.Fatalf("migration applied %d times, want 1", migrations)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close before reopen: %v", err)
	}
	reopened, err := NewCognitiveEventStore(path)
	if err != nil {
		t.Fatalf("reopen migrated store: %v", err)
	}
	defer func() { _ = reopened.Close() }()
	if err := reopened.db.QueryRow(`SELECT COUNT(*) FROM cognitive_event_schema_migrations WHERE version = ?`, cognitiveEventSchemaMigrationVersion).Scan(&migrations); err != nil {
		t.Fatalf("count migrations after reopen: %v", err)
	}
	if migrations != 1 {
		t.Fatalf("migration applied %d times after reopen, want 1", migrations)
	}
}

func TestCognitiveEventStoreStableOrderingReplayAndCount(t *testing.T) {
	store, _ := newTestCognitiveEventStore(t)
	ctx := context.Background()
	base := time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC)
	ids := []string{"event-3", "event-1", "event-2"}
	for index, id := range ids {
		event := testCognitiveEvent(id, base.Add(time.Duration(2-index)*time.Hour))
		if _, err := store.Append(ctx, event); err != nil {
			t.Fatalf("append %s: %v", id, err)
		}
	}

	events, err := store.Replay(ctx, 0, 10)
	if err != nil {
		t.Fatalf("Replay(): %v", err)
	}
	if len(events) != len(ids) {
		t.Fatalf("Replay length = %d, want %d", len(events), len(ids))
	}
	for i, event := range events {
		if event.EventID != ids[i] {
			t.Fatalf("replay[%d].EventID = %q, want insertion-order %q", i, event.EventID, ids[i])
		}
		if event.Sequence != int64(i+1) {
			t.Fatalf("replay[%d].Sequence = %d, want %d", i, event.Sequence, i+1)
		}
	}
	next, err := store.Replay(ctx, events[0].Sequence, 10)
	if err != nil {
		t.Fatalf("Replay after cursor: %v", err)
	}
	if len(next) != 2 || next[0].EventID != "event-1" {
		t.Fatalf("Replay after cursor = %#v, want event-1 then event-2", next)
	}
	count, err := store.Count(ctx, CognitiveEventQuery{SessionID: "session-test"})
	if err != nil {
		t.Fatalf("Count(): %v", err)
	}
	if count != 3 {
		t.Fatalf("Count() = %d, want 3", count)
	}
	health, err := store.Health(ctx)
	if err != nil {
		t.Fatalf("Health(): %v", err)
	}
	if !health.Ready || health.EventCount != 3 || !health.WAL || !health.ForeignKeys || !health.SynchronousFull || health.BusyTimeoutMS < sqliteBusyTimeoutMilliseconds {
		t.Fatalf("unexpected health: %#v", health)
	}
}

func TestCognitiveEventStoreDuplicateAndIdempotencySemantics(t *testing.T) {
	store, _ := newTestCognitiveEventStore(t)
	ctx := context.Background()
	event := testCognitiveEvent("event-idempotent", time.Now())
	event.IdempotencyKey = "idem-1"
	first, err := store.Append(ctx, event)
	if err != nil {
		t.Fatalf("first Append(): %v", err)
	}
	if first.Duplicate || first.Sequence == 0 {
		t.Fatalf("first append result = %#v, want new sequence", first)
	}

	duplicate, err := store.Append(ctx, event)
	if err != nil {
		t.Fatalf("duplicate Append(): %v", err)
	}
	if !duplicate.Duplicate || duplicate.Sequence != first.Sequence {
		t.Fatalf("duplicate append result = %#v, want original sequence %#v", duplicate, first)
	}
	count, err := store.Count(ctx, CognitiveEventQuery{})
	if err != nil || count != 1 {
		t.Fatalf("count after duplicate = %d, %v; want 1, nil", count, err)
	}

	conflictingID := event
	conflictingID.PayloadJSON = json.RawMessage(`{"result":"different"}`)
	if _, err := store.Append(ctx, conflictingID); !errors.Is(err, ErrCognitiveEventConflict) {
		t.Fatalf("same event ID conflict error = %v, want ErrCognitiveEventConflict", err)
	}
	conflictingKey := testCognitiveEvent("event-other", event.OccurredAt)
	conflictingKey.IdempotencyKey = event.IdempotencyKey
	if _, err := store.Append(ctx, conflictingKey); !errors.Is(err, ErrCognitiveEventConflict) {
		t.Fatalf("same idempotency key conflict error = %v, want ErrCognitiveEventConflict", err)
	}
}

func TestCognitiveEventStoreRejectsUpdateAndDelete(t *testing.T) {
	store, _ := newTestCognitiveEventStore(t)
	ctx := context.Background()
	result, err := store.Append(ctx, testCognitiveEvent("event-immutable", time.Now()))
	if err != nil {
		t.Fatalf("Append(): %v", err)
	}
	if _, err := store.db.Exec(`UPDATE cognitive_events SET event_type = ? WHERE sequence = ?`, EventTypeActionFailed, result.Sequence); err == nil {
		t.Fatal("UPDATE cognitive_events unexpectedly succeeded")
	}
	if _, err := store.db.Exec(`DELETE FROM cognitive_events WHERE sequence = ?`, result.Sequence); err == nil {
		t.Fatal("DELETE cognitive_events unexpectedly succeeded")
	}
	count, err := store.Count(ctx, CognitiveEventQuery{})
	if err != nil || count != 1 {
		t.Fatalf("count after rejected mutations = %d, %v; want 1, nil", count, err)
	}
}

func TestCognitiveEventValidationAndCanonicalHash(t *testing.T) {
	valid := testCognitiveEvent("event-valid", time.Date(2026, 9, 11, 8, 0, 0, 123, time.FixedZone("offset", 7200)))
	hashed, err := valid.WithCanonicalHash()
	if err != nil {
		t.Fatalf("WithCanonicalHash(): %v", err)
	}
	if len(hashed.ContentSHA256) != 64 {
		t.Fatalf("hash length = %d, want 64", len(hashed.ContentSHA256))
	}
	if err := hashed.Validate(); err != nil {
		t.Fatalf("hashed event Validate(): %v", err)
	}
	permuted := valid
	permuted.PayloadJSON = json.RawMessage(` { "value" : 1, "result" : "ok" } `)
	permutedHash, err := permuted.CanonicalHash()
	if err != nil {
		t.Fatalf("permuted CanonicalHash(): %v", err)
	}
	if permutedHash != hashed.ContentSHA256 {
		t.Fatalf("canonical JSON hashes differ: %s != %s", permutedHash, hashed.ContentSHA256)
	}

	cases := []struct {
		name   string
		mutate func(*CognitiveEvent)
		err    error
	}{
		{"missing event ID", func(e *CognitiveEvent) { e.EventID = "" }, ErrInvalidCognitiveEvent},
		{"unsupported schema", func(e *CognitiveEvent) { e.SchemaVersion = 99 }, ErrUnsupportedCognitiveEventSchemaVersion},
		{"unsupported type", func(e *CognitiveEvent) { e.EventType = "future.unknown" }, ErrInvalidCognitiveEventType},
		{"unsupported evidence", func(e *CognitiveEvent) { e.EvidenceClass = "secret" }, ErrInvalidEvidenceClass},
		{"bad JSON", func(e *CognitiveEvent) { e.PayloadJSON = json.RawMessage(`{"x":`) }, ErrInvalidCognitiveEventPayload},
		{"non-object JSON", func(e *CognitiveEvent) { e.PayloadJSON = json.RawMessage(`[1]`) }, ErrInvalidCognitiveEventPayload},
		{"hidden thought JSON", func(e *CognitiveEvent) { e.PayloadJSON = json.RawMessage(`{"chain_of_thought":"private"}`) }, ErrInvalidCognitiveEventPayload},
		{"bad supplied hash", func(e *CognitiveEvent) { e.ContentSHA256 = "wrong" }, ErrInvalidCognitiveEvent},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			event := valid
			test.mutate(&event)
			if err := event.Validate(); !errors.Is(err, test.err) {
				t.Fatalf("Validate() error = %v, want errors.Is(_, %v)", err, test.err)
			}
		})
	}
}

func TestCognitiveEventStoreConcurrentAppend(t *testing.T) {
	store, _ := newTestCognitiveEventStore(t)
	ctx := context.Background()
	const writers = 24
	var group sync.WaitGroup
	errs := make(chan error, writers)
	for i := 0; i < writers; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			event := testCognitiveEvent(fmt.Sprintf("event-concurrent-%02d", i), time.Now().Add(time.Duration(i)*time.Nanosecond))
			result, err := store.Append(ctx, event)
			if err == nil && result.Duplicate {
				err = fmt.Errorf("concurrent append %d unexpectedly duplicated", i)
			}
			errs <- err
		}(i)
	}
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	events, err := store.Replay(ctx, 0, 100)
	if err != nil {
		t.Fatalf("Replay(): %v", err)
	}
	if len(events) != writers {
		t.Fatalf("replayed %d concurrent events, want %d", len(events), writers)
	}
	for i, event := range events {
		if event.Sequence != int64(i+1) {
			t.Fatalf("sequence[%d] = %d, want %d", i, event.Sequence, i+1)
		}
	}
}

func TestSQLiteOwnerOnlyPermissionsAndReopenReplay(t *testing.T) {
	store, path := newTestCognitiveEventStore(t)
	parentInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat parent: %v", err)
	}
	if parentInfo.Mode().Perm() != 0o700 {
		t.Fatalf("parent mode = %04o, want 0700", parentInfo.Mode().Perm())
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat database: %v", err)
	}
	if fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf("database mode = %04o, want 0600", fileInfo.Mode().Perm())
	}
	result, err := store.Append(context.Background(), testCognitiveEvent("event-reopen", time.Now()))
	if err != nil {
		t.Fatalf("Append(): %v", err)
	}
	for _, sidecar := range []string{path + "-wal", path + "-shm"} {
		if info, err := os.Stat(sidecar); err == nil && info.Mode().Perm() != 0o600 {
			t.Fatalf("sidecar %s mode = %04o, want 0600", sidecar, info.Mode().Perm())
		} else if err != nil && !os.IsNotExist(err) {
			t.Fatalf("stat sidecar %s: %v", sidecar, err)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	reopened, err := NewCognitiveEventStore(path)
	if err != nil {
		t.Fatalf("NewCognitiveEventStore after close: %v", err)
	}
	defer func() { _ = reopened.Close() }()
	events, err := reopened.Replay(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("Replay after reopen: %v", err)
	}
	if len(events) != 1 || events[0].EventID != "event-reopen" || events[0].Sequence != result.Sequence {
		t.Fatalf("replay after reopen = %#v, want saved event sequence %d", events, result.Sequence)
	}
	cursor, err := reopened.AdvanceProjectionCursor(context.Background(), "projection-test", result.Sequence)
	if err != nil {
		t.Fatalf("AdvanceProjectionCursor(): %v", err)
	}
	if cursor.Sequence != result.Sequence {
		t.Fatalf("cursor sequence = %d, want %d", cursor.Sequence, result.Sequence)
	}
	if _, err := reopened.AdvanceProjectionCursor(context.Background(), "projection-test", result.Sequence-1); !errors.Is(err, ErrProjectionCursorRegression) {
		t.Fatalf("cursor regression error = %v, want ErrProjectionCursorRegression", err)
	}
}
