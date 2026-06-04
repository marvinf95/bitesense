# BiteSense

> A privacy-first food diary app for spotting intolerances and food-symptom correlations.

BiteSense lets you log meals (by text, photo, or barcode) and symptoms (with timing and severity), then surfaces statistical correlations to help you and your doctor identify likely triggers.

[![CI](https://github.com/marvinf95/bitesense/actions/workflows/ci.yml/badge.svg)](https://github.com/marvinf95/bitesense/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Features

- **Meal capture** — text entry, photo with AI vision (Gemini 2.5 Flash, Claude Sonnet 4.6 fallback), or barcode scan (Open Food Facts).
- **Symptom logging** — type, severity (1–10), Bristol stool scale, duration, notes.
- **Correlation analysis** — risk ratio + Fisher's exact test over a 12 h window flags likely trigger foods.
- **Allergen tagging** — gluten, lactose, histamine, FODMAP, nuts, egg, soy, fructose, and more.
- **PDF export** — date-ranged report for your doctor or nutritionist.
- **Bilingual UI** — German and English, auto-detected from device locale.
- **Offline-first** — works without connectivity, syncs when online.
- **Self-hostable** — runs on a single Raspberry Pi or any small k3s cluster.

## Tech Stack

| Layer | Choice |
|-------|--------|
| Mobile / Web client | Flutter (iOS, Android, Web) |
| State management | Riverpod |
| Local DB | Drift (SQLite) |
| Backend | Go 1.22, chi router |
| Server DB | SQLite |
| Auth | Email + password (argon2id), JWT (HS256) |
| Vision (primary) | Google Gemini 2.5 Flash |
| Vision (fallback) | Anthropic Claude Sonnet 4.6 |
| Food database | Open Food Facts |
| PDF | gofpdf v2 |
| Deployment | k3s on Raspberry Pi 5 (or any Kubernetes) |

## Quickstart (local development)

### Prerequisites

- Go 1.22+
- Flutter 3.24+
- An Anthropic API key (`ANTHROPIC_API_KEY`)
- A Google AI Studio key for Gemini (`GEMINI_API_KEY`)

### Backend

```bash
cd backend
cp .env.example .env       # then fill in keys
go run ./cmd/server
# listens on http://localhost:8080
```

### Frontend

```bash
cd frontend
flutter pub get
flutter run -d chrome      # or -d <device-id> for mobile
```

## Project Layout

```
backend/    Go API server (chi, SQLite, vision, correlation, PDF)
frontend/   Flutter client (iOS, Android, Web)
k8s/        Kustomize manifests for self-hosted deployment
docs/       Architecture, API spec, data model, roadmap
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for a system overview.

## Medical Disclaimer

BiteSense is a personal tracking tool. It surfaces statistical signals between foods and symptoms but **does not diagnose** food intolerances, allergies, or any other medical condition. Always discuss findings with a qualified healthcare professional before changing your diet.

## License

[MIT](LICENSE) © 2026 Marvin Forster
