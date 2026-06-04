# Vision pipeline

Goal: extract structured `{name, ingredients, allergens, estimated_quantity_g, confidence}` from a photo, at the lowest cost compatible with usable quality.

## Flow

```
photo bytes ──► sha-256 ──► vision_cache hit? ──► yes ──► return cached Result
                                  │
                                  no
                                  ▼
                          Gemini 2.5 Flash
                          (responseMimeType=application/json)
                                  │
                          scene_confidence ≥ 0.6  ──► yes ──► OFF enrichment ──► persist
                                  │
                                  no  (or 429 / 5xx / parse error)
                                  ▼
                          Claude Sonnet 4.6
                          (vision content block, JSON-instruct prompt)
                                  │
                                  ▼
                          OFF enrichment ──► persist
```

## Prompt contract

Both providers receive the exact same English instructions in `internal/vision/types.go`:

- Output must be minified JSON only — no markdown, no commentary.
- Allergens come from a closed enum that matches `internal/models.AllTags`.
- `estimated_quantity_g` is a single number for one serving.
- `scene_confidence` is the model's confidence that the image depicts food.

Keeping the prompt identical means the same parser handles both providers.

## Configuration

| Env var | Default | Purpose |
|---------|---------|---------|
| `GEMINI_API_KEY` | — | Primary provider key (free tier sufficient for low volume) |
| `ANTHROPIC_API_KEY` | — | Fallback provider key |
| `BITESENSE_VISION_FALLBACK_THRESHOLD` | `0.6` | Gemini results below this `scene_confidence` trigger Claude fallback |

## Caching

`vision_cache(image_hash, response_json, provider)` is keyed by SHA-256 of the
raw upload bytes. Identical photos (same SHA-256) reuse the previous result
and skip both LLM calls.

## Enrichment

After the LLM returns, each detected food is looked up in Open Food Facts
via the search endpoint:

- Canonical product name replaces the LLM's guess when a strong match exists.
- Allergens from OFF are merged with the LLM's list, deduplicated against the
  `models.AllTags` allowlist.

## Failure modes

| Failure | Behaviour |
|---------|-----------|
| Gemini 429 | Fall back to Claude |
| Gemini 5xx | Fall back to Claude |
| Gemini parse error | Fall back to Claude |
| Gemini low confidence | Fall back to Claude |
| Both providers fail | Surface 502 to client with combined error |
| Fallback succeeds, primary failed | Use fallback result, log primary failure |
| Fallback fails, primary succeeded with low confidence | Keep primary result, mark provider `gemini:low-conf` |

## Privacy

- Photos are stored on the backend PVC under `/data/uploads/<user_id>/<sha256>.<ext>`.
- EXIF is **not** stripped on the server yet — see [PRIVACY.md](PRIVACY.md) for the planned client-side strip before upload.
- Account deletion (`DELETE /auth/account`) cascades meals and removes their photos on next pruning run.
