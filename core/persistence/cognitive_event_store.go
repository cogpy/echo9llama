package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const cognitiveEventSchemaMigrationVersion = 1

// CognitiveEventStore extends SQLiteStore with an append-only, durable event
// ledger. The embedded SQLiteStore remains available for lifecycle management;
// this type intentionally exposes no event update or delete operation.
type CognitiveEventStore struct {
	*SQLiteStore
}

// StoredCognitiveEvent is a persisted envelope together with its monotonically
// increasing ledger sequence, which is the canonical replay order.
type StoredCognitiveEvent struct {
	Sequence int64 `json:"sequence"`
	CognitiveEvent
}

// CognitiveEventQuery limits an ordered event query. Zero-value filters are
// omitted. Ordering is always ascending sequence and cannot be caller-defined.
type CognitiveEventQuery struct {
	AfterSequence int64
	Limit         int
	IdentityID    string
	SessionID     string
	CorrelationID string
	GoalID        string
	ActionID      string
	EventType     string
	EvidenceClass string
}

// ProjectionCursor records a projection's durable high-water mark. Projection
// cursors are mutable; cognitive_events are not.
type ProjectionCursor struct {
	ProjectionName string    `json:"projection_name"`
	Sequence       int64     `json:"sequence"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CognitiveEventHealth describes the event ledger's readiness and durable
// configuration without returning event payloads.
type CognitiveEventHealth struct {
	Ready           bool  `json:"ready"`
	EventCount      int64 `json:"event_count"`
	SchemaVersion   int   `json:"schema_version"`
	ForeignKeys     bool  `json:"foreign_keys"`
	WAL             bool  `json:"wal"`
	SynchronousFull bool  `json:"synchronous_full"`
	BusyTimeoutMS   int   `json:"busy_timeout_ms"`
}

// NewCognitiveEventStore creates a hardened SQLite store and applies the event
// ledger migration transactionally.
func NewCognitiveEventStore(dbPath string) (*CognitiveEventStore, error) {
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		return nil, err
	}
	eventStore, err := NewCognitiveEventStoreFromSQLite(store)
	if err != nil {
		_ = store.Close()
		return nil, err
	}
	return eventStore, nil
}

// NewCognitiveEventStoreFromSQLite adds the event ledger to an already-open
// SQLiteStore. Ownership of the supplied store remains with the caller.
func NewCognitiveEventStoreFromSQLite(store *SQLiteStore) (*CognitiveEventStore, error) {
	if store == nil {
		return nil, fmt.Errorf("SQLite store is required")
	}
	eventStore := &CognitiveEventStore{SQLiteStore: store}
	if err := eventStore.migrate(); err != nil {
		return nil, err
	}
	return eventStore, nil
}

// NewCognitiveEventStoreWithStore is an explicit alias for integration code.
func NewCognitiveEventStoreWithStore(store *SQLiteStore) (*CognitiveEventStore, error) {
	return NewCognitiveEventStoreFromSQLite(store)
}

func (s *CognitiveEventStore) migrate() error {
	if s == nil || s.SQLiteStore == nil {
		return fmt.Errorf("cognitive event store is not initialized")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isOpen || s.db == nil {
		return fmt.Errorf("database not open")
	}

	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("begin cognitive event migration: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS cognitive_event_schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		)`); err != nil {
		return fmt.Errorf("create cognitive event migration ledger: %w", err)
	}

	var applied int
	err = tx.QueryRow(`SELECT COUNT(*) FROM cognitive_event_schema_migrations WHERE version = ?`, cognitiveEventSchemaMigrationVersion).Scan(&applied)
	if err != nil {
		return fmt.Errorf("read cognitive event migration ledger: %w", err)
	}
	if applied == 0 {
		if err := applyCognitiveEventSchemaV1(tx); err != nil {
			return err
		}
		if _, err := tx.Exec(
			`INSERT INTO cognitive_event_schema_migrations (version, applied_at) VALUES (?, ?)`,
			cognitiveEventSchemaMigrationVersion,
			time.Now().UTC().Format(time.RFC3339Nano),
		); err != nil {
			return fmt.Errorf("record cognitive event migration: %w", err)
		}
	} else if applied != 1 {
		return fmt.Errorf("invalid cognitive event migration ledger state for version %d", cognitiveEventSchemaMigrationVersion)
	}

	// These idempotent statements verify append-only protections also exist when
	// opening a database created by an earlier binary with the same migration.
	if err := ensureCognitiveEventTriggers(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit cognitive event migration: %w", err)
	}
	return s.secureSQLiteFiles()
}

