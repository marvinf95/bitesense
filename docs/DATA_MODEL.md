# Data Model

All entities are user-scoped (`user_id` FK). The authoritative SQL schema lives in [backend/migrations/0001_init.up.sql](../backend/migrations/0001_init.up.sql).

```
┌──────────┐     1   N   ┌────────┐  1   N   ┌────────────┐  1   N   ┌────────────────┐
│  users   │─────────────│ meals  │──────────│ meal_items │──────────│ meal_item_tags │
└──────────┘             └────────┘          └────────────┘          └────────────────┘
     │ 1                                            │
     │ N                                            │
┌───────────────┐                          ┌────────────────┐
│   symptoms    │                          │ meal_favorites │
└───────────────┘                          └────────────────┘
                                                    
┌──────────────┐    ┌─────────────────┐    ┌───────────────┐
│ vision_cache │    │ refresh_tokens  │    │   sync_log    │
└──────────────┘    └─────────────────┘    └───────────────┘
```

## Key invariants

- `symptoms.severity` is 1–10. `bristol_stool` is 1–7 or NULL.
- `meal_item_tags.tag` is constrained by the allowlist in `internal/models/AllTags`. Backend rejects unknown values.
- `meals.source` is one of `text | image | barcode | favorite`.
- `vision_cache.image_hash` is SHA-256 of the raw upload bytes; identical photos skip the LLM call.
- `refresh_tokens.token_hash` is SHA-256 of the plain token; only the hash is stored.

## Sample queries

Top suspect foods for one user (12 h window):

```sql
WITH symptomatic AS (
  SELECT m.id AS meal_id, mi.name AS food, s.type AS symptom, s.severity
  FROM symptoms s
  JOIN meals m       ON m.user_id = s.user_id
                     AND m.eaten_at <= s.occurred_at
                     AND s.occurred_at <= datetime(m.eaten_at, '+12 hours')
  JOIN meal_items mi ON mi.meal_id = m.id
  WHERE s.user_id = ?
)
SELECT food, symptom, COUNT(*) AS hits, AVG(severity) AS avg_severity
FROM symptomatic
GROUP BY food, symptom
HAVING hits >= 3
ORDER BY hits DESC;
```

The Go correlation engine extends this with risk ratio and Fisher's exact test
to suppress coincidental signals.
