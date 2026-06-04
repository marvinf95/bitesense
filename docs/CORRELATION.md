# Correlation engine

For each user, the engine looks at the last 180 days of meals + symptoms and builds a 2×2 contingency table per `(food, symptom_type)` pair:

|                | Symptom within window | No symptom within window |
|----------------|-----------------------|--------------------------|
| Food eaten     | a                     | b                        |
| Food not eaten | c                     | d                        |

- **Window** defaults to 12 h after the meal; tunable via `?window_hours=N` (1..72).
- **Risk Ratio (RR)**: `(a / (a+b)) / (c / (c+d))`. Infinity is reported when the unexposed rate is zero.
- **p-value**: two-sided **Fisher's exact test** via the hypergeometric distribution. No chi-square approximation — small samples are common.
- **Tiering**:
  - `STRONG_SUSPECT` if RR ≥ 3 **and** p < 0.05 **and** a ≥ 5
  - `SUSPECT` if RR ≥ 2 **and** a ≥ 3
  - `WEAK_SIGNAL` if RR ≥ 1.5
  - Anything weaker is dropped.

## Why Fisher's exact

The data is sparse, often with single-digit `a`. Chi-square is unreliable when expected cell counts are below 5; Fisher's exact stays valid regardless of cell counts and avoids continuity-correction guesswork.

## Limitations & honest caveats

- This is **observational** data with confounders the user hasn't tagged (stress, sleep, hormones). The disclaimer in [PRIVACY.md](PRIVACY.md) and in-app is not decoration.
- A food and a symptom can be correlated without being causal — e.g. you eat the suspect food when you're already feeling unwell.
- Multiple testing: dozens of `(food × symptom)` pairs are evaluated; we mitigate with the tiered thresholds (effect size + n), but a formal Holm/Benjamini-Hochberg correction is on the roadmap.
- The "unexposed + symptomatic" cell is currently approximated by union across foods of the same symptom type. For low-data periods this slightly over-counts — improvement tracked in [ROADMAP.md](ROADMAP.md).

## Frontend rendering

Surfaced in the Analytics tab with one card per suspect (food ↔ symptom), badge color matching the tier, and the textual delay (`Δ3.4 h`). The medical disclaimer is rendered above the list.
