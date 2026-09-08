package control

import (
	"encoding/json"
	"net/http"

	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/bootstrap"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/config"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/runtime"
)

type Handler struct {
	cfg   config.Config
	state *runtime.State
}

func New(cfg config.Config, state *runtime.State) *Handler {
	return &Handler{cfg: cfg, state: state}
}

func (h *Handler) HandleMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Mode == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	switch req.Mode {
	case "auto_accept", "auto_reject", "silent", "broken_webhook":
		h.state.SetMode(req.Mode)
	default:
		http.Error(w, "unknown mode", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"mode": req.Mode})
}

func (h *Handler) HandleStoplist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ExternalID  string `json:"external_id"`
		IsAvailable *bool  `json:"is_available"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ExternalID == "" || req.IsAvailable == nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := bootstrap.SetItemAvailability(h.cfg, req.ExternalID, *req.IsAvailable); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"external_id":   req.ExternalID,
		"is_available": *req.IsAvailable,
	})
}

func (h *Handler) HandleMenuReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := bootstrap.SyncMenu(h.cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
