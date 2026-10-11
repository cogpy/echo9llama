package progress

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Spinner struct {
	mu           sync.Mutex
	message      atomic.Value
	messageWidth int
	parts        []string
	started      time.Time
	stopped      time.Time
}

func NewSpinner(message string) *Spinner {
	s := &Spinner{parts: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}, started: time.Now()}
	s.SetMessage(message)
	return s
}

func (s *Spinner) SetMessage(message string) { s.message.Store(message) }

func (s *Spinner) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var sb strings.Builder
	if message, ok := s.message.Load().(string); ok && len(message) > 0 {
		message = strings.TrimSpace(message)
		if s.messageWidth > 0 && len(message) > s.messageWidth {
			message = message[:s.messageWidth]
		}
		fmt.Fprint(&sb, message)
		if padding := s.messageWidth - sb.Len(); padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(" ")
	}
	if s.stopped.IsZero() && len(s.parts) > 0 {
		// Derive animation from time rather than a detached ticker goroutine.
		frame := int(time.Since(s.started)/(100*time.Millisecond)) % len(s.parts)
		sb.WriteString(s.parts[frame])
		sb.WriteString(" ")
	}
	return sb.String()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped.IsZero() {
		s.stopped = time.Now()
	}
}
