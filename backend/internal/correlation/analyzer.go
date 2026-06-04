// Package correlation implements the food↔symptom statistical engine.
package correlation

import (
	"database/sql"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/marvinf95/bitesense/backend/internal/api"
)

const (
	defaultWindowHours = 12
	tierStrong         = "STRONG_SUSPECT"
	tierSuspect        = "SUSPECT"
	tierWeak           = "WEAK_SIGNAL"
)

type Analyzer struct {
	DB *sql.DB
}

// Compile-time check the Analyzer satisfies the api interface.
var _ api.CorrelationAnalyzer = (*Analyzer)(nil)

// TopSuspects returns ranked food/symptom associations for the authenticated user.
// Window defaults to 12 hours and can be overridden with ?window_hours.
func (a *Analyzer) TopSuspects(r *http.Request, userID string) ([]api.Suspect, error) {
	window := defaultWindowHours
	if v := r.URL.Query().Get("window_hours"); v != "" {
		if n, err := time.ParseDuration(v + "h"); err == nil && n > 0 && n <= 72*time.Hour {
			window = int(n.Hours())
		}
	}
	symptomTypeFilter := r.URL.Query().Get("type")

	// Pull all meals + items for this user (last 180 days).
	cutoff := time.Now().AddDate(0, 0, -180)
	mealRows, err := a.DB.QueryContext(r.Context(), `
		SELECT m.id, m.eaten_at, mi.name
		FROM meals m
		JOIN meal_items mi ON mi.meal_id = m.id
		WHERE m.user_id = ? AND m.eaten_at >= ?`,
		userID, cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer mealRows.Close()

	type mealItem struct {
		mealID  string
		eatenAt time.Time
		food    string
	}
	var items []mealItem
	mealsCount := map[string]struct{}{}
	for mealRows.Next() {
		var it mealItem
		if err := mealRows.Scan(&it.mealID, &it.eatenAt, &it.food); err != nil {
			return nil, err
		}
		items = append(items, it)
		mealsCount[it.mealID] = struct{}{}
	}

	// All symptoms in same window.
	q := `SELECT id, occurred_at, type, severity FROM symptoms WHERE user_id = ? AND occurred_at >= ?`
	args := []any{userID, cutoff}
	if symptomTypeFilter != "" {
		q += " AND type = ?"
		args = append(args, symptomTypeFilter)
	}
	symRows, err := a.DB.QueryContext(r.Context(), q, args...)
	if err != nil {
		return nil, err
	}
	defer symRows.Close()

	type sym struct {
		id       string
		at       time.Time
		typeName string
		severity int
	}
	var symptoms []sym
	for symRows.Next() {
		var s sym
		if err := symRows.Scan(&s.id, &s.at, &s.typeName, &s.severity); err != nil {
			return nil, err
		}
		symptoms = append(symptoms, s)
	}

	// Build per (food, symptomType) stats.
	type key struct {
		food, symptom string
	}
	type bucket struct {
		exposedSymptomatic  int
		mealsExposed        map[string]struct{}
		mealsWithMatchedSym map[string]struct{}
		hoursLag            []float64
		severities          []int
	}
	buckets := map[key]*bucket{}

	// Index meals by food for quick lookups (set of meal_ids per food).
	foodMeals := map[string]map[string]time.Time{}
	for _, it := range items {
		if _, ok := foodMeals[it.food]; !ok {
			foodMeals[it.food] = map[string]time.Time{}
		}
		foodMeals[it.food][it.mealID] = it.eatenAt
	}

	allSymTypes := map[string]struct{}{}
	for _, s := range symptoms {
		allSymTypes[s.typeName] = struct{}{}
	}

	winDur := time.Duration(window) * time.Hour

	for food, meals := range foodMeals {
		for typeName := range allSymTypes {
			k := key{food: food, symptom: typeName}
			if buckets[k] == nil {
				buckets[k] = &bucket{
					mealsExposed:        map[string]struct{}{},
					mealsWithMatchedSym: map[string]struct{}{},
				}
			}
			for mealID, eatenAt := range meals {
				buckets[k].mealsExposed[mealID] = struct{}{}
				for _, s := range symptoms {
					if s.typeName != typeName {
						continue
					}
					if s.at.Before(eatenAt) {
						continue
					}
					delta := s.at.Sub(eatenAt)
					if delta > winDur {
						continue
					}
					buckets[k].exposedSymptomatic++
					buckets[k].mealsWithMatchedSym[mealID] = struct{}{}
					buckets[k].hoursLag = append(buckets[k].hoursLag, delta.Hours())
					buckets[k].severities = append(buckets[k].severities, s.severity)
				}
			}
		}
	}

	totalMeals := len(mealsCount)
	out := make([]api.Suspect, 0, len(buckets))

	for k, b := range buckets {
		exposedMealCount := len(b.mealsExposed)
		if exposedMealCount == 0 {
			continue
		}
		unexposedMealCount := totalMeals - exposedMealCount
		// Count symptomatic meals overall (any meal followed by a symptom of this type within window).
		// Approximated by union of mealsWithMatchedSym across all foods for the same symptom type.
		symptomaticAny := 0
		for kk, bb := range buckets {
			if kk.symptom != k.symptom {
				continue
			}
			for m := range bb.mealsWithMatchedSym {
				_ = m
			}
			symptomaticAny += len(bb.mealsWithMatchedSym)
		}
		// Risk Ratio and Fisher's exact.
		a11 := len(b.mealsWithMatchedSym)                                 // exposed + symptomatic
		a12 := exposedMealCount - a11                                     // exposed + asymptomatic
		a21 := max0(symptomaticAny - a11)                                 // unexposed + symptomatic (rough)
		a22 := max0(unexposedMealCount - a21)                             // unexposed + asymptomatic
		n := a11 + a12 + a21 + a22
		if n < 5 || a11 < 1 {
			continue
		}
		var p1, p2 float64
		if exposedMealCount > 0 {
			p1 = float64(a11) / float64(exposedMealCount)
		}
		if unexposedMealCount > 0 {
			p2 = float64(a21) / float64(unexposedMealCount)
		}
		rr := 0.0
		if p2 > 0 {
			rr = p1 / p2
		} else if p1 > 0 {
			rr = math.Inf(1)
		}
		pVal := fishersExact(a11, a12, a21, a22)

		tier := tierWeak
		switch {
		case rr >= 3 && pVal < 0.05 && a11 >= 5:
			tier = tierStrong
		case rr >= 2 && a11 >= 3:
			tier = tierSuspect
		case rr < 1.5:
			continue
		}

		out = append(out, api.Suspect{
			Food:        k.food,
			SymptomType: k.symptom,
			RiskRatio:   roundTo(rr, 2),
			PValue:      roundTo(pVal, 4),
			N:           a11,
			AvgHoursLag: roundTo(mean(b.hoursLag), 1),
			Severity:    roundTo(meanInt(b.severities), 1),
			Tier:        tier,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Tier != out[j].Tier {
			return tierWeight(out[i].Tier) > tierWeight(out[j].Tier)
		}
		if out[i].RiskRatio != out[j].RiskRatio {
			return out[i].RiskRatio > out[j].RiskRatio
		}
		return out[i].N > out[j].N
	})
	if len(out) > 25 {
		out = out[:25]
	}
	return out, nil
}

func tierWeight(t string) int {
	switch t {
	case tierStrong:
		return 3
	case tierSuspect:
		return 2
	default:
		return 1
	}
}

// fishersExact returns the two-sided p-value via the hypergeometric distribution.
// For 2x2 tables this is exact (no chi-square approximation).
func fishersExact(a, b, c, d int) float64 {
	n := a + b + c + d
	r1, r2 := a+b, c+d
	c1 := a + c
	observed := logHGProb(r1, r2, c1, a)
	total := 0.0
	low := max(0, c1-r2)
	high := min(c1, r1)
	for k := low; k <= high; k++ {
		p := logHGProb(r1, r2, c1, k)
		if p <= observed+1e-9 {
			total += math.Exp(p)
		}
		_ = n
	}
	if total > 1 {
		return 1
	}
	if total < 0 {
		return 0
	}
	return total
}

func logHGProb(r1, r2, c1, k int) float64 {
	return logChoose(r1, k) + logChoose(r2, c1-k) - logChoose(r1+r2, c1)
}

func logChoose(n, k int) float64 {
	if k < 0 || k > n {
		return math.Inf(-1)
	}
	return logFact(n) - logFact(k) - logFact(n-k)
}

func logFact(n int) float64 {
	if n < 0 {
		return math.Inf(-1)
	}
	// stable: lgamma(n+1)
	v, _ := math.Lgamma(float64(n) + 1)
	return v
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func meanInt(xs []int) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0
	for _, x := range xs {
		s += x
	}
	return float64(s) / float64(len(xs))
}

func roundTo(v float64, digits int) float64 {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return v
	}
	m := math.Pow(10, float64(digits))
	return math.Round(v*m) / m
}

func max0(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
