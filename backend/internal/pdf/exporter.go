// Package pdf renders a localized PDF report for a given date range.
package pdf

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"

	"github.com/marvinf95/bitesense/backend/internal/api"
	"github.com/marvinf95/bitesense/backend/internal/correlation"
	"github.com/marvinf95/bitesense/backend/internal/i18n"
)

type Exporter struct {
	DB   *sql.DB
	Corr *correlation.Analyzer
}

var _ api.PDFExporter = (*Exporter)(nil)

func (e *Exporter) Export(w http.ResponseWriter, r *http.Request, userID, locale string) error {
	from, to := parseRange(r)
	strs := i18n.For(locale)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	pdf.SetTitle(strs.PDFTitle, true)
	pdf.SetCreator("BiteSense", true)

	// Cover header
	pdf.SetFont("Helvetica", "B", 18)
	pdf.MultiCell(0, 9, strs.PDFTitle, "", "L", false)
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "", 11)
	pdf.MultiCell(0, 6, fmt.Sprintf(strs.PDFSubtitle, from.Format("2006-01-02"), to.Format("2006-01-02")), "", "L", false)
	pdf.MultiCell(0, 6, fmt.Sprintf(strs.PDFGeneratedAt, time.Now().UTC().Format(time.RFC3339)), "", "L", false)
	pdf.Ln(6)

	if err := e.renderMeals(pdf, r, userID, from, to, strs); err != nil {
		return err
	}
	if err := e.renderSymptoms(pdf, r, userID, from, to, strs); err != nil {
		return err
	}
	if err := e.renderSuspects(pdf, r, userID, strs); err != nil {
		return err
	}

	pdf.AddPage()
	pdf.SetFont("Helvetica", "I", 10)
	pdf.MultiCell(0, 5, strs.PDFDisclaimer, "", "L", false)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="bitesense-report.pdf"`)
	return pdf.Output(w)
}

func (e *Exporter) renderMeals(pdf *gofpdf.Fpdf, r *http.Request, userID string, from, to time.Time, s i18n.Strings) error {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.MultiCell(0, 8, s.PDFSectionMeals, "", "L", false)
	pdf.SetFont("Helvetica", "", 9)
	rows, err := e.DB.QueryContext(r.Context(), `
		SELECT m.id, m.eaten_at, COALESCE(m.title, ''),
		       (SELECT COALESCE(GROUP_CONCAT(mi.display_name, ', '), '')
		          FROM meal_items mi WHERE mi.meal_id = m.id),
		       (SELECT COALESCE(GROUP_CONCAT(DISTINCT mit.tag, ', '), '')
		          FROM meal_items mi
		          LEFT JOIN meal_item_tags mit ON mit.meal_item_id = mi.id
		          WHERE mi.meal_id = m.id)
		FROM meals m
		WHERE m.user_id = ? AND m.eaten_at BETWEEN ? AND ?
		ORDER BY m.eaten_at`,
		userID, from, to,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	hasRows := false
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(35, 6, s.PDFColTime, "1", 0, "L", true, 0, "")
	pdf.CellFormat(40, 6, s.PDFColTitle, "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 6, s.PDFColItems, "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 6, s.PDFColTags, "1", 1, "L", true, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	for rows.Next() {
		var id, title, items, tags string
		var eatenAt time.Time
		if err := rows.Scan(&id, &eatenAt, &title, &items, &tags); err != nil {
			return err
		}
		hasRows = true
		pdf.CellFormat(35, 5, eatenAt.Format("2006-01-02 15:04"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 5, truncate(title, 22), "1", 0, "L", false, 0, "")
		pdf.CellFormat(70, 5, truncate(items, 50), "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 5, truncate(tags, 22), "1", 1, "L", false, 0, "")
	}
	if !hasRows {
		pdf.MultiCell(0, 6, s.PDFEmptyMeals, "", "L", false)
	}
	pdf.Ln(6)
	return nil
}

func (e *Exporter) renderSymptoms(pdf *gofpdf.Fpdf, r *http.Request, userID string, from, to time.Time, s i18n.Strings) error {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.MultiCell(0, 8, s.PDFSectionSymptoms, "", "L", false)
	rows, err := e.DB.QueryContext(r.Context(), `
		SELECT occurred_at, type, severity, COALESCE(bristol_stool, 0), COALESCE(notes, '')
		FROM symptoms WHERE user_id = ? AND occurred_at BETWEEN ? AND ?
		ORDER BY occurred_at`,
		userID, from, to,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(35, 6, s.PDFColTime, "1", 0, "L", true, 0, "")
	pdf.CellFormat(40, 6, s.PDFColSymptom, "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 6, s.PDFColSeverity, "1", 0, "L", true, 0, "")
	pdf.CellFormat(80, 6, "Notes", "1", 1, "L", true, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	hasRows := false
	for rows.Next() {
		var occurredAt time.Time
		var typ string
		var severity int
		var br int
		var notes string
		if err := rows.Scan(&occurredAt, &typ, &severity, &br, &notes); err != nil {
			return err
		}
		hasRows = true
		extra := ""
		if br > 0 {
			extra = fmt.Sprintf(" (Bristol %d)", br)
		}
		pdf.CellFormat(35, 5, occurredAt.Format("2006-01-02 15:04"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 5, typ+extra, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("%d/10", severity), "1", 0, "L", false, 0, "")
		pdf.CellFormat(80, 5, truncate(notes, 55), "1", 1, "L", false, 0, "")
	}
	if !hasRows {
		pdf.MultiCell(0, 6, s.PDFEmptySymptoms, "", "L", false)
	}
	pdf.Ln(6)
	return nil
}

func (e *Exporter) renderSuspects(pdf *gofpdf.Fpdf, r *http.Request, userID string, s i18n.Strings) error {
	if e.Corr == nil {
		return nil
	}
	suspects, err := e.Corr.TopSuspects(r, userID)
	if err != nil {
		return err
	}
	pdf.SetFont("Helvetica", "B", 14)
	pdf.MultiCell(0, 8, s.PDFSectionSuspects, "", "L", false)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(50, 6, s.PDFColFood, "1", 0, "L", true, 0, "")
	pdf.CellFormat(40, 6, s.PDFColSymptom, "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 6, s.PDFColRR, "1", 0, "L", true, 0, "")
	pdf.CellFormat(20, 6, s.PDFColN, "1", 0, "L", true, 0, "")
	pdf.CellFormat(40, 6, s.PDFColTier, "1", 1, "L", true, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	for _, sp := range suspects {
		pdf.CellFormat(50, 5, truncate(sp.Food, 30), "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 5, sp.SymptomType, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 5, fmt.Sprintf("%.2f", sp.RiskRatio), "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 5, fmt.Sprintf("%d", sp.N), "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 5, sp.Tier, "1", 1, "L", false, 0, "")
	}
	pdf.Ln(6)
	return nil
}

func parseRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now()
	from := now.AddDate(0, 0, -30)
	to := now
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}
	return from, to
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimRight(string(r[:n-1]), " ,;") + "…"
}
