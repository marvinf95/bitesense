# HTTP API

Base URL: `/api/v1`. All non-`/auth/*` routes require `Authorization: Bearer <access_token>`.

## Auth

| Method | Path | Body | Returns |
|--------|------|------|---------|
| POST | `/auth/register` | `{email, password, locale}` | `{access_token, refresh_token, user_id}` |
| POST | `/auth/login` | `{email, password}` | same |
| POST | `/auth/refresh` | `{refresh_token}` | rotates both tokens |
| GET  | `/auth/me` | — | current user |
| DELETE | `/auth/account` | — | cascades all user data |

## Meals

| Method | Path | Notes |
|--------|------|-------|
| GET    | `/meals?from=&to=` | RFC3339 range, returns items + tags |
| POST   | `/meals` | text/manual entry — `{eaten_at, source: "text", title?, notes?, items: [...]}` |
| POST   | `/meals/from-image` | `multipart/form-data` with `photo` field |
| POST   | `/meals/from-barcode/{ean}` | EAN-13/UPC lookup via Open Food Facts |
| GET    | `/meals/{id}` | detail |
| PATCH  | `/meals/{id}` | partial update of title/notes/eaten_at |
| DELETE | `/meals/{id}` | hard delete (cascades items + tags) |

## Symptoms

| Method | Path | Notes |
|--------|------|-------|
| POST   | `/symptoms` | `{occurred_at, type, severity (1-10), duration_min?, bristol_stool?, notes?}` |
| GET    | `/symptoms?from=&to=&type=` | filter by type |
| PATCH  | `/symptoms/{id}` | |
| DELETE | `/symptoms/{id}` | |

## Favorites

`/favorites` — list/create/delete reusable meal templates (`{label, template}`).

## Analytics

`GET /analytics/correlations?window_hours=12&type=heartburn`
returns up to 25 ranked suspects:

```json
{
  "suspects": [
    {
      "food": "tomato sauce",
      "symptom_type": "heartburn",
      "risk_ratio": 4.2,
      "p_value": 0.003,
      "n": 7,
      "avg_hours_lag": 3.4,
      "avg_severity": 6.2,
      "tier": "STRONG_SUSPECT"
    }
  ]
}
```

## Export

`GET /export/pdf?from=&to=&locale=de` — streams `application/pdf`.

## Health probes

`GET /livez` and `GET /readyz` (latter pings SQLite).
