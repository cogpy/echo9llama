package progress

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

const (
	defaultTermWidth  = 80
	defaultTermHeight = 24
)

type State interface{ String() string }

type Progress struct {
	mu      sync.Mutex
	w       *bufio.Writer
	pos     int
	ticker  *time.Ticker
	states  []State
	done    chan struct{}
	exited  chan struct{}
	stopped bool
}

func NewProgress(w io.Writer) *Progress {
	p := &Progress{w: bufio.NewWriter(w), ticker: time.NewTicker(100 * time.Millisecond), done: make(chan struct{}), exited: make(chan struct{})}
	go p.start()
	return p
}

// stopLocked owns all renderer state and buffered output. Caller holds p.mu.
func (p *Progress) stopLocked() bool {
	if p.stopped {
		return false
	}
	p.stopped = true
	p.ticker.Stop()
	close(p.done)
	for _, state := range p.states {
		if spinner, ok := state.(*Spinner); ok {
			spinner.Stop()
		}
	}
	p.renderLocked()
	return true
}

func (p *Progress) Stop() bool {
	p.mu.Lock()
	stopped := p.stopLocked()
	if stopped {
		fmt.Fprint(p.w, "\n")
		p.w.Flush()
	}
	p.mu.Unlock()
	<-p.exited
	return stopped
}

func (p *Progress) StopAndClear() bool {
	p.mu.Lock()
	stopped := p.stopLocked()
	if stopped {
		fmt.Fprint(p.w, "\033[?25l")
		for i := range p.pos {
			if i > 0 {
				fmt.Fprint(p.w, "\033[A")
			}
			fmt.Fprint(p.w, "\033[2K\033[1G")
		}
		fmt.Fprint(p.w, "\033[?25h")
		p.w.Flush()
	}
	p.mu.Unlock()
	<-p.exited
	return stopped
}

func (p *Progress) Add(key string, state State) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		if spinner, ok := state.(*Spinner); ok {
			spinner.Stop()
		}
		return
	}
	p.states = append(p.states, state)
}

func (p *Progress) render() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.stopped {
		p.renderLocked()
	}
}

func (p *Progress) renderLocked() {
	_, termHeight, err := term.GetSize(int(os.Stderr.Fd()))
	if err != nil {
		termHeight = defaultTermHeight
	}
	defer p.w.Flush()
	fmt.Fprint(p.w, "\033[?2026h")
	defer fmt.Fprint(p.w, "\033[?2026l")
	fmt.Fprint(p.w, "\033[?25l")
	defer fmt.Fprint(p.w, "\033[?25h")
	for range p.pos - 1 {
		fmt.Fprint(p.w, "\033[A")
	}
	fmt.Fprint(p.w, "\033[1G")
	maxHeight := min(len(p.states), termHeight)
	for i := len(p.states) - maxHeight; i < len(p.states); i++ {
		fmt.Fprint(p.w, p.states[i].String(), "\033[K")
		if i < len(p.states)-1 {
			fmt.Fprint(p.w, "\n")
		}
	}
	p.pos = len(p.states)
}

func (p *Progress) start() {
	defer close(p.exited)
	for {
		select {
		case <-p.done:
			return
		case <-p.ticker.C:
			p.render()
		}
	}
}
