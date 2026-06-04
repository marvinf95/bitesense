// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr        string
	DataDir     string
	UploadDir   string
	DBPath      string
	LogLevel    string
	LogFormat   string

	JWTSecret     []byte
	AccessTTL     time.Duration
	RefreshTTL    time.Duration

	GeminiAPIKey    string
	AnthropicAPIKey string
	VisionFallback  float64
	OFFUserAgent    string

	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	c := &Config{
		Addr:               env("BITESENSE_ADDR", ":8080"),
		DataDir:            env("BITESENSE_DATA_DIR", "./data"),
		UploadDir:          env("BITESENSE_UPLOAD_DIR", "./data/uploads"),
		DBPath:             env("BITESENSE_DB_PATH", "./data/bitesense.db"),
		LogLevel:           env("BITESENSE_LOG_LEVEL", "info"),
		LogFormat:          env("BITESENSE_LOG_FORMAT", "json"),
		GeminiAPIKey:       os.Getenv("GEMINI_API_KEY"),
		AnthropicAPIKey:    os.Getenv("ANTHROPIC_API_KEY"),
		OFFUserAgent:       env("BITESENSE_OFF_USER_AGENT", "BiteSense/0.1"),
		CORSAllowedOrigins: splitCSV(env("BITESENSE_CORS_ALLOWED_ORIGINS", "*")),
	}

	secret := os.Getenv("BITESENSE_JWT_SECRET")
	if len(secret) < 32 {
		return nil, fmt.Errorf("BITESENSE_JWT_SECRET must be at least 32 bytes")
	}
	c.JWTSecret = []byte(secret)

	access, err := time.ParseDuration(env("BITESENSE_JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("BITESENSE_JWT_ACCESS_TTL: %w", err)
	}
	c.AccessTTL = access

	refresh, err := time.ParseDuration(env("BITESENSE_JWT_REFRESH_TTL", "720h"))
	if err != nil {
		return nil, fmt.Errorf("BITESENSE_JWT_REFRESH_TTL: %w", err)
	}
	c.RefreshTTL = refresh

	threshold := env("BITESENSE_VISION_FALLBACK_THRESHOLD", "0.6")
	t, err := strconv.ParseFloat(threshold, 64)
	if err != nil || t < 0 || t > 1 {
		return nil, fmt.Errorf("BITESENSE_VISION_FALLBACK_THRESHOLD must be in [0,1]")
	}
	c.VisionFallback = t

	return c, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
