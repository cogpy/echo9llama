// Package tools contains small, locally-scoped tools used by the core runtime.
package tools

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const (
	// WorkspaceCreateNoteToolName is the stable name exposed to a tool caller.
	WorkspaceCreateNoteToolName = "workspace.create_note"

	// DefaultWorkspaceNoteMaxBytes bounds note payloads when MaxBytes is zero.
	DefaultWorkspaceNoteMaxBytes int64 = 1 << 20 // 1 MiB
)

var (
	// ErrInvalidWorkspaceRoot means Root is not an absolute, private directory.
	ErrInvalidWorkspaceRoot = errors.New("invalid workspace root")
	// ErrInvalidRelativePath means a request path is not already clean and relative.
	ErrInvalidRelativePath = errors.New("invalid relative path")
	// ErrPathTraversal means a request path includes a parent-directory component.
	ErrPathTraversal = errors.New("path traversal is not allowed")
	// ErrSymlink means a root, path component, or target was a symbolic link.
	ErrSymlink = errors.New("symbolic links are not allowed")
	// ErrNonRegular means an existing target or path component has an unsafe type.
	ErrNonRegular = errors.New("non-regular filesystem object")
	// ErrInsecureDirectory means a required directory is not owned and mode 0700.
	ErrInsecureDirectory = errors.New("directory is not owner-only mode 0700")
	// ErrWrongOwner means a filesystem object is not owned by the effective user.
	ErrWrongOwner = errors.New("filesystem object is not owned by the effective user")
	// ErrAlreadyExists means create-only publication found an existing target.
	ErrAlreadyExists = errors.New("workspace note already exists")
	// ErrTooLarge means a note exceeded the configured byte limit.
	ErrTooLarge = errors.New("workspace note exceeds byte limit")
	// ErrReadBackMismatch means the durable target did not exactly match the request.
	ErrReadBackMismatch = errors.New("workspace note read-back does not match request")
)

var errPathAbsent = errors.New("workspace note path is absent")

// WorkspaceNoteRequest is the complete input to workspace.create_note. Path must
// be a clean relative slash-separated path. Content is retained as bytes so that
// the SHA-256 evidence and the mandatory read-back comparison are unambiguous.
type WorkspaceNoteRequest struct {
	Path    string
	Content []byte
}

// WorkspaceNoteStatus describes the durable state observed by Create or Reconcile.
type WorkspaceNoteStatus string

const (
	// WorkspaceNoteCreated means this call atomically created and read back the note.
	WorkspaceNoteCreated WorkspaceNoteStatus = "created"
	// WorkspaceNoteMatchingExisting means Reconcile found the requested bytes already present.
	WorkspaceNoteMatchingExisting WorkspaceNoteStatus = "matching_existing"
	// WorkspaceNoteAbsentRetry means no target exists; the caller may attempt exactly one create retry.
	WorkspaceNoteAbsentRetry WorkspaceNoteStatus = "absent_retry"
	// WorkspaceNoteHashConflict means a target exists with a different SHA-256 and must not be overwritten.
	WorkspaceNoteHashConflict WorkspaceNoteStatus = "hash_conflict"
)

// WorkspaceNoteObservation is the evidence returned by a workspace-note operation.
// SHA256 always identifies ReadBack when ReadBack is present. RequestedSHA256 is
// the digest of request.Content, which makes reconciliation conflicts explicit.
type WorkspaceNoteObservation struct {
	ToolName        string
	Path            string
	Status          WorkspaceNoteStatus
	Created         bool
	RetryAllowed    bool
	RetryLimit      int
	Blocked         bool
	Bytes           int64
	SHA256          string
	RequestedSHA256 string
	ReadBack        []byte
}

// WorkspaceNoteTool writes small notes beneath one private workspace root. Root
// must be absolute. Existing roots and all parent directories must be owned by
// the effective user with exact mode 0700. A zero MaxBytes uses
// DefaultWorkspaceNoteMaxBytes; a negative MaxBytes is invalid. Timeout, when
// positive, bounds each operation in addition to its caller context.
type WorkspaceNoteTool struct {
	Root     string
	MaxBytes int64
	Timeout  time.Duration
}

// Name returns the stable external tool name.
func (WorkspaceNoteTool) Name() string { return WorkspaceCreateNoteToolName }

