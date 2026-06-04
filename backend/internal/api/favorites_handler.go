package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/marvinf95/bitesense/backend/internal/models"
)

type FavoritesHandler struct {
	DB *sql.DB
}

type favoriteReq struct {
	Label    string `json:"label"`
	Template string `json:"template"` // raw JSON describing items + tags
}

func (h *FavoritesHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	var req favoriteReq
	if err := decodeJSON(r, &req); err != nil || req.Label == "" || req.Template == "" {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.DB.ExecContext(r.Context(),
		`INSERT INTO meal_favorites (id, user_id, label, template, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, uid, req.Label, req.Template, now,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "insert")
		return
	}
	writeJSON(w, http.StatusCreated, models.MealFavorite{
		ID: id, UserID: uid, Label: req.Label, Template: req.Template, CreatedAt: now,
	})
}

func (h *FavoritesHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	rows, err := h.DB.QueryContext(r.Context(),
		`SELECT id, label, template, created_at FROM meal_favorites WHERE user_id = ? ORDER BY label`, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query")
		return
	}
	defer rows.Close()
	out := []models.MealFavorite{}
	for rows.Next() {
		var f models.MealFavorite
		f.UserID = uid
		if err := rows.Scan(&f.ID, &f.Label, &f.Template, &f.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "scan")
			return
		}
		out = append(out, f)
	}
	writeJSON(w, http.StatusOK, map[string]any{"favorites": out})
}

func (h *FavoritesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	res, err := h.DB.ExecContext(r.Context(), `DELETE FROM meal_favorites WHERE id = ? AND user_id = ?`, id, uid)
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
