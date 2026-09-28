package discover

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func IsNUMA() bool {
	if runtime.GOOS != "linux" {
		// numa support in llama.cpp is linux only
		return false
	}
	ids := map[string]any{}
	packageIds, _ := filepath.Glob("/sys/devices/system/cpu/cpu*/topology/physical_package_id")
	for _, packageID := range packageIds {
		id, err := os.ReadFile(packageID)
		if err == nil {
			ids[strings.TrimSpace(string(id))] = struct{}{}
		}
	}
	return len(ids) > 1
}
