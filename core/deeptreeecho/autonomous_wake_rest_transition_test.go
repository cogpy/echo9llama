package deeptreeecho

import (
	"testing"
	"time"
)

func TestRestStatePublishesOnlyAfterQuiescenceCallback(t *testing.T) {
	manager := NewAutonomousWakeRestManager()
	entered := make(chan WakeRestState, 1)
	release := make(chan struct{})
	done := make(chan struct{})
	manager.SetCallbacks(nil, func() error {
		entered <- manager.GetState()
		<-release
		return nil
	}, nil, nil)

	go func() {
		manager.transitionToRest()
		close(done)
	}()

	select {
	case state := <-entered:
		if state != StateAwake {
			t.Fatalf("rest became observable before quiescence: %s", state)
		}
	case <-time.After(time.Second):
		t.Fatal("rest callback was not entered")
	}
	if state := manager.GetState(); state != StateAwake {
		t.Fatalf("state changed while quiescence callback was blocked: %s", state)
	}
	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("rest transition did not finish")
	}
	if state := manager.GetState(); state != StateResting {
		t.Fatalf("rest state was not published after quiescence: %s", state)
	}
}
