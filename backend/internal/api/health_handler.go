package api

import (
	"database/sql"
	"net/http"
)

type HealthHandler struct {
	DB *sql.DB
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.PingContext(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "db down")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
