// Package foodfacts is a thin Open Food Facts client (search + barcode lookup).
package foodfacts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/marvinf95/bitesense/backend/internal/models"
)

const baseURL = "https://world.openfoodfacts.org"

type Client struct {
	HTTP      *http.Client
	UserAgent string
}

func New(userAgent string) *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: 10 * time.Second},
		UserAgent: userAgent,
	}
}

type Product struct {
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Brands       string   `json:"brands"`
	Ingredients  []string `json:"ingredients"`
	Allergens    []string `json:"allergens"`
	ServingSizeG float64  `json:"serving_size_g"`
}

func (c *Client) Lookup(ctx context.Context, ean string) (*Product, error) {
	if !isDigits(ean) || len(ean) < 8 {
		return nil, fmt.Errorf("invalid EAN")
	}
	u := fmt.Sprintf("%s/api/v2/product/%s.json", baseURL, ean)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("off lookup status %d", resp.StatusCode)
	}
	var body struct {
		Status  int `json:"status"`
		Product struct {
			Code            string   `json:"code"`
			ProductName     string   `json:"product_name"`
			ProductNameEN   string   `json:"product_name_en"`
			ProductNameDE   string   `json:"product_name_de"`
			Brands          string   `json:"brands"`
			IngredientsText string   `json:"ingredients_text"`
			AllergensTags   []string `json:"allergens_tags"`
			ServingSize     string   `json:"serving_size"`
		} `json:"product"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.Status != 1 {
		return nil, fmt.Errorf("product not found")
	}
	name := body.Product.ProductName
	if name == "" {
		name = body.Product.ProductNameEN
	}
	if name == "" {
		name = body.Product.ProductNameDE
	}
	return &Product{
		Code:         body.Product.Code,
		Name:         name,
		Brands:       body.Product.Brands,
		Ingredients:  splitIngredients(body.Product.IngredientsText),
		Allergens:    canonicalAllergens(body.Product.AllergensTags),
		ServingSizeG: parseGrams(body.Product.ServingSize),
	}, nil
}

type SearchHit struct {
	Name      string
	Allergens []string
}

func (c *Client) Search(ctx context.Context, q string, limit int) ([]SearchHit, error) {
	if strings.TrimSpace(q) == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 1
	}
	params := url.Values{}
	params.Set("search_terms", q)
	params.Set("search_simple", "1")
	params.Set("action", "process")
	params.Set("json", "1")
	params.Set("page_size", fmt.Sprintf("%d", limit))
	params.Set("fields", "product_name,product_name_en,allergens_tags")
	u := baseURL + "/cgi/search.pl?" + params.Encode()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("off search status %d", resp.StatusCode)
	}
	var body struct {
		Products []struct {
			ProductName   string   `json:"product_name"`
			ProductNameEN string   `json:"product_name_en"`
			AllergensTags []string `json:"allergens_tags"`
		} `json:"products"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	out := make([]SearchHit, 0, len(body.Products))
	for _, p := range body.Products {
		name := p.ProductName
		if name == "" {
			name = p.ProductNameEN
		}
		if name == "" {
			continue
		}
		out = append(out, SearchHit{
			Name:      name,
			Allergens: canonicalAllergens(p.AllergensTags),
		})
	}
	return out, nil
}

func splitIngredients(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, strings.ToLower(t))
		}
	}
	return out
}

// canonicalAllergens maps OFF tags (e.g. "en:gluten") to our internal tag namespace.
func canonicalAllergens(tags []string) []string {
	mapping := map[string]string{
		"en:gluten":             models.TagGluten,
		"en:milk":               models.TagLactose,
		"en:lactose":            models.TagLactose,
		"en:nuts":               models.TagNuts,
		"en:peanuts":            models.TagNuts,
		"en:tree-nuts":          models.TagNuts,
		"en:eggs":               models.TagEgg,
		"en:soybeans":           models.TagSoy,
		"en:soy":                models.TagSoy,
		"en:fructose":           models.TagFructose,
		"en:fish":               models.TagFish,
		"en:crustaceans":        models.TagShellfish,
		"en:molluscs":           models.TagShellfish,
		"en:sulphur-dioxide-and-sulphites": models.TagSulphites,
		"en:sesame-seeds":       models.TagSesame,
		"en:mustard":            models.TagMustard,
		"en:celery":             models.TagCelery,
	}
	seen := map[string]bool{}
	out := []string{}
	for _, t := range tags {
		if v, ok := mapping[t]; ok && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func parseGrams(s string) float64 {
	if s == "" {
		return 0
	}
	s = strings.ToLower(strings.TrimSpace(s))
	var n float64
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			n = n*10 + float64(ch-'0')
		} else if ch == '.' || ch == ',' {
			continue
		} else if ch == ' ' {
			continue
		} else {
			break
		}
	}
	return n
}

func isDigits(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return s != ""
}
