package deeptreeecho

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cogpy/echo9llama/core/coreself"
)

func (uao *UnifiedAutonomousOrchestrator) initializeCoreSelf() {
	if !uao.coreSelfRequired {
		return
	}
	if !uao.config.EnablePersistence || uao.persistentState == nil {
		uao.coreSelfInitErr = fmt.Errorf("core-self requires private durable persistence")
		return
	}

	directory := strings.TrimSpace(uao.config.CoreSelfDirectory)
	if directory == "" {
		directory = filepath.Join(uao.config.StateDirectory, "core-self")
	}
	absolute, err := filepath.Abs(directory)
	if err != nil {
		uao.coreSelfInitErr = err
		return
	}
	uao.config.CoreSelfDirectory = filepath.Clean(absolute)
	kernel, err := coreself.Open(uao.config.CoreSelfDirectory, coreself.BootstrapConfig{
		ReviewerPublicKey: strings.TrimSpace(uao.config.CoreSelfReviewerPublicKey),
	})
	if err != nil {
		uao.coreSelfInitErr = err
		return
	}
	uao.coreSelf = kernel
	fmt.Println("   ✓ E3 deterministic core-self verified")
}
