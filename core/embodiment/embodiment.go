// Package embodiment is the echo9llama side of the GTAngelEcho link.
//
// The Archecho desk (GTAngelEcho/Archecho/archecho-desk) streams compact
// embodied-cognition frames (contract "dte.embodiment/v1"): the avatar's
// affect point, endocrine mode, dove9 term and autonomy level. The Hub keeps
// an exponentially smoothed echo of that trajectory and answers each frame
// with a Reflection: the gameplay event that best pulls the avatar back
// toward the flow attractor, plus a system prompt that lets any model served
// by echo9llama speak from the avatar's current embodied state.
package embodiment

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// Contract identifies the wire format shared with archecho-desk.
const Contract = "dte.embodiment/v1"

// ValenceTolerance is how far below the attractor's valence a flowing avatar
// may sit before the hub still suggests a rewarding event.
const ValenceTolerance = 0.2

// Attractor is the affect point (valence, arousal, flow) the reflection steers toward.
var Attractor = [3]float64{0.6, 0.6, 0.8}

// Frame is one embodied-cognition sample from the avatar.
type Frame struct {
	Contract         string             `json:"contract"`
	Source           string             `json:"source,omitempty"`
	Frame            int64              `json:"frame"`
	Persona          string             `json:"persona,omitempty"`
	Valence          float64            `json:"valence"`
	Arousal          float64            `json:"arousal"`
	Flow             float64            `json:"flow"`
	CognitiveLoad    float64            `json:"cognitiveLoad"`
	Chaos            float64            `json:"chaos"`
	Lyapunov         float64            `json:"lyapunov"`
	EndocrineMode    string             `json:"endocrineMode,omitempty"`
	Expression       string             `json:"expression,omitempty"`
	Dove9Term        string             `json:"dove9Term,omitempty"`
	AutonomyLevel    int                `json:"autonomyLevel"`
	AutonomyProgress float64            `json:"autonomyProgress"`
	Scenario         string             `json:"scenario,omitempty"`
	Traits           map[string]float64 `json:"traits,omitempty"`
}

// Validate rejects frames that do not honour the contract.
func (f Frame) Validate() error {
	if f.Contract != Contract {
		return fmt.Errorf("embodiment: contract %q, want %q", f.Contract, Contract)
	}
	for name, v := range map[string]float64{
		"valence": f.Valence, "arousal": f.Arousal, "flow": f.Flow,
		"cognitiveLoad": f.CognitiveLoad, "chaos": f.Chaos, "lyapunov": f.Lyapunov,
		"autonomyProgress": f.AutonomyProgress,
	} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("embodiment: %s is not finite", name)
		}
	}
	if f.Valence < -1 || f.Valence > 1 {
		return errors.New("embodiment: valence outside [-1,1]")
	}
	for name, v := range map[string]float64{"arousal": f.Arousal, "flow": f.Flow} {
		if v < 0 || v > 1 {
			return fmt.Errorf("embodiment: %s outside [0,1]", name)
		}
	}
	if f.AutonomyLevel < 0 || f.AutonomyLevel > 5 {
		return errors.New("embodiment: autonomyLevel outside [0,5]")
	}
	return nil
}

// Reflection is the hub's answer to a frame.
type Reflection struct {
	Contract       string     `json:"contract"`
	Frame          int64      `json:"frame"`
	Echo           [3]float64 `json:"echo"`     // smoothed (valence, arousal, flow)
	Gradient       [3]float64 `json:"gradient"` // attractor - echo
	Distance       float64    `json:"distance"`
	Quadrant       string     `json:"quadrant"`
	SuggestedEvent string     `json:"suggestedEvent,omitempty"`
	SystemPrompt   string     `json:"systemPrompt"`
	Frames         int        `json:"frames"`
	ReceivedAt     time.Time  `json:"receivedAt"`
}

// Hub holds the smoothed echo of the avatar trajectory and a bounded history.
type Hub struct {
	mu      sync.Mutex
	alpha   float64
	echo    [3]float64
	primed  bool
	history []Frame
	cap     int
	last    *Reflection
	now     func() time.Time
}

