# Roadmap

## Shipped in v0.1 (MVP)

- Multi-user auth (email + argon2id, JWT + refresh rotation)
- Meal capture: text, photo (Gemini → Claude fallback), barcode (Open Food Facts)
- Symptom logging with severity 1–10, Bristol stool scale, duration, notes
- Correlation engine (RR + Fisher's exact, tiered output)
- PDF export (English + German)
- Bilingual UI (de, en) — auto detection + manual override
- k3s deployment manifests (PSA `restricted`, NetworkPolicy, Sealed Secrets)

## Next (v0.2)

- **Elimination-diet helper**: start a cycle, compare symptom frequency before/after.
- **Confounders tracking**: sleep hours, stress (short scale), alcohol, caffeine, medications, menstrual cycle.
- **Voice input** for meal/symptom entry via native STT.
- **Doctor share link** (signed, read-only, 7-day TTL).
- **CSV + JSON exports** alongside PDF.
- **Image EXIF strip** on the client before upload + backend backfill.
- **Weekly review push** ("This week you had heartburn 5×; tomatoes appear in 4 of those meals").

## Later

- **Offline-first sync layer** for the Flutter client (Drift mutation queue, last-write-wins).
- **Multiple testing correction** for correlations (Benjamini-Hochberg).
- **WatchOS/WearOS companion** for one-tap symptom logging.
- **OpenFoodFacts contribution flow** when a scanned product is missing.
- **Family / partner mode**: switch profile within one account.
- **Push notifications** for meal-logging reminders.

## Known limitations to address

- Correlation engine's "unexposed + symptomatic" cell is an approximation (see [CORRELATION.md](CORRELATION.md)).
- Photos remain on disk after `DELETE /auth/account` until the next pruning run.
- No rate limit on `/meals/from-image` yet (covered by global `httprate` only).
