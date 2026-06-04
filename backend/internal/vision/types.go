package vision

// Result is the canonical JSON schema returned by both Gemini and Claude.
type Result struct {
	Foods           []Food  `json:"foods"`
	SceneConfidence float64 `json:"scene_confidence"`
	LanguageHint    string  `json:"language_hint"`
	Provider        string  `json:"provider"` // "gemini" | "claude"
}

type Food struct {
	Name               string   `json:"name"`
	Ingredients        []string `json:"ingredients"`
	EstimatedQuantityG float64  `json:"estimated_quantity_g"`
	Allergens          []string `json:"allergens"`
	Confidence         float64  `json:"confidence"`
}

// The shared JSON schema instructed in both prompts. Kept as one source of truth.
const jsonSchemaInstruction = `You receive a photo of food.
Return ONLY valid minified JSON matching this schema:
{
  "foods": [
    {
      "name": "lowercase canonical food name in English",
      "ingredients": ["array of likely ingredients"],
      "estimated_quantity_g": 0,
      "allergens": ["gluten"|"lactose"|"histamine"|"fodmap_high"|"nuts"|"egg"|"soy"|"fructose"|"fish"|"shellfish"|"sulphites"|"sesame"|"mustard"|"celery"],
      "confidence": 0.0
    }
  ],
  "scene_confidence": 0.0,
  "language_hint": "de"|"en"
}
Rules:
- Only output allergens from the enum above; omit unknowns.
- estimated_quantity_g is a single number (best guess for one serving).
- scene_confidence is your overall confidence that the photo depicts food (0..1).
- No prose, no markdown fences, no trailing commas.`
