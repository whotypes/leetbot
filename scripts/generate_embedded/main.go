package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Problem struct {
	ID         int
	URL        string
	Title      string
	Difficulty string
	Acceptance float64
	Frequency  float64
}

func normalizeVarName(name string) string {
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	reg := regexp.MustCompile("[^a-zA-Z0-9_]")
	name = reg.ReplaceAllString(name, "")

	parts := strings.Split(name, "_")
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	name = strings.Join(parts, "")

	if len(name) > 0 && unicode.IsDigit(rune(name[0])) {
		name = "_" + name
	}

	return name
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run scripts/generate_embedded/main.go <data-directory>")
		os.Exit(1)
	}

	dataDir := os.Args[1]
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Printf("Data directory %s does not exist\n", dataDir)
		os.Exit(1)
	}

	fmt.Printf("Generating embedded data from %s...\n", dataDir)
	companies, err := findCompanies(dataDir)
	if err != nil {
		fmt.Printf("Error finding companies: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d companies\n", len(companies))

	output := `package data


`

	for _, company := range companies {
		companyData, err := loadCompanyData(dataDir, company)
		if err != nil {
			fmt.Printf("Error loading data for %s: %v\n", company, err)
			continue
		}

		for timeframe, problems := range companyData {
			companyName := normalizeVarName(company)
			timeframeName := normalizeVarName(timeframe)
			varName := fmt.Sprintf("%s%sCSV", companyName, timeframeName)
			csvContent, err := generateCSVContent(problems)
			if err != nil {
				fmt.Printf("Error generating embedded CSV for %s/%s: %v\n", company, timeframe, err)
				continue
			}

			output += fmt.Sprintf("var %s = `%s`\n\n", varName, csvContent)
		}
	}

	output += `// embeddedCSVs maps company and timeframe to their embedded CSV data
var embeddedCSVs = map[string]map[string][]byte{
`

	for _, company := range companies {
		output += fmt.Sprintf("\t%q: {\n", company)

		companyData, err := loadCompanyData(dataDir, company)
		if err != nil {
			continue
		}

		for timeframe := range companyData {
			companyName := normalizeVarName(company)
			timeframeName := normalizeVarName(timeframe)
			varName := fmt.Sprintf("%s%sCSV", companyName, timeframeName)
			output += fmt.Sprintf("\t\t%q: []byte(%s),\n", timeframe, varName)
		}

		output += "\t},\n"
	}

	output += "}\n"

	err = os.WriteFile("internal/data/parser_generated.go", []byte(output), 0644)
	if err != nil {
		fmt.Printf("Error writing generated file: %v\n", err)
		os.Exit(1)
	}

	if err := writeCompaniesManifest(dataDir); err != nil {
		fmt.Printf("Error writing companies manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated embedded data for %d companies\n", len(companies))
	fmt.Printf("Generated file: internal/data/parser_generated.go\n")
	fmt.Printf("Generated file: internal/data/companies_manifest.json\n")
}

type manifestCompany struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type manifestFile struct {
	GeneratedAt string            `json:"generatedAt"`
	Source      string            `json:"source,omitempty"`
	Companies   []manifestCompany `json:"companies"`
}

func writeCompaniesManifest(dataDir string) error {
	root := filepath.Dir(dataDir)
	if root == "." || root == "" {
		root = "."
	}
	outPath := filepath.Join(root, "internal", "data", "companies_manifest.json")
	leetPath := filepath.Join(dataDir, "companies_manifest.json")

	var base manifestFile
	if raw, err := os.ReadFile(leetPath); err == nil {
		_ = json.Unmarshal(raw, &base)
	}

	seen := make(map[string]bool)
	var merged []manifestCompany
	for _, c := range base.Companies {
		slug := strings.ToLower(strings.TrimSpace(c.Slug))
		if slug == "" {
			continue
		}
		name := strings.TrimSpace(c.Name)
		if name == "" {
			name = humanizeCompanySlug(slug)
		}
		seen[slug] = true
		merged = append(merged, manifestCompany{Slug: slug, Name: name})
	}

	dirs, err := findCompanies(dataDir)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		if seen[d] {
			continue
		}
		seen[d] = true
		merged = append(merged, manifestCompany{Slug: d, Name: humanizeCompanySlug(d)})
	}

	sort.Slice(merged, func(i, j int) bool {
		ai := strings.ToLower(merged[i].Name)
		aj := strings.ToLower(merged[j].Name)
		if ai != aj {
			return ai < aj
		}
		return merged[i].Slug < merged[j].Slug
	})

	genAt := time.Now().UTC().Format(time.RFC3339)
	if strings.TrimSpace(base.GeneratedAt) != "" {
		genAt = strings.TrimSpace(base.GeneratedAt)
	}
	src := base.Source
	if src == "" {
		src = "data-directories"
	}

	enc, err := json.MarshalIndent(manifestFile{
		GeneratedAt: genAt,
		Source:      src,
		Companies:   merged,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, enc, 0o644)
}

func humanizeCompanySlug(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, " ")
}