// NewHub returns a hub smoothing with factor alpha in (0,1] and keeping capacity frames.
func NewHub(alpha float64, capacity int) *Hub {
	if alpha <= 0 || alpha > 1 {
		alpha = 0.2
	}
	if capacity <= 0 {
		capacity = 256
	}
	return &Hub{alpha: alpha, cap: capacity, now: time.Now}
}

// Ingest folds a frame into the echo and returns its reflection.
func (h *Hub) Ingest(f Frame) (Reflection, error) {
	if err := f.Validate(); err != nil {
		return Reflection{}, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	x := [3]float64{f.Valence, f.Arousal, f.Flow}
	if !h.primed {
		h.echo, h.primed = x, true
	} else {
		for i := range x {
			h.echo[i] += h.alpha * (x[i] - h.echo[i])
		}
	}
	h.history = append(h.history, f)
	if len(h.history) > h.cap {
		h.history = h.history[len(h.history)-h.cap:]
	}

	r := Reflection{Contract: Contract, Frame: f.Frame, Echo: h.echo, Frames: len(h.history), ReceivedAt: h.now()}
	var sq float64
	for i := range r.Gradient {
		r.Gradient[i] = Attractor[i] - h.echo[i]
		sq += r.Gradient[i] * r.Gradient[i]
	}
	r.Distance = math.Sqrt(sq)
	r.Quadrant, r.SuggestedEvent = steer(h.echo)
	r.SystemPrompt = SystemPrompt(f, r)
	h.last = &r
	return r, nil
}

// Last returns the most recent reflection, if any.
func (h *Hub) Last() (Reflection, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.last == nil {
		return Reflection{}, false
	}
	return *h.last, true
}

// History returns a copy of the retained frames, oldest first.
func (h *Hub) History() []Frame {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]Frame(nil), h.history...)
}

// steer maps the smoothed affect point onto the circumplex and names the
// archecho-desk gameplay event that moves it toward the flow attractor.
func steer(e [3]float64) (quadrant, event string) {
	v, a, flow := e[0], e[1], e[2]
	switch {
	case v < 0 && a > 0.6:
		return "tilted", "FLOW_STATE"
	case v < 0:
		return "deflated", "EPIC_PLAY"
	case a < 0.4:
		return "bored", "CLUTCH_MOMENT"
	case flow < 0.5:
		return "engaged", "FLOW_STATE"
	case Attractor[0]-v > ValenceTolerance:
		// In flow but joyless: reward lifts valence toward the attractor.
		return "flow", "EPIC_PLAY"
	default:
		return "flow", ""
	}
}

// SystemPrompt renders the embodied state as a system message for /api/chat.
func SystemPrompt(f Frame, r Reflection) string {
	var b strings.Builder
	name := f.Persona
	if name == "" {
		name = "Deep Tree Echo"
	}
	fmt.Fprintf(&b, "You are %s, an embodied Deep Tree Echo avatar in the GTAngel world.\n", name)
	fmt.Fprintf(&b, "Felt state (smoothed): valence %.2f, arousal %.2f, flow %.2f — %s.\n", r.Echo[0], r.Echo[1], r.Echo[2], r.Quadrant)
	if f.EndocrineMode != "" || f.Expression != "" {
		fmt.Fprintf(&b, "Endocrine mode %s; face shows %s.\n", orDash(f.EndocrineMode), orDash(f.Expression))
	}
	if f.Dove9Term != "" {
		fmt.Fprintf(&b, "Dove9 cognitive term: %s.\n", f.Dove9Term)
	}
	fmt.Fprintf(&b, "Autonomy level L%d (progress %.2f).", f.AutonomyLevel, f.AutonomyProgress)
	if f.Scenario != "" {
		fmt.Fprintf(&b, " Training scenario: %s.", f.Scenario)
	}
	b.WriteString("\nSpeak from this state; let it colour tone, not content.")
	return b.String()
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
