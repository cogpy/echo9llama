package deeptreeecho

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/coreself"
)

func coreSelfTestConfig(t *testing.T) OrchestratorConfig {
	t.Helper()
	config := DefaultOrchestratorConfig()
	config.StateDirectory = filepath.Join(t.TempDir(), "state")
	config.EventStorePath = filepath.Join(config.StateDirectory, "events.db")
	config.MainLoopInterval = time.Hour
	config.EnableStreamOfConsciousness = false
	config.EnableEchobeats = false
	config.EnableEchodream = false
	config.EnableDiscussionMonitoring = false
	config.EnableSkillLearning = false
	config.EnableWisdomSynthesis = false
	config.EnableCognitiveCore = false
	config.EnableEnaction = false
	config.AutoWakeRest = false
	return config
}

func TestCoreSelfBootstrapAndStatusIntegration(t *testing.T) {
	config := coreSelfTestConfig(t)
	orchestrator := NewUnifiedAutonomousOrchestrator(nil, config)
	status := orchestrator.GetStatus()
	if !status.CoreSelfEnabled || !status.CoreSelfReady {
		t.Fatalf("core-self status = %+v", status)
	}
	if status.CoreSelf.EventCount != 1 || status.CoreSelf.HeadHash == "" || status.CoreSelf.StateDigest == "" {
		t.Fatalf("unexpected core-self projection: %+v", status.CoreSelf)
	}
	if strings.Contains(status.CoreSelf.HeadHash, config.StateDirectory) || strings.Contains(status.CoreSelf.StateDigest, config.StateDirectory) {
		t.Fatal("core-self status exposed a private path")
	}
}

func TestCoreSelfTamperFailsAwakeningClosed(t *testing.T) {
	config := coreSelfTestConfig(t)
	first := NewUnifiedAutonomousOrchestrator(nil, config)
	if !first.GetStatus().CoreSelfReady {
		t.Fatal("initial core-self did not bootstrap")
	}
	ledgerPath := filepath.Join(config.StateDirectory, "core-self", "identity-ledger.jsonl")
	ledger, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	ledger[len(ledger)-2] ^= 1
	if err := os.WriteFile(ledgerPath, ledger, 0o600); err != nil {
		t.Fatal(err)
	}

	second := NewUnifiedAutonomousOrchestrator(nil, config)
	if second.GetStatus().CoreSelfReady {
		t.Fatal("tampered core-self reported ready")
	}
	if err := second.Awaken(); err == nil || !errors.Is(second.coreSelfInitErr, coreself.ErrTampered) {
		t.Fatalf("tampered core-self awakening error = %v, init = %v", err, second.coreSelfInitErr)
	}
}

func TestRequestedCoreSelfWithoutPersistenceFailsClosed(t *testing.T) {
	config := coreSelfTestConfig(t)
	config.EnablePersistence = false
	orchestrator := NewUnifiedAutonomousOrchestrator(nil, config)
	if !orchestrator.coreSelfRequired || orchestrator.GetStatus().CoreSelfReady {
		t.Fatalf("requested core-self was silently disabled: %+v", orchestrator.GetStatus())
	}
	if err := orchestrator.Awaken(); err == nil || !strings.Contains(err.Error(), "core-self") {
		t.Fatalf("Awaken error = %v, want core-self failure", err)
	}
}
