package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/marvinf95/bitesense/backend/internal/models"
)

type MealsHandler struct {
	DB *sql.DB
}

type mealRequest struct {
	EatenAt time.Time          `json:"eaten_at"`
	Title   *string            `json:"title,omitempty"`
	Notes   *string            `json:"notes,omitempty"`
	Source  string             `json:"source"` // text|image|barcode|favorite
	Items   []mealItemRequest  `json:"items"`
}

type mealItemRequest struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	QuantityG   *float64 `json:"quantity_g,omitempty"`
	OFFID       *string  `json:"off_id,omitempty"`
	Confidence  *float64 `json:"confidence,omitempty"`
	Tags        []string `json:"tags"`
}

var validSources = map[string]bool{"text": true, "image": true, "barcode": true, "favorite": true}

func (h *MealsHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	var req mealRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if !validSources[req.Source] {
		writeError(w, http.StatusBadRequest, "invalid source")
		return
	}
	if req.EatenAt.IsZero() {
		req.EatenAt = time.Now()
	}
	for _, it := range req.Items {
		if strings.TrimSpace(it.Name) == "" {
			writeError(w, http.StatusBadRequest, "item name required")
			return
		}
		for _, t := range it.Tags {
			if !models.AllTags[t] {
				writeError(w, http.StatusBadRequest, "unknown tag: "+t)
				return
			}
		}
	}

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "begin")
		return
	}
	defer func() { _ = tx.Rollback() }()

	mealID := uuid.New().String()
	now := time.Now()
	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO meals (id, user_id, eaten_at, title, notes, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		mealID, uid, req.EatenAt, req.Title, req.Notes, req.Source, now, now,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "insert meal")
		return
	}
	items := make([]models.MealItem, 0, len(req.Items))
	for _, it := range req.Items {
		itemID := uuid.New().String()
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO meal_items (id, meal_id, name, display_name, quantity_g, off_id, confidence)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			itemID, mealID, strings.ToLower(strings.TrimSpace(it.Name)), it.DisplayName, it.QuantityG, it.OFFID, it.Confidence,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "insert item")
			return
		}
		for _, t := range it.Tags {
			_, err = tx.ExecContext(r.Context(), `INSERT INTO meal_item_tags (meal_item_id, tag) VALUES (?, ?)`, itemID, t)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "insert tag")
				return
			}
		}
		items = append(items, models.MealItem{
			ID: itemID, MealID: mealID, Name: strings.ToLower(it.Name),
			DisplayName: it.DisplayName, QuantityG: it.QuantityG, OFFID: it.OFFID,
			Confidence: it.Confidence, Tags: it.Tags,
		})
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "commit")
		return
	}
	writeJSON(w, http.StatusCreated, models.Meal{
		ID: mealID, UserID: uid, EatenAt: req.EatenAt, Title: req.Title, Notes: req.Notes,
		Source: req.Source, Items: items, CreatedAt: now, UpdatedAt: now,
	})
}

func (h *MealsHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())

	from, to := parseRange(r)
	rows, err := h.DB.QueryContext(r.Context(), `
		SELECT id, eaten_at, title, notes, source, photo_path, created_at, updated_at
		FROM meals
		WHERE user_id = ? AND eaten_at BETWEEN ? AND ?
		ORDER BY eaten_at DESC
		LIMIT 500`,
		uid, from, to,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query meals")
		return
	}
	defer rows.Close()

	meals := []models.Meal{}
	mealIDs := []any{}
	for rows.Next() {
		var m models.Meal
		var photo sql.NullString
		if err := rows.Scan(&m.ID, &m.EatenAt, &m.Title, &m.Notes, &m.Source, &photo, &m.CreatedAt, &m.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "scan")
			return
		}
		if photo.Valid {
			s := photo.String
			m.PhotoPath = &s
		}
		m.UserID = uid
		m.Items = []models.MealItem{}
		meals = append(meals, m)
		mealIDs = append(mealIDs, m.ID)
	}
	if len(mealIDs) > 0 {
		items, err := loadItemsForMeals(r, h.DB, mealIDs)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "items")
			return
		}
		for i := range meals {
			meals[i].Items = items[meals[i].ID]
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"meals": meals, "from": from, "to": to})
}