// Create validates, writes, fsyncs, atomically publishes, fsyncs again, and then
// exactly reads back a new note. It never replaces an existing destination.
func (t WorkspaceNoteTool) Create(ctx context.Context, request WorkspaceNoteRequest) (WorkspaceNoteObservation, error) {
	ctx, cancel := t.operationContext(ctx)
	defer cancel()

	payload, err := t.validateRequest(ctx, request)
	if err != nil {
		return WorkspaceNoteObservation{}, err
	}

	rootFD, err := t.openRoot(ctx, true)
	if err != nil {
		return WorkspaceNoteObservation{}, err
	}
	defer unix.Close(rootFD)

	parentFD, err := t.openParent(ctx, rootFD, payload.parts, true)
	if err != nil {
		return WorkspaceNoteObservation{}, err
	}
	defer unix.Close(parentFD)

	if err := checkContext(ctx); err != nil {
		return WorkspaceNoteObservation{}, err
	}
	if exists, err := inspectTarget(parentFD, payload.leaf); err != nil {
		return WorkspaceNoteObservation{}, err
	} else if exists {
		return WorkspaceNoteObservation{}, fmt.Errorf("%w: %q", ErrAlreadyExists, payload.path)
	}

	tempName, err := writeTemporary(ctx, parentFD, payload.content)
	if err != nil {
		return WorkspaceNoteObservation{}, err
	}
	published := false
	defer func() {
		// The temporary name is in the already-open private parent directory. It
		// is safe to remove after either link success or a failed operation.
		if tempName != "" {
			_ = unix.Unlinkat(parentFD, tempName, 0)
			_ = unix.Fsync(parentFD)
		}
	}()

	if err := checkContext(ctx); err != nil {
		return WorkspaceNoteObservation{}, err
	}
	// linkat is create-only: unlike rename it fails with EEXIST and cannot replace
	// an existing target. The temporary file is in the same directory/filesystem.
	if err := unix.Linkat(parentFD, tempName, parentFD, payload.leaf, 0); err != nil {
		if errors.Is(err, unix.EEXIST) {
			if _, inspectErr := inspectTarget(parentFD, payload.leaf); inspectErr != nil {
				return WorkspaceNoteObservation{}, inspectErr
			}
			return WorkspaceNoteObservation{}, fmt.Errorf("%w: %q", ErrAlreadyExists, payload.path)
		}
		return WorkspaceNoteObservation{}, fmt.Errorf("link temporary workspace note: %w", err)
	}
	published = true
	if err := unix.Fsync(parentFD); err != nil {
		return WorkspaceNoteObservation{}, fmt.Errorf("fsync workspace note parent: %w", err)
	}
	if err := unix.Unlinkat(parentFD, tempName, 0); err != nil {
		return WorkspaceNoteObservation{}, fmt.Errorf("remove workspace note temporary file: %w", err)
	}
	tempName = ""
	if err := unix.Fsync(parentFD); err != nil {
		return WorkspaceNoteObservation{}, fmt.Errorf("fsync workspace note parent cleanup: %w", err)
	}

	readBack, readHash, err := readRegularAt(ctx, parentFD, payload.leaf, t.maxBytes())
	if err != nil {
		return WorkspaceNoteObservation{}, err
	}
	if !bytes.Equal(readBack, payload.content) || readHash != payload.sha256 {
		return WorkspaceNoteObservation{}, ErrReadBackMismatch
	}
	if !published {
		// Kept as a defensive assertion beside the no-overwrite publication path.
		return WorkspaceNoteObservation{}, ErrReadBackMismatch
	}
	return observation(payload, WorkspaceNoteCreated, true, false, false, readBack, readHash), nil
}

// Execute is an explicit tool-dispatch alias for Create.
func (t WorkspaceNoteTool) Execute(ctx context.Context, request WorkspaceNoteRequest) (WorkspaceNoteObservation, error) {
	return t.Create(ctx, request)
}

// Reconcile observes a prior create attempt without rewriting it. It returns
// WorkspaceNoteMatchingExisting for identical bytes, WorkspaceNoteAbsentRetry
// with a retry limit of one when the target is absent, and
// WorkspaceNoteHashConflict (Blocked=true) when a different note exists.
func (t WorkspaceNoteTool) Reconcile(request WorkspaceNoteRequest) (WorkspaceNoteObservation, error) {
	return t.ReconcileContext(context.Background(), request)
}

