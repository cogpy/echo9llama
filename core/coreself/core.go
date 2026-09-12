// Package coreself implements a deterministic, file-backed identity kernel.
//
// It deliberately has no inference, networking, scheduler, shell, tool, or
// external-action capability. Identity changes are accepted only from a
// separately stored proposal after review by an authorized human steward.
package coreself

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	// SchemaVersion is the only accepted contract version.
	SchemaVersion = "1.0.0"

	// DefaultGenesisTime is the fixed origin of the initial identity record.
	DefaultGenesisTime = "2026-09-12T03:30:55Z"
	// DefaultSourceCommit binds the bootstrap record to its reviewed source.
	DefaultSourceCommit = "2143b1ebcab4b36b6050f1f33e260d4feb3de9f1"

	// RootSubjectID is the protected root identity subject.
	RootSubjectID = "core:agent:cogpy/echo9llama"
	// StewardSubjectID is the protected human steward subject.
	StewardSubjectID = "core:steward:human/dan"
	// RuntimeSubjectID is the protected Deep Tree Echo runtime subject.
	RuntimeSubjectID = "core:agent:runtime/deep-tree-echo"

	maxLedgerBytes   int64 = 64 << 20
	maxHeadBytes     int64 = 16 << 10
	maxProposalBytes int64 = 1 << 20
)

var (
	// ErrInvalidContract reports a syntactically or semantically invalid contract.
	ErrInvalidContract = errors.New("coreself: invalid contract")
	// ErrTampered reports non-canonical or digest-invalid persisted material.
	ErrTampered = errors.New("coreself: tampered persisted material")
	// ErrFork reports a broken ledger predecessor or sequence chain.
	ErrFork = errors.New("coreself: ledger fork")
	// ErrStaleHead reports a writer attempting to persist from an obsolete state.
	ErrStaleHead = errors.New("coreself: stale ledger head")
	// ErrImmutableCore reports an attempted mutation of protected core identity.
	ErrImmutableCore = errors.New("coreself: immutable core identity")
	// ErrUnauthorized reports an actor or reviewer without the required authority.
	ErrUnauthorized = errors.New("coreself: unauthorized")
	// ErrProposalOnly reports a prohibited direct source path.
	ErrProposalOnly = errors.New("coreself: proposal-only source")
	// ErrNotFound reports a missing proposal or record.
	ErrNotFound = errors.New("coreself: not found")
	// ErrAlreadyExists reports an immutable record identifier collision.
	ErrAlreadyExists = errors.New("coreself: already exists")
	// ErrAlreadyDecided reports a proposal already represented by the ledger.
	ErrAlreadyDecided = errors.New("coreself: proposal already decided")
)

var (
	identifierRE = regexp.MustCompile(`^[a-z0-9][a-z0-9:._/-]{2,160}$`)
	digestRE     = regexp.MustCompile(`^[a-f0-9]{64}$`)
	signatureRE  = regexp.MustCompile(`^[a-f0-9]{128}$`)
	commitRE     = regexp.MustCompile(`^[a-f0-9]{40}$`)
)

// ActorRole declares a narrowly scoped identity authority role.
type ActorRole string

const (
	// RoleCoreIdentity identifies the protected root subject.
	RoleCoreIdentity ActorRole = "core_identity"
	// RoleHumanSteward identifies a human responsible for stewardship.
	RoleHumanSteward ActorRole = "human_steward"
	// RoleAuthorizedReviewer may accept a pending proposal.
	RoleAuthorizedReviewer ActorRole = "authorized_reviewer"
	// RoleRuntime identifies a runtime-only identity endpoint.
	RoleRuntime ActorRole = "runtime"
	// RoleExternalContributor may submit a proposal but may not review one.
	RoleExternalContributor ActorRole = "external_contributor"
)

// SourceClass classifies provenance without granting external action authority.
type SourceClass string

const (
	// SourceLocal is a local human-stewardship context.
	SourceLocal SourceClass = "local"
	// SourceExternal is an external source and is always proposal-only.
	SourceExternal SourceClass = "external"
	// SourceExperimental is an experimental source and is always proposal-only.
	SourceExperimental SourceClass = "experimental"
)

// SubjectKind is a typed hypergraph endpoint kind.
type SubjectKind string

const (
	// KindCoreAgent identifies the protected root agent subject.
	KindCoreAgent SubjectKind = "core_agent"
	// KindHumanSteward identifies a human stewardship subject.
	KindHumanSteward SubjectKind = "human_steward"
	// KindRuntimeAgent identifies the restricted runtime subject.
	KindRuntimeAgent SubjectKind = "runtime_agent"
	// KindExternalAgent identifies a non-core external identity.
	KindExternalAgent SubjectKind = "external_agent"
	// KindService identifies a non-core service identity.
	KindService SubjectKind = "service"
)

// TruthValue controls the evidence requirements of a hypergraph relation.
type TruthValue string

const (
	// TruthAsserted is a reviewable assertion and may have no evidence reference.
	TruthAsserted TruthValue = "asserted"
	// TruthVerified requires at least one existing evidence reference.
	TruthVerified TruthValue = "verified"
)

// DecisionKind is the reviewer disposition for a proposal.
type DecisionKind string

const (
	// DecisionAccepted records an authorized accepted proposal.
	DecisionAccepted DecisionKind = "accepted"
	// DecisionDenied records an authorized denied proposal.
	DecisionDenied DecisionKind = "denied"
)

// EventType distinguishes bootstrap from authorized proposal acceptance.
type EventType string

const (
	// EventBootstrap is the one allowed sequence-zero event.
	EventBootstrap EventType = "bootstrap"
	// EventProposalAccepted records a proposal accepted by a reviewer.
	EventProposalAccepted EventType = "proposal_accepted"
)

// MutationOperation is the closed set of state transitions.
type MutationOperation string

const (
	// MutationBootstrap initializes the deterministic core state.
	MutationBootstrap MutationOperation = "bootstrap"
	// MutationAddSubject adds one non-core subject.
	MutationAddSubject MutationOperation = "add_subject"
	// MutationAddRelation adds one typed relation.
	MutationAddRelation MutationOperation = "add_relation"
	// MutationAddEvidence adds one digest-addressed evidence reference.
	MutationAddEvidence MutationOperation = "add_evidence"
	// MutationAmendSubject is retained for explicit denial of protected mutations.
	MutationAmendSubject MutationOperation = "amend_subject"
)

// Subject is an identity node with closed roles and no free-form attributes.
type Subject struct {
	SchemaVersion string      `json:"schema_version"`
	ID            string      `json:"id"`
	Kind          SubjectKind `json:"kind"`
	DisplayName   string      `json:"display_name"`
	Immutable     bool        `json:"immutable"`
	Roles         []ActorRole `json:"roles"`
}

// Endpoint identifies a typed relation endpoint by both subject ID and kind.
type Endpoint struct {
	SchemaVersion string      `json:"schema_version"`
	SubjectID     string      `json:"subject_id"`
	Kind          SubjectKind `json:"kind"`
}

// Truth records a relation's closed truth classification and evidence IDs.
type Truth struct {
	SchemaVersion string     `json:"schema_version"`
	Value         TruthValue `json:"value"`
	EvidenceIDs   []string   `json:"evidence_ids"`
}

// Relation is a typed and evidence-constrained hypergraph edge.
type Relation struct {
	SchemaVersion string   `json:"schema_version"`
	ID            string   `json:"id"`
	From          Endpoint `json:"from"`
	Predicate     string   `json:"predicate"`
	To            Endpoint `json:"to"`
	Truth         Truth    `json:"truth"`
}

// EvidenceRef points at content by immutable SHA-256 digest; it stores no content.
type EvidenceRef struct {
	SchemaVersion string      `json:"schema_version"`
	ID            string      `json:"id"`
	Locator       string      `json:"locator"`
	SHA256        string      `json:"sha256"`
	MediaType     string      `json:"media_type"`
	Source        SourceClass `json:"source"`
}

// PolicyContext records an actor, a closed role, provenance, and a bounded purpose.
type PolicyContext struct {
	SchemaVersion string      `json:"schema_version"`
	ActorID       string      `json:"actor_id"`
	Role          ActorRole   `json:"role"`
	Source        SourceClass `json:"source"`
	Purpose       string      `json:"purpose"`
}

// BootstrapRecord is the complete deterministic sequence-zero identity projection.
type BootstrapRecord struct {
	SchemaVersion         string     `json:"schema_version"`
	GenesisTime           string     `json:"genesis_time"`
	SourceCommit          string     `json:"source_commit"`
	Subjects              []Subject  `json:"subjects"`
	Relations             []Relation `json:"relations"`
	AuthorizedReviewerIDs []string   `json:"authorized_reviewer_ids"`
}

// Mutation is a closed tagged union. Exactly one operation-specific payload is valid.
type Mutation struct {
	SchemaVersion string            `json:"schema_version"`
	Operation     MutationOperation `json:"operation"`
	Subject       *Subject          `json:"subject,omitempty"`
	Relation      *Relation         `json:"relation,omitempty"`
	Evidence      *EvidenceRef      `json:"evidence,omitempty"`
	TargetID      string            `json:"target_id,omitempty"`
	Bootstrap     *BootstrapRecord  `json:"bootstrap,omitempty"`
}

// Proposal is stored outside the ledger until an authorized reviewer accepts it.
type Proposal struct {
	SchemaVersion string        `json:"schema_version"`
	ID            string        `json:"id"`
	SubmittedAt   string        `json:"submitted_at"`
	Submitter     PolicyContext `json:"submitter"`
	Mutation      Mutation      `json:"mutation"`
	Digest        string        `json:"digest,omitempty"`
}

