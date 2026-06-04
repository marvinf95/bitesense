package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/marvinf95/bitesense/backend/internal/models"
)

type SymptomsHandler struct {
	DB *sql.DB
}

type symptomRequest struct {
	OccurredAt   time.Time `json:"occurred_at"`
	Type         string    `json:"type"`
	Severity     int       `json:"severity"`
	DurationMin  *int      `json:"duration_min,omitempty"`
	BristolStool *int      `json:"bristol_stool,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
}

func (h *SymptomsHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	var req symptomRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if !models.AllSymptomTypes[req.Type] {
		writeError(w, http.StatusBadRequest, "unknown symptom type")
		return
	}
	if req.Severity < 1 || req.Severity > 10 {
		writeError(w, http.StatusBadRequest, "severity must be 1..10")
		return
	}
	if req.BristolStool != nil && (*req.BristolStool < 1 || *req.BristolStool > 7) {
		writeError(w, http.StatusBadRequest, "bristol_stool must be 1..7")
		return
	}
	if req.OccurredAt.IsZero() {
		req.OccurredAt = time.Now()
	}

	id := uuid.New().String()
	_, err := h.DB.ExecContext(r.Context(), `
		INSERT INTO symptoms (id, user_id, occurred_at, type, severity, duration_min, bristol_stool, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, uid, req.OccurredAt, req.Type, req.Severity, req.DurationMin, req.BristolStool, req.Notes,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "insert symptom")
		return
	}
	writeJSON(w, http.StatusCreated, models.Symptom{
		ID: id, UserID: uid, OccurredAt: req.OccurredAt, Type: req.Type,
		Severity: req.Severity, DurationMin: req.DurationMin, BristolStool: req.BristolStool,
		Notes: req.Notes, CreatedAt: time.Now(),
	})
}

func (h *SymptomsHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	from, to := parseRange(r)
	typeFilter := r.URL.Query().Get("type")

	q := `SELECT id, occurred_at, type, severity, duration_min, bristol_stool, notes, created_at
	      FROM symptoms WHERE user_id = ? AND occurred_at BETWEEN ? AND ?`
	args := []any{uid, from, to}
	if typeFilter != "" {
		q += " AND type = ?"
		args = append(args, typeFilter)
	}
	q += " ORDER BY occurred_at DESC LIMIT 500"

	rows, err := h.DB.QueryContext(r.Context(), q, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query")
		return
	}
	defer rows.Close()

	out := []models.Symptom{}
	for rows.Next() {
		var s models.Symptom
		s.UserID = uid
		var dur, br sql.NullInt64
		var notes sql.NullString
		if err := rows.Scan(&s.ID, &s.OccurredAt, &s.Type, &s.Severity, &dur, &br, &notes, &s.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "scan")
			return
		}
		if dur.Valid {
			v := int(dur.Int64)
			s.DurationMin = &v
		}
		if br.Valid {
			v := int(br.Int64)
			s.BristolStool = &v
		}
		if notes.Valid {
			v := notes.String
			s.Notes = &v
		}
		out = append(out, s)
	}
	writeJSON(w, http.StatusOK, map[string]any{"symptoms": out, "from": from, "to": to})
}

func (h *SymptomsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	res, err := h.DB.ExecContext(r.Context(), `DELETE FROM symptoms WHERE id = ? AND user_id = ?`, id, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "delete")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Patch updates mutable fields.
func (h *SymptomsHandler) Patch(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	var req symptomRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Severity != 0 && (req.Severity < 1 || req.Severity > 10) {
		writeError(w, http.StatusBadRequest, "severity must be 1..10")
		return
	}
	res, err := h.DB.ExecContext(r.Context(), `
		UPDATE symptoms SET
			occurred_at = COALESCE(NULLIF(?, ''), occurred_at),
			type = COALESCE(NULLIF(?, ''), type),
			severity = COALESCE(NULLIF(?, 0), severity),
			duration_min = ?,
			bristol_stool = ?,
			notes = ?
		WHERE id = ? AND user_id = ?`,
		req.OccurredAt, req.Type, req.Severity, req.DurationMin, req.BristolStool, req.Notes, id, uid,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

var _ = errors.Is
