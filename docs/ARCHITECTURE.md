# Architecture

BiteSense is a two-tier system: a Flutter client (iOS, Android, Web) and a Go API server backed by SQLite. The server runs in a single namespace on Kubernetes (designed for k3s on a Raspberry Pi 5), exposed over Tailscale rather than the public internet.

## Components

| Component | Folder | Responsibility |
|-----------|--------|----------------|
| Flutter client | `frontend/` | UI, local cache, camera, barcode, PDF download |
| Go API | `backend/cmd/server` | HTTP entrypoint, lifecycle, graceful shutdown |
| Auth | `backend/internal/auth` | argon2id passwords, JWT (HS256) access + opaque refresh tokens |
| Data | `backend/internal/db` | SQLite (modernc) + `golang-migrate` |
| Vision | `backend/internal/vision` | Gemini 2.5 Flash primary, Claude Sonnet 4.6 fallback, image-hash cache |
| Food facts | `backend/internal/foodfacts` | Open Food Facts client (barcode + search) |
| Correlation | `backend/internal/correlation` | Per-user contingency tables, risk ratio, Fisher's exact |
| PDF | `backend/internal/pdf` | Localized PDF report (gofpdf) |
| i18n | `backend/internal/i18n` | Backend-side strings for PDF + system content |
| API surface | `backend/internal/api` | chi router, handlers, middleware, request/response shapes |

## Request flow: photo to meal

```
Client takes photo
  → POST /api/v1/meals/from-image (multipart)
    → handler reads bytes (≤8 MiB), computes SHA-256
    → vision.Analyzer.runWithCache
       → cache hit? return cached result
       → Gemini.Analyze (JSON-mode)
       → if low confidence or error → Claude.Analyze (fallback)
       → enrich via Open Food Facts (canonical names + allergens)
    → persist meal + meal_items + meal_item_tags inside a single tx
    → respond 201 { meal_id }
```

## Why these choices

- **Flutter + Riverpod + go_router**: matches the Goalbooze pattern already in `~/projects/goalbooze`; same toolchain on the dev machine.
- **Go + chi + SQLite**: low-RAM footprint, zero external DB to run on a Pi, easy backups (single file).
- **JWT access + refresh rotation**: stateless access tokens, refresh tokens are hashed in the DB so they can be revoked.
- **Gemini primary**: free tier covers our expected volume; Claude fallback keeps quality when confidence dips.
- **k8s manifests**: same security profile (PSA `restricted`, network policies, sealed secrets) as the rest of the Pi homelab. See `k8s/base/`.

## Data model

See [DATA_MODEL.md](DATA_MODEL.md) and migrations under `backend/migrations/`.
