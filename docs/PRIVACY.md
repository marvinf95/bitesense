# Privacy & data handling

A food diary is **special-category personal data** under the GDPR (information about health). This document describes how BiteSense handles it.

## What is stored

| Data | Where | Purpose |
|------|-------|---------|
| Email + argon2id password hash | SQLite (`users`) | Authentication |
| Meals, items, tags | SQLite | Core feature |
| Symptoms | SQLite | Core feature |
| Uploaded photos | PVC `/data/uploads/<user_id>/<sha256>.<ext>` | Display + cache lookup |
| Vision provider responses | SQLite (`vision_cache`) | Skip duplicate LLM calls |
| Refresh-token hashes | SQLite (`refresh_tokens`) | Revocable sessions |

We do **not** store:

- Plain passwords.
- Plain refresh tokens.
- Analytics SDK data (no Firebase, no Sentry SaaS).
- Photos outside the user's directory.

## Third parties

| Service | When | Data sent | Why |
|---------|------|-----------|-----|
| Google Generative Language API (Gemini) | Photo upload | Image bytes only | Primary food recognition |
| Anthropic Messages API (Claude) | Photo upload fallback | Image bytes only | Secondary food recognition |
| Open Food Facts | Barcode + canonicalisation | EAN or food name | Allergen + ingredient enrichment |

No personally identifying information (email, name, user ID) is forwarded to any third party. Photos may incidentally contain identifying context (faces, surroundings) — see the EXIF note below.

## EXIF

Uploaded images keep their EXIF metadata on the server today. The roadmap includes a client-side EXIF strip before upload (`exif_remove` in Flutter) and a one-off backend migration to scrub existing entries.

## User rights

- **Access**: `GET /api/v1/auth/me` and the data is yours via PDF export.
- **Erasure**: `DELETE /api/v1/auth/account` cascades meals, items, symptoms, refresh tokens. Uploaded photos are removed on the next pruning run (run nightly).
- **Portability**: PDF export today; CSV/JSON in the roadmap.

## Transport

- HTTPS only (Tailscale-issued certs on the Pi-Homelab deployment).
- JWT access tokens have a 15 min TTL; refresh tokens are rotated on every use and revoked atomically.

## Self-hosting recommendation

This codebase is licensed MIT — feel free to host your own instance if you don't want any third party involvement except for the vision models you explicitly enable.
