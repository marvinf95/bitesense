// Package auth handles password hashing, JWT issuance, and refresh-token rotation.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	db         *sql.DB
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(db *sql.DB, secret []byte, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{db: db, secret: secret, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

var argonParams = &argon2id.Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

func HashPassword(plain string) (string, error) {
	if len(plain) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}
	return argon2id.CreateHash(plain, argonParams)
}

func VerifyPassword(plain, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(plain, hash)
}

type Claims struct {
	UserID string `json:"sub"`
	Locale string `json:"locale"`
	jwt.RegisteredClaims
}

func (s *Service) IssueAccess(userID, locale string) (string, error) {
	claims := Claims{
		UserID: userID,
		Locale: locale,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.secret)
}

func (s *Service) ParseAccess(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// IssueRefresh creates a refresh token, stores its hash, and returns the plain token.
func (s *Service) IssueRefresh(ctx context.Context, userID string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	plain := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(plain))
	hash := hex.EncodeToString(sum[:])

	id := uuid.New().String()
	expiresAt := time.Now().Add(s.refreshTTL)

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at) VALUES (?, ?, ?, ?)`,
		id, userID, hash, expiresAt,
	)
	if err != nil {
		return "", err
	}
	return plain, nil
}

// RotateRefresh validates the old refresh token, revokes it, and issues a new pair.
func (s *Service) RotateRefresh(ctx context.Context, presented string) (userID, newAccess, newRefresh string, err error) {
	sum := sha256.Sum256([]byte(presented))
	hash := hex.EncodeToString(sum[:])

	var (
		id        string
		uid       string
		locale    string
		expiresAt time.Time
		revoked   sql.NullTime
	)
	err = s.db.QueryRowContext(ctx, `
		SELECT rt.id, rt.user_id, u.locale, rt.expires_at, rt.revoked_at
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token_hash = ?`, hash).Scan(&id, &uid, &locale, &expiresAt, &revoked)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", "", errors.New("unknown refresh token")
	}
	if err != nil {
		return "", "", "", err
	}
	if revoked.Valid {
		return "", "", "", errors.New("refresh token revoked")
	}
	if time.Now().After(expiresAt) {
		return "", "", "", errors.New("refresh token expired")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", "", "", err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err = tx.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE id = ?`, id); err != nil {
		return "", "", "", err
	}
	if err = tx.Commit(); err != nil {
		return "", "", "", err
	}

	access, err := s.IssueAccess(uid, locale)
	if err != nil {
		return "", "", "", err
	}
	refresh, err := s.IssueRefresh(ctx, uid)
	if err != nil {
		return "", "", "", err
	}
	return uid, access, refresh, nil
}