// PolicyDecision is an append-only reviewer decision embedded in an identity event.
type PolicyDecision struct {
	SchemaVersion  string        `json:"schema_version"`
	ProposalID     string        `json:"proposal_id"`
	ProposalDigest string        `json:"proposal_digest"`
	Reviewer       PolicyContext `json:"reviewer"`
	Decision       DecisionKind  `json:"decision"`
	DecidedAt      string        `json:"decided_at"`
	Digest         string        `json:"digest,omitempty"`
	Signature      string        `json:"signature,omitempty"`
}

// IdentityEvent is a SHA-256 hash-linked, canonical ledger event.
type IdentityEvent struct {
	SchemaVersion  string          `json:"schema_version"`
	ID             string          `json:"id"`
	Sequence       uint64          `json:"sequence"`
	OccurredAt     string          `json:"occurred_at"`
	PreviousHash   string          `json:"previous_hash"`
	Type           EventType       `json:"type"`
	ProposalID     string          `json:"proposal_id,omitempty"`
	ProposalDigest string          `json:"proposal_digest,omitempty"`
	Proposal       *Proposal       `json:"proposal,omitempty"`
	Decision       *PolicyDecision `json:"decision,omitempty"`
	Mutation       Mutation        `json:"mutation"`
	Hash           string          `json:"hash,omitempty"`
}

// State is the wholly projected, sorted identity hypergraph state.
type State struct {
	SchemaVersion         string        `json:"schema_version"`
	GenesisTime           string        `json:"genesis_time"`
	SourceCommit          string        `json:"source_commit"`
	Subjects              []Subject     `json:"subjects"`
	Relations             []Relation    `json:"relations"`
	Evidence              []EvidenceRef `json:"evidence"`
	AuthorizedReviewerIDs []string      `json:"authorized_reviewer_ids"`
	HeadHash              string        `json:"head_hash"`
	EventCount            uint64        `json:"event_count"`
}

// Snapshot binds a canonical State to its SHA-256 digest.
type Snapshot struct {
	SchemaVersion string `json:"schema_version"`
	State         State  `json:"state"`
	StateDigest   string `json:"state_digest"`
}

// Capsule contains all material needed to verify an identity state without local files.
type Capsule struct {
	SchemaVersion     string          `json:"schema_version"`
	ReviewerPublicKey string          `json:"reviewer_public_key,omitempty"`
	Snapshot          Snapshot        `json:"snapshot"`
	Ledger            []IdentityEvent `json:"ledger"`
	LedgerDigest      string          `json:"ledger_digest"`
	Digest            string          `json:"digest,omitempty"`
}

// Status is a compact immutable view of the current verified state.
type Status struct {
	SchemaVersion string `json:"schema_version"`
	HeadHash      string `json:"head_hash"`
	EventCount    uint64 `json:"event_count"`
	StateDigest   string `json:"state_digest"`
	ProposalCount uint64 `json:"proposal_count"`
}

// BootstrapConfig optionally repeats the fixed bootstrap values. Empty fields select defaults.
type BootstrapConfig struct {
	GenesisTime           string   `json:"genesis_time,omitempty"`
	SourceCommit          string   `json:"source_commit,omitempty"`
	AuthorizedReviewerIDs []string `json:"authorized_reviewer_ids,omitempty"`
	ReviewerPublicKey     string   `json:"reviewer_public_key,omitempty"`
}

// Kernel owns a verified on-disk ledger and has no cognitive or external-action authority.
type Kernel struct {
	directory   string
	rootName    string
	parent      *os.Root
	root        *os.Root
	proposals   *os.Root
	quarantine  *os.Root
	lock        *os.File
	ledger      *os.File
	head        *os.File
	reviewerKey ed25519.PublicKey
	state       State
	events      []IdentityEvent
	closed      bool
	mu          sync.Mutex
}

type headRecord struct {
	SchemaVersion string `json:"schema_version"`
	HeadHash      string `json:"head_hash"`
	EventCount    uint64 `json:"event_count"`
	StateDigest   string `json:"state_digest"`
}

// Open verifies an existing ledger or creates the deterministic bootstrap ledger.
func Open(directory string, config BootstrapConfig) (*Kernel, error) {
	if directory == "" || !utf8.ValidString(directory) {
		return nil, invalid("directory is required")
	}
	absolute, err := filepath.Abs(directory)
	if err != nil {
		return nil, err
	}
	cfg, err := normalizeBootstrap(config)
	if err != nil {
		return nil, err
	}
	reviewerKey, err := decodeReviewerPublicKey(cfg.ReviewerPublicKey)
	if err != nil {
		return nil, err
	}
	if err := ensurePrivateDir(absolute); err != nil {
		return nil, err
	}
	parent, err := os.OpenRoot(filepath.Dir(absolute))
	if err != nil {
		return nil, err
	}
	keepParent := false
	defer func() {
		if !keepParent {
			_ = parent.Close()
		}
	}()
	lock, err := openDirectoryHandle(parent)
	if err != nil {
		return nil, err
	}
	keepLock := false
	defer func() {
		if !keepLock {
			_ = lock.Close()
		}
	}()
	if err := lockFile(lock); err != nil {
		return nil, err
	}
	defer func() { _ = unlockFile(lock) }()
	rootName := filepath.Base(absolute)
	root, err := parent.OpenRoot(rootName)
	if err != nil {
		return nil, err
	}
	keepRoot := false
	defer func() {
		if !keepRoot {
			_ = root.Close()
		}
	}()
	if err := verifyRootEntry(parent, rootName, root); err != nil {
		return nil, err
	}
	for _, name := range []string{"proposals", "quarantine"} {
		if err := ensurePrivateRootDir(root, name); err != nil {
			return nil, err
		}
	}
	proposals, err := openPrivateSubroot(root, "proposals")
	if err != nil {
		return nil, err
	}
	keepProposals := false
	defer func() {
		if !keepProposals {
			_ = proposals.Close()
		}
	}()
	quarantineRoot, err := openPrivateSubroot(root, "quarantine")
	if err != nil {
		return nil, err
	}
	keepQuarantine := false
	defer func() {
		if !keepQuarantine {
			_ = quarantineRoot.Close()
		}
	}()
	ledgerExists, err := regularPrivateExists(root, "identity-ledger.jsonl")
	if err != nil {
		return nil, err
	}
	headExists, err := regularPrivateExists(root, "head.json")
	if err != nil {
		return nil, err
	}
	if !ledgerExists && !headExists {
		if err := bootstrap(root, cfg); err != nil {
			return nil, err
		}
		ledgerExists = true
		headExists = true
	}
	if !ledgerExists {
		err := fmt.Errorf("%w: ledger and head presence disagree", ErrFork)
		return nil, quarantined(quarantineRoot, err)
	}
	ledger, err := openPrivateRegular(root, "identity-ledger.jsonl", os.O_RDWR|os.O_APPEND)
	if err != nil {
		return nil, err
	}
	keepLedger := false
	defer func() {
		if !keepLedger {
			_ = ledger.Close()
		}
	}()
	var head *os.File
	if headExists {
		head, err = openPrivateRegular(root, "head.json", os.O_RDONLY)
		if err != nil {
			return nil, err
		}
	}
	keepHead := false
	defer func() {
		if head != nil && !keepHead {
			_ = head.Close()
		}
	}()
	kernel := &Kernel{
		directory:   absolute,
		rootName:    rootName,
		parent:      parent,
		root:        root,
		proposals:   proposals,
		quarantine:  quarantineRoot,
		lock:        lock,
		ledger:      ledger,
		head:        head,
		reviewerKey: reviewerKey,
	}
	state, events, err := loadVerified(kernel, true)
	if err != nil {
		return nil, quarantined(quarantineRoot, err)
	}
	if err := verifyBootstrapState(state, cfg); err != nil {
		return nil, quarantined(quarantineRoot, err)
	}
	if err := verifyRootEntry(parent, rootName, root); err != nil {
		return nil, quarantined(quarantineRoot, err)
	}
	keepParent = true
	keepRoot = true
	keepProposals = true
	keepQuarantine = true
	keepLock = true
	keepLedger = true
	keepHead = true
	kernel.state = state
	kernel.events = events
	return kernel, nil
}

// Status returns the current verified head and counts valid pending proposals without exposing content.
func (k *Kernel) Status() Status {
	k.mu.Lock()
	defer k.mu.Unlock()
	proposalCount := uint64(0)
	if !k.closed {
		proposalCount = k.proposalCount()
	}
	return Status{
		SchemaVersion: SchemaVersion,
		HeadHash:      k.state.HeadHash,
		EventCount:    k.state.EventCount,
		StateDigest:   stateDigest(k.state),
		ProposalCount: proposalCount,
	}
}

// Close deterministically releases retained persistence descriptors. It is
// idempotent and does not alter accepted identity state.
func (k *Kernel) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.closed {
		return nil
	}
	k.closed = true
	var errs []error
	for _, closeFile := range []*os.File{k.head, k.ledger, k.lock} {
		if closeFile != nil {
			errs = append(errs, closeFile.Close())
		}
	}
	for _, closeRoot := range []*os.Root{k.quarantine, k.proposals, k.root, k.parent} {
		if closeRoot != nil {
			errs = append(errs, closeRoot.Close())
		}
	}
	k.head = nil
	k.ledger = nil
	k.lock = nil
	k.quarantine = nil
	k.proposals = nil
	k.root = nil
	k.parent = nil
	return errors.Join(errs...)
}

