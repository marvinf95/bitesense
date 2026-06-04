package vision

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/marvinf95/bitesense/backend/internal/foodfacts"
	"github.com/marvinf95/bitesense/backend/internal/models"
)

type Analyzer struct {
	DB                *sql.DB
	Gemini            *GeminiClient
	Claude            *ClaudeClient
	OFF               *foodfacts.Client
	UploadDir         string
	FallbackThreshold float64
}

const maxUploadBytes = 8 << 20 // 8 MiB

// AnalyzeFromUpload processes a multipart upload, persists the image,
// runs the vision pipeline, and creates a meal with its items.
func (a *Analyzer) AnalyzeFromUpload(r *http.Request, userID string) (string, error) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		return "", fmt.Errorf("parse multipart: %w", err)
	}
	file, hdr, err := r.FormFile("photo")
	if err != nil {
		return "", fmt.Errorf("photo field: %w", err)
	}
	defer file.Close()

	raw, err := io.ReadAll(io.LimitReader(file, maxUploadBytes))
	if err != nil {
		return "", err
	}
	mimeType := detectMime(hdr.Filename, raw)

	hash := sha256.Sum256(raw)
	hashHex := hex.EncodeToString(hash[:])

	result, err := a.runWithCache(r.Context(), hashHex, raw, mimeType)
	if err != nil {
		return "", err
	}

	// Persist the image on disk (user-scoped subdir).
	userDir := filepath.Join(a.UploadDir, userID)
	if err := os.MkdirAll(userDir, 0o750); err != nil {
		return "", err
	}
	ext := extFromMime(mimeType)
	relPath := filepath.Join(userID, hashHex+ext)
	if err := os.WriteFile(filepath.Join(a.UploadDir, relPath), raw, 0o640); err != nil {
		return "", err
	}

	// Enrich + persist.
	if err := a.enrich(r.Context(), result); err != nil {
		log.Warn().Err(err).Msg("vision enrichment partial failure")
	}
	mealID, err := a.persistMeal(r.Context(), userID, relPath, result)
	if err != nil {
		return "", err
	}
	return mealID, nil
}

// AnalyzeFromBarcode skips vision and resolves an EAN directly via Open Food Facts.
func (a *Analyzer) AnalyzeFromBarcode(r *http.Request, userID, ean string) (string, error) {
	if a.OFF == nil {
		return "", errors.New("food facts client unavailable")
	}
	prod, err := a.OFF.Lookup(r.Context(), ean)
	if err != nil {
		return "", err
	}
	food := Food{
		Name:               strings.ToLower(prod.Name),
		Ingredients:        prod.Ingredients,
		EstimatedQuantityG: prod.ServingSizeG,
		Allergens:          prod.Allergens,
		Confidence:         1.0,
	}
	result := &Result{
		Foods:           []Food{food},
		SceneConfidence: 1.0,
		LanguageHint:    "",
		Provider:        "off",
	}
	return a.persistMeal(r.Context(), userID, "", result)
}

func (a *Analyzer) runWithCache(ctx context.Context, hash string, raw []byte, mimeType string) (*Result, error) {
	var cached string
	err := a.DB.QueryRowContext(ctx, `SELECT response_json FROM vision_cache WHERE image_hash = ?`, hash).Scan(&cached)
	if err == nil && cached != "" {
		var r Result
		if jsonErr := json.Unmarshal([]byte(cached), &r); jsonErr == nil {
			r.Provider = r.Provider + ":cache"
			return &r, nil
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()

	result, primaryErr := a.Gemini.Analyze(ctx, raw, mimeType)
	if primaryErr != nil || result == nil || result.SceneConfidence < a.FallbackThreshold {
		log.Info().
			Err(primaryErr).
			Float64("scene_confidence", confOr(result)).
			Msg("vision: falling back to claude")
		fallback, err := a.Claude.Analyze(ctx, raw, mimeType)
		if err != nil {
			if primaryErr != nil {
				return nil, fmt.Errorf("both providers failed: gemini=%v claude=%v", primaryErr, err)
			}
			// keep low-confidence primary if fallback also fails.
			result.Provider = "gemini:low-conf"
			return result, nil
		}
		result = fallback
	}

	buf, _ := json.Marshal(result)
	_, _ = a.DB.ExecContext(ctx,
		`INSERT OR REPLACE INTO vision_cache (image_hash, response_json, provider) VALUES (?, ?, ?)`,
		hash, string(buf), result.Provider,
	)
	return result, nil
}

func confOr(r *Result) float64 {
	if r == nil {
		return 0
	}
	return r.SceneConfidence
}

// enrich pulls canonical names + allergens from Open Food Facts where possible.
func (a *Analyzer) enrich(ctx context.Context, r *Result) error {
	if a.OFF == nil {
		return nil
	}
	for i := range r.Foods {
		f := &r.Foods[i]
		hits, err := a.OFF.Search(ctx, f.Name, 1)
		if err != nil || len(hits) == 0 {
			continue
		}
		hit := hits[0]
		// Replace with canonical name; merge allergens.
		f.Name = strings.ToLower(hit.Name)
		seen := map[string]bool{}
		merged := []string{}
		for _, t := range append(f.Allergens, hit.Allergens...) {
			if models.AllTags[t] && !seen[t] {
				seen[t] = true
				merged = append(merged, t)
			}
		}
		f.Allergens = merged
	}
	return nil
}

func (a *Analyzer) persistMeal(ctx context.Context, userID, photoPath string, r *Result) (string, error) {
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	mealID := uuid.New().String()
	now := time.Now()
	visionJSON, _ := json.Marshal(r)

	source := "image"
	if r.Provider == "off" {
		source = "barcode"
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO meals (id, user_id, eaten_at, source, photo_path, vision_raw, created_at, updated_at)
		VALUES (?, ?, ?, ?, NULLIF(?, ''), ?, ?, ?)`,
		mealID, userID, now, source, photoPath, string(visionJSON), now, now,
	)
	if err != nil {
		return "", err
	}
	for _, f := range r.Foods {
		itemID := uuid.New().String()
		var qty *float64
		if f.EstimatedQuantityG > 0 {
			v := f.EstimatedQuantityG
			qty = &v
		}
		var conf *float64
		if f.Confidence > 0 {
			c := f.Confidence
			conf = &c
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO meal_items (id, meal_id, name, display_name, quantity_g, confidence)
			VALUES (?, ?, ?, ?, ?, ?)`,
			itemID, mealID, strings.ToLower(f.Name), f.Name, qty, conf,
		)
		if err != nil {
			return "", err
		}
		for _, t := range f.Allergens {
			if !models.AllTags[t] {
				continue
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO meal_item_tags (meal_item_id, tag) VALUES (?, ?)`, itemID, t)
			if err != nil {
				return "", err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return mealID, nil
}

func detectMime(filename string, raw []byte) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".heic":
		return "image/heic"
	}
	if len(raw) >= 12 {
		// quick magic-byte sniff
		if raw[0] == 0xFF && raw[1] == 0xD8 {
			return "image/jpeg"
		}
		if raw[0] == 0x89 && raw[1] == 0x50 {
			return "image/png"
		}
		if string(raw[8:12]) == "WEBP" {
			return "image/webp"
		}
	}
	return http.DetectContentType(raw)
}

func extFromMime(m string) string {
	switch m {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/heic":
		return ".heic"
	}
	return ".bin"
}
