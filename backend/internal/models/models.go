// Package models defines the domain entities serialized over the API.
package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Locale    string    `json:"locale"`
	CreatedAt time.Time `json:"created_at"`
}

type Meal struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	EatenAt    time.Time  `json:"eaten_at"`
	Title      *string    `json:"title,omitempty"`
	Notes      *string    `json:"notes,omitempty"`
	Source     string     `json:"source"`
	PhotoPath  *string    `json:"photo_path,omitempty"`
	Items      []MealItem `json:"items"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type MealItem struct {
	ID          string   `json:"id"`
	MealID      string   `json:"meal_id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	QuantityG   *float64 `json:"quantity_g,omitempty"`
	OFFID       *string  `json:"off_id,omitempty"`
	Confidence  *float64 `json:"confidence,omitempty"`
	Tags        []string `json:"tags"`
}

type Symptom struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	OccurredAt   time.Time `json:"occurred_at"`
	Type         string    `json:"type"`
	Severity     int       `json:"severity"`
	DurationMin  *int      `json:"duration_min,omitempty"`
	BristolStool *int      `json:"bristol_stool,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type MealFavorite struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Label     string    `json:"label"`
	Template  string    `json:"template"`
	CreatedAt time.Time `json:"created_at"`
}

// Allergen/category tag canonical values.
const (
	TagGluten     = "gluten"
	TagLactose    = "lactose"
	TagHistamine  = "histamine"
	TagFODMAPHigh = "fodmap_high"
	TagNuts       = "nuts"
	TagEgg        = "egg"
	TagSoy        = "soy"
	TagFructose   = "fructose"
	TagFish       = "fish"
	TagShellfish  = "shellfish"
	TagSulphites  = "sulphites"
	TagSesame     = "sesame"
	TagMustard    = "mustard"
	TagCelery     = "celery"
)

// AllTags is the canonical allowlist; new entries must be added here for validation.
var AllTags = map[string]bool{
	TagGluten: true, TagLactose: true, TagHistamine: true, TagFODMAPHigh: true,
	TagNuts: true, TagEgg: true, TagSoy: true, TagFructose: true,
	TagFish: true, TagShellfish: true, TagSulphites: true, TagSesame: true,
	TagMustard: true, TagCelery: true,
}

// Symptom type canonical values.
const (
	SymHeartburn    = "heartburn"
	SymBloating     = "bloating"
	SymDiarrhea     = "diarrhea"
	SymConstipation = "constipation"
	SymHeadache     = "headache"
	SymFatigue      = "fatigue"
	SymBrainFog     = "brain_fog"
	SymSkin         = "skin"
	SymJointPain    = "joint_pain"
	SymMood         = "mood"
	SymNausea       = "nausea"
	SymOther        = "other"
)

var AllSymptomTypes = map[string]bool{
	SymHeartburn: true, SymBloating: true, SymDiarrhea: true, SymConstipation: true,
	SymHeadache: true, SymFatigue: true, SymBrainFog: true, SymSkin: true,
	SymJointPain: true, SymMood: true, SymNausea: true, SymOther: true,
}