// StateSnapshot returns a deep, deterministic snapshot of the projected state.
func (k *Kernel) StateSnapshot() Snapshot {
	k.mu.Lock()
	defer k.mu.Unlock()
	state := cloneState(k.state)
	return Snapshot{SchemaVersion: SchemaVersion, State: state, StateDigest: stateDigest(state)}
}

// ExportCapsule creates a portable capsule from the already verified in-memory ledger.
func (k *Kernel) ExportCapsule() (Capsule, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	ledger, err := cloneEvents(k.events)
	if err != nil {
		return Capsule{}, err
	}
	reviewerPublicKey := ""
	if len(ledger) > 1 {
		reviewerPublicKey = hex.EncodeToString(k.reviewerKey)
	}
	capsule := Capsule{
		SchemaVersion:     SchemaVersion,
		ReviewerPublicKey: reviewerPublicKey,
		Snapshot:          Snapshot{SchemaVersion: SchemaVersion, State: cloneState(k.state), StateDigest: stateDigest(k.state)},
		Ledger:            ledger,
	}
	capsule.LedgerDigest = digestValue(capsule.Ledger)
	digest, err := capsuleDigest(capsule)
	if err != nil {
		return Capsule{}, err
	}
	capsule.Digest = digest
	return capsule, nil
}

// VerifyCapsule independently verifies a bootstrap-only capsule. Accepted
// post-genesis history requires an externally supplied reviewer trust anchor.
func VerifyCapsule(c Capsule) error {
	if len(c.Ledger) > 1 || c.ReviewerPublicKey != "" {
		return fmt.Errorf("%w: accepted capsule requires an external reviewer public key", ErrUnauthorized)
	}
	return verifyCapsule(c, nil)
}

// VerifyCapsuleWithReviewerKey verifies an accepted capsule against a reviewer
// key obtained outside the capsule and accepted ledger.
func VerifyCapsuleWithReviewerKey(c Capsule, encodedReviewerPublicKey string) error {
	reviewerKey, err := decodeReviewerPublicKey(encodedReviewerPublicKey)
	if err != nil || len(reviewerKey) != ed25519.PublicKeySize {
		return fmt.Errorf("%w: external reviewer public key", ErrUnauthorized)
	}
	if c.ReviewerPublicKey != hex.EncodeToString(reviewerKey) {
		return fmt.Errorf("%w: capsule reviewer key differs from external trust anchor", ErrTampered)
	}
	return verifyCapsule(c, reviewerKey)
}

func verifyCapsule(c Capsule, reviewerKey ed25519.PublicKey) error {
	if err := validateCapsule(c, true); err != nil {
		return err
	}
	state, err := replay(c.Ledger, reviewerKey)
	if err != nil {
		return err
	}
	if err := verifyBootstrapState(state, mustDefaultBootstrap()); err != nil {
		return fmt.Errorf("%w: capsule bootstrap: %v", ErrTampered, err)
	}
	if stateDigest(state) != c.Snapshot.StateDigest || !statesEqual(state, c.Snapshot.State) {
		return fmt.Errorf("%w: capsule snapshot does not equal replay", ErrTampered)
	}
	return nil
}

// VerifyCapsule verifies a caller-supplied portable capsule against this
// kernel's externally configured reviewer trust anchor without reading disk.
func (k *Kernel) VerifyCapsule(c Capsule) error {
	if len(c.Ledger) == 1 {
		if c.ReviewerPublicKey != "" {
			return fmt.Errorf("%w: bootstrap capsule carries an unnecessary reviewer key", ErrTampered)
		}
		return verifyCapsule(c, nil)
	}
	k.mu.Lock()
	reviewerKey := append(ed25519.PublicKey(nil), k.reviewerKey...)
	k.mu.Unlock()
	if len(c.Ledger) > 1 && len(reviewerKey) != ed25519.PublicKeySize {
		return fmt.Errorf("%w: kernel reviewer public key unavailable", ErrUnauthorized)
	}
	if c.ReviewerPublicKey != hex.EncodeToString(reviewerKey) {
		return fmt.Errorf("%w: capsule reviewer key differs from kernel trust anchor", ErrTampered)
	}
	return verifyCapsule(c, reviewerKey)
}

// submitProposal validates and atomically stores a proposal outside the identity ledger.
// E3 deliberately exposes no public persistence method.
func (k *Kernel) submitProposal(proposal Proposal) (Proposal, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.closed {
		return Proposal{}, fmt.Errorf("%w: core-self kernel is closed", ErrUnauthorized)
	}
	if err := validateProposal(proposal, false); err != nil {
		return Proposal{}, err
	}
	if proposal.Submitter.Source == SourceExternal || proposal.Submitter.Source == SourceExperimental {
		if proposal.Submitter.Role != RoleExternalContributor {
			return Proposal{}, fmt.Errorf("%w: external and experimental submitters must use the contributor role", ErrUnauthorized)
		}
	} else if err := k.validateActor(proposal.Submitter, false); err != nil {
		return Proposal{}, err
	}
	if err := k.preflight(proposal.Mutation); err != nil {
		return Proposal{}, err
	}
	proposal.Digest = ""
	digest, err := proposalDigest(proposal)
	if err != nil {
		return Proposal{}, err
	}
	proposal.Digest = digest
	data, err := CanonicalJSON(proposal)
	if err != nil {
		return Proposal{}, err
	}
	unlock, err := acquireKernelLock(k)
	if err != nil {
		return Proposal{}, err
	}
	defer unlock()
	name := proposalFileName(proposal.ID)
	if err := writePrivateExclusive(k.proposals, name, append(data, '\n')); err != nil {
		if errors.Is(err, os.ErrExist) {
			return Proposal{}, fmt.Errorf("%w: proposal %s", ErrAlreadyExists, proposal.ID)
		}
		return Proposal{}, err
	}
	if err := verifyRootEntry(k.parent, k.rootName, k.root); err != nil {
		return Proposal{}, err
	}
	return proposal, nil
}

// acceptProposal accepts one pending proposal only when a configured human reviewer authorizes it.
// E3 intentionally exposes no public acceptance API; a future local review
// adapter must be designed as a separate authority-bearing iteration.
func (k *Kernel) acceptProposal(proposalID string, reviewer PolicyContext, privateKey ed25519.PrivateKey) (PolicyDecision, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.closed {
		return PolicyDecision{}, fmt.Errorf("%w: core-self kernel is closed", ErrUnauthorized)
	}
	if !identifierRE.MatchString(proposalID) {
		return PolicyDecision{}, invalid("proposal ID")
	}
	if err := k.validateActor(reviewer, true); err != nil {
		return PolicyDecision{}, err
	}
	if reviewer.Source != SourceLocal {
		return PolicyDecision{}, fmt.Errorf("%w: reviewer source must be local", ErrUnauthorized)
	}
	if !containsString(k.state.AuthorizedReviewerIDs, reviewer.ActorID) {
		return PolicyDecision{}, fmt.Errorf("%w: reviewer is not configured", ErrUnauthorized)
	}
	if len(k.reviewerKey) != ed25519.PublicKeySize || len(privateKey) != ed25519.PrivateKeySize || !bytes.Equal(privateKey.Public().(ed25519.PublicKey), k.reviewerKey) {
		return PolicyDecision{}, fmt.Errorf("%w: reviewer signing key", ErrUnauthorized)
	}
	unlock, err := acquireKernelLock(k)
	if err != nil {
		if errors.Is(err, ErrTampered) {
			return PolicyDecision{}, quarantined(k.quarantine, err)
		}
		return PolicyDecision{}, err
	}
	defer unlock()

	// A live kernel never rebinds an anchored head descriptor. Atomic head
	// replacement by this kernel updates the handle inside replaceHead; any
	// other inode substitution is tampering, even when its bytes are a valid
	// historical ancestor.
	if k.head != nil {
		if err := verifyAnchoredFile(k.root, "head.json", k.head); err != nil {
			return PolicyDecision{}, quarantined(k.quarantine, err)
		}
	}
	current, events, err := loadVerified(k, false)
	if err != nil {
		return PolicyDecision{}, quarantined(k.quarantine, err)
	}
	if current.HeadHash != k.state.HeadHash || current.EventCount != k.state.EventCount {
		return PolicyDecision{}, fmt.Errorf("%w: reopen the kernel before accepting", ErrStaleHead)
	}
	if proposalAccepted(events, proposalID) {
		return PolicyDecision{}, fmt.Errorf("%w: %s", ErrAlreadyDecided, proposalID)
	}
	proposal, err := readProposal(k.proposals, proposalFileName(proposalID))
	if err != nil {
		return PolicyDecision{}, err
	}
	if err := k.preflight(proposal.Mutation); err != nil {
		return PolicyDecision{}, err
	}
	decision := PolicyDecision{
		SchemaVersion:  SchemaVersion,
		ProposalID:     proposal.ID,
		ProposalDigest: proposal.Digest,
		Reviewer:       reviewer,
		Decision:       DecisionAccepted,
		DecidedAt:      proposal.SubmittedAt,
	}
	decisionDigest, err := decisionDigest(decision)
	if err != nil {
		return PolicyDecision{}, err
	}
	decision.Digest = decisionDigest
	event := IdentityEvent{
		SchemaVersion:  SchemaVersion,
		ID:             "event:" + proposal.ID,
		Sequence:       current.EventCount,
		OccurredAt:     proposal.SubmittedAt,
		PreviousHash:   current.HeadHash,
		Type:           EventProposalAccepted,
		ProposalID:     proposal.ID,
		ProposalDigest: proposal.Digest,
		Proposal:       &proposal,
		Decision:       &decision,
		Mutation:       proposal.Mutation,
	}
	authorization, err := eventAuthorizationBytes(event)
	if err != nil {
		return PolicyDecision{}, err
	}
	event.Decision.Signature = hex.EncodeToString(ed25519.Sign(privateKey, authorization))
	eventHash, err := eventDigest(event)
	if err != nil {
		return PolicyDecision{}, err
	}
	event.Hash = eventHash
	projected, err := applyEvent(current, event)
	if err != nil {
		return PolicyDecision{}, err
	}
	// Projection is complete before persistence changes. The ledger append is
	// authoritative; the atomically replaced head is only a recoverable cache.
	line, err := CanonicalJSON(event)
	if err != nil {
		return PolicyDecision{}, err
	}
	if err := appendPrivateFile(k.ledger, append(line, '\n')); err != nil {
		return PolicyDecision{}, err
	}
	if err := verifyAnchoredFile(k.root, "identity-ledger.jsonl", k.ledger); err != nil {
		return PolicyDecision{}, quarantined(k.quarantine, err)
	}
	if err := verifyRootEntry(k.parent, k.rootName, k.root); err != nil {
		return PolicyDecision{}, quarantined(k.quarantine, err)
	}
	if err := k.replaceHead(projected); err != nil {
		return PolicyDecision{}, err
	}
	if err := verifyRootEntry(k.parent, k.rootName, k.root); err != nil {
		return PolicyDecision{}, quarantined(k.quarantine, err)
	}
	k.state = projected
	k.events = append(k.events, event)
	return decision, nil
}

