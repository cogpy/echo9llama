//go:build darwin

package coreself

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDarwinSystemTempAliasCanonicalizesBeforePrivateBoundaryChecks(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel, err := Open(directory, BootstrapConfig{})
	if err != nil {
		t.Fatalf("Open through Darwin temp alias: %v", err)
	}
	defer kernel.Close()

	if strings.HasPrefix(directory, "/var/") && strings.HasPrefix(kernel.directory, "/var/") {
		t.Fatalf("Darwin /var alias was not canonicalized: %q", kernel.directory)
	}
	if kernel.Status().EventCount != 1 {
		t.Fatalf("bootstrap event count = %d, want 1", kernel.Status().EventCount)
	}
}
