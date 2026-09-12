//go:build !unix && !windows

package coreself

import (
	"fmt"
	"os"
)

func lockFile(_ *os.File) error {
	return fmt.Errorf("%w: advisory file locking is unsupported on this platform", ErrUnauthorized)
}

func unlockFile(_ *os.File) error {
	return nil
}