// Preflight projects a mutation against the current state without persistent side effects.
func (k *Kernel) Preflight(m Mutation) (State, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if err := k.preflight(m); err != nil {
		return State{}, err
	}
	candidate := IdentityEvent{SchemaVersion: SchemaVersion, ID: "event:preflight", Sequence: k.state.EventCount, OccurredAt: DefaultGenesisTime, PreviousHash: k.state.HeadHash, Type: EventProposalAccepted, ProposalID: "proposal:preflight", ProposalDigest: strings.Repeat("0", 64), Mutation: m, Decision: &PolicyDecision{SchemaVersion: SchemaVersion, ProposalID: "proposal:preflight", ProposalDigest: strings.Repeat("0", 64), Reviewer: PolicyContext{SchemaVersion: SchemaVersion, ActorID: StewardSubjectID, Role: RoleAuthorizedReviewer, Source: SourceLocal, Purpose: "preflight"}, Decision: DecisionAccepted, DecidedAt: DefaultGenesisTime, Digest: strings.Repeat("0", 64)}}
	// The public preflight exposes only projection validity, not a synthetic ledger event.
	return projectMutation(k.state, candidate.Mutation)
}

func (k *Kernel) preflight(m Mutation) error {
	_, err := projectMutation(k.state, m)
	return err
}

func bootstrap(root *os.Root, cfg BootstrapConfig) error {
	record := expectedBootstrap(cfg)
	mutation := Mutation{SchemaVersion: SchemaVersion, Operation: MutationBootstrap, Bootstrap: &record}
	event := IdentityEvent{
		SchemaVersion: SchemaVersion,
		ID:            "event:bootstrap",
		Sequence:      0,
		OccurredAt:    record.GenesisTime,
		PreviousHash:  "",
		Type:          EventBootstrap,
		Mutation:      mutation,
	}
	hash, err := eventDigest(event)
	if err != nil {
		return err
	}
	event.Hash = hash
	state, err := applyEvent(State{}, event)
	if err != nil {
		return err
	}
	line, err := CanonicalJSON(event)
	if err != nil {
		return err
	}
	if err := writePrivateExclusive(root, "identity-ledger.jsonl", append(line, '\n')); err != nil {
		return err
	}
	if err := writePrivateExclusive(root, "head.json", append(mustCanonical(headFor(state)), '\n')); err != nil {
		return err
	}
	return nil
}

func expectedBootstrap(cfg BootstrapConfig) BootstrapRecord {
	steward := Subject{SchemaVersion: SchemaVersion, ID: StewardSubjectID, Kind: KindHumanSteward, DisplayName: "Dan", Immutable: true, Roles: []ActorRole{RoleAuthorizedReviewer, RoleHumanSteward}}
	root := Subject{SchemaVersion: SchemaVersion, ID: RootSubjectID, Kind: KindCoreAgent, DisplayName: "Deep Tree Echo Core", Immutable: true, Roles: []ActorRole{RoleCoreIdentity}}
	runtime := Subject{SchemaVersion: SchemaVersion, ID: RuntimeSubjectID, Kind: KindRuntimeAgent, DisplayName: "Deep Tree Echo Runtime", Immutable: true, Roles: []ActorRole{RoleRuntime}}
	return BootstrapRecord{
		SchemaVersion: SchemaVersion,
		GenesisTime:   cfg.GenesisTime,
		SourceCommit:  cfg.SourceCommit,
		Subjects:      []Subject{root, steward, runtime},
		Relations: []Relation{
			{SchemaVersion: SchemaVersion, ID: "relation:root-stewarded-by-dan", From: endpoint(root), Predicate: "stewarded_by", To: endpoint(steward), Truth: Truth{SchemaVersion: SchemaVersion, Value: TruthAsserted, EvidenceIDs: []string{}}},
			{SchemaVersion: SchemaVersion, ID: "relation:root-executes-as-runtime", From: endpoint(root), Predicate: "executes_as", To: endpoint(runtime), Truth: Truth{SchemaVersion: SchemaVersion, Value: TruthAsserted, EvidenceIDs: []string{}}},
		},
		AuthorizedReviewerIDs: append([]string(nil), cfg.AuthorizedReviewerIDs...),
	}
}

func mustDefaultBootstrap() BootstrapConfig {
	cfg, err := normalizeBootstrap(BootstrapConfig{})
	if err != nil {
		panic(err)
	}
	return cfg
}

func normalizeBootstrap(in BootstrapConfig) (BootstrapConfig, error) {
	out := in
	if out.GenesisTime == "" {
		out.GenesisTime = DefaultGenesisTime
	}
	if out.SourceCommit == "" {
		out.SourceCommit = DefaultSourceCommit
	}
	if out.AuthorizedReviewerIDs == nil {
		out.AuthorizedReviewerIDs = []string{StewardSubjectID}
	}
	if out.GenesisTime != DefaultGenesisTime || out.SourceCommit != DefaultSourceCommit {
		return BootstrapConfig{}, invalid("bootstrap genesis time and source commit are fixed")
	}
	if err := validateTime(out.GenesisTime); err != nil || !commitRE.MatchString(out.SourceCommit) {
		return BootstrapConfig{}, invalid("bootstrap fields")
	}
	if !sortedUniqueIDs(out.AuthorizedReviewerIDs) || !equalStrings(out.AuthorizedReviewerIDs, []string{StewardSubjectID}) {
		return BootstrapConfig{}, invalid("authorized reviewers are fixed to the human steward")
	}
	return out, nil
}

func verifyBootstrapState(state State, cfg BootstrapConfig) error {
	expected := expectedBootstrap(cfg)
	if state.GenesisTime != expected.GenesisTime || state.SourceCommit != expected.SourceCommit || !equalStrings(state.AuthorizedReviewerIDs, expected.AuthorizedReviewerIDs) {
		return fmt.Errorf("%w: bootstrap configuration differs", ErrTampered)
	}
	for _, s := range expected.Subjects {
		actual, ok := findSubject(state, s.ID)
		if !ok || !subjectsEqual(actual, s) {
			return fmt.Errorf("%w: protected bootstrap subject %s", ErrTampered, s.ID)
		}
	}
	for _, r := range expected.Relations {
		actual, ok := findRelation(state, r.ID)
		if !ok || !relationsEqual(actual, r) {
			return fmt.Errorf("%w: bootstrap relation %s", ErrTampered, r.ID)
		}
	}
	return nil
}