func applyCognitiveEventSchemaV1(tx *sql.Tx) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS cognitive_events (
			sequence INTEGER PRIMARY KEY AUTOINCREMENT,
			schema_version INTEGER NOT NULL,
			event_id TEXT NOT NULL UNIQUE,
			event_type TEXT NOT NULL,
			occurred_at TEXT NOT NULL,
			identity_id TEXT NOT NULL,
			session_id TEXT NOT NULL,
			correlation_id TEXT NOT NULL DEFAULT '',
			causation_id TEXT NOT NULL DEFAULT '',
			goal_id TEXT NOT NULL DEFAULT '',
			action_id TEXT NOT NULL DEFAULT '',
			idempotency_key TEXT NOT NULL DEFAULT '',
			evidence_class TEXT NOT NULL,
			payload_json TEXT NOT NULL,
			provider TEXT NOT NULL DEFAULT '',
			backend_kind TEXT NOT NULL DEFAULT '',
			model_id TEXT NOT NULL DEFAULT '',
			route_reason TEXT NOT NULL DEFAULT '',
			degraded INTEGER NOT NULL CHECK (degraded IN (0, 1)),
			policy_version TEXT NOT NULL,
			content_sha256 TEXT NOT NULL CHECK (length(content_sha256) = 64),
			persisted_at TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_cognitive_events_idempotency_key_material
			ON cognitive_events(idempotency_key) WHERE idempotency_key <> ''`,
		`CREATE INDEX IF NOT EXISTS idx_cognitive_events_replay
			ON cognitive_events(sequence ASC)`,
		`CREATE INDEX IF NOT EXISTS idx_cognitive_events_identity_sequence
			ON cognitive_events(identity_id, sequence ASC)`,
		`CREATE INDEX IF NOT EXISTS idx_cognitive_events_session_sequence
			ON cognitive_events(session_id, sequence ASC)`,
		`CREATE INDEX IF NOT EXISTS idx_cognitive_events_type_sequence
			ON cognitive_events(event_type, sequence ASC)`,
		`CREATE TABLE IF NOT EXISTS cognitive_projection_cursors (
			projection_name TEXT PRIMARY KEY,
			sequence INTEGER NOT NULL CHECK (sequence >= 0),
			updated_at TEXT NOT NULL
		)`,
	}
	for _, statement := range statements {
		if _, err := tx.Exec(statement); err != nil {
			return fmt.Errorf("apply cognitive event schema v1: %w", err)
		}
	}
	return ensureCognitiveEventTriggers(tx)
}

func ensureCognitiveEventTriggers(tx *sql.Tx) error {
	for _, statement := range []string{
		`CREATE TRIGGER IF NOT EXISTS cognitive_events_reject_update
			BEFORE UPDATE ON cognitive_events
			BEGIN
				SELECT RAISE(ABORT, 'cognitive_events is append-only: updates are forbidden');
			END`,
		`CREATE TRIGGER IF NOT EXISTS cognitive_events_reject_delete
			BEFORE DELETE ON cognitive_events
			BEGIN
				SELECT RAISE(ABORT, 'cognitive_events is append-only: deletes are forbidden');
			END`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			return fmt.Errorf("install cognitive event append-only trigger: %w", err)
		}
	}
	return nil
}

// Append validates, canonicalizes, and atomically appends an event. Retrying
// the exact immutable event returns Duplicate=true and its original sequence;
// reusing an ID or material idempotency key for different material is an error.
func (s *CognitiveEventStore) Append(ctx context.Context, event CognitiveEvent) (DuplicateCognitiveEventResult, error) {
	ctx = nonNilContext(ctx)
	normalized, err := normalizeCognitiveEvent(event)
	if err != nil {
		return DuplicateCognitiveEventResult{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.isOpen || s.db == nil {
		return DuplicateCognitiveEventResult{}, fmt.Errorf("database not open")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DuplicateCognitiveEventResult{}, fmt.Errorf("begin cognitive event append: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if result, found, err := findEquivalentOrConflict(ctx, tx, normalized); err != nil {
		return DuplicateCognitiveEventResult{}, err
	} else if found {
		if err := tx.Commit(); err != nil {
			return DuplicateCognitiveEventResult{}, fmt.Errorf("commit duplicate cognitive event append: %w", err)
		}
		return result, nil
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO cognitive_events (
			schema_version, event_id, event_type, occurred_at, identity_id, session_id,
			correlation_id, causation_id, goal_id, action_id, idempotency_key,
			evidence_class, payload_json, provider, backend_kind, model_id,
			route_reason, degraded, policy_version, content_sha256, persisted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		normalized.SchemaVersion,
		normalized.EventID,
		normalized.EventType,
		normalized.OccurredAt.UTC().Format(time.RFC3339Nano),
		normalized.IdentityID,
		normalized.SessionID,
		normalized.CorrelationID,
		normalized.CausationID,
		normalized.GoalID,
		normalized.ActionID,
		normalized.IdempotencyKey,
		normalized.EvidenceClass,
		string(normalized.PayloadJSON),
		normalized.Provider,
		normalized.BackendKind,
		normalized.ModelID,
		normalized.RouteReason,
		boolToSQLite(normalized.Degraded),
		normalized.PolicyVersion,
		normalized.ContentSHA256,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		// A unique violation can occur only if a different process bypasses the
		// single local connection. Re-check to retain the typed contract.
		if isSQLiteConstraint(err) {
			if duplicate, found, lookupErr := findEquivalentOrConflict(ctx, tx, normalized); lookupErr != nil {
				return DuplicateCognitiveEventResult{}, lookupErr
			} else if found {
				if commitErr := tx.Commit(); commitErr != nil {
					return DuplicateCognitiveEventResult{}, fmt.Errorf("commit duplicate cognitive event append: %w", commitErr)
				}
				return duplicate, nil
			}
		}
		return DuplicateCognitiveEventResult{}, fmt.Errorf("append cognitive event: %w", err)
	}
	sequence, err := result.LastInsertId()
	if err != nil {
		return DuplicateCognitiveEventResult{}, fmt.Errorf("read cognitive event sequence: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return DuplicateCognitiveEventResult{}, fmt.Errorf("commit cognitive event append: %w", err)
	}
	if err := s.secureSQLiteFiles(); err != nil {
		return DuplicateCognitiveEventResult{}, err
	}
	return DuplicateCognitiveEventResult{Sequence: sequence, Duplicate: false}, nil
}

// AppendCognitiveEvent is a descriptive alias for Append.
func (s *CognitiveEventStore) AppendCognitiveEvent(ctx context.Context, event CognitiveEvent) (DuplicateCognitiveEventResult, error) {
	return s.Append(ctx, event)
}

// AppendEvent is a concise alias for Append.
func (s *CognitiveEventStore) AppendEvent(ctx context.Context, event CognitiveEvent) (DuplicateCognitiveEventResult, error) {
	return s.Append(ctx, event)
}

func normalizeCognitiveEvent(event CognitiveEvent) (CognitiveEvent, error) {
	if err := event.Validate(); err != nil {
		return CognitiveEvent{}, err
	}
	payload, err := canonicalPayloadJSON(event.PayloadJSON)
	if err != nil {
		return CognitiveEvent{}, err
	}
	event.PayloadJSON = payload
	event.OccurredAt = event.OccurredAt.UTC()
	hash, err := event.CanonicalHash()
	if err != nil {
		return CognitiveEvent{}, err
	}
	if event.ContentSHA256 != "" && event.ContentSHA256 != hash {
		return CognitiveEvent{}, fmt.Errorf("%w: content_sha256 does not match canonical event", ErrInvalidCognitiveEvent)
	}
	event.ContentSHA256 = hash
	return event, nil
}

func findEquivalentOrConflict(ctx context.Context, tx *sql.Tx, incoming CognitiveEvent) (DuplicateCognitiveEventResult, bool, error) {
	if stored, found, err := findStoredEventBy(ctx, tx, "event_id", incoming.EventID); err != nil {
		return DuplicateCognitiveEventResult{}, false, err
	} else if found {
		if stored.ContentSHA256 == incoming.ContentSHA256 {
			return DuplicateCognitiveEventResult{Sequence: stored.Sequence, Duplicate: true}, true, nil
		}
		return DuplicateCognitiveEventResult{}, false, &CognitiveEventConflictError{Kind: "event_id", Key: incoming.EventID}
	}
	if incoming.IdempotencyKey != "" {
		if stored, found, err := findStoredEventBy(ctx, tx, "idempotency_key", incoming.IdempotencyKey); err != nil {
			return DuplicateCognitiveEventResult{}, false, err
		} else if found {
			if stored.ContentSHA256 == incoming.ContentSHA256 {
				return DuplicateCognitiveEventResult{Sequence: stored.Sequence, Duplicate: true}, true, nil
			}
			return DuplicateCognitiveEventResult{}, false, &CognitiveEventConflictError{Kind: "idempotency_key", Key: incoming.IdempotencyKey}
		}
	}
	return DuplicateCognitiveEventResult{}, false, nil
}

func findStoredEventBy(ctx context.Context, tx *sql.Tx, column, value string) (StoredCognitiveEvent, bool, error) {
	if column != "event_id" && column != "idempotency_key" {
		return StoredCognitiveEvent{}, false, fmt.Errorf("unsupported cognitive event lookup column")
	}
	row := tx.QueryRowContext(ctx, selectCognitiveEvent+` WHERE `+column+` = ?`, value)
	event, err := scanStoredCognitiveEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return StoredCognitiveEvent{}, false, nil
	}
	if err != nil {
		return StoredCognitiveEvent{}, false, fmt.Errorf("lookup cognitive event by %s: %w", column, err)
	}
	return event, true, nil
}

const selectCognitiveEvent = `SELECT sequence, schema_version, event_id, event_type, occurred_at, identity_id, session_id,
	correlation_id, causation_id, goal_id, action_id, idempotency_key, evidence_class,
	payload_json, provider, backend_kind, model_id, route_reason, degraded, policy_version,
	content_sha256 FROM cognitive_events`

// Get retrieves one event by immutable event ID.
func (s *CognitiveEventStore) Get(ctx context.Context, eventID string) (StoredCognitiveEvent, error) {
	if err := validateRequiredIdentifier("event_id", eventID); err != nil {
		return StoredCognitiveEvent{}, err
	}
	return s.queryOne(nonNilContext(ctx), selectCognitiveEvent+` WHERE event_id = ?`, eventID)
}

// GetCognitiveEvent is a descriptive alias for Get.
func (s *CognitiveEventStore) GetCognitiveEvent(ctx context.Context, eventID string) (StoredCognitiveEvent, error) {
	return s.Get(ctx, eventID)
}

// Replay returns events strictly after afterSequence in stable ledger order.
func (s *CognitiveEventStore) Replay(ctx context.Context, afterSequence int64, limit int) ([]StoredCognitiveEvent, error) {
	return s.Query(ctx, CognitiveEventQuery{AfterSequence: afterSequence, Limit: limit})
}

// ReplayCognitiveEvents is a descriptive alias for Replay.
func (s *CognitiveEventStore) ReplayCognitiveEvents(ctx context.Context, afterSequence int64, limit int) ([]StoredCognitiveEvent, error) {
	return s.Replay(ctx, afterSequence, limit)
}

// Query returns a bounded, ascending-sequence replay. The query supports only
// equality filters so it cannot weaken the immutable ordering contract.
func (s *CognitiveEventStore) Query(ctx context.Context, query CognitiveEventQuery) ([]StoredCognitiveEvent, error) {
	ctx = nonNilContext(ctx)
	if err := validateCognitiveEventQuery(query); err != nil {
		return nil, err
	}
	limit := normalizedEventQueryLimit(query.Limit)

	clauses := []string{"sequence > ?"}
	args := []any{query.AfterSequence}
	for _, filter := range []struct {
		column string
		value  string
	}{
		{"identity_id", query.IdentityID},
		{"session_id", query.SessionID},
		{"correlation_id", query.CorrelationID},
		{"goal_id", query.GoalID},
		{"action_id", query.ActionID},
		{"event_type", query.EventType},
		{"evidence_class", query.EvidenceClass},
	} {
		if filter.value != "" {
			clauses = append(clauses, filter.column+" = ?")
			args = append(args, filter.value)
		}
	}
	args = append(args, limit)

	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.isOpen || s.db == nil {
		return nil, fmt.Errorf("database not open")
	}
	rows, err := s.db.QueryContext(ctx, selectCognitiveEvent+` WHERE `+strings.Join(clauses, ` AND `)+` ORDER BY sequence ASC LIMIT ?`, args...)
	if err != nil {
		return nil, fmt.Errorf("query cognitive events: %w", err)
	}
	defer rows.Close()

	events := make([]StoredCognitiveEvent, 0)
	for rows.Next() {
		event, err := scanStoredCognitiveEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan cognitive event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cognitive events: %w", err)
	}
	return events, nil
}

// QueryCognitiveEvents is a descriptive alias for Query.
func (s *CognitiveEventStore) QueryCognitiveEvents(ctx context.Context, query CognitiveEventQuery) ([]StoredCognitiveEvent, error) {
	return s.Query(ctx, query)
}

// Count returns the number of events matching the supplied immutable filters.
func (s *CognitiveEventStore) Count(ctx context.Context, query CognitiveEventQuery) (int64, error) {
	ctx = nonNilContext(ctx)
	if err := validateCognitiveEventQuery(query); err != nil {
		return 0, err
	}
	clauses := []string{"sequence > ?"}
	args := []any{query.AfterSequence}
	for _, filter := range []struct {
		column string
		value  string
	}{
		{"identity_id", query.IdentityID},
		{"session_id", query.SessionID},
		{"correlation_id", query.CorrelationID},
		{"goal_id", query.GoalID},
		{"action_id", query.ActionID},
		{"event_type", query.EventType},
		{"evidence_class", query.EvidenceClass},
	} {
		if filter.value != "" {
			clauses = append(clauses, filter.column+" = ?")
			args = append(args, filter.value)
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.isOpen || s.db == nil {
		return 0, fmt.Errorf("database not open")
	}
	var count int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM cognitive_events WHERE `+strings.Join(clauses, ` AND `), args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count cognitive events: %w", err)
	}
	return count, nil
}

