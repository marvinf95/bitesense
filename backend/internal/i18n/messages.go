// Package i18n provides minimal localised strings for backend-generated content
// (PDF headers, email subjects, etc.). Frontend handles its own l10n via ARB.
package i18n

import "strings"

type Strings struct {
	PDFTitle           string
	PDFSubtitle        string
	PDFGeneratedAt     string
	PDFSectionMeals    string
	PDFSectionSymptoms string
	PDFSectionSuspects string
	PDFColTime         string
	PDFColTitle        string
	PDFColItems        string
	PDFColTags         string
	PDFColSymptom      string
	PDFColSeverity     string
	PDFColFood         string
	PDFColRR           string
	PDFColN            string
	PDFColTier         string
	PDFEmptyMeals      string
	PDFEmptySymptoms   string
	PDFDisclaimer      string
}

var en = Strings{
	PDFTitle:           "BiteSense — Food Diary Report",
	PDFSubtitle:        "Period: %s — %s",
	PDFGeneratedAt:     "Generated %s",
	PDFSectionMeals:    "Meals",
	PDFSectionSymptoms: "Symptoms",
	PDFSectionSuspects: "Likely trigger foods",
	PDFColTime:         "Time",
	PDFColTitle:        "Title",
	PDFColItems:        "Items",
	PDFColTags:         "Tags",
	PDFColSymptom:      "Symptom",
	PDFColSeverity:     "Severity",
	PDFColFood:         "Food",
	PDFColRR:           "Risk ratio",
	PDFColN:            "n",
	PDFColTier:         "Tier",
	PDFEmptyMeals:      "No meals in the selected range.",
	PDFEmptySymptoms:   "No symptoms in the selected range.",
	PDFDisclaimer:      "Statistical signal, not a medical diagnosis. Discuss with a healthcare professional before changing your diet.",
}

var de = Strings{
	PDFTitle:           "BiteSense — Essenstagebuch-Bericht",
	PDFSubtitle:        "Zeitraum: %s — %s",
	PDFGeneratedAt:     "Erstellt am %s",
	PDFSectionMeals:    "Mahlzeiten",
	PDFSectionSymptoms: "Symptome",
	PDFSectionSuspects: "Mögliche Auslöser",
	PDFColTime:         "Zeit",
	PDFColTitle:        "Titel",
	PDFColItems:        "Bestandteile",
	PDFColTags:         "Tags",
	PDFColSymptom:      "Symptom",
	PDFColSeverity:     "Intensität",
	PDFColFood:         "Lebensmittel",
	PDFColRR:           "Risiko-Verhältnis",
	PDFColN:            "n",
	PDFColTier:         "Stufe",
	PDFEmptyMeals:      "Keine Mahlzeiten im gewählten Zeitraum.",
	PDFEmptySymptoms:   "Keine Symptome im gewählten Zeitraum.",
	PDFDisclaimer:      "Statistisches Signal, keine medizinische Diagnose. Sprich mit medizinischem Fachpersonal, bevor du deine Ernährung umstellst.",
}

func For(locale string) Strings {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "de", "de-de", "de_de":
		return de
	default:
		return en
	}
}