func (h *MealsHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	var m models.Meal
	var photo sql.NullString
	err := h.DB.QueryRowContext(r.Context(), `
		SELECT id, eaten_at, title, notes, source, photo_path, created_at, updated_at
		FROM meals WHERE id = ? AND user_id = ?`, id, uid,
	).Scan(&m.ID, &m.EatenAt, &m.Title, &m.Notes, &m.Source, &photo, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup")
		return
	}
	if photo.Valid {
		s := photo.String
		m.PhotoPath = &s
	}
	m.UserID = uid
	items, err := loadItemsForMeals(r, h.DB, []any{m.ID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "items")
		return
	}
	m.Items = items[m.ID]
	writeJSON(w, http.StatusOK, m)
}

func (h *MealsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	res, err := h.DB.ExecContext(r.Context(), `DELETE FROM meals WHERE id = ? AND user_id = ?`, id, uid)
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

type patchMealReq struct {
	Title   *string  `json:"title,omitempty"`
	Notes   *string  `json:"notes,omitempty"`
	EatenAt *time.Time `json:"eaten_at,omitempty"`
}

func (h *MealsHandler) Patch(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	var req patchMealReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	sets := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []any{}
	if req.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Notes != nil {
		sets = append(sets, "notes = ?")
		args = append(args, *req.Notes)
	}
	if req.EatenAt != nil {
		sets = append(sets, "eaten_at = ?")
		args = append(args, *req.EatenAt)
	}
	args = append(args, id, uid)
	q := fmt.Sprintf(`UPDATE meals SET %s WHERE id = ? AND user_id = ?`, strings.Join(sets, ", "))
	res, err := h.DB.ExecContext(r.Context(), q, args...)
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

func loadItemsForMeals(r *http.Request, db *sql.DB, mealIDs []any) (map[string][]models.MealItem, error) {
	if len(mealIDs) == 0 {
		return map[string][]models.MealItem{}, nil
	}
	placeholders := strings.Repeat("?,", len(mealIDs))
	placeholders = strings.TrimRight(placeholders, ",")
	q := fmt.Sprintf(`
		SELECT mi.id, mi.meal_id, mi.name, mi.display_name, mi.quantity_g, mi.off_id, mi.confidence,
		       COALESCE(GROUP_CONCAT(t.tag), '')
		FROM meal_items mi
		LEFT JOIN meal_item_tags t ON t.meal_item_id = mi.id
		WHERE mi.meal_id IN (%s)
		GROUP BY mi.id`, placeholders)
	rows, err := db.QueryContext(r.Context(), q, mealIDs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string][]models.MealItem{}
	for rows.Next() {
		var it models.MealItem
		var qty sql.NullFloat64
		var off sql.NullString
		var conf sql.NullFloat64
		var tags string
		if err := rows.Scan(&it.ID, &it.MealID, &it.Name, &it.DisplayName, &qty, &off, &conf, &tags); err != nil {
			return nil, err
		}
		if qty.Valid {
			v := qty.Float64
			it.QuantityG = &v
		}
		if off.Valid {
			v := off.String
			it.OFFID = &v
		}
		if conf.Valid {
			v := conf.Float64
			it.Confidence = &v
		}
		if tags != "" {
			it.Tags = strings.Split(tags, ",")
		} else {
			it.Tags = []string{}
		}
		out[it.MealID] = append(out[it.MealID], it)
	}
	return out, nil
}

func parseRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now()
	from := now.AddDate(0, 0, -30)
	to := now.Add(24 * time.Hour)
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}
	return from, to
}

// MarshalIndent compat trick: avoid pulling json for prod paths.
var _ = json.Marshal