func loadVerified(kernel *Kernel, allowHeadRecovery bool) (State, []IdentityEvent, error) {
	data, err := readPrivateFileHandle(kernel.ledger, "identity-ledger.jsonl", maxLedgerBytes)
	if err != nil {
		return State{}, nil, fmt.Errorf("%w: read ledger: %v", ErrTampered, err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		return State{}, nil, fmt.Errorf("%w: ledger must end in one newline", ErrTampered)
	}
	lines := bytes.Split(data[:len(data)-1], []byte{'\n'})
	if len(lines) == 0 || len(lines[0]) == 0 {
		return State{}, nil, fmt.Errorf("%w: empty ledger", ErrTampered)
	}
	events := make([]IdentityEvent, 0, len(lines))
	for i, line := range lines {
		if len(line) == 0 {
			return State{}, nil, fmt.Errorf("%w: blank ledger line", ErrTampered)
		}
		var event IdentityEvent
		if err := decodeStrict(line, &event); err != nil {
			return State{}, nil, fmt.Errorf("%w: event %d: %v", ErrTampered, i, err)
		}
		canonical, err := CanonicalJSON(event)
		if err != nil || !bytes.Equal(line, canonical) {
			return State{}, nil, fmt.Errorf("%w: event %d is not canonical", ErrTampered, i)
		}
		events = append(events, event)
	}
	state, err := replay(events, kernel.reviewerKey)
	if err != nil {
		return State{}, nil, err
	}
	if kernel.head == nil {
		if !allowHeadRecovery {
			return State{}, nil, fmt.Errorf("%w: live kernel head descriptor unavailable", ErrTampered)
		}
		if err := kernel.replaceHead(state); err != nil {
			return State{}, nil, fmt.Errorf("%w: recover missing head: %v", ErrTampered, err)
		}
		return state, events, nil
	}
	if err := verifyAnchoredFile(kernel.root, "head.json", kernel.head); err != nil {
		return State{}, nil, err
	}
	headData, err := readPrivateFileHandle(kernel.head, "head.json", maxHeadBytes)
	if err != nil {
		return State{}, nil, fmt.Errorf("%w: read head: %v", ErrTampered, err)
	}
	var head headRecord
	if err := decodeStrict(bytes.TrimSuffix(headData, []byte{'\n'}), &head); err != nil {
		return State{}, nil, fmt.Errorf("%w: head: %v", ErrTampered, err)
	}
	canonicalHead, err := CanonicalJSON(head)
	if err != nil || !bytes.Equal(bytes.TrimSuffix(headData, []byte{'\n'}), canonicalHead) || (len(headData) > 0 && headData[len(headData)-1] != '\n') {
		return State{}, nil, fmt.Errorf("%w: non-canonical head", ErrTampered)
	}
	if err := validateHead(head); err != nil {
		return State{}, nil, fmt.Errorf("%w: invalid head", ErrTampered)
	}
	if headMatchesState(head, state) {
		return state, events, nil
	}
	if head.EventCount == 0 || head.EventCount >= uint64(len(events)) {
		return State{}, nil, fmt.Errorf("%w: head does not match ledger", ErrFork)
	}
	ancestor, err := replay(events[:head.EventCount], kernel.reviewerKey)
	if err != nil || !headMatchesState(head, ancestor) {
		return State{}, nil, fmt.Errorf("%w: head is not a valid ledger ancestor", ErrFork)
	}
	if !allowHeadRecovery {
		return State{}, nil, fmt.Errorf("%w: live kernel head differs from verified ledger", ErrStaleHead)
	}
	if err := kernel.replaceHead(state); err != nil {
		return State{}, nil, fmt.Errorf("%w: recover stale head: %v", ErrTampered, err)
	}
	return state, events, nil
}

func headMatchesState(head headRecord, state State) bool {
	return head.HeadHash == state.HeadHash && head.EventCount == state.EventCount && head.StateDigest == stateDigest(state)
}

func replay(events []IdentityEvent, reviewerKey ed25519.PublicKey) (State, error) {
	if len(events) == 0 {
		return State{}, fmt.Errorf("%w: no bootstrap event", ErrTampered)
	}
	var state State
	for index, event := range events {
		if err := validateEvent(event, true); err != nil {
			return State{}, fmt.Errorf("%w: event %d: %v", ErrTampered, index, err)
		}
		if event.Sequence != uint64(index) {
			return State{}, fmt.Errorf("%w: sequence %d", ErrFork, index)
		}
		if index == 0 {
			if event.Type != EventBootstrap || event.PreviousHash != "" {
				return State{}, fmt.Errorf("%w: invalid genesis", ErrFork)
			}
		} else {
			if event.Type == EventBootstrap || event.PreviousHash != state.HeadHash {
				return State{}, fmt.Errorf("%w: predecessor %d", ErrFork, index)
			}
			if err := verifyEventAuthorization(event, reviewerKey); err != nil {
				return State{}, fmt.Errorf("%w: event authorization %d: %w", ErrTampered, index, err)
			}
		}
		calculated, err := eventDigest(event)
		if err != nil || calculated != event.Hash {
			return State{}, fmt.Errorf("%w: event digest %d", ErrTampered, index)
		}
		next, err := applyEvent(state, event)
		if err != nil {
			return State{}, err
		}
		state = next
	}
	return state, nil
}

func applyEvent(state State, event IdentityEvent) (State, error) {
	if event.Type == EventBootstrap {
		if state.EventCount != 0 || event.Sequence != 0 || event.Mutation.Operation != MutationBootstrap || event.Mutation.Bootstrap == nil {
			return State{}, fmt.Errorf("%w: invalid bootstrap event", ErrInvalidContract)
		}
		return projectBootstrap(event.Mutation.Bootstrap, event.Hash)
	}
	if event.Type != EventProposalAccepted || event.Decision == nil || event.Decision.Decision != DecisionAccepted {
		return State{}, invalid("unsupported identity event")
	}
	if event.Sequence != state.EventCount || event.PreviousHash != state.HeadHash {
		return State{}, fmt.Errorf("%w: event head differs", ErrFork)
	}
	if err := validateDecision(*event.Decision, true); err != nil {
		return State{}, err
	}
	if event.ProposalID != event.Decision.ProposalID || event.ProposalDigest != event.Decision.ProposalDigest {
		return State{}, invalid("event decision linkage")
	}
	if err := validateReviewer(state, event.Decision.Reviewer); err != nil {
		return State{}, err
	}
	next, err := projectMutation(state, event.Mutation)
	if err != nil {
		return State{}, err
	}
	next.HeadHash = event.Hash
	next.EventCount = state.EventCount + 1
	return canonicalState(next), nil
}

func projectBootstrap(record *BootstrapRecord, eventHash string) (State, error) {
	if err := validateBootstrap(*record); err != nil {
		return State{}, err
	}
	state := State{SchemaVersion: SchemaVersion, GenesisTime: record.GenesisTime, SourceCommit: record.SourceCommit, Subjects: append([]Subject(nil), record.Subjects...), Relations: append([]Relation(nil), record.Relations...), Evidence: []EvidenceRef{}, AuthorizedReviewerIDs: append([]string(nil), record.AuthorizedReviewerIDs...), HeadHash: eventHash, EventCount: 1}
	state = canonicalState(state)
	if err := validateState(state); err != nil {
		return State{}, err
	}
	return state, nil
}

func projectMutation(state State, mutation Mutation) (State, error) {
	if err := validateMutation(mutation); err != nil {
		return State{}, err
	}
	next := cloneState(state)
	switch mutation.Operation {
	case MutationAddSubject:
		if mutation.Subject.Immutable ||
			(mutation.Subject.Kind != KindExternalAgent && mutation.Subject.Kind != KindService) ||
			!equalRoles(mutation.Subject.Roles, []ActorRole{RoleExternalContributor}) {
			return State{}, fmt.Errorf("%w: new core subjects are prohibited", ErrImmutableCore)
		}
		if _, exists := findSubject(next, mutation.Subject.ID); exists {
			return State{}, fmt.Errorf("%w: subject %s", ErrAlreadyExists, mutation.Subject.ID)
		}
		next.Subjects = append(next.Subjects, *mutation.Subject)
	case MutationAddEvidence:
		if _, exists := findEvidence(next, mutation.Evidence.ID); exists {
			return State{}, fmt.Errorf("%w: evidence %s", ErrAlreadyExists, mutation.Evidence.ID)
		}
		next.Evidence = append(next.Evidence, *mutation.Evidence)
	case MutationAddRelation:
		if _, exists := findRelation(next, mutation.Relation.ID); exists {
			return State{}, fmt.Errorf("%w: relation %s", ErrAlreadyExists, mutation.Relation.ID)
		}
		fromSubject, ok := findSubject(next, mutation.Relation.From.SubjectID)
		if !ok {
			return State{}, invalid("relation source subject does not exist")
		}
		toSubject, ok := findSubject(next, mutation.Relation.To.SubjectID)
		if !ok {
			return State{}, invalid("relation target subject does not exist")
		}
		if fromSubject.Kind != mutation.Relation.From.Kind || toSubject.Kind != mutation.Relation.To.Kind {
			return State{}, invalid("relation endpoint kind differs from subject")
		}
		for _, evidenceID := range mutation.Relation.Truth.EvidenceIDs {
			if _, ok := findEvidence(next, evidenceID); !ok {
				return State{}, invalid("relation evidence does not exist")
			}
		}
		if err := validateProjectedRelation(*mutation.Relation); err != nil {
			return State{}, err
		}
		next.Relations = append(next.Relations, *mutation.Relation)
	case MutationAmendSubject:
		return State{}, fmt.Errorf("%w: subject amendment is not supported", ErrImmutableCore)
	default:
		return State{}, invalid("unknown mutation")
	}
	next = canonicalState(next)
	if err := validateState(next); err != nil {
		return State{}, err
	}
	return next, nil
}

func validateSubject(s Subject) error {
	if s.SchemaVersion != SchemaVersion || !validID(s.ID) || !validText(s.DisplayName, 1, 120) || !validKind(s.Kind) || !sortedUniqueRoles(s.Roles) {
		return invalid("subject")
	}
	if s.Immutable && s.Kind != KindCoreAgent && s.Kind != KindHumanSteward && s.Kind != KindRuntimeAgent {
		return invalid("only fixed core kinds may be immutable")
	}
	return nil
}

func validateEndpoint(e Endpoint) error {
	if e.SchemaVersion != SchemaVersion || !validID(e.SubjectID) || !validKind(e.Kind) {
		return invalid("endpoint")
	}
	return nil
}

func validateTruth(t Truth) error {
	if t.SchemaVersion != SchemaVersion || !validTruth(t.Value) || !sortedUniqueIDs(t.EvidenceIDs) {
		return invalid("truth")
	}
	if t.Value == TruthVerified && len(t.EvidenceIDs) == 0 {
		return invalid("verified truth requires evidence")
	}
	return nil
}

func validateRelation(r Relation) error {
	if r.SchemaVersion != SchemaVersion || !validID(r.ID) || !validText(r.Predicate, 1, 80) || !identifierRE.MatchString(r.Predicate) || r.From.SubjectID == r.To.SubjectID || validateEndpoint(r.From) != nil || validateEndpoint(r.To) != nil || validateTruth(r.Truth) != nil {
		return invalid("relation")
	}
	return nil
}

func validateProjectedRelation(relation Relation) error {
	allowed := map[string]struct{}{
		"contradicts":   {},
		"derived_from":  {},
		"references":    {},
		"supports":      {},
		"verified_with": {},
	}
	if _, ok := allowed[relation.Predicate]; !ok {
		return fmt.Errorf("%w: relation predicate %s is not an evidence predicate", ErrUnauthorized, relation.Predicate)
	}
	protected := map[string]struct{}{RootSubjectID: {}, StewardSubjectID: {}, RuntimeSubjectID: {}}
	_, fromProtected := protected[relation.From.SubjectID]
	_, toProtected := protected[relation.To.SubjectID]
	if !fromProtected && !toProtected {
		return nil
	}
	if relation.Predicate != "verified_with" ||
		relation.From.SubjectID != RootSubjectID ||
		relation.To.SubjectID != RuntimeSubjectID ||
		relation.Truth.Value != TruthVerified {
		return fmt.Errorf("%w: protected subjects accept only evidence-backed root-to-runtime verification", ErrImmutableCore)
	}
	return nil
}

func validateEvidence(e EvidenceRef) error {
	if e.SchemaVersion != SchemaVersion || !validID(e.ID) || !validLocator(e.Locator) || !digestRE.MatchString(e.SHA256) || !validMediaType(e.MediaType) || !validSource(e.Source) {
		return invalid("evidence reference")
	}
	return nil
}

func validatePolicy(p PolicyContext) error {
	if p.SchemaVersion != SchemaVersion || !validID(p.ActorID) || !validRole(p.Role) || !validSource(p.Source) || !validText(p.Purpose, 1, 160) || containsSecretMarker(p.Purpose) {
		return invalid("policy context")
	}
	return nil
}

func validateBootstrap(b BootstrapRecord) error {
	if b.SchemaVersion != SchemaVersion || b.GenesisTime != DefaultGenesisTime || b.SourceCommit != DefaultSourceCommit || validateTime(b.GenesisTime) != nil || !commitRE.MatchString(b.SourceCommit) || !sortedUniqueIDs(b.AuthorizedReviewerIDs) || !equalStrings(b.AuthorizedReviewerIDs, []string{StewardSubjectID}) {
		return invalid("bootstrap record")
	}
	if len(b.Subjects) != 3 || len(b.Relations) != 2 {
		return invalid("bootstrap cardinality")
	}
	seen := make(map[string]bool, len(b.Subjects))
	for _, s := range b.Subjects {
		if validateSubject(s) != nil || seen[s.ID] {
			return invalid("bootstrap subject")
		}
		seen[s.ID] = true
	}
	for _, r := range b.Relations {
		if validateRelation(r) != nil || !seen[r.From.SubjectID] || !seen[r.To.SubjectID] {
			return invalid("bootstrap relation")
		}
	}
	return nil
}

func validateMutation(m Mutation) error {
	if m.SchemaVersion != SchemaVersion {
		return invalid("mutation version")
	}
	n := 0
	if m.Subject != nil {
		n++
	}
	if m.Relation != nil {
		n++
	}
	if m.Evidence != nil {
		n++
	}
	if m.Bootstrap != nil {
		n++
	}
	switch m.Operation {
	case MutationBootstrap:
		if n != 1 || m.Bootstrap == nil || m.TargetID != "" || validateBootstrap(*m.Bootstrap) != nil {
			return invalid("bootstrap mutation")
		}
	case MutationAddSubject:
		if n != 1 || m.Subject == nil || m.TargetID != "" || validateSubject(*m.Subject) != nil {
			return invalid("add subject mutation")
		}
	case MutationAddRelation:
		if n != 1 || m.Relation == nil || m.TargetID != "" || validateRelation(*m.Relation) != nil {
			return invalid("add relation mutation")
		}
	case MutationAddEvidence:
		if n != 1 || m.Evidence == nil || m.TargetID != "" || validateEvidence(*m.Evidence) != nil {
			return invalid("add evidence mutation")
		}
	case MutationAmendSubject:
		if n != 0 || !validID(m.TargetID) {
			return invalid("amend subject mutation")
		}
	default:
		return invalid("mutation operation")
	}
	return nil
}

func validateProposal(p Proposal, requireDigest bool) error {
	if p.SchemaVersion != SchemaVersion || !validID(p.ID) || validateTime(p.SubmittedAt) != nil || validatePolicy(p.Submitter) != nil || validateMutation(p.Mutation) != nil || p.Mutation.Operation == MutationBootstrap {
		return invalid("proposal")
	}
	if requireDigest {
		if !digestRE.MatchString(p.Digest) {
			return invalid("proposal digest")
		}
		expected, err := proposalDigest(p)
		if err != nil || expected != p.Digest {
			return fmt.Errorf("%w: proposal digest", ErrTampered)
		}
	} else if p.Digest != "" && !digestRE.MatchString(p.Digest) {
		return invalid("proposal digest")
	}
	return nil
}

func validateDecision(d PolicyDecision, requireDigest bool) error {
	if d.SchemaVersion != SchemaVersion || !validID(d.ProposalID) || !digestRE.MatchString(d.ProposalDigest) || validatePolicy(d.Reviewer) != nil || d.Decision != DecisionAccepted || validateTime(d.DecidedAt) != nil {
		return invalid("policy decision")
	}
	if requireDigest {
		if !digestRE.MatchString(d.Digest) || !signatureRE.MatchString(d.Signature) {
			return invalid("decision digest")
		}
		expected, err := decisionDigest(d)
		if err != nil || expected != d.Digest {
			return fmt.Errorf("%w: decision digest", ErrTampered)
		}
	} else if (d.Digest != "" && !digestRE.MatchString(d.Digest)) || (d.Signature != "" && !signatureRE.MatchString(d.Signature)) {
		return invalid("decision digest")
	}
	return nil
}

func validateEvent(e IdentityEvent, requireHash bool) error {
	if e.SchemaVersion != SchemaVersion || !validID(e.ID) || validateTime(e.OccurredAt) != nil || validateMutation(e.Mutation) != nil || (e.Sequence > 0 && !digestRE.MatchString(e.PreviousHash)) {
		return invalid("identity event")
	}
	if e.Sequence == 0 && e.PreviousHash != "" {
		return invalid("genesis previous hash")
	}
	switch e.Type {
	case EventBootstrap:
		if e.Sequence != 0 || e.ProposalID != "" || e.ProposalDigest != "" || e.Proposal != nil || e.Decision != nil || e.Mutation.Operation != MutationBootstrap {
			return invalid("bootstrap event")
		}
	case EventProposalAccepted:
		if e.Sequence == 0 || !validID(e.ProposalID) || !digestRE.MatchString(e.ProposalDigest) || e.Proposal == nil || e.Decision == nil || e.Mutation.Operation == MutationBootstrap || validateProposal(*e.Proposal, true) != nil || validateDecision(*e.Decision, true) != nil {
			return invalid("accepted proposal event")
		}
		if e.Proposal.ID != e.ProposalID || e.Proposal.Digest != e.ProposalDigest || e.Decision.ProposalID != e.ProposalID || e.Decision.ProposalDigest != e.ProposalDigest || !reflect.DeepEqual(e.Proposal.Mutation, e.Mutation) {
			return fmt.Errorf("%w: accepted proposal linkage", ErrTampered)
		}
	default:
		return invalid("event type")
	}
	if requireHash && !digestRE.MatchString(e.Hash) {
		return invalid("event hash")
	}
	if !requireHash && e.Hash != "" && !digestRE.MatchString(e.Hash) {
		return invalid("event hash")
	}
	return nil
}

func validateState(s State) error {
	if s.SchemaVersion != SchemaVersion || validateTime(s.GenesisTime) != nil || !commitRE.MatchString(s.SourceCommit) || !digestRE.MatchString(s.HeadHash) || s.EventCount == 0 || !sortedUniqueIDs(s.AuthorizedReviewerIDs) || !equalStrings(s.AuthorizedReviewerIDs, []string{StewardSubjectID}) {
		return invalid("state")
	}
	subjects := make(map[string]Subject, len(s.Subjects))
	for _, subject := range s.Subjects {
		if validateSubject(subject) != nil || subjects[subject.ID].ID != "" {
			return invalid("state subjects")
		}
		subjects[subject.ID] = subject
	}
	evidence := make(map[string]bool, len(s.Evidence))
	for _, item := range s.Evidence {
		if validateEvidence(item) != nil || evidence[item.ID] {
			return invalid("state evidence")
		}
		evidence[item.ID] = true
	}
	relations := make(map[string]bool, len(s.Relations))
	for _, relation := range s.Relations {
		if validateRelation(relation) != nil || relations[relation.ID] || subjects[relation.From.SubjectID].Kind != relation.From.Kind || subjects[relation.To.SubjectID].Kind != relation.To.Kind {
			return invalid("state relations")
		}
		for _, evidenceID := range relation.Truth.EvidenceIDs {
			if !evidence[evidenceID] {
				return invalid("state relation evidence")
			}
		}
		relations[relation.ID] = true
	}
	return nil
}

func validateHead(h headRecord) error {
	if h.SchemaVersion != SchemaVersion || !digestRE.MatchString(h.HeadHash) || h.EventCount == 0 || !digestRE.MatchString(h.StateDigest) {
		return invalid("head")
	}
	return nil
}

func validateCapsule(c Capsule, requireDigest bool) error {
	if c.SchemaVersion != SchemaVersion || c.Snapshot.SchemaVersion != SchemaVersion || validateState(c.Snapshot.State) != nil || !digestRE.MatchString(c.Snapshot.StateDigest) || stateDigest(c.Snapshot.State) != c.Snapshot.StateDigest || len(c.Ledger) == 0 || !digestRE.MatchString(c.LedgerDigest) || digestValue(c.Ledger) != c.LedgerDigest {
		return invalid("capsule")
	}
	key, err := decodeReviewerPublicKey(c.ReviewerPublicKey)
	if err != nil || (len(c.Ledger) > 1 && len(key) != ed25519.PublicKeySize) {
		return invalid("capsule reviewer public key")
	}
	if requireDigest {
		if !digestRE.MatchString(c.Digest) {
			return invalid("capsule digest")
		}
		expected, err := capsuleDigest(c)
		if err != nil || expected != c.Digest {
			return fmt.Errorf("%w: capsule digest", ErrTampered)
		}
	}
	return nil
}

func validateReviewer(state State, reviewer PolicyContext) error {
	if err := validatePolicy(reviewer); err != nil {
		return err
	}
	if reviewer.Role != RoleAuthorizedReviewer || reviewer.Source != SourceLocal || !containsString(state.AuthorizedReviewerIDs, reviewer.ActorID) {
		return fmt.Errorf("%w: reviewer", ErrUnauthorized)
	}
	subject, ok := findSubject(state, reviewer.ActorID)
	if !ok || !containsRole(subject.Roles, RoleAuthorizedReviewer) || subject.Kind != KindHumanSteward {
		return fmt.Errorf("%w: reviewer subject", ErrUnauthorized)
	}
	return nil
}

func (k *Kernel) validateActor(actor PolicyContext, reviewer bool) error {
	if err := validatePolicy(actor); err != nil {
		return err
	}
	if reviewer {
		return validateReviewer(k.state, actor)
	}
	if actor.Source == SourceExternal || actor.Source == SourceExperimental {
		return nil
	}
	subject, ok := findSubject(k.state, actor.ActorID)
	if !ok || !containsRole(subject.Roles, actor.Role) {
		return fmt.Errorf("%w: actor role", ErrUnauthorized)
	}
	return nil
}

// CanonicalJSON serializes standard-library JSON deterministically for this closed, map-free contract family.
func CanonicalJSON(value any) ([]byte, error) {
	if value == nil {
		return nil, invalid("nil canonical value")
	}
	return json.Marshal(value)
}

func eventDigest(event IdentityEvent) (string, error) {
	event.Hash = ""
	data, err := CanonicalJSON(event)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}

func proposalDigest(proposal Proposal) (string, error) {
	proposal.Digest = ""
	data, err := CanonicalJSON(proposal)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}

func decisionDigest(decision PolicyDecision) (string, error) {
	decision.Digest = ""
	decision.Signature = ""
	data, err := CanonicalJSON(decision)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}

func eventAuthorizationBytes(event IdentityEvent) ([]byte, error) {
	event.Hash = ""
	if event.Decision == nil {
		return nil, invalid("event authorization decision")
	}
	decision := *event.Decision
	decision.Signature = ""
	event.Decision = &decision
	return CanonicalJSON(event)
}

func verifyEventAuthorization(event IdentityEvent, publicKey ed25519.PublicKey) error {
	if len(publicKey) != ed25519.PublicKeySize || event.Decision == nil {
		return fmt.Errorf("%w: reviewer public key unavailable", ErrUnauthorized)
	}
	signature, err := hex.DecodeString(event.Decision.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("%w: reviewer signature encoding", ErrTampered)
	}
	message, err := eventAuthorizationBytes(event)
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, message, signature) {
		return fmt.Errorf("%w: reviewer signature", ErrTampered)
	}
	return nil
}

