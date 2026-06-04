package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/marvinf95/bitesense/backend/internal/auth"
)

type AuthHandler struct {
	DB      *sql.DB
	Service *auth.Service
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Locale   string `json:"locale"`
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if req.Locale != "en" && req.Locale != "de" {
		req.Locale = "en"
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id := uuid.New().String()
	now := time.Now()
	_, err = h.DB.ExecContext(r.Context(),
		`INSERT INTO users (id, email, pwd_hash, locale, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, req.Email, hash, req.Locale, now, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "create user")
		return
	}

	access, err := h.Service.IssueAccess(id, req.Locale)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "issue access")
		return
	}
	refresh, err := h.Service.IssueRefresh(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "issue refresh")
		return
	}
	writeJSON(w, http.StatusCreated, authResponse{AccessToken: access, RefreshToken: refresh, UserID: id})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	var (
		id     string
		hash   string
		locale string
	)
	err := h.DB.QueryRowContext(r.Context(),
		`SELECT id, pwd_hash, locale FROM users WHERE email = ?`, req.Email,
	).Scan(&id, &hash, &locale)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup")
		return
	}
	ok, err := auth.VerifyPassword(req.Password, hash)
	if err != nil || !ok {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	access, err := h.Service.IssueAccess(id, locale)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "issue access")
		return
	}
	refresh, err := h.Service.IssueRefresh(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "issue refresh")
		return
	}
	writeJSON(w, http.StatusOK, authResponse{AccessToken: access, RefreshToken: refresh, UserID: id})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	uid, access, refresh, err := h.Service.RotateRefresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh")
		return
	}
	writeJSON(w, http.StatusOK, authResponse{AccessToken: access, RefreshToken: refresh, UserID: uid})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	var email, locale string
	var createdAt time.Time
	err := h.DB.QueryRowContext(r.Context(),
		`SELECT email, locale, created_at FROM users WHERE id = ?`, uid,
	).Scan(&email, &locale, &createdAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":         uid,
		"email":      email,
		"locale":     locale,
		"created_at": createdAt,
	})
}

func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserIDFrom(r.Context())
	_, err := h.DB.ExecContext(r.Context(), `DELETE FROM users WHERE id = ?`, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "delete")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
