//go:build unix

package coreself

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func lockFile(file *os.File) error {
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
			return fmt.Errorf("%w: ledger lock held", ErrStaleHead)
		}
		return err
	}
	return nil
}

func unlockFile(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_UN)
}