func decodeReviewerPublicKey(encoded string) (ed25519.PublicKey, error) {
	if encoded == "" {
		return nil, nil
	}
	decoded, err := hex.DecodeString(encoded)
	if err != nil || len(decoded) != ed25519.PublicKeySize {
		return nil, invalid("reviewer public key")
	}
	return ed25519.PublicKey(append([]byte(nil), decoded...)), nil
}

func cloneEvents(events []IdentityEvent) ([]IdentityEvent, error) {
	data, err := CanonicalJSON(events)
	if err != nil {
		return nil, err
	}
	var clone []IdentityEvent
	if err := decodeStrict(data, &clone); err != nil {
		return nil, err
	}
	return clone, nil
}

func capsuleDigest(capsule Capsule) (string, error) {
	capsule.Digest = ""
	data, err := CanonicalJSON(capsule)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}
func stateDigest(state State) string { return digestValue(canonicalState(state)) }
func digestValue(value any) string {
	data, err := CanonicalJSON(value)
	if err != nil {
		return ""
	}
	return digestBytes(data)
}
func digestBytes(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func canonicalState(state State) State {
	state.Subjects = append([]Subject{}, state.Subjects...)
	state.Relations = append([]Relation{}, state.Relations...)
	state.Evidence = append([]EvidenceRef{}, state.Evidence...)
	state.AuthorizedReviewerIDs = append([]string{}, state.AuthorizedReviewerIDs...)
	for i := range state.Subjects {
		state.Subjects[i].Roles = append([]ActorRole(nil), state.Subjects[i].Roles...)
	}
	for i := range state.Relations {
		state.Relations[i].Truth.EvidenceIDs = append([]string{}, state.Relations[i].Truth.EvidenceIDs...)
	}
	sort.Slice(state.Subjects, func(i, j int) bool { return state.Subjects[i].ID < state.Subjects[j].ID })
	sort.Slice(state.Relations, func(i, j int) bool { return state.Relations[i].ID < state.Relations[j].ID })
	sort.Slice(state.Evidence, func(i, j int) bool { return state.Evidence[i].ID < state.Evidence[j].ID })
	sort.Strings(state.AuthorizedReviewerIDs)
	return state
}
func cloneState(state State) State { return canonicalState(state) }
func headFor(state State) headRecord {
	return headRecord{SchemaVersion: SchemaVersion, HeadHash: state.HeadHash, EventCount: state.EventCount, StateDigest: stateDigest(state)}
}

func (k *Kernel) replaceHead(state State) error {
	if k.head != nil {
		if err := verifyAnchoredFile(k.root, "head.json", k.head); err != nil {
			return err
		}
	}
	if err := writePrivateAtomic(k.root, "head.json", append(mustCanonical(headFor(state)), '\n')); err != nil {
		return err
	}
	next, err := openPrivateRegular(k.root, "head.json", os.O_RDONLY)
	if err != nil {
		return err
	}
	previous := k.head
	k.head = next
	if previous != nil {
		_ = previous.Close()
	}
	return nil
}

func mustCanonical(value any) []byte {
	data, err := CanonicalJSON(value)
	if err != nil {
		panic(err)
	}
	return data
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("trailing JSON data")
	}
	return nil
}

