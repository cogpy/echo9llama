package progress

import (
	"bytes"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestProgressConcurrentLifecycle(t *testing.T) {
	for range 20 {
		var output bytes.Buffer
		p := NewProgress(&output)
		spinner := NewSpinner("working")
		bar := NewBar("bytes", 100, 0)
		p.Add("spinner", spinner)
		p.Add("bar", bar)
		var wg sync.WaitGroup
		var stopped atomic.Int32
		for worker := range 8 {
			wg.Add(1)
			go func(worker int) {
				defer wg.Done()
				for i := range 25 {
					spinner.SetMessage("updated")
					bar.Set(int64(i))
					p.Add("late", NewSpinner("late"))
					p.render()
				}
				var result bool
				if worker%2 == 0 {
					result = p.Stop()
				} else {
					result = p.StopAndClear()
				}
				if result {
					stopped.Add(1)
				}
			}(worker)
		}
		wg.Wait()
		if stopped.Load() != 1 {
			t.Fatalf("Stop was not idempotent: %d", stopped.Load())
		}
		select {
		case <-p.exited:
		case <-time.After(time.Second):
			t.Fatal("progress worker did not exit")
		}
		before := output.Len()
		p.render()
		p.Add("after", NewSpinner("ignored"))
		if p.Stop() || p.StopAndClear() || output.Len() != before {
			t.Fatal("stopped renderer changed output")
		}
	}
}

func TestSpinnerAndBarConcurrentState(t *testing.T) {
	s := NewSpinner("ready")
	b := NewBar("ready", 100, 0)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for i := range 100 {
				s.SetMessage("ready")
				_ = s.String()
				s.Stop()
				b.Set(int64(i))
				_ = b.String()
			}
		})
	}
	wg.Wait()
	if strings.TrimSpace(s.String()) != "ready" {
		t.Fatalf("stopped spinner still animates: %q", s.String())
	}
}
