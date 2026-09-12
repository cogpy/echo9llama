package coreself

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const testTime = "2026-09-12T03:30:56Z"

func testReviewerPrivateKey() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x42}, ed25519.SeedSize))
}

func testBootstrapConfig() BootstrapConfig {
	publicKey := testReviewerPrivateKey().Public().(ed25519.PublicKey)
	return BootstrapConfig{ReviewerPublicKey: hex.EncodeToString(publicKey)}
}

func externalSubmitter() PolicyContext {
	return PolicyContext{
		SchemaVersion: SchemaVersion,
		ActorID:       "external:source:test",
		Role:          RoleExternalContributor,
		Source:        SourceExternal,
		Purpose:       "submit identity evidence",
	}
}

func stewardReviewer() PolicyContext {
	return PolicyContext{
		SchemaVersion: SchemaVersion,
		ActorID:       StewardSubjectID,
		Role:          RoleAuthorizedReviewer,
		Source:        SourceLocal,
		Purpose:       "review identity proposal",
	}
}

func evidenceProposal(id, evidenceID string) Proposal {
	return Proposal{
		SchemaVersion: SchemaVersion,
		ID:            id,
		SubmittedAt:   testTime,
		Submitter:     externalSubmitter(),
		Mutation: Mutation{
			SchemaVersion: SchemaVersion,
			Operation:     MutationAddEvidence,
			Evidence: &EvidenceRef{
				SchemaVersion: SchemaVersion,
				ID:            evidenceID,
				Locator:       "sha256:identity-evidence",
				SHA256:        strings.Repeat("a", 64),
				MediaType:     "application/json",
				Source:        SourceExternal,
			},
		},
	}
}

func relationProposal(id string, relation Relation) Proposal {
	return Proposal{
		SchemaVersion: SchemaVersion,
		ID:            id,
		SubmittedAt:   testTime,
		Submitter:     externalSubmitter(),
		Mutation: Mutation{
			SchemaVersion: SchemaVersion,
			Operation:     MutationAddRelation,
			Relation:      &relation,
		},
	}
}

func mustOpen(t *testing.T, directory string) *Kernel {
	t.Helper()
	kernel, err := Open(directory, testBootstrapConfig())
	if err != nil {
		t.Fatalf("Open(%s): %v", directory, err)
	}
	return kernel
}

func mustSubmit(t *testing.T, kernel *Kernel, proposal Proposal) Proposal {
	t.Helper()
	stored, err := kernel.submitProposal(proposal)
	if err != nil {
		t.Fatalf("submitProposal(%s): %v", proposal.ID, err)
	}
	if stored.Digest == "" {
		t.Fatalf("proposal %s was returned without digest", proposal.ID)
	}
	return stored
}

func mustAccept(t *testing.T, kernel *Kernel, proposalID string) PolicyDecision {
	t.Helper()
	decision, err := kernel.acceptProposal(proposalID, stewardReviewer(), testReviewerPrivateKey())
	if err != nil {
		t.Fatalf("AcceptProposal(%s): %v", proposalID, err)
	}
	return decision
}