// ReconcileContext is Reconcile with caller cancellation support.
func (t WorkspaceNoteTool) ReconcileContext(ctx context.Context, request WorkspaceNoteRequest) (WorkspaceNoteObservation, error) {
	ctx, cancel := t.operationContext(ctx)
	defer cancel()

	payload, err := t.validateRequest(ctx, request)
	if err != nil {
		return WorkspaceNoteObservation{}, err
	}

	rootFD, err := t.openRoot(ctx, false)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return observation(payload, WorkspaceNoteAbsentRetry, false, true, false, nil, ""), nil
		}
		return WorkspaceNoteObservation{}, err
	}
	defer unix.Close(rootFD)

	parentFD, err := t.openParent(ctx, rootFD, payload.parts, false)
	if err != nil {
		if errors.Is(err, errPathAbsent) {
			return observation(payload, WorkspaceNoteAbsentRetry, false, true, false, nil, ""), nil
		}
		return WorkspaceNoteObservation{}, err
	}
	defer unix.Close(parentFD)

	exists, err := inspectTarget(parentFD, payload.leaf)
	if err != nil {
		return WorkspaceNoteObservation{}, err
	}
	if !exists {
		return observation(payload, WorkspaceNoteAbsentRetry, false, true, false, nil, ""), nil
	}

	readBack, readHash, err := readRegularAt(ctx, parentFD, payload.leaf, t.maxBytes())
	if err != nil {
		return WorkspaceNoteObservation{}, err
	}
	if readHash == payload.sha256 && bytes.Equal(readBack, payload.content) {
		return observation(payload, WorkspaceNoteMatchingExisting, false, false, false, readBack, readHash), nil
	}
	return observation(payload, WorkspaceNoteHashConflict, false, false, true, readBack, readHash), nil
}

type notePayload struct {
	path    string
	parts   []string
	leaf    string
	content []byte
	sha256  string
}

func (t WorkspaceNoteTool) validateRequest(ctx context.Context, request WorkspaceNoteRequest) (notePayload, error) {
	if err := checkContext(ctx); err != nil {
		return notePayload{}, err
	}
	parts, clean, err := cleanRelativePath(request.Path)
	if err != nil {
		return notePayload{}, err
	}
	maxBytes := t.maxBytes()
	if maxBytes < 0 {
		return notePayload{}, fmt.Errorf("%w: negative maximum", ErrTooLarge)
	}
	if int64(len(request.Content)) > maxBytes {
		return notePayload{}, fmt.Errorf("%w: got %d, limit %d", ErrTooLarge, len(request.Content), maxBytes)
	}
	sum := sha256.Sum256(request.Content)
	return notePayload{
		path:    clean,
		parts:   parts,
		leaf:    parts[len(parts)-1],
		content: append([]byte(nil), request.Content...),
		sha256:  hex.EncodeToString(sum[:]),
	}, nil
}

func (t WorkspaceNoteTool) maxBytes() int64 {
	if t.MaxBytes == 0 {
		return DefaultWorkspaceNoteMaxBytes
	}
	return t.MaxBytes
}

func (t WorkspaceNoteTool) operationContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if t.Timeout > 0 {
		return context.WithTimeout(parent, t.Timeout)
	}
	return parent, func() {}
}

