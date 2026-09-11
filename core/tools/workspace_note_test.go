package tools

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newWorkspaceNoteTool(t *testing.T) WorkspaceNoteTool {
	t.Helper()
	root := filepath.Join(t.TempDir(), "workspace")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatalf("create workspace root: %v", err)
	}
	return WorkspaceNoteTool{Root: root, MaxBytes: 1024}
}

func TestWorkspaceNoteCreateAtomicReadBack(t *testing.T) {
	tool := newWorkspaceNoteTool(t)
	content := []byte("an exact, durable note\n")

	got, err := tool.Create(context.Background(), WorkspaceNoteRequest{
		Path:    "notes/first.txt",
		Content: content,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.ToolName != WorkspaceCreateNoteToolName {
		t.Errorf("ToolName = %q, want %q", got.ToolName, WorkspaceCreateNoteToolName)
	}
	if got.Status != WorkspaceNoteCreated || !got.Created || got.Blocked || got.RetryAllowed {
		t.Errorf("unexpected creation state: %+v", got)
	}
	if !bytes.Equal(got.ReadBack, content) {
		t.Errorf("ReadBack = %q, want %q", got.ReadBack, content)
	}
	sum := sha256.Sum256(content)
	if got.SHA256 != hex.EncodeToString(sum[:]) || got.RequestedSHA256 != got.SHA256 {
		t.Errorf("hash evidence = (%q, %q), want %q", got.SHA256, got.RequestedSHA256, hex.EncodeToString(sum[:]))
	}

	stored, err := os.ReadFile(filepath.Join(tool.Root, "notes", "first.txt"))
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if !bytes.Equal(stored, content) {
		t.Errorf("stored note = %q, want %q", stored, content)
	}
	info, err := os.Stat(filepath.Join(tool.Root, "notes"))
	if err != nil {
		t.Fatalf("stat parent: %v", err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Errorf("parent mode = %04o, want 0700", info.Mode().Perm())
	}
	info, err = os.Stat(filepath.Join(tool.Root, "notes", "first.txt"))
	if err != nil {
		t.Fatalf("stat target: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("target mode = %04o, want 0600", info.Mode().Perm())
	}
}

func TestWorkspaceNoteRejectsUnsafePathsAndOverwrite(t *testing.T) {
	tool := newWorkspaceNoteTool(t)
	valid := WorkspaceNoteRequest{Path: "safe.txt", Content: []byte("safe")}
	if _, err := tool.Create(context.Background(), valid); err != nil {
		t.Fatalf("initial Create() error = %v", err)
	}

	cases := []struct {
		name string
		path string
		want error
	}{
		{name: "absolute", path: "/tmp/note.txt", want: ErrInvalidRelativePath},
		{name: "traversal", path: "notes/../note.txt", want: ErrPathTraversal},
		{name: "unclean", path: "notes//note.txt", want: ErrInvalidRelativePath},
		{name: "overwrite", path: "safe.txt", want: ErrAlreadyExists},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tool.Create(context.Background(), WorkspaceNoteRequest{Path: tc.path, Content: []byte("changed")})
			if !errors.Is(err, tc.want) {
				t.Fatalf("Create(%q) error = %v, want errors.Is(..., %v)", tc.path, err, tc.want)
			}
		})
	}

	stored, err := os.ReadFile(filepath.Join(tool.Root, "safe.txt"))
	if err != nil {
		t.Fatalf("read original after overwrite rejection: %v", err)
	}
	if !bytes.Equal(stored, valid.Content) {
		t.Errorf("overwrite changed existing target: got %q, want %q", stored, valid.Content)
	}
}

func TestWorkspaceNoteRejectsSymlinkAndOversizeAndCancellation(t *testing.T) {
	tool := newWorkspaceNoteTool(t)
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.Mkdir(outside, 0o700); err != nil {
		t.Fatalf("create outside: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(tool.Root, "linked")); err != nil {
		t.Fatalf("create parent symlink: %v", err)
	}
	if _, err := tool.Create(context.Background(), WorkspaceNoteRequest{Path: "linked/note.txt", Content: []byte("x")}); !errors.Is(err, ErrSymlink) {
		t.Fatalf("parent symlink error = %v, want ErrSymlink", err)
	}
	if err := os.Symlink(filepath.Join(outside, "other.txt"), filepath.Join(tool.Root, "target-link.txt")); err != nil {
		t.Fatalf("create target symlink: %v", err)
	}
	if _, err := tool.Create(context.Background(), WorkspaceNoteRequest{Path: "target-link.txt", Content: []byte("x")}); !errors.Is(err, ErrSymlink) {
		t.Fatalf("target symlink error = %v, want ErrSymlink", err)
	}

	tool.MaxBytes = 3
	if _, err := tool.Create(context.Background(), WorkspaceNoteRequest{Path: "large.txt", Content: []byte("four")}); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("oversize error = %v, want ErrTooLarge", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := tool.Create(ctx, WorkspaceNoteRequest{Path: "cancelled.txt", Content: []byte("x")}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled create error = %v, want context.Canceled", err)
	}
	if _, err := os.Stat(filepath.Join(tool.Root, "cancelled.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("cancelled target stat error = %v, want not exist", err)
	}

	timed := newWorkspaceNoteTool(t)
	timed.Timeout = time.Nanosecond
	if _, err := timed.Create(context.Background(), WorkspaceNoteRequest{Path: "timed-out.txt", Content: []byte("x")}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timed Create() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestWorkspaceNoteReconcileMatchingAndConflict(t *testing.T) {
	tool := newWorkspaceNoteTool(t)
	request := WorkspaceNoteRequest{Path: "notes/reconcile.txt", Content: []byte("canonical note")}

	absent, err := tool.Reconcile(request)
	if err != nil {
		t.Fatalf("absent Reconcile() error = %v", err)
	}
	if absent.Status != WorkspaceNoteAbsentRetry || !absent.RetryAllowed || absent.RetryLimit != 1 || absent.Blocked {
		t.Errorf("absent reconciliation = %+v, want one allowed retry", absent)
	}

	if _, err := tool.Create(context.Background(), request); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	matching, err := tool.Reconcile(request)
	if err != nil {
		t.Fatalf("matching Reconcile() error = %v", err)
	}
	if matching.Status != WorkspaceNoteMatchingExisting || matching.Created || matching.RetryAllowed || matching.Blocked {
		t.Errorf("matching reconciliation = %+v", matching)
	}
	if !bytes.Equal(matching.ReadBack, request.Content) {
		t.Errorf("matching read-back = %q, want %q", matching.ReadBack, request.Content)
	}

	conflict, err := tool.Reconcile(WorkspaceNoteRequest{Path: request.Path, Content: []byte("different note")})
	if err != nil {
		t.Fatalf("conflict Reconcile() error = %v", err)
	}
	if conflict.Status != WorkspaceNoteHashConflict || !conflict.Blocked || conflict.RetryAllowed || conflict.RetryLimit != 0 {
		t.Errorf("conflict reconciliation = %+v, want blocked hash conflict", conflict)
	}
	if conflict.SHA256 == conflict.RequestedSHA256 {
		t.Errorf("conflict hashes unexpectedly match: %+v", conflict)
	}
}