func TestDeterministicBootstrapAndReplay(t *testing.T) {
	left := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	right := mustOpen(t, filepath.Join(t.TempDir(), "identity"))

	leftSnapshot := left.StateSnapshot()
	rightSnapshot := right.StateSnapshot()
	if !reflect.DeepEqual(leftSnapshot, rightSnapshot) {
		t.Fatalf("bootstrap snapshots differ:\nleft=%+v\nright=%+v", leftSnapshot, rightSnapshot)
	}
	for _, id := range []string{RootSubjectID, StewardSubjectID, RuntimeSubjectID} {
		if _, ok := findSubject(leftSnapshot.State, id); !ok {
			t.Fatalf("missing bootstrap subject %q", id)
		}
	}

	leftLedger, err := os.ReadFile(filepath.Join(left.directory, "identity-ledger.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	rightLedger, err := os.ReadFile(filepath.Join(right.directory, "identity-ledger.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(leftLedger, rightLedger) {
		t.Fatalf("deterministic bootstrap ledger bytes differ")
	}
	if err := VerifyCapsule(mustCapsule(t, left)); err != nil {
		t.Fatalf("bootstrap capsule verification: %v", err)
	}

	stored := mustSubmit(t, left, evidenceProposal("proposal:evidence-one", "evidence:one"))
	mustAccept(t, left, stored.ID)
	reopened := mustOpen(t, left.directory)
	if !reflect.DeepEqual(left.StateSnapshot(), reopened.StateSnapshot()) {
		t.Fatal("reopened state is not replay-equivalent")
	}
	if got := reopened.Status().EventCount; got != 2 {
		t.Fatalf("event count = %d, want 2", got)
	}
}

func TestUnprojectableEventRejectedBeforePersistence(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	before, err := os.ReadFile(filepath.Join(kernel.directory, "identity-ledger.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	missing := Subject{SchemaVersion: SchemaVersion, ID: "external:agent:missing", Kind: KindExternalAgent}
	root, _ := findSubject(kernel.StateSnapshot().State, RootSubjectID)
	proposal := relationProposal("proposal:bad-edge", Relation{
		SchemaVersion: SchemaVersion,
		ID:            "relation:bad-edge",
		From:          endpoint(root),
		Predicate:     "relates_to",
		To:            endpoint(missing),
		Truth:         Truth{SchemaVersion: SchemaVersion, Value: TruthAsserted, EvidenceIDs: []string{}},
	})
	if _, err := kernel.submitProposal(proposal); !errors.Is(err, ErrInvalidContract) {
		t.Fatalf("SubmitProposal error = %v, want unprojectable invalid contract", err)
	}
	after, err := os.ReadFile(filepath.Join(kernel.directory, "identity-ledger.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("ledger changed before rejecting unprojectable proposal")
	}
	if _, err := os.Stat(filepath.Join(kernel.directory, filepath.FromSlash(proposalName(proposal.ID)))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unprojectable proposal was persisted: %v", err)
	}
}

func TestStaleHeadAndForkQuarantine(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	first := mustOpen(t, directory)
	second := mustOpen(t, directory)
	proposal := mustSubmit(t, first, evidenceProposal("proposal:stale", "evidence:stale"))
	mustAccept(t, first, proposal.ID)
	if _, err := second.acceptProposal(proposal.ID, stewardReviewer(), testReviewerPrivateKey()); !errors.Is(err, ErrTampered) {
		t.Fatalf("stale live-kernel head replacement error = %v, want ErrTampered", err)
	}

	head := filepath.Join(directory, "head.json")
	if err := os.WriteFile(head, []byte(`{"schema_version":"1.0.0","head_hash":"`+strings.Repeat("0", 64)+`","event_count":2,"state_digest":"`+strings.Repeat("0", 64)+`"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(directory, testBootstrapConfig()); !errors.Is(err, ErrFork) {
		t.Fatalf("forked head error = %v, want ErrFork", err)
	}
	if _, err := os.Stat(filepath.Join(directory, "quarantine", "QUARANTINED")); err != nil {
		t.Fatalf("fork was not quarantined: %v", err)
	}
}

func TestImmutableCoreAndReviewerGate(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	immutable := Proposal{
		SchemaVersion: SchemaVersion,
		ID:            "proposal:mutate-root",
		SubmittedAt:   testTime,
		Submitter:     externalSubmitter(),
		Mutation: Mutation{
			SchemaVersion: SchemaVersion,
			Operation:     MutationAmendSubject,
			TargetID:      RootSubjectID,
		},
	}
	if _, err := kernel.submitProposal(immutable); !errors.Is(err, ErrImmutableCore) {
		t.Fatalf("immutable mutation error = %v, want ErrImmutableCore", err)
	}

	stored := mustSubmit(t, kernel, evidenceProposal("proposal:review-gate", "evidence:review-gate"))
	unauthorized := PolicyContext{
		SchemaVersion: SchemaVersion,
		ActorID:       RuntimeSubjectID,
		Role:          RoleRuntime,
		Source:        SourceLocal,
		Purpose:       "attempt review",
	}
	if _, err := kernel.acceptProposal(stored.ID, unauthorized, testReviewerPrivateKey()); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("unauthorized reviewer error = %v, want ErrUnauthorized", err)
	}
	if got := kernel.Status().EventCount; got != 1 {
		t.Fatalf("unauthorized review changed event count to %d", got)
	}
	mustAccept(t, kernel, stored.ID)
}

func TestProposalCountReportsOnlyPendingValidatedProposals(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	stored := mustSubmit(t, kernel, evidenceProposal("proposal:pending-count", "evidence:pending-count"))
	if got := kernel.Status().ProposalCount; got != 1 {
		t.Fatalf("pending proposal count = %d, want 1", got)
	}
	mustAccept(t, kernel, stored.ID)
	if got := kernel.Status().ProposalCount; got != 0 {
		t.Fatalf("accepted proposal remained pending: %d", got)
	}
}

func TestTypedHypergraphIntegrity(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	root, _ := findSubject(kernel.StateSnapshot().State, RootSubjectID)
	runtime, _ := findSubject(kernel.StateSnapshot().State, RuntimeSubjectID)

	wrongKind := endpoint(runtime)
	wrongKind.Kind = KindExternalAgent
	badKind := relationProposal("proposal:wrong-kind", Relation{
		SchemaVersion: SchemaVersion,
		ID:            "relation:wrong-kind",
		From:          endpoint(root),
		Predicate:     "references",
		To:            wrongKind,
		Truth:         Truth{SchemaVersion: SchemaVersion, Value: TruthAsserted, EvidenceIDs: []string{}},
	})
	if _, err := kernel.submitProposal(badKind); !errors.Is(err, ErrInvalidContract) {
		t.Fatalf("wrong endpoint kind error = %v, want ErrInvalidContract", err)
	}

	verifiedWithoutEvidence := relationProposal("proposal:missing-evidence", Relation{
		SchemaVersion: SchemaVersion,
		ID:            "relation:missing-evidence",
		From:          endpoint(root),
		Predicate:     "verified_with",
		To:            endpoint(runtime),
		Truth:         Truth{SchemaVersion: SchemaVersion, Value: TruthVerified, EvidenceIDs: []string{"evidence:not-present"}},
	})
	if _, err := kernel.submitProposal(verifiedWithoutEvidence); !errors.Is(err, ErrInvalidContract) {
		t.Fatalf("missing evidence error = %v, want ErrInvalidContract", err)
	}

	evidence := mustSubmit(t, kernel, evidenceProposal("proposal:graph-evidence", "evidence:graph"))
	mustAccept(t, kernel, evidence.ID)
	authorityEdge := relationProposal("proposal:authority-edge", Relation{
		SchemaVersion: SchemaVersion,
		ID:            "relation:authority-edge",
		From:          endpoint(root),
		Predicate:     "authorizes",
		To:            endpoint(runtime),
		Truth:         Truth{SchemaVersion: SchemaVersion, Value: TruthVerified, EvidenceIDs: []string{"evidence:graph"}},
	})
	if _, err := kernel.submitProposal(authorityEdge); !errors.Is(err, ErrImmutableCore) && !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("authority-like relation error = %v, want protected semantic rejection", err)
	}
	validRelation := relationProposal("proposal:valid-edge", Relation{
		SchemaVersion: SchemaVersion,
		ID:            "relation:validated-edge",
		From:          endpoint(root),
		Predicate:     "verified_with",
		To:            endpoint(runtime),
		Truth:         Truth{SchemaVersion: SchemaVersion, Value: TruthVerified, EvidenceIDs: []string{"evidence:graph"}},
	})
	mustSubmit(t, kernel, validRelation)
	mustAccept(t, kernel, validRelation.ID)
	state := kernel.StateSnapshot().State
	if _, ok := findRelation(state, "relation:validated-edge"); !ok {
		t.Fatal("verified relation was not projected")
	}
	if err := validateState(state); err != nil {
		t.Fatalf("projected graph is invalid: %v", err)
	}
}

func TestCapsuleVerificationAndTamperDetection(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:capsule", "evidence:capsule"))
	mustAccept(t, kernel, proposal.ID)
	capsule := mustCapsule(t, kernel)
	if err := kernel.VerifyCapsule(capsule); err != nil {
		t.Fatalf("VerifyCapsule: %v", err)
	}

	tampered := capsule
	tampered.Ledger = append([]IdentityEvent(nil), capsule.Ledger...)
	tampered.Ledger[0].Hash = strings.Repeat("f", 64)
	if err := kernel.VerifyCapsule(tampered); err == nil {
		t.Fatal("tampered capsule verified")
	}
	tampered = capsule
	tampered.Snapshot.State.Subjects = append([]Subject(nil), capsule.Snapshot.State.Subjects...)
	tampered.Snapshot.State.Subjects[0].DisplayName = "Changed"
	if err := kernel.VerifyCapsule(tampered); err == nil {
		t.Fatal("snapshot-tampered capsule verified")
	}
}

func TestAcceptedCapsuleRequiresExternalReviewerTrustAnchor(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:capsule-anchor", "evidence:capsule-anchor"))
	mustAccept(t, kernel, proposal.ID)
	capsule := mustCapsule(t, kernel)
	if err := VerifyCapsule(capsule); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("self-anchored accepted capsule error = %v, want ErrUnauthorized", err)
	}
	if err := VerifyCapsuleWithReviewerKey(capsule, strings.Repeat("d", 64)); !errors.Is(err, ErrTampered) {
		t.Fatalf("wrong external reviewer key error = %v, want ErrTampered", err)
	}
	if err := VerifyCapsuleWithReviewerKey(capsule, hex.EncodeToString(testReviewerPrivateKey().Public().(ed25519.PublicKey))); err != nil {
		t.Fatalf("externally anchored capsule verification: %v", err)
	}
}

func TestPrivatePermissionsAndProposalSeparation(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:separate", "evidence:separate"))
	storedProposalPath := filepath.Join(kernel.directory, filepath.FromSlash(proposalName(proposal.ID)))
	for _, path := range []string{filepath.Join(kernel.directory, "identity-ledger.jsonl"), filepath.Join(kernel.directory, "head.json"), storedProposalPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("%s permissions = %o, want 0600", path, got)
		}
	}
	for _, path := range []string{kernel.directory, filepath.Join(kernel.directory, "proposals"), filepath.Join(kernel.directory, "quarantine")} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0o700 {
			t.Fatalf("%s permissions = %o, want 0700", path, got)
		}
	}
	ledger, err := os.ReadFile(filepath.Join(kernel.directory, "identity-ledger.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(ledger, []byte(proposal.ID)) || bytes.Contains(ledger, []byte(proposal.Digest)) {
		t.Fatal("pending proposal leaked into ledger")
	}
	if _, err := readProposal(kernel.root, proposalName(proposal.ID)); err != nil {
		t.Fatalf("separate proposal cannot be independently verified: %v", err)
	}
}

func TestProposalIDsCannotEscapeProposalStoreOrImpersonatePrivilegedRoles(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	proposal := evidenceProposal("a/../../../proposal:escape", "evidence:escape")
	stored := mustSubmit(t, kernel, proposal)
	if _, err := os.Stat(filepath.Join(directory, filepath.FromSlash(proposalName(stored.ID)))); err != nil {
		t.Fatalf("digest-addressed proposal was not stored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(directory), "proposal:escape.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("proposal ID escaped the proposal directory: %v", err)
	}

	privileged := evidenceProposal("proposal:privileged-source", "evidence:privileged-source")
	privileged.Submitter.Role = RoleAuthorizedReviewer
	if _, err := kernel.submitProposal(privileged); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("privileged external role error = %v, want ErrUnauthorized", err)
	}
}

func TestNewSubjectsCannotClaimCoreRoles(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	proposal := Proposal{
		SchemaVersion: SchemaVersion,
		ID:            "proposal:privileged-subject",
		SubmittedAt:   testTime,
		Submitter:     externalSubmitter(),
		Mutation: Mutation{
			SchemaVersion: SchemaVersion,
			Operation:     MutationAddSubject,
			Subject: &Subject{
				SchemaVersion: SchemaVersion,
				ID:            "external:agent:privileged",
				Kind:          KindExternalAgent,
				DisplayName:   "Untrusted privileged subject",
				Roles:         []ActorRole{RoleCoreIdentity},
			},
		},
	}
	if _, err := kernel.submitProposal(proposal); !errors.Is(err, ErrImmutableCore) {
		t.Fatalf("privileged subject error = %v, want ErrImmutableCore", err)
	}
}

func TestStrictContractsRejectSecretsAndUnknownFields(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	secret := evidenceProposal("proposal:secret", "evidence:secret")
	secret.Mutation.Evidence.Locator = "contains-secret-value"
	if _, err := kernel.submitProposal(secret); !errors.Is(err, ErrInvalidContract) {
		t.Fatalf("secret-bearing evidence error = %v, want ErrInvalidContract", err)
	}
	data := []byte(`{"schema_version":"1.0.0","id":"evidence:one","locator":"sha256:x","sha256":"` + strings.Repeat("a", 64) + `","media_type":"application/json","source":"external","unexpected":true}`)
	var evidence EvidenceRef
	if err := decodeStrict(data, &evidence); err == nil {
		t.Fatal("unknown JSON field was accepted")
	}
}

func TestCoreSelfRejectsSymlinkedPersistenceBoundaries(t *testing.T) {
	realDirectory := t.TempDir()
	linkParent := t.TempDir()
	directoryLink := filepath.Join(linkParent, "linked-core-self")
	if err := os.Symlink(realDirectory, directoryLink); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := Open(directoryLink, BootstrapConfig{}); err == nil {
		t.Fatal("symlinked core-self directory was accepted")
	}
	if _, err := Open(filepath.Join(directoryLink, "nested"), BootstrapConfig{}); err == nil {
		t.Fatal("core-self path with a symlinked ancestor was accepted")
	}

	directory := filepath.Join(t.TempDir(), "identity")
	if _, err := Open(directory, BootstrapConfig{}); err != nil {
		t.Fatal(err)
	}
	ledgerPath := filepath.Join(directory, "identity-ledger.jsonl")
	realLedger := filepath.Join(t.TempDir(), "ledger.jsonl")
	if err := os.Rename(ledgerPath, realLedger); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realLedger, ledgerPath); err != nil {
		t.Skipf("ledger symlink unavailable: %v", err)
	}
	if _, err := Open(directory, BootstrapConfig{}); !errors.Is(err, ErrTampered) {
		t.Fatalf("symlinked ledger error = %v, want ErrTampered", err)
	}
}

func TestCoreSelfRejectsPostOpenRootPathReplacement(t *testing.T) {
	parent := t.TempDir()
	directory := filepath.Join(parent, "identity")
	kernel := mustOpen(t, directory)
	anchoredDirectory := filepath.Join(parent, "anchored-identity")
	if err := os.Rename(directory, anchoredDirectory); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(directory, "proposals")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	proposal := evidenceProposal("proposal:anchored-root", "evidence:anchored-root")
	if _, err := kernel.submitProposal(proposal); !errors.Is(err, ErrTampered) {
		t.Fatalf("replaced root error = %v, want ErrTampered", err)
	}
	if entries, err := os.ReadDir(outside); err != nil || len(entries) != 0 {
		t.Fatalf("replacement path received proposal data: entries=%d err=%v", len(entries), err)
	}
	if _, err := os.Stat(filepath.Join(anchoredDirectory, filepath.FromSlash(proposalName(proposal.ID)))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("rejected proposal was persisted in displaced root: %v", err)
	}
}

func TestParentNamespaceLockCannotSplitAfterRootReplacement(t *testing.T) {
	parent := t.TempDir()
	directory := filepath.Join(parent, "identity")
	first := mustOpen(t, directory)
	if err := os.Rename(directory, filepath.Join(parent, "identity-displaced")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	second := mustOpen(t, directory)
	releaseSecond, err := acquireKernelLock(second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := acquireKernelLock(first); !errors.Is(err, ErrStaleHead) {
		releaseSecond()
		t.Fatalf("displaced root escaped shared parent lock: %v", err)
	}
	releaseSecond()
	if _, err := acquireKernelLock(first); !errors.Is(err, ErrTampered) {
		t.Fatalf("displaced root was not rejected after lock release: %v", err)
	}
}

func TestAnchoredProposalStoreRejectsPostOpenInRootRedirection(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	anchored := filepath.Join(directory, "proposals-anchored")
	if err := os.Rename(filepath.Join(directory, "proposals"), anchored); err != nil {
		t.Fatal(err)
	}
	trap := filepath.Join(directory, "proposal-trap")
	if err := os.Mkdir(trap, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("proposal-trap", filepath.Join(directory, "proposals")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:child-anchor", "evidence:child-anchor"))
	if entries, err := os.ReadDir(trap); err != nil || len(entries) != 0 {
		t.Fatalf("replacement proposal directory received data: entries=%d err=%v", len(entries), err)
	}
	if _, err := os.Stat(filepath.Join(anchored, proposalFileName(proposal.ID))); err != nil {
		t.Fatalf("anchored proposal store did not receive proposal: %v", err)
	}
}

func TestAnchoredQuarantineAndLedgerRejectPostOpenInRootRedirection(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:ledger-anchor", "evidence:ledger-anchor"))

	anchoredQuarantine := filepath.Join(directory, "quarantine-anchored")
	if err := os.Rename(filepath.Join(directory, "quarantine"), anchoredQuarantine); err != nil {
		t.Fatal(err)
	}
	quarantineTrap := filepath.Join(directory, "quarantine-trap")
	if err := os.Mkdir(quarantineTrap, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("quarantine-trap", filepath.Join(directory, "quarantine")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	anchoredLedger := filepath.Join(directory, "identity-ledger-anchored.jsonl")
	if err := os.Rename(filepath.Join(directory, "identity-ledger.jsonl"), anchoredLedger); err != nil {
		t.Fatal(err)
	}
	ledgerTrap := filepath.Join(directory, "ledger-trap.jsonl")
	trapBytes := []byte("do-not-touch\n")
	if err := os.WriteFile(ledgerTrap, trapBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("ledger-trap.jsonl", filepath.Join(directory, "identity-ledger.jsonl")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	if _, err := kernel.acceptProposal(proposal.ID, stewardReviewer(), testReviewerPrivateKey()); !errors.Is(err, ErrTampered) {
		t.Fatalf("replaced ledger error = %v, want ErrTampered", err)
	}
	if got, err := os.ReadFile(ledgerTrap); err != nil || !bytes.Equal(got, trapBytes) {
		t.Fatalf("replacement ledger was modified: %q err=%v", got, err)
	}
	if entries, err := os.ReadDir(quarantineTrap); err != nil || len(entries) != 0 {
		t.Fatalf("replacement quarantine received data: entries=%d err=%v", len(entries), err)
	}
	if _, err := os.Stat(filepath.Join(anchoredQuarantine, "QUARANTINED")); err != nil {
		t.Fatalf("anchored quarantine did not receive marker: %v", err)
	}
}

func TestQuarantineWriteFailureIsSurfaced(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:quarantine-error", "evidence:quarantine-error"))
	quarantinePath := filepath.Join(directory, "quarantine")
	if err := os.Chmod(quarantinePath, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(quarantinePath, 0o700) })
	anchoredLedger := filepath.Join(directory, "identity-ledger-anchored.jsonl")
	if err := os.Rename(filepath.Join(directory, "identity-ledger.jsonl"), anchoredLedger); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "identity-ledger.jsonl"), []byte("replacement\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := kernel.acceptProposal(proposal.ID, stewardReviewer(), testReviewerPrivateKey()); !errors.Is(err, ErrTampered) || !strings.Contains(err.Error(), "write quarantine marker") {
		t.Fatalf("quarantine failure error = %v, want joined ErrTampered and quarantine error", err)
	}
}

func TestQuarantineMarkerCollisionsAreSurfaced(t *testing.T) {
	cases := []struct {
		name  string
		plant func(*testing.T, string)
	}{
		{name: "symlink", plant: func(t *testing.T, marker string) {
			if err := os.Symlink("unrelated-target", marker); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
		}},
		{name: "directory", plant: func(t *testing.T, marker string) {
			if err := os.Mkdir(marker, 0o700); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "unrelated-regular-file", plant: func(t *testing.T, marker string) {
			if err := os.WriteFile(marker, []byte("not a quarantine marker\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			directory := filepath.Join(t.TempDir(), "identity")
			kernel := mustOpen(t, directory)
			proposal := mustSubmit(t, kernel, evidenceProposal("proposal:quarantine-collision", "evidence:quarantine-collision"))
			if err := os.Rename(filepath.Join(directory, "identity-ledger.jsonl"), filepath.Join(directory, "ledger-displaced.jsonl")); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "identity-ledger.jsonl"), []byte("replacement\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			testCase.plant(t, filepath.Join(directory, "quarantine", "QUARANTINED"))
			_, err := kernel.acceptProposal(proposal.ID, stewardReviewer(), testReviewerPrivateKey())
			if !errors.Is(err, ErrTampered) || !strings.Contains(err.Error(), "write quarantine marker") {
				t.Fatalf("quarantine marker collision error = %v, want joined ErrTampered and marker-write error", err)
			}
		})
	}
}

func TestAnchoredHeadRejectsPostOpenInRootRedirection(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:head-anchor", "evidence:head-anchor"))
	anchoredHead := filepath.Join(directory, "head-anchored.json")
	if err := os.Rename(filepath.Join(directory, "head.json"), anchoredHead); err != nil {
		t.Fatal(err)
	}
	trap := filepath.Join(directory, "head-trap.json")
	trapBytes := []byte("do-not-touch\n")
	if err := os.WriteFile(trap, trapBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("head-trap.json", filepath.Join(directory, "head.json")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	if _, err := kernel.acceptProposal(proposal.ID, stewardReviewer(), testReviewerPrivateKey()); !errors.Is(err, ErrTampered) {
		t.Fatalf("replaced head error = %v, want ErrTampered", err)
	}
	if got, err := os.ReadFile(trap); err != nil || !bytes.Equal(got, trapBytes) {
		t.Fatalf("replacement head was modified: %q err=%v", got, err)
	}
	if info, err := os.Stat(anchoredHead); err != nil || info.Size() == 0 {
		t.Fatalf("anchored head was lost: info=%v err=%v", info, err)
	}
}

func TestLiveKernelRejectsRegularStaleHeadReplacement(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	first := mustSubmit(t, kernel, evidenceProposal("proposal:head-first", "evidence:head-first"))
	mustAccept(t, kernel, first.ID)
	staleHead, err := os.ReadFile(filepath.Join(directory, "head.json"))
	if err != nil {
		t.Fatal(err)
	}
	second := mustSubmit(t, kernel, evidenceProposal("proposal:head-second", "evidence:head-second"))
	mustAccept(t, kernel, second.ID)
	ledgerBefore, err := os.ReadFile(filepath.Join(directory, "identity-ledger.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(directory, "head.json"), filepath.Join(directory, "head-displaced.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "head.json"), staleHead, 0o600); err != nil {
		t.Fatal(err)
	}
	third := mustSubmit(t, kernel, evidenceProposal("proposal:head-third", "evidence:head-third"))
	if _, err := kernel.acceptProposal(third.ID, stewardReviewer(), testReviewerPrivateKey()); !errors.Is(err, ErrTampered) {
		t.Fatalf("regular replacement head error = %v, want ErrTampered", err)
	}
	ledgerAfter, err := os.ReadFile(filepath.Join(directory, "identity-ledger.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ledgerBefore, ledgerAfter) || kernel.Status().EventCount != 3 {
		t.Fatal("live kernel adopted a replacement stale head or advanced accepted history")
	}
	if _, err := os.Stat(filepath.Join(directory, "quarantine", "QUARANTINED")); err != nil {
		t.Fatalf("replacement head was not quarantined: %v", err)
	}
}

func TestCoreSelfAdvisoryLockReleasesWithDescriptor(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	firstRelease, err := acquireKernelLock(kernel)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(kernel.directory, BootstrapConfig{}); !errors.Is(err, ErrStaleHead) {
		firstRelease()
		t.Fatalf("concurrent lock error = %v, want ErrStaleHead", err)
	}
	decoy := filepath.Join(kernel.directory, ".identity.lock")
	if err := os.WriteFile(decoy, []byte("decoy"), 0o600); err != nil {
		firstRelease()
		t.Fatal(err)
	}
	if err := os.Remove(decoy); err != nil {
		firstRelease()
		t.Fatal(err)
	}
	trap := filepath.Join(kernel.directory, "lock-trap")
	if err := os.WriteFile(trap, []byte("replacement"), 0o600); err != nil {
		firstRelease()
		t.Fatal(err)
	}
	if err := os.Symlink("lock-trap", decoy); err != nil {
		firstRelease()
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := Open(kernel.directory, BootstrapConfig{}); !errors.Is(err, ErrStaleHead) {
		firstRelease()
		t.Fatalf("decoy lock replacement split namespace: %v", err)
	}
	firstRelease()
	secondKernel, err := Open(kernel.directory, BootstrapConfig{})
	if err != nil {
		t.Fatalf("lock did not recover after descriptor release: %v", err)
	}
	secondRelease, err := acquireKernelLock(secondKernel)
	if err != nil {
		t.Fatalf("recovered kernel could not acquire lock: %v", err)
	}
	secondRelease()
}

func TestOpenRecoversValidStaleHeadAfterDurableLedgerAppend(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	staleHead, err := os.ReadFile(filepath.Join(directory, "head.json"))
	if err != nil {
		t.Fatal(err)
	}
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:stale-head", "evidence:stale-head"))
	mustAccept(t, kernel, proposal.ID)
	if err := os.WriteFile(filepath.Join(directory, "head.json"), staleHead, 0o600); err != nil {
		t.Fatal(err)
	}

	recovered := mustOpen(t, directory)
	if got := recovered.Status().EventCount; got != 2 {
		t.Fatalf("recovered event count = %d, want 2", got)
	}
	if _, ok := findEvidence(recovered.StateSnapshot().State, "evidence:stale-head"); !ok {
		t.Fatal("valid ledger event was not recovered")
	}
	headBytes, err := os.ReadFile(filepath.Join(directory, "head.json"))
	if err != nil {
		t.Fatal(err)
	}
	var head headRecord
	if err := decodeStrict(bytes.TrimSpace(headBytes), &head); err != nil || !headMatchesState(head, recovered.state) {
		t.Fatalf("repaired head does not match replay: head=%+v err=%v", head, err)
	}
	if _, err := os.Stat(filepath.Join(directory, "quarantine", "QUARANTINED")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ordinary stale-head recovery was quarantined: %v", err)
	}
}

func TestOpenRecoversMissingHeadCacheFromVerifiedLedger(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:missing-head", "evidence:missing-head"))
	mustAccept(t, kernel, proposal.ID)
	if err := os.Remove(filepath.Join(directory, "head.json")); err != nil {
		t.Fatal(err)
	}
	recovered := mustOpen(t, directory)
	if recovered.Status().EventCount != 2 {
		t.Fatalf("missing-head recovery lost ledger state: %+v", recovered.Status())
	}
	if _, err := os.Stat(filepath.Join(directory, "head.json")); err != nil {
		t.Fatalf("missing head cache was not rebuilt: %v", err)
	}
}

func TestOpenRejectsMalformedAndForgedStaleHeads(t *testing.T) {
	t.Run("malformed", func(t *testing.T) {
		directory := filepath.Join(t.TempDir(), "identity")
		mustOpen(t, directory)
		if err := os.WriteFile(filepath.Join(directory, "head.json"), []byte("{malformed}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Open(directory, BootstrapConfig{}); !errors.Is(err, ErrTampered) {
			t.Fatalf("malformed head error = %v, want ErrTampered", err)
		}
		if _, err := os.Stat(filepath.Join(directory, "quarantine", "QUARANTINED")); err != nil {
			t.Fatalf("malformed head was not quarantined: %v", err)
		}
	})

	t.Run("forged-ancestor", func(t *testing.T) {
		directory := filepath.Join(t.TempDir(), "identity")
		kernel := mustOpen(t, directory)
		proposal := mustSubmit(t, kernel, evidenceProposal("proposal:forged-head", "evidence:forged-head"))
		mustAccept(t, kernel, proposal.ID)
		forged := headFor(State{SchemaVersion: SchemaVersion, HeadHash: strings.Repeat("f", 64), EventCount: 1})
		forged.EventCount = 1
		if err := os.WriteFile(filepath.Join(directory, "head.json"), append(mustCanonical(forged), '\n'), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Open(directory, testBootstrapConfig()); !errors.Is(err, ErrFork) {
			t.Fatalf("forged stale head error = %v, want ErrFork", err)
		}
	})
}

func TestCoreSelfRejectsOversizedPrivateState(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	if _, err := Open(directory, BootstrapConfig{}); err != nil {
		t.Fatal(err)
	}
	headPath := filepath.Join(directory, "head.json")
	if err := os.WriteFile(headPath, bytes.Repeat([]byte{'x'}, int(maxHeadBytes+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(directory, BootstrapConfig{}); !errors.Is(err, ErrTampered) || !strings.Contains(err.Error(), "size bound") {
		t.Fatalf("oversized head error = %v, want bounded ErrTampered", err)
	}
}

func mustCapsule(t *testing.T, kernel *Kernel) Capsule {
	t.Helper()
	capsule, err := kernel.ExportCapsule()
	if err != nil {
		t.Fatalf("ExportCapsule: %v", err)
	}
	return capsule
}

func TestLedgerTamperDetectionAndQuarantine(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:tamper", "evidence:tamper"))
	mustAccept(t, kernel, proposal.ID)

	ledgerPath := filepath.Join(directory, "identity-ledger.jsonl")
	ledger, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	ledger[len(ledger)-2] ^= 1
	if err := os.WriteFile(ledgerPath, ledger, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(directory, BootstrapConfig{}); !errors.Is(err, ErrTampered) {
		t.Fatalf("tampered ledger error = %v, want ErrTampered", err)
	}
	if _, err := os.Stat(filepath.Join(directory, "quarantine", "QUARANTINED")); err != nil {
		t.Fatalf("tampered ledger was not quarantined: %v", err)
	}
}

func rewriteAcceptedEvent(t *testing.T, directory string, mutate func(*IdentityEvent)) {
	t.Helper()
	ledgerPath := filepath.Join(directory, "identity-ledger.jsonl")
	ledger, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSuffix(ledger, []byte{'\n'}), []byte{'\n'})
	if len(lines) != 2 {
		t.Fatalf("ledger lines = %d, want 2", len(lines))
	}
	var event IdentityEvent
	if err := decodeStrict(lines[1], &event); err != nil {
		t.Fatal(err)
	}
	mutate(&event)
	event.Hash = ""
	event.Hash, err = eventDigest(event)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := CanonicalJSON(event)
	if err != nil {
		t.Fatal(err)
	}
	content := append(append(append([]byte{}, lines[0]...), '\n'), append(changed, '\n')...)
	if err := os.WriteFile(ledgerPath, content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestSignedProposalCommitmentRejectsRehashedMutation(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	staleHead, err := os.ReadFile(filepath.Join(directory, "head.json"))
	if err != nil {
		t.Fatal(err)
	}
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:signed-tamper", "evidence:signed-tamper"))
	mustAccept(t, kernel, proposal.ID)
	rewriteAcceptedEvent(t, directory, func(event *IdentityEvent) {
		event.Mutation.Evidence.SHA256 = strings.Repeat("b", 64)
	})
	if err := os.WriteFile(filepath.Join(directory, "head.json"), staleHead, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(directory, testBootstrapConfig()); !errors.Is(err, ErrTampered) {
		t.Fatalf("rehashed mutation error = %v, want ErrTampered", err)
	}
}

func TestReviewerSignatureRejectsRecomputedAcceptedEvent(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	staleHead, err := os.ReadFile(filepath.Join(directory, "head.json"))
	if err != nil {
		t.Fatal(err)
	}
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:forged-signature", "evidence:forged-signature"))
	mustAccept(t, kernel, proposal.ID)
	rewriteAcceptedEvent(t, directory, func(event *IdentityEvent) {
		event.Mutation.Evidence.SHA256 = strings.Repeat("c", 64)
		event.Proposal.Mutation = event.Mutation
		event.Proposal.Digest = ""
		event.Proposal.Digest, _ = proposalDigest(*event.Proposal)
		event.ProposalDigest = event.Proposal.Digest
		event.Decision.ProposalDigest = event.Proposal.Digest
		event.Decision.Digest = ""
		event.Decision.Digest, _ = decisionDigest(*event.Decision)
	})
	if err := os.WriteFile(filepath.Join(directory, "head.json"), staleHead, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(directory, testBootstrapConfig()); !errors.Is(err, ErrTampered) {
		t.Fatalf("recomputed accepted event error = %v, want ErrTampered", err)
	}
}

func TestAcceptedLedgerRequiresExternallyPinnedReviewerKey(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:key-anchor", "evidence:key-anchor"))
	mustAccept(t, kernel, proposal.ID)
	if _, err := Open(directory, BootstrapConfig{}); !errors.Is(err, ErrTampered) || !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("missing reviewer trust anchor error = %v", err)
	}
}

func TestReviewerSignatureRejectsAppendedForgedSuffix(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:signed-suffix", "evidence:signed-suffix"))
	mustAccept(t, kernel, proposal.ID)
	ledgerPath := filepath.Join(directory, "identity-ledger.jsonl")
	ledger, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSuffix(ledger, []byte{'\n'}), []byte{'\n'})
	var forged IdentityEvent
	if err := decodeStrict(lines[1], &forged); err != nil {
		t.Fatal(err)
	}
	forged.ID = "event:forged-suffix"
	forged.Sequence = 2
	forged.PreviousHash = forged.Hash
	forged.Hash = ""
	forged.Hash, err = eventDigest(forged)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := CanonicalJSON(forged)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ledgerPath, append(append(ledger, encoded...), '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(directory, testBootstrapConfig()); !errors.Is(err, ErrTampered) {
		t.Fatalf("forged suffix error = %v, want ErrTampered", err)
	}
}

func TestExportCapsuleReturnsDeepCopiedLedger(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	exported := mustCapsule(t, kernel)
	exported.Ledger[0].Mutation.Bootstrap.Subjects[0].DisplayName = "caller-mutated"
	again := mustCapsule(t, kernel)
	if again.Ledger[0].Mutation.Bootstrap.Subjects[0].DisplayName == "caller-mutated" {
		t.Fatal("exported capsule aliases the kernel ledger")
	}
	if err := kernel.VerifyCapsule(again); err != nil {
		t.Fatalf("fresh capsule did not verify: %v", err)
	}
}

func TestKernelCloseIsIdempotentAndNonAuthoritative(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	before := kernel.StateSnapshot()
	if err := kernel.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := kernel.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if after := kernel.StateSnapshot(); !reflect.DeepEqual(after, before) {
		t.Fatalf("Close changed identity state: before=%+v after=%+v", before, after)
	}
	if status := kernel.Status(); status.EventCount != 1 || status.ProposalCount != 0 {
		t.Fatalf("closed status = %+v", status)
	}
	if _, err := kernel.submitProposal(evidenceProposal("proposal:after-close", "evidence:after-close")); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("post-close persistence error = %v, want ErrUnauthorized", err)
	}
	capsule := mustCapsule(t, kernel)
	if err := VerifyCapsule(capsule); err != nil {
		t.Fatalf("closed kernel did not retain verifiable in-memory identity: %v", err)
	}
}

func TestExternalAndExperimentalRemainProposalOnly(t *testing.T) {
	kernel := mustOpen(t, filepath.Join(t.TempDir(), "identity"))
	experimental := evidenceProposal("proposal:experimental", "evidence:experimental")
	experimental.Submitter.Source = SourceExperimental
	stored := mustSubmit(t, kernel, experimental)
	if got := kernel.Status().EventCount; got != 1 {
		t.Fatalf("proposal-only submission changed event count to %d", got)
	}
	if _, err := kernel.acceptProposal(stored.ID, PolicyContext{
		SchemaVersion: SchemaVersion,
		ActorID:       "external:reviewer:test",
		Role:          RoleAuthorizedReviewer,
		Source:        SourceExternal,
		Purpose:       "attempt external review",
	}, testReviewerPrivateKey()); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("external reviewer error = %v, want ErrUnauthorized", err)
	}
	mustAccept(t, kernel, stored.ID)
	if got := kernel.Status().EventCount; got != 2 {
		t.Fatalf("authorized acceptance event count = %d, want 2", got)
	}
}
