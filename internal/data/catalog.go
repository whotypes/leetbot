package data

import (
	"cmp"
	"maps"
	"slices"
	"strings"
)

// DefaultTimeframes is the set of timeframes a company can show when the site has
// no per-timeframe files yet, but the company is in the LeetCode catalog.
var DefaultTimeframes = []string{
	"thirty-days",
	"three-months",
	"six-months",
	"more-than-six-months",
	"all",
}

// CompanyEntry is one row in the /api/companies list.
type CompanyEntry struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	HasData bool   `json:"hasData"`
}

// CompanyCatalog merges the LeetCode manifest (when present) with any extra
// companies that exist only in embedded data, sorted by display name.
func (pbc *ProblemsByCompany) CompanyCatalog() []CompanyEntry {
	bySlug := make(map[string]CompanyEntry)
	for _, c := range globalCompaniesManifest.Companies {
		slug := strings.ToLower(strings.TrimSpace(c.Slug))
		if slug == "" {
			continue
		}
		bySlug[slug] = CompanyEntry{
			Slug:    slug,
			Name:    c.Name,
			HasData: pbc.CompanyHasLocalData(slug),
		}
	}
	for _, slug := range pbc.GetAvailableCompanies() {
		if _, ok := bySlug[slug]; ok {
			continue
		}
		bySlug[slug] = CompanyEntry{
			Slug:    slug,
			Name:    humanizeSlug(slug),
			HasData: pbc.CompanyHasLocalData(slug),
		}
	}
	rows := slices.Collect(maps.Values(bySlug))
	slices.SortFunc(rows, func(a, b CompanyEntry) int {
		return cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return rows
}

// DataLastUpdatedRFC3339 is the last manifest refresh time, or empty if unknown.
func DataLastUpdatedRFC3339() string {
	return strings.TrimSpace(globalCompaniesManifest.GeneratedAt)
}

// InWebCatalog is true if the company appears in the manifest or in embedded data.
func InWebCatalog(pbc *ProblemsByCompany, company string) bool {
	company = strings.ToLower(strings.TrimSpace(company))
	if globalCatalogSlug[company] {
		return true
	}
	return pbc.CompanyExists(company)
}

// TimeframesForWeb returns known timeframes for a company, or DefaultTimeframes
// when the company is listed but has no CSV files yet.
func TimeframesForWeb(pbc *ProblemsByCompany, company string) []string {
	company = strings.ToLower(strings.TrimSpace(company))
	if t := pbc.GetAvailableTimeframes(company); len(t) > 0 {
		return t
	}
	if InWebCatalog(pbc, company) {
		return slices.Clone(DefaultTimeframes)
	}
	return nil
}

func humanizeSlug(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, " ")
}