func cleanRelativePath(path string) ([]string, string, error) {
	if path == "" || strings.IndexByte(path, 0) >= 0 {
		return nil, "", ErrInvalidRelativePath
	}
	// Reject backslashes explicitly. This keeps the wire format deterministic and
	// prevents a Windows-style traversal spelling from acquiring meaning elsewhere.
	if strings.Contains(path, `\`) || filepath.IsAbs(path) || filepath.VolumeName(path) != "" {
		return nil, "", ErrInvalidRelativePath
	}
	for _, component := range strings.Split(path, "/") {
		if component == ".." {
			return nil, "", ErrPathTraversal
		}
	}
	clean := filepath.Clean(path)
	if clean == "." || clean != path || filepath.IsAbs(clean) {
		return nil, "", ErrInvalidRelativePath
	}
	parts := strings.Split(clean, string(filepath.Separator))
	if len(parts) == 0 {
		return nil, "", ErrInvalidRelativePath
	}
	return parts, clean, nil
}

func (t WorkspaceNoteTool) openRoot(ctx context.Context, create bool) (int, error) {
	if err := checkContext(ctx); err != nil {
		return -1, err
	}
	if t.Root == "" || !filepath.IsAbs(t.Root) || filepath.Clean(t.Root) != t.Root {
		return -1, ErrInvalidWorkspaceRoot
	}

	_, err := os.Lstat(t.Root)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) || !create {
			return -1, err
		}
		if err := os.Mkdir(t.Root, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
			return -1, fmt.Errorf("create workspace root: %w", err)
		}
	}
	if err := checkContext(ctx); err != nil {
		return -1, err
	}

	info, err := os.Lstat(t.Root)
	if err != nil {
		return -1, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return -1, fmt.Errorf("%w: workspace root", ErrSymlink)
	}
	fd, err := unix.Open(t.Root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		if errors.Is(err, unix.ELOOP) {
			return -1, fmt.Errorf("%w: workspace root", ErrSymlink)
		}
		return -1, fmt.Errorf("open workspace root: %w", err)
	}
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		unix.Close(fd)
		return -1, fmt.Errorf("stat workspace root: %w", err)
	}
	if err := validatePrivateDirectory(&stat); err != nil {
		unix.Close(fd)
		return -1, fmt.Errorf("%w: workspace root", err)
	}
	return fd, nil
}

// openParent returns an owned file descriptor for the directory containing the
// target. It follows no symlinks and creates missing parents at exact mode 0700
// only when create is true.
func (t WorkspaceNoteTool) openParent(ctx context.Context, rootFD int, parts []string, create bool) (int, error) {
	current, err := unix.Dup(rootFD)
	if err != nil {
		return -1, fmt.Errorf("duplicate workspace root descriptor: %w", err)
	}
	for _, part := range parts[:len(parts)-1] {
		if err := checkContext(ctx); err != nil {
			unix.Close(current)
			return -1, err
		}
		next, err := openPrivateChildDirectory(ctx, current, part, create)
		unix.Close(current)
		if err != nil {
			return -1, err
		}
		current = next
	}
	return current, nil
}

func openPrivateChildDirectory(ctx context.Context, parentFD int, name string, create bool) (int, error) {
	for attempts := 0; attempts < 4; attempts++ {
		var stat unix.Stat_t
		err := unix.Fstatat(parentFD, name, &stat, unix.AT_SYMLINK_NOFOLLOW)
		if err != nil {
			if !errors.Is(err, unix.ENOENT) {
				return -1, fmt.Errorf("lstat workspace path component %q: %w", name, err)
			}
			if !create {
				return -1, fmt.Errorf("%w: %q", errPathAbsent, name)
			}
			if err := checkContext(ctx); err != nil {
				return -1, err
			}
			if err := unix.Mkdirat(parentFD, name, 0o700); err != nil {
				if errors.Is(err, unix.EEXIST) {
					continue
				}
				return -1, fmt.Errorf("create workspace parent %q: %w", name, err)
			}
			if err := unix.Fsync(parentFD); err != nil {
				return -1, fmt.Errorf("fsync new workspace parent: %w", err)
			}
			continue
		}
		if stat.Mode&unix.S_IFMT == unix.S_IFLNK {
			return -1, fmt.Errorf("%w: %q", ErrSymlink, name)
		}
		if err := validatePrivateDirectory(&stat); err != nil {
			return -1, fmt.Errorf("%w: path component %q", err, name)
		}
		fd, err := unix.Openat(parentFD, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if err != nil {
			if errors.Is(err, unix.ELOOP) {
				return -1, fmt.Errorf("%w: %q", ErrSymlink, name)
			}
			return -1, fmt.Errorf("open workspace path component %q: %w", name, err)
		}
		var opened unix.Stat_t
		if err := unix.Fstat(fd, &opened); err != nil {
			unix.Close(fd)
			return -1, fmt.Errorf("stat workspace path component %q: %w", name, err)
		}
		if err := validatePrivateDirectory(&opened); err != nil {
			unix.Close(fd)
			return -1, fmt.Errorf("%w: path component %q", err, name)
		}
		return fd, nil
	}
	return -1, fmt.Errorf("create workspace parent %q: too much concurrent mutation", name)
}

func validatePrivateDirectory(stat *unix.Stat_t) error {
	if stat.Mode&unix.S_IFMT != unix.S_IFDIR {
		return ErrNonRegular
	}
	if int(stat.Uid) != os.Geteuid() {
		return ErrWrongOwner
	}
	if stat.Mode&0o777 != 0o700 {
		return ErrInsecureDirectory
	}
	return nil
}

// inspectTarget returns true only for an existing regular target. It deliberately
// lstat's rather than stat's so a symlink is rejected rather than followed.
func inspectTarget(parentFD int, name string) (bool, error) {
	var stat unix.Stat_t
	if err := unix.Fstatat(parentFD, name, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		if errors.Is(err, unix.ENOENT) {
			return false, nil
		}
		return false, fmt.Errorf("lstat workspace note target: %w", err)
	}
	switch stat.Mode & unix.S_IFMT {
	case unix.S_IFLNK:
		return false, fmt.Errorf("%w: target", ErrSymlink)
	case unix.S_IFREG:
		if int(stat.Uid) != os.Geteuid() {
			return false, fmt.Errorf("%w: target", ErrWrongOwner)
		}
		return true, nil
	default:
		return false, fmt.Errorf("%w: target", ErrNonRegular)
	}
}

func writeTemporary(ctx context.Context, parentFD int, content []byte) (string, error) {
	for attempts := 0; attempts < 16; attempts++ {
		if err := checkContext(ctx); err != nil {
			return "", err
		}
		var nonce [16]byte
		if _, err := rand.Read(nonce[:]); err != nil {
			return "", fmt.Errorf("generate temporary workspace note name: %w", err)
		}
		name := ".workspace-note-" + hex.EncodeToString(nonce[:]) + ".tmp"
		fd, err := unix.Openat(parentFD, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o600)
		if err != nil {
			if errors.Is(err, unix.EEXIST) {
				continue
			}
			return "", fmt.Errorf("create temporary workspace note: %w", err)
		}
		file := os.NewFile(uintptr(fd), name)
		err = writeAndSync(ctx, file, content)
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
		if err != nil {
			_ = unix.Unlinkat(parentFD, name, 0)
			_ = unix.Fsync(parentFD)
			return "", fmt.Errorf("write temporary workspace note: %w", err)
		}
		return name, nil
	}
	return "", errors.New("unable to allocate unique temporary workspace note name")
}

func writeAndSync(ctx context.Context, file *os.File, content []byte) error {
	if err := file.Chmod(0o600); err != nil {
		return fmt.Errorf("set temporary workspace note mode: %w", err)
	}
	for len(content) > 0 {
		if err := checkContext(ctx); err != nil {
			return err
		}
		chunk := content
		if len(chunk) > 32*1024 {
			chunk = chunk[:32*1024]
		}
		n, err := file.Write(chunk)
		content = content[n:]
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	if err := checkContext(ctx); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("fsync temporary workspace note: %w", err)
	}
	return checkContext(ctx)
}

func readRegularAt(ctx context.Context, parentFD int, name string, maxBytes int64) ([]byte, string, error) {
	if err := checkContext(ctx); err != nil {
		return nil, "", err
	}
	fd, err := unix.Openat(parentFD, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		if errors.Is(err, unix.ELOOP) {
			return nil, "", fmt.Errorf("%w: target", ErrSymlink)
		}
		return nil, "", fmt.Errorf("open workspace note for read-back: %w", err)
	}
	file := os.NewFile(uintptr(fd), name)
	defer file.Close()

	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return nil, "", fmt.Errorf("stat workspace note for read-back: %w", err)
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return nil, "", fmt.Errorf("%w: target", ErrNonRegular)
	}
	if int(stat.Uid) != os.Geteuid() {
		return nil, "", fmt.Errorf("%w: target", ErrWrongOwner)
	}
	if stat.Size > maxBytes {
		return nil, "", fmt.Errorf("%w: target", ErrTooLarge)
	}

	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("read workspace note: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, "", fmt.Errorf("%w: target", ErrTooLarge)
	}
	if err := checkContext(ctx); err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), nil
}

func observation(payload notePayload, status WorkspaceNoteStatus, created, retryAllowed, blocked bool, readBack []byte, actualHash string) WorkspaceNoteObservation {
	return WorkspaceNoteObservation{
		ToolName:        WorkspaceCreateNoteToolName,
		Path:            payload.path,
		Status:          status,
		Created:         created,
		RetryAllowed:    retryAllowed,
		RetryLimit:      retryLimit(retryAllowed),
		Blocked:         blocked,
		Bytes:           int64(len(readBack)),
		SHA256:          actualHash,
		RequestedSHA256: payload.sha256,
		ReadBack:        append([]byte(nil), readBack...),
	}
}

func retryLimit(allowed bool) int {
	if allowed {
		return 1
	}
	return 0
}

func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
