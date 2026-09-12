//go:build unix

package coreself

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestQuarantineFIFOCollisionIsSurfaced(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	kernel := mustOpen(t, directory)
	proposal := mustSubmit(t, kernel, evidenceProposal("proposal:quarantine-fifo", "evidence:quarantine-fifo"))
	if err := os.Rename(filepath.Join(directory, "identity-ledger.jsonl"), filepath.Join(directory, "ledger-displaced.jsonl")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "identity-ledger.jsonl"), []byte("replacement\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := unix.Mkfifo(filepath.Join(directory, "quarantine", "QUARANTINED"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := kernel.acceptProposal(proposal.ID, stewardReviewer(), testReviewerPrivateKey())
	if !errors.Is(err, ErrTampered) || !strings.Contains(err.Error(), "write quarantine marker") {
		t.Fatalf("quarantine FIFO collision error = %v, want joined ErrTampered and marker-write error", err)
	}
}

func TestCoreSelfLockHolderProcess(t *testing.T) {
	if os.Getenv("ECHO_CORE_SELF_LOCK_HELPER") != "1" {
		return
	}
	kernel, err := Open(os.Getenv("ECHO_CORE_SELF_LOCK_DIRECTORY"), BootstrapConfig{})
	if err != nil {
		t.Fatal(err)
	}
	release, err := acquireKernelLock(kernel)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	fmt.Println("LOCKED")
	select {}
}

func TestCoreSelfLockReleasesAfterKilledProcess(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "identity")
	mustOpen(t, directory)

	command := exec.Command(os.Args[0], "-test.run=^TestCoreSelfLockHolderProcess$")
	command.Env = append(os.Environ(),
		"ECHO_CORE_SELF_LOCK_HELPER=1",
		"ECHO_CORE_SELF_LOCK_DIRECTORY="+directory,
	)
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	command.Stderr = command.Stdout
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "LOCKED" {
		_ = command.Process.Kill()
		_ = command.Wait()
		t.Fatalf("lock helper did not become ready: %q (%v)", scanner.Text(), scanner.Err())
	}
	if _, err := Open(directory, BootstrapConfig{}); !errors.Is(err, ErrStaleHead) {
		_ = command.Process.Kill()
		_ = command.Wait()
		t.Fatalf("live helper lock error = %v, want ErrStaleHead", err)
	}
	if err := command.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := command.Wait(); err == nil {
		t.Fatal("killed lock helper exited without an error")
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := Open(directory, BootstrapConfig{}); err == nil {
			return
		} else if time.Now().After(deadline) {
			t.Fatalf("writer lock remained held after process death: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCoreSelfParentLockSurvivesRootReplacementAcrossProcesses(t *testing.T) {
	parent := t.TempDir()
	directory := filepath.Join(parent, "identity")
	first := mustOpen(t, directory)
	if err := os.Rename(directory, filepath.Join(parent, "identity-displaced")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(os.Args[0], "-test.run=^TestCoreSelfLockHolderProcess$")
	command.Env = append(os.Environ(), "ECHO_CORE_SELF_LOCK_HELPER=1", "ECHO_CORE_SELF_LOCK_DIRECTORY="+directory)
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	command.Stderr = command.Stdout
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "LOCKED" {
		_ = command.Process.Kill()
		_ = command.Wait()
		t.Fatalf("lock helper did not report readiness: %q (%v)", scanner.Text(), scanner.Err())
	}
	if _, err := acquireKernelLock(first); !errors.Is(err, ErrStaleHead) {
		_ = command.Process.Kill()
		_ = command.Wait()
		t.Fatalf("displaced root entered a split process lock namespace: %v", err)
	}
	if err := command.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := command.Wait(); err == nil {
		t.Fatal("killed lock helper exited without an error")
	}
	if _, err := acquireKernelLock(first); !errors.Is(err, ErrTampered) {
		t.Fatalf("displaced root was not rejected after helper exit: %v", err)
	}
}