func readProposal(root *os.Root, name string) (Proposal, error) {
	if exists, err := regularPrivateExists(root, name); err != nil {
		return Proposal{}, err
	} else if !exists {
		return Proposal{}, fmt.Errorf("%w: proposal", ErrNotFound)
	}
	data, err := readPrivateFile(root, name, maxProposalBytes)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Proposal{}, fmt.Errorf("%w: proposal", ErrNotFound)
		}
		return Proposal{}, err
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		return Proposal{}, fmt.Errorf("%w: proposal newline", ErrTampered)
	}
	line := data[:len(data)-1]
	var proposal Proposal
	if err := decodeStrict(line, &proposal); err != nil {
		return Proposal{}, fmt.Errorf("%w: proposal: %v", ErrTampered, err)
	}
	canonical, err := CanonicalJSON(proposal)
	if err != nil || !bytes.Equal(line, canonical) {
		return Proposal{}, fmt.Errorf("%w: proposal non-canonical", ErrTampered)
	}
	if err := validateProposal(proposal, true); err != nil {
		return Proposal{}, err
	}
	return proposal, nil
}

func proposalAccepted(events []IdentityEvent, proposalID string) bool {
	for _, event := range events {
		if event.ProposalID == proposalID {
			return true
		}
	}
	return false
}

func pendingProposalCount(root *os.Root, events []IdentityEvent) uint64 {
	proposalDirectory, err := root.Open(".")
	if err != nil {
		return 0
	}
	defer proposalDirectory.Close()
	entries, err := proposalDirectory.ReadDir(-1)
	if err != nil {
		return 0
	}
	var count uint64
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		proposal, err := readProposal(root, entry.Name())
		if err == nil && !proposalAccepted(events, proposal.ID) {
			count++
		}
	}
	return count
}
func (k *Kernel) proposalCount() uint64 { return pendingProposalCount(k.proposals, k.events) }
func proposalName(proposalID string) string {
	return "proposals/" + proposalFileName(proposalID)
}
func proposalFileName(proposalID string) string { return digestBytes([]byte(proposalID)) + ".json" }

func ensurePrivateDir(path string) error {
	if err := rejectSymlinkComponents(path); err != nil {
		return err
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("%w: directory %s is not private", ErrUnauthorized, path)
	}
	return nil
}

func ensurePrivateRootDir(root *os.Root, name string) error {
	if err := root.MkdirAll(name, 0o700); err != nil {
		return err
	}
	info, err := root.Lstat(name)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("%w: directory %s is not private", ErrUnauthorized, name)
	}
	return nil
}

