package data

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed companies_manifest.json
var embeddedCompaniesManifest []byte

// companiesManifest is the on-disk / refresh JSON shape written by
// scripts/refresh_leetcode_data and scripts/generate_embedded.
type companiesManifest struct {
	GeneratedAt string              `json:"generatedAt"`
	Source      string              `json:"source,omitempty"`
	Companies   []manifestCompany `json:"companies"`
}

type manifestCompany struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

var (
	globalCompaniesManifest companiesManifest
	globalCatalogSlug       map[string]bool
)

func init() {
	if err := json.Unmarshal(embeddedCompaniesManifest, &globalCompaniesManifest); err != nil {
		panic("data: companies_manifest.json: " + err.Error())
	}
	globalCatalogSlug = make(map[string]bool, len(globalCompaniesManifest.Companies))
	for _, c := range globalCompaniesManifest.Companies {
		slug := strings.ToLower(strings.TrimSpace(c.Slug))
		if slug != "" {
			globalCatalogSlug[slug] = true
		}
	}
}