func findCompanies(dataDir string) ([]string, error) {
	var companies []string

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			companies = append(companies, entry.Name())
		}
	}

	sort.Strings(companies)
	return companies, nil
}

func loadCompanyData(dataDir, company string) (map[string][]Problem, error) {
	companyDir := filepath.Join(dataDir, company)
	timeframes := map[string][]Problem{}

	entries, err := os.ReadDir(companyDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".csv") {
			timeframe := strings.TrimSuffix(entry.Name(), ".csv")
			problems, err := loadCSV(filepath.Join(companyDir, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("error loading %s: %v", entry.Name(), err)
			}

			sort.Slice(problems, func(i, j int) bool {
				return problems[i].Frequency > problems[j].Frequency
			})

			timeframes[timeframe] = problems
		}
	}

	return timeframes, nil
}

func loadCSV(filename string) ([]Problem, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(strings.NewReader(string(content)))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least header and one data row")
	}
	expectedHeader := []string{"ID", "URL", "Title", "Difficulty", "Acceptance %", "Frequency %"}
	for i := range expectedHeader {
		if i >= len(records[0]) {
			return nil, fmt.Errorf("invalid header: missing columns")
		}
		field := strings.TrimSpace(records[0][i])
		if field != expectedHeader[i] {
			return nil, fmt.Errorf("invalid header[%d]: expected %q, got %q", i, expectedHeader[i], field)
		}
	}
	var problems []Problem
	for rowIdx, rec := range records[1:] {
		if len(rec) < 6 {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSpace(rec[0]))
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid ID: %w", rowIdx+2, err)
		}
		acceptance, err := parsePercentage(rec[4])
		if err != nil {
			return nil, fmt.Errorf("row %d: acceptance: %w", rowIdx+2, err)
		}
		frequency, err := parsePercentage(rec[5])
		if err != nil {
			return nil, fmt.Errorf("row %d: frequency: %w", rowIdx+2, err)
		}
		problems = append(problems, Problem{
			ID:         id,
			URL:        strings.TrimSpace(rec[1]),
			Title:      strings.TrimSpace(rec[2]),
			Difficulty: strings.TrimSpace(rec[3]),
			Acceptance: acceptance,
			Frequency:  frequency,
		})
	}
	return problems, nil
}

func parsePercentage(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

func generateCSVContent(problems []Problem) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	header := []string{"ID", "URL", "Title", "Difficulty", "Acceptance %", "Frequency %"}
	if err := w.Write(header); err != nil {
		return "", err
	}
	for _, problem := range problems {
		// backticks would break the generated Go raw string literals
		title := strings.ReplaceAll(problem.Title, "`", "'")
		title = strings.ReplaceAll(title, "\n", " ")
		title = strings.ReplaceAll(title, "\r", " ")
		title = strings.ReplaceAll(title, "\t", " ")
		rec := []string{
			strconv.Itoa(problem.ID),
			problem.URL,
			title,
			problem.Difficulty,
			fmt.Sprintf("%.1f%%", problem.Acceptance),
			fmt.Sprintf("%.1f%%", problem.Frequency),
		}
		if err := w.Write(rec); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}