// CountCognitiveEvents is a descriptive alias for Count.
func (s *CognitiveEventStore) CountCognitiveEvents(ctx context.Context, query CognitiveEventQuery) (int64, error) {
	return s.Count(ctx, query)
}

// Health checks event-ledger reachability and the SQLite invariants required
// for durable append-only use.
func (s *CognitiveEventStore) Health(ctx context.Context) (CognitiveEventHealth, error) {
	ctx = nonNilContext(ctx)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.isOpen || s.db == nil {
		return CognitiveEventHealth{}, fmt.Errorf("database not open")
	}
	health := CognitiveEventHealth{SchemaVersion: cognitiveEventSchemaMigrationVersion}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM cognitive_events`).Scan(&health.EventCount); err != nil {
		return CognitiveEventHealth{}, fmt.Errorf("count cognitive events for health: %w", err)
	}
	var foreignKeys, synchronous, busyTimeout int
	var journalMode string
	if err := s.db.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		return CognitiveEventHealth{}, fmt.Errorf("read SQLite foreign_keys for health: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `PRAGMA synchronous`).Scan(&synchronous); err != nil {
		return CognitiveEventHealth{}, fmt.Errorf("read SQLite synchronous for health: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `PRAGMA busy_timeout`).Scan(&busyTimeout); err != nil {
		return CognitiveEventHealth{}, fmt.Errorf("read SQLite busy_timeout for health: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `PRAGMA journal_mode`).Scan(&journalMode); err != nil {
		return CognitiveEventHealth{}, fmt.Errorf("read SQLite journal_mode for health: %w", err)
	}
	health.ForeignKeys = foreignKeys == 1
	health.SynchronousFull = synchronous == 2
	health.BusyTimeoutMS = busyTimeout
	health.WAL = strings.EqualFold(journalMode, "wal")
	health.Ready = health.ForeignKeys && health.SynchronousFull && health.WAL && health.BusyTimeoutMS >= sqliteBusyTimeoutMilliseconds
	if !health.Ready {
		return health, fmt.Errorf("cognitive event SQLite hardening is not ready")
	}
	return health, nil
}

// CognitiveEventHealthCheck is a descriptive alias for Health.
func (s *CognitiveEventStore) CognitiveEventHealthCheck(ctx context.Context) (CognitiveEventHealth, error) {
	return s.Health(ctx)
}

// GetProjectionCursor retrieves a projection high-water mark.
func (s *CognitiveEventStore) GetProjectionCursor(ctx context.Context, projectionName string) (ProjectionCursor, error) {
	if err := validateRequiredIdentifier("projection_name", projectionName); err != nil {
		return ProjectionCursor{}, err
	}
	ctx = nonNilContext(ctx)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.isOpen || s.db == nil {
		return ProjectionCursor{}, fmt.Errorf("database not open")
	}
	var cursor ProjectionCursor
	var updatedAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT projection_name, sequence, updated_at FROM cognitive_projection_cursors WHERE projection_name = ?`, projectionName,
	).Scan(&cursor.ProjectionName, &cursor.Sequence, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ProjectionCursor{}, fmt.Errorf("%w: %s", ErrProjectionCursorNotFound, projectionName)
	}
	if err != nil {
		return ProjectionCursor{}, fmt.Errorf("get projection cursor: %w", err)
	}
	parsed, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return ProjectionCursor{}, fmt.Errorf("parse projection cursor timestamp: %w", err)
	}
	cursor.UpdatedAt = parsed
	return cursor, nil
}