func openPrivateSubroot(root *os.Root, name string) (*os.Root, error) {
	pathInfo, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if pathInfo.Mode()&os.ModeSymlink != 0 || !pathInfo.IsDir() || pathInfo.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("%w: directory %s is not private", ErrUnauthorized, name)
	}
	child, err := root.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	opened, err := child.Open(".")
	if err != nil {
		_ = child.Close()
		return nil, err
	}
	openedInfo, statErr := opened.Stat()
	closeErr := opened.Close()
	if statErr != nil || closeErr != nil || !os.SameFile(pathInfo, openedInfo) {
		_ = child.Close()
		return nil, fmt.Errorf("%w: directory changed while anchoring %s", ErrTampered, name)
	}
	return child, nil
}

func verifyRootEntry(parent *os.Root, name string, root *os.Root) error {
	pathInfo, err := parent.Lstat(name)
	if err != nil {
		return fmt.Errorf("%w: core-self root entry unavailable: %v", ErrTampered, err)
	}
	if pathInfo.Mode()&os.ModeSymlink != 0 || !pathInfo.IsDir() || pathInfo.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("%w: core-self root entry is not a private directory", ErrTampered)
	}
	opened, err := root.Open(".")
	if err != nil {
		return err
	}
	openedInfo, statErr := opened.Stat()
	closeErr := opened.Close()
	if statErr != nil || closeErr != nil || !os.SameFile(pathInfo, openedInfo) {
		return fmt.Errorf("%w: core-self root entry was replaced", ErrTampered)
	}
	return nil
}

func openDirectoryHandle(root *os.Root) (*os.File, error) {
	file, err := root.Open(".")
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil || !info.IsDir() {
		_ = file.Close()
		return nil, fmt.Errorf("%w: lock anchor is not a directory", ErrTampered)
	}
	return file, nil
}

func rejectSymlinkComponents(path string) error {
	for current := filepath.Clean(path); ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: symbolic link in persistence path %s", ErrUnauthorized, current)
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return nil
		}
	}
}

func regularPrivateExists(root *os.Root, name string) (bool, error) {
	info, err := root.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return false, fmt.Errorf("%w: non-private regular file %s", ErrTampered, name)
	}
	return true, nil
}

func openPrivateRegular(root *os.Root, name string, flag int) (*os.File, error) {
	pathInfo, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !pathInfo.Mode().IsRegular() || pathInfo.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("%w: non-private regular file %s", ErrTampered, name)
	}
	file, err := root.OpenFile(name, flag, 0o600)
	if err != nil {
		return nil, err
	}
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || opened.Mode().Perm()&0o077 != 0 || !os.SameFile(pathInfo, opened) {
		_ = file.Close()
		return nil, fmt.Errorf("%w: file changed while anchoring %s", ErrTampered, name)
	}
	return file, nil
}

func verifyAnchoredFile(root *os.Root, name string, file *os.File) error {
	pathInfo, err := root.Lstat(name)
	if err != nil {
		return fmt.Errorf("%w: anchored file name unavailable %s: %v", ErrTampered, name, err)
	}
	opened, err := file.Stat()
	if err != nil || !pathInfo.Mode().IsRegular() || !opened.Mode().IsRegular() || opened.Mode().Perm()&0o077 != 0 || !os.SameFile(pathInfo, opened) {
		return fmt.Errorf("%w: anchored file replaced %s", ErrTampered, name)
	}
	return nil
}

func readPrivateFile(root *os.Root, name string, maximum int64) ([]byte, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("%w: non-private regular file %s", ErrTampered, name)
	}
	if info.Size() > maximum {
		return nil, fmt.Errorf("%w: file exceeds size bound %s", ErrTampered, name)
	}
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !os.SameFile(info, opened) {
		return nil, fmt.Errorf("%w: file changed while opening %s", ErrTampered, name)
	}
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maximum {
		return nil, fmt.Errorf("%w: file exceeds size bound %s", ErrTampered, name)
	}
	return data, nil
}

func readPrivateFileHandle(file *os.File, name string, maximum int64) ([]byte, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("%w: non-private regular file %s", ErrTampered, name)
	}
	if info.Size() > maximum {
		return nil, fmt.Errorf("%w: file exceeds size bound %s", ErrTampered, name)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maximum {
		return nil, fmt.Errorf("%w: file exceeds size bound %s", ErrTampered, name)
	}
	return data, nil
}

func writePrivateExclusive(root *os.Root, name string, data []byte) error {
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return syncRoot(root)
}

func writePrivateAtomic(root *os.Root, name string, data []byte) error {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return err
	}
	temporary := ".identity-head-" + hex.EncodeToString(random)
	file, err := root.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer root.Remove(temporary)
	err = file.Chmod(0o600)
	if err == nil {
		_, err = file.Write(data)
	}
	if err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := root.Rename(temporary, name); err != nil {
		return err
	}
	return syncRoot(root)
}

func syncRoot(root *os.Root) error {
	directory, err := root.Open(".")
	if err != nil {
		return err
	}
	err = directory.Sync()
	closeErr := directory.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func appendPrivateFile(file *os.File, data []byte) error {
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("%w: ledger is not private", ErrTampered)
	}
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	return file.Sync()
}

func acquireKernelLock(kernel *Kernel) (func(), error) {
	if err := lockFile(kernel.lock); err != nil {
		return nil, err
	}
	if err := verifyRootEntry(kernel.parent, kernel.rootName, kernel.root); err != nil {
		_ = unlockFile(kernel.lock)
		return nil, err
	}
	if err := verifyAnchoredFile(kernel.root, "identity-ledger.jsonl", kernel.ledger); err != nil {
		_ = unlockFile(kernel.lock)
		return nil, err
	}
	return func() { _ = unlockFile(kernel.lock) }, nil
}

func quarantine(root *os.Root, cause error) error {
	return writePrivateExclusive(root, "QUARANTINED", []byte("identity ledger rejected: "+cause.Error()+"\n"))
}

func quarantined(root *os.Root, cause error) error {
	if err := quarantine(root, cause); err != nil {
		return errors.Join(cause, fmt.Errorf("write quarantine marker: %w", err))
	}
	return cause
}

func validID(value string) bool {
	return utf8.ValidString(value) && identifierRE.MatchString(value) && !containsSecretMarker(value)
}

func validText(value string, minimum, maximum int) bool {
	return utf8.ValidString(value) && len(value) >= minimum && len(value) <= maximum && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "\r\n\x00") && !containsSecretMarker(value)
}

func validLocator(value string) bool {
	return validText(value, 1, 512) && !containsSecretMarker(value) && !strings.Contains(value, "@")
}

func validMediaType(value string) bool {
	return validText(value, 3, 127) && strings.Count(value, "/") == 1 && !strings.ContainsAny(value, " ;")
}

func validKind(value SubjectKind) bool {
	switch value {
	case KindCoreAgent, KindHumanSteward, KindRuntimeAgent, KindExternalAgent, KindService:
		return true
	}
	return false
}

func validRole(value ActorRole) bool {
	switch value {
	case RoleCoreIdentity, RoleHumanSteward, RoleAuthorizedReviewer, RoleRuntime, RoleExternalContributor:
		return true
	}
	return false
}

func validSource(value SourceClass) bool {
	return value == SourceLocal || value == SourceExternal || value == SourceExperimental
}
func validTruth(value TruthValue) bool { return value == TruthAsserted || value == TruthVerified }
func validateTime(value string) error {
	if !utf8.ValidString(value) {
		return invalid("time encoding")
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil || parsed.UTC().Format(time.RFC3339) != value {
		return invalid("time")
	}
	return nil
}

func containsSecretMarker(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{"secret", "password", "passwd", "api_key", "apikey", "bearer", "authorization", "token=", "private_key", "credential"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func sortedUniqueIDs(values []string) bool {
	if values == nil {
		return false
	}
	for i, value := range values {
		if !validID(value) || (i > 0 && values[i-1] >= value) {
			return false
		}
	}
	return true
}

func sortedUniqueRoles(values []ActorRole) bool {
	if len(values) == 0 {
		return false
	}
	for i, value := range values {
		if !validRole(value) || (i > 0 && values[i-1] >= value) {
			return false
		}
	}
	return true
}

func containsString(values []string, expected string) bool {
	i := sort.SearchStrings(values, expected)
	return i < len(values) && values[i] == expected
}

func containsRole(values []ActorRole, expected ActorRole) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func endpoint(subject Subject) Endpoint {
	return Endpoint{SchemaVersion: SchemaVersion, SubjectID: subject.ID, Kind: subject.Kind}
}

func findSubject(state State, id string) (Subject, bool) {
	for _, subject := range state.Subjects {
		if subject.ID == id {
			return subject, true
		}
	}
	return Subject{}, false
}

func findRelation(state State, id string) (Relation, bool) {
	for _, relation := range state.Relations {
		if relation.ID == id {
			return relation, true
		}
	}
	return Relation{}, false
}

func findEvidence(state State, id string) (EvidenceRef, bool) {
	for _, evidence := range state.Evidence {
		if evidence.ID == id {
			return evidence, true
		}
	}
	return EvidenceRef{}, false
}
func equalStrings(left, right []string) bool   { return reflect.DeepEqual(left, right) }
func equalRoles(left, right []ActorRole) bool  { return reflect.DeepEqual(left, right) }
func subjectsEqual(left, right Subject) bool   { return reflect.DeepEqual(left, right) }
func relationsEqual(left, right Relation) bool { return reflect.DeepEqual(left, right) }
func statesEqual(left, right State) bool {
	return reflect.DeepEqual(canonicalState(left), canonicalState(right))
}
func invalid(field string) error { return fmt.Errorf("%w: %s", ErrInvalidContract, field) }
