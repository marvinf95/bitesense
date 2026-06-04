package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/marvinf95/bitesense/backend/internal/auth"
	"github.com/marvinf95/bitesense/backend/internal/config"
)

type Deps struct {
	Cfg     *config.Config
	DB      *sql.DB
	Auth    *auth.Service
	Vision  VisionAnalyzer
	PDF     PDFExporter
	Corr    CorrelationAnalyzer
}

// VisionAnalyzer wraps the per-image analysis pipeline. Implemented in internal/vision.
type VisionAnalyzer interface {
	AnalyzeFromUpload(r *http.Request, userID string) (mealID string, err error)
	AnalyzeFromBarcode(r *http.Request, userID, ean string) (mealID string, err error)
}

type PDFExporter interface {
	Export(w http.ResponseWriter, r *http.Request, userID, locale string) error
}

type CorrelationAnalyzer interface {
	TopSuspects(r *http.Request, userID string) ([]Suspect, error)
}

type Suspect struct {
	Food         string  `json:"food"`
	SymptomType  string  `json:"symptom_type"`
	RiskRatio    float64 `json:"risk_ratio"`
	PValue       float64 `json:"p_value"`
	N            int     `json:"n"`
	AvgHoursLag  float64 `json:"avg_hours_lag"`
	Severity     float64 `json:"avg_severity"`
	Tier         string  `json:"tier"` // STRONG_SUSPECT|SUSPECT|WEAK_SIGNAL
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   d.Cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	health := &HealthHandler{DB: d.DB}
	r.Get("/livez", health.Live)
	r.Get("/readyz", health.Ready)

	authH := &AuthHandler{DB: d.DB, Service: d.Auth}

	r.Route("/api/v1", func(api chi.Router) {
		// Public auth endpoints (with per-IP rate limit).
		api.Group(func(public chi.Router) {
			public.Use(httprate.LimitByIP(20, time.Minute))
			public.Post("/auth/register", authH.Register)
			public.Post("/auth/login", authH.Login)
			public.Post("/auth/refresh", authH.Refresh)
		})

		// Authenticated.
		api.Group(func(priv chi.Router) {
			priv.Use(RequireAuth(d.Auth))

			priv.Get("/auth/me", authH.Me)
			priv.Delete("/auth/account", authH.DeleteAccount)

			meals := &MealsHandler{DB: d.DB}
			priv.Post("/meals", meals.Create)
			priv.Get("/meals", meals.List)
			priv.Get("/meals/{id}", meals.Get)
			priv.Patch("/meals/{id}", meals.Patch)
			priv.Delete("/meals/{id}", meals.Delete)

			if d.Vision != nil {
				priv.Post("/meals/from-image", func(w http.ResponseWriter, r *http.Request) {
					uid, _ := UserIDFrom(r.Context())
					mealID, err := d.Vision.AnalyzeFromUpload(r, uid)
					if err != nil {
						writeError(w, http.StatusBadGateway, err.Error())
						return
					}
					writeJSON(w, http.StatusCreated, map[string]string{"meal_id": mealID})
				})
				priv.Post("/meals/from-barcode/{ean}", func(w http.ResponseWriter, r *http.Request) {
					uid, _ := UserIDFrom(r.Context())
					ean := chi.URLParam(r, "ean")
					mealID, err := d.Vision.AnalyzeFromBarcode(r, uid, ean)
					if err != nil {
						writeError(w, http.StatusBadGateway, err.Error())
						return
					}
					writeJSON(w, http.StatusCreated, map[string]string{"meal_id": mealID})
				})
			}

			symptoms := &SymptomsHandler{DB: d.DB}
			priv.Post("/symptoms", symptoms.Create)
			priv.Get("/symptoms", symptoms.List)
			priv.Patch("/symptoms/{id}", symptoms.Patch)
			priv.Delete("/symptoms/{id}", symptoms.Delete)

			favs := &FavoritesHandler{DB: d.DB}
			priv.Post("/favorites", favs.Create)
			priv.Get("/favorites", favs.List)
			priv.Delete("/favorites/{id}", favs.Delete)

			if d.Corr != nil {
				priv.Get("/analytics/correlations", func(w http.ResponseWriter, r *http.Request) {
					uid, _ := UserIDFrom(r.Context())
					suspects, err := d.Corr.TopSuspects(r, uid)
					if err != nil {
						writeError(w, http.StatusInternalServerError, err.Error())
						return
					}
					writeJSON(w, http.StatusOK, map[string]any{"suspects": suspects})
				})
			}

			if d.PDF != nil {
				priv.Get("/export/pdf", func(w http.ResponseWriter, r *http.Request) {
					uid, _ := UserIDFrom(r.Context())
					locale := r.URL.Query().Get("locale")
					if locale == "" {
						locale = LocaleFrom(r.Context())
					}
					if err := d.PDF.Export(w, r, uid, locale); err != nil {
						writeError(w, http.StatusInternalServerError, err.Error())
						return
					}
				})
			}
		})
	})
	return r
}
