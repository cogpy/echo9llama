package embodiment

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func fixture(t *testing.T) Frame {
	t.Helper()
	b, err := os.ReadFile("testdata/frame.json")
	if err != nil {
		t.Fatal(err)
	}
	var f Frame
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFixtureReflectsTilt(t *testing.T) {
	r, err := NewHub(0.2, 8).Ingest(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.Quadrant != "tilted" || r.SuggestedEvent != "FLOW_STATE" {
		t.Fatalf("got %s/%s, want tilted/FLOW_STATE", r.Quadrant, r.SuggestedEvent)
	}
	if !strings.Contains(r.SystemPrompt, "Lucy") || !strings.Contains(r.SystemPrompt, "L1") {
		t.Fatalf("prompt lacks persona or autonomy: %q", r.SystemPrompt)
	}
}

func TestEchoConvergesToAttractorAndGoesQuiet(t *testing.T) {
	h := NewHub(0.5, 4)
	f := fixture(t)
	f.Valence, f.Arousal, f.Flow = Attractor[0], Attractor[1], Attractor[2]
	var r Reflection
	for i := 0; i < 20; i++ {
		f.Frame = int64(i)
		var err error
		if r, err = h.Ingest(f); err != nil {
			t.Fatal(err)
		}
	}
	if r.Distance > 1e-9 || r.SuggestedEvent != "" || r.Quadrant != "flow" {
		t.Fatalf("at attractor got distance=%g event=%q quadrant=%s", r.Distance, r.SuggestedEvent, r.Quadrant)
	}
	if got := len(h.History()); got != 4 {
		t.Fatalf("history len %d, want capacity 4", got)
	}
}

func TestSteerQuadrants(t *testing.T) {
	for _, c := range []struct {
		e     [3]float64
		q, ev string
	}{
		{[3]float64{-0.5, 0.9, 0}, "tilted", "FLOW_STATE"},
		{[3]float64{-0.5, 0.2, 0}, "deflated", "EPIC_PLAY"},
		{[3]float64{0.5, 0.1, 0}, "bored", "CLUTCH_MOMENT"},
		{[3]float64{0.5, 0.5, 0.2}, "engaged", "FLOW_STATE"},
		{[3]float64{0.5, 0.5, 0.9}, "flow", ""},
	} {
		if q, ev := steer(c.e); q != c.q || ev != c.ev {
			t.Errorf("steer(%v) = %s/%q, want %s/%q", c.e, q, ev, c.q, c.ev)
		}
	}
}

func TestValidateRejects(t *testing.T) {
	for name, mut := range map[string]func(*Frame){
		"contract": func(f *Frame) { f.Contract = "v0" },
		"nan":      func(f *Frame) { f.Lyapunov = math.NaN() },
		"valence":  func(f *Frame) { f.Valence = 2 },
		"flow":     func(f *Frame) { f.Flow = -0.1 },
		"level":    func(f *Frame) { f.AutonomyLevel = 6 },
	} {
		f := fixture(t)
		mut(&f)
		if f.Validate() == nil {
			t.Errorf("%s: expected rejection", name)
		}
	}
}

func TestRoutes(t *testing.T) {
	r := http.NewServeMux()
	NewHub(0.2, 8).Register(r)

	do := func(method, path string, body []byte) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		return w
	}

	if w := do(http.MethodGet, "/api/dte/embodiment", nil); w.Code != http.StatusNotFound {
		t.Fatalf("empty GET = %d", w.Code)
	}
	body, _ := os.ReadFile("testdata/frame.json")
	w := do(http.MethodPost, "/api/dte/embodiment", body)
	if w.Code != http.StatusOK {
		t.Fatalf("POST = %d %s", w.Code, w.Body)
	}
	var ref Reflection
	if err := json.Unmarshal(w.Body.Bytes(), &ref); err != nil || ref.SuggestedEvent != "FLOW_STATE" {
		t.Fatalf("reflection %+v err=%v", ref, err)
	}
	if w := do(http.MethodPost, "/api/dte/embodiment", []byte(`{"contract":"nope"}`)); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("bad contract = %d", w.Code)
	}
	if w := do(http.MethodGet, "/api/dte/embodiment", nil); w.Code != http.StatusOK {
		t.Fatalf("GET after POST = %d", w.Code)
	}
	if w := do(http.MethodDelete, "/api/dte/embodiment", nil); w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("DELETE = %d", w.Code)
	}
	if w := do(http.MethodGet, "/api/dte/embodiment/history", nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"frame":31`) {
		t.Fatalf("history = %d %s", w.Code, w.Body)
	}
}
