package deeptreeecho

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPersistentConsciousnessStateUsesPrivateAtomicStorage(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not enforced on Windows")
	}

	stateDir := filepath.Join(t.TempDir(), "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatalf("prepare state directory: %v", err)
	}

	state, err := NewPersistentConsciousnessState(stateDir, "Echo")
	if err != nil {
		t.Fatalf("NewPersistentConsciousnessState failed: %v", err)
	}
	if err := state.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	directoryInfo, err := os.Stat(stateDir)
	if err != nil {
		t.Fatalf("stat state directory: %v", err)
	}
	if got := directoryInfo.Mode().Perm(); got != 0700 {
		t.Fatalf("expected state directory mode 0700, got %04o", got)
	}

	stateFile := filepath.Join(stateDir, "consciousness_state.json")
	fileInfo, err := os.Stat(stateFile)
	if err != nil {
		t.Fatalf("stat state file: %v", err)
	}
	if got := fileInfo.Mode().Perm(); got != 0600 {
		t.Fatalf("expected state file mode 0600, got %04o", got)
	}

	temporaryFiles, err := filepath.Glob(filepath.Join(stateDir, ".consciousness-state-*.tmp"))
	if err != nil {
		t.Fatalf("glob temporary files: %v", err)
	}
	if len(temporaryFiles) != 0 {
		t.Fatalf("atomic save left temporary files behind: %v", temporaryFiles)
	}
}