// AdvanceProjectionCursor inserts a new cursor or atomically advances an
// existing one. A lower sequence is rejected, while replaying the same sequence
// is an idempotent no-op that returns the existing cursor.
func (s *CognitiveEventStore) AdvanceProjectionCursor(ctx context.Context, projectionName string, sequence int64) (ProjectionCursor, error) {
	if err := validateRequiredIdentifier("projection_name", projectionName); err != nil {
		return ProjectionCursor{}, err
	}
	if sequence < 0 {
		return ProjectionCursor{}, fmt.Errorf("%w: sequence must be non-negative", ErrInvalidCognitiveEvent)
	}
	ctx = nonNilContext(ctx)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.isOpen || s.db == nil {
		return ProjectionCursor{}, fmt.Errorf("database not open")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ProjectionCursor{}, fmt.Errorf("begin projection cursor advance: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var current int64
	var updatedAt string
	err = tx.QueryRowContext(ctx,
		`SELECT sequence, updated_at FROM cognitive_projection_cursors WHERE projection_name = ?`, projectionName,
	).Scan(&current, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		now := time.Now().UTC()
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO cognitive_projection_cursors (projection_name, sequence, updated_at) VALUES (?, ?, ?)`,
			projectionName, sequence, now.Format(time.RFC3339Nano),
		); err != nil {
			return ProjectionCursor{}, fmt.Errorf("insert projection cursor: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return ProjectionCursor{}, fmt.Errorf("commit projection cursor insert: %w", err)
		}
		return ProjectionCursor{ProjectionName: projectionName, Sequence: sequence, UpdatedAt: now}, nil
	}
	if err != nil {
		return ProjectionCursor{}, fmt.Errorf("read projection cursor: %w", err)
	}
	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return ProjectionCursor{}, fmt.Errorf("parse projection cursor timestamp: %w", err)
	}
	if sequence < current {
		return ProjectionCursor{}, fmt.Errorf("%w: current=%d requested=%d", ErrProjectionCursorRegression, current, sequence)
	}
	if sequence == current {
		if err := tx.Commit(); err != nil {
			return ProjectionCursor{}, fmt.Errorf("commit unchanged projection cursor: %w", err)
		}
		return ProjectionCursor{ProjectionName: projectionName, Sequence: current, UpdatedAt: parsedUpdatedAt}, nil
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx,
		`UPDATE cognitive_projection_cursors SET sequence = ?, updated_at = ? WHERE projection_name = ?`,
		sequence, now.Format(time.RFC3339Nano), projectionName,
	); err != nil {
		return ProjectionCursor{}, fmt.Errorf("advance projection cursor: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return ProjectionCursor{}, fmt.Errorf("commit projection cursor advance: %w", err)
	}
	return ProjectionCursor{ProjectionName: projectionName, Sequence: sequence, UpdatedAt: now}, nil
}

func (s *CognitiveEventStore) queryOne(ctx context.Context, query string, args ...any) (StoredCognitiveEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.isOpen || s.db == nil {
		return StoredCognitiveEvent{}, fmt.Errorf("database not open")
	}
	event, err := scanStoredCognitiveEvent(s.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return StoredCognitiveEvent{}, fmt.Errorf("%w", ErrCognitiveEventNotFound)
	}
	if err != nil {
		return StoredCognitiveEvent{}, fmt.Errorf("get cognitive event: %w", err)
	}
	return event, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanStoredCognitiveEvent(row rowScanner) (StoredCognitiveEvent, error) {
	var stored StoredCognitiveEvent
	var occurredAt string
	var payload string
	var degraded int
	if err := row.Scan(
		&stored.Sequence,
		&stored.SchemaVersion,
		&stored.EventID,
		&stored.EventType,
		&occurredAt,
		&stored.IdentityID,
		&stored.SessionID,
		&stored.CorrelationID,
		&stored.CausationID,
		&stored.GoalID,
		&stored.ActionID,
		&stored.IdempotencyKey,
		&stored.EvidenceClass,
		&payload,
		&stored.Provider,
		&stored.BackendKind,
		&stored.ModelID,
		&stored.RouteReason,
		&degraded,
		&stored.PolicyVersion,
		&stored.ContentSHA256,
	); err != nil {
		return StoredCognitiveEvent{}, err
	}
	occurred, err := time.Parse(time.RFC3339Nano, occurredAt)
	if err != nil {
		return StoredCognitiveEvent{}, fmt.Errorf("parse occurred_at: %w", err)
	}
	stored.OccurredAt = occurred
	stored.PayloadJSON = []byte(payload)
	stored.Degraded = degraded != 0
	if err := stored.CognitiveEvent.Validate(); err != nil {
		return StoredCognitiveEvent{}, fmt.Errorf("stored cognitive event validation failed: %w", err)
	}
	return stored, nil
}

func validateCognitiveEventQuery(query CognitiveEventQuery) error {
	if query.AfterSequence < 0 {
		return fmt.Errorf("%w: after_sequence must be non-negative", ErrInvalidCognitiveEvent)
	}
	if query.Limit < 0 || query.Limit > 1000 {
		return fmt.Errorf("%w: limit must be between 0 and 1000", ErrInvalidCognitiveEvent)
	}
	for field, value := range map[string]string{
		"identity_id":    query.IdentityID,
		"session_id":     query.SessionID,
		"correlation_id": query.CorrelationID,
		"goal_id":        query.GoalID,
		"action_id":      query.ActionID,
	} {
		if err := validateOptionalIdentifier(field, value); err != nil {
			return err
		}
	}
	if query.EventType != "" {
		if _, ok := stableEventTypes[query.EventType]; !ok {
			return fmt.Errorf("%w: %q", ErrInvalidCognitiveEventType, query.EventType)
		}
	}
	if query.EvidenceClass != "" {
		if _, ok := stableEvidenceClasses[query.EvidenceClass]; !ok {
			return fmt.Errorf("%w: %q", ErrInvalidEvidenceClass, query.EvidenceClass)
		}
	}
	return nil
}

func normalizedEventQueryLimit(limit int) int {
	if limit == 0 {
		return 100
	}
	return limit
}

func boolToSQLite(value bool) int {
	if value {
		return 1
	}
	return 0
}

func isSQLiteConstraint(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "constraint") || strings.Contains(message, "unique")
}

func nonNilContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
