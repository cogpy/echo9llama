package embodiment

import (
	"encoding/json"
	"net/http"
)

const (
	// Path accepts POST (ingest a Frame → Reflection) and GET (latest Reflection).
	Path = "/api/dte/embodiment"
	// HistoryPath answers GET with the retained frames, oldest first.
	HistoryPath = Path + "/history"
)

// Register mounts the embodiment endpoints on mux.
func (h *Hub) Register(mux *http.ServeMux) {
	mux.HandleFunc(Path, h.handleFrame)
	mux.HandleFunc(HistoryPath, h.handleHistory)
}

func (h *Hub) handleFrame(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		var f Frame
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 1<<20)).Decode(&f); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		r, err := h.Ingest(f)
		if err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, r)
	case http.MethodGet:
		r, ok := h.Last()
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "no embodiment frame received yet"})
			return
		}
		writeJSON(w, http.StatusOK, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *Hub) handleHistory(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"contract": Contract, "frames": h.History()})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
