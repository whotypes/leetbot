// Command refresh_leetcode_data refetches LeetCode company problem lists via GraphQL and
// merges into data/<slug>/<timeframe>.csv, then runs scripts/generate_embedded.
//
// Run from repository root with a valid leetcode_cookies_netscape.txt:
//
//	go run ./scripts/refresh_leetcode_data
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	graphqlURL     = "https://leetcode.com/graphql"
	cookieFile     = "leetcode_cookies_netscape.txt"
	dataDir        = "data"
	userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
	requestSleep   = 350 * time.Millisecond
	maxRetries     = 4
	pageLimit      = 100
	generateScript = "scripts/generate_embedded/main.go"
)

// filtersV2 from HAR (favoriteQuestionList on company page).
const favoriteListFiltersV2JSON = `{
  "filterCombineType": "ALL",
  "statusFilter": {"questionStatuses": [], "operator": "IS"},
  "difficultyFilter": {"difficulties": [], "operator": "IS"},
  "languageFilter": {"languageSlugs": [], "operator": "IS"},
  "topicFilter": {"topicSlugs": [], "operator": "IS"},
  "acceptanceFilter": {},
  "frequencyFilter": {},
  "frontendIdFilter": {},
  "lastSubmittedFilter": {},
  "publishedFilter": {},
  "companyFilter": {"companySlugs": [], "operator": "IS"},
  "positionFilter": {"positionSlugs": [], "operator": "IS"},
  "positionLevelFilter": {"positionLevelSlugs": [], "operator": "IS"},
  "contestPointFilter": {"contestPoints": [], "operator": "IS"},
  "premiumFilter": {"premiumStatus": [], "operator": "IS"}
}`

var favoriteListFiltersV2 any

func init() {
	if err := json.Unmarshal([]byte(favoriteListFiltersV2JSON), &favoriteListFiltersV2); err != nil {
		panic(err)
	}
}

const queryProblemsetCompanyTags = `query problemsetCompanyTags {
  problemsetCompanyTags {
    name
    slug
  }
}`

const queryFavoriteDetailV2ForCompany = `query favoriteDetailV2ForCompany($favoriteSlug: String!) {
  favoriteDetailV2(favoriteSlug: $favoriteSlug) {
    slug
    generatedFavoritesInfo {
      defaultFavoriteSlug
      categoriesToSlugs {
        categoryName
        favoriteSlug
        displayName
      }
    }
  }
}`

const queryFavoriteQuestionList = `query favoriteQuestionList($favoriteSlug: String!, $filter: FavoriteQuestionFilterInput, $filtersV2: QuestionFilterInput, $searchKeyword: String, $sortBy: QuestionSortByInput, $limit: Int, $skip: Int, $version: String = "v2") {
  favoriteQuestionList(
    favoriteSlug: $favoriteSlug
    filter: $filter
    filtersV2: $filtersV2
    searchKeyword: $searchKeyword
    sortBy: $sortBy
    limit: $limit
    skip: $skip
    version: $version
  ) {
    questions {
      difficulty
      questionFrontendId
      title
      titleSlug
      frequency
      acRate
    }
    totalLength
    hasMore
  }
}`

type graphQLPayload struct {
	Query         string `json:"query"`
	Variables     any    `json:"variables,omitempty"`
	OperationName string `json:"operationName"`
}

type graphQLErrors struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type problemRow struct {
	ID         int
	URL        string
	Title      string
	Difficulty string
	Acceptance float64
	Frequency  float64
}

func main() {
	log.SetFlags(0)
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cookiePath := filepath.Join(root, cookieFile)
	cookieHeader, csrf, err := loadNetscapeCookies(cookiePath)
	if err != nil {
		log.Fatalf("cookies: %v (expected %s in repo root)", err, cookieFile)
	}
	if csrf == "" {
		log.Fatal("csrftoken not found in cookie file")
	}

	client := &http.Client{Timeout: 120 * time.Second}

	tags, err := fetchProblemsetCompanyTags(client, cookieHeader, csrf)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("problemsetCompanyTags: %d companies\n", len(tags))

	manifestPath := filepath.Join(root, dataDir, "companies_manifest.json")
	if err := writeCompaniesManifestFile(manifestPath, tags); err != nil {
		log.Fatalf("write companies manifest: %v", err)
	}
	log.Printf("wrote %s\n", manifestPath)

	success := 0
	skipped := 0
	for i, tag := range tags {
		if err := refreshCompany(root, client, cookieHeader, csrf, tag.Slug); err != nil {
			log.Printf("[%d/%d] %s: skip: %v\n", i+1, len(tags), tag.Slug, err)
			skipped++
			continue
		}
		success++
		log.Printf("[%d/%d] %s: ok\n", i+1, len(tags), tag.Slug)
	}
	log.Printf("done: %d ok, %d skipped, %d total\n", success, skipped, len(tags))

	gen := filepath.Join(root, generateScript)
	dataPath := filepath.Join(root, dataDir)
	cmd := exec.Command("go", "run", gen, dataPath)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println("running generate_embedded...")
	if err := cmd.Run(); err != nil {
		log.Fatalf("generate_embedded: %v", err)
	}
	log.Println("internal/data/parser_generated.go updated")
}

type companyTag struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func writeCompaniesManifestFile(path string, tags []companyTag) error {
	type row struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	var companies []row
	for _, t := range tags {
		slug := strings.ToLower(strings.TrimSpace(t.Slug))
		if slug == "" {
			continue
		}
		name := strings.TrimSpace(t.Name)
		if name == "" {
			name = slug
		}
		companies = append(companies, row{Slug: slug, Name: name})
	}
	sort.Slice(companies, func(i, j int) bool {
		ai := strings.ToLower(companies[i].Name)
		aj := strings.ToLower(companies[j].Name)
		if ai != aj {
			return ai < aj
		}
		return companies[i].Slug < companies[j].Slug
	})
	doc := struct {
		GeneratedAt string `json:"generatedAt"`
		Source      string `json:"source"`
		Companies   []row  `json:"companies"`
	}{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Source:      "leetcode-problemsetCompanyTags",
		Companies:   companies,
	}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func fetchProblemsetCompanyTags(client *http.Client, cookieHeader, csrf string) ([]companyTag, error) {
	body, err := postGraphQL(client, cookieHeader, csrf, "problemsetCompanyTags",
		"https://leetcode.com/problemset/", queryProblemsetCompanyTags, map[string]any{})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data struct {
			Tags []companyTag `json:"problemsetCompanyTags"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Data.Tags, nil
}

func refreshCompany(root string, client *http.Client, cookieHeader, csrf, companySlug string) error {
	ref := "https://leetcode.com/company/" + companySlug + "/"
	detailBody, err := postGraphQL(client, cookieHeader, csrf, "favoriteDetailV2ForCompany", ref, queryFavoriteDetailV2ForCompany, map[string]string{"favoriteSlug": companySlug})
	if err != nil {
		return err
	}
	var detail struct {
		Data *struct {
			Favorite *struct {
				Slug string `json:"slug"`
				GFI  *struct {
					Categories []struct {
						CategoryName string `json:"categoryName"`
						FavoriteSlug string `json:"favoriteSlug"`
					} `json:"categoriesToSlugs"`
				} `json:"generatedFavoritesInfo"`
			} `json:"favoriteDetailV2"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailBody, &detail); err != nil {
		return fmt.Errorf("decode detail: %w", err)
	}
	if detail.Data == nil || detail.Data.Favorite == nil || detail.Data.Favorite.GFI == nil {
		return fmt.Errorf("no favorite detail")
	}
	cats := detail.Data.Favorite.GFI.Categories
	if len(cats) == 0 {
		return fmt.Errorf("empty categoriesToSlugs")
	}

	dir := filepath.Join(root, dataDir, companySlug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	for _, cat := range cats {
		tf, ok := categoryToTimeframe(cat.CategoryName)
		if !ok {
			log.Printf("  unknown categoryName %q, skip\n", cat.CategoryName)
			continue
		}
		favSlug := cat.FavoriteSlug
		if favSlug == "" {
			continue
		}
		questions, err := fetchAllFavoriteQuestions(client, cookieHeader, csrf, ref, favSlug)
		if err != nil {
			return fmt.Errorf("%s: %w", tf, err)
		}
		csvPath := filepath.Join(dir, tf+".csv")
		if len(questions) == 0 {
			// No questions for this list on LeetCode: do not create or keep an empty/hdr-only CSV.
			if err := os.Remove(csvPath); err == nil {
				log.Printf("  %s: 0 questions, removed %s\n", tf, filepath.Base(csvPath))
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("%s: remove empty csv: %w", tf, err)
			} else {
				log.Printf("  %s: 0 questions, no file to write\n", tf)
			}
			continue
		}
		if err := mergeAndWriteCSV(csvPath, questions); err != nil {
			return fmt.Errorf("%s: %w", tf, err)
		}
	}
	return nil
}

func categoryToTimeframe(categoryName string) (string, bool) {
	switch categoryName {
	case "thirty_days":
		return "thirty-days", true
	case "three_months":
		return "three-months", true
	case "six_months":
		return "six-months", true
	case "more_than_six_months":
		return "more-than-six-months", true
	case "all":
		return "all", true
	default:
		return "", false
	}
}

type gqlQuestion struct {
	Difficulty         string  `json:"difficulty"`
	QuestionFrontendID string  `json:"questionFrontendId"`
	Title              string  `json:"title"`
	TitleSlug          string  `json:"titleSlug"`
	Frequency          float64 `json:"frequency"`
	AcRate             float64 `json:"acRate"`
}

func fetchAllFavoriteQuestions(client *http.Client, cookieHeader, csrf, referer, favoriteSlug string) ([]problemRow, error) {
	var out []problemRow
	skip := 0
	for {
		time.Sleep(requestSleep)
		vars := map[string]any{
			"skip":          skip,
			"limit":         pageLimit,
			"favoriteSlug":  favoriteSlug,
			"filtersV2":     favoriteListFiltersV2,
			"searchKeyword": "",
			"sortBy":        map[string]string{"sortField": "CUSTOM", "sortOrder": "ASCENDING"},
		}
		body, err := postGraphQL(client, cookieHeader, csrf, "favoriteQuestionList", referer, queryFavoriteQuestionList, vars)
		if err != nil {
			return nil, err
		}
		var resp struct {
			Data *struct {
				List *struct {
					Questions []gqlQuestion `json:"questions"`
					HasMore   bool            `json:"hasMore"`
				} `json:"favoriteQuestionList"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, err
		}
		if resp.Data == nil || resp.Data.List == nil {
			return nil, fmt.Errorf("empty favoriteQuestionList response")
		}
		for _, q := range resp.Data.List.Questions {
			id, err := strconv.Atoi(strings.TrimSpace(q.QuestionFrontendID))
			if err != nil {
				continue
			}
			out = append(out, problemRow{
				ID:         id,
				URL:        "https://leetcode.com/problems/" + q.TitleSlug,
				Title:      q.Title,
				Difficulty: normalizeDifficulty(q.Difficulty),
				Acceptance: q.AcRate * 100,
				Frequency:  q.Frequency,
			})
		}
		if !resp.Data.List.HasMore || len(resp.Data.List.Questions) == 0 {
			break
		}
		skip += pageLimit
	}
	return out, nil
}

func normalizeDifficulty(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "EASY":
		return "Easy"
	case "MEDIUM":
		return "Medium"
	case "HARD":
		return "Hard"
	default:
		return s
	}
}

func postGraphQL(client *http.Client, cookieHeader, csrf, opName, referer, query string, variables any) ([]byte, error) {
	payload := graphQLPayload{
		Query:         query,
		Variables:     variables,
		OperationName: opName,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
		req, err := http.NewRequest(http.MethodPost, graphqlURL, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Cookie", cookieHeader)
		req.Header.Set("x-csrftoken", csrf)
		req.Header.Set("x-operation-name", opName)
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Origin", "https://leetcode.com")
		req.Header.Set("Referer", referer)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
		}
		var gerr graphQLErrors
		_ = json.Unmarshal(body, &gerr)
		if len(gerr.Errors) > 0 {
			return nil, fmt.Errorf("graphql: %s", gerr.Errors[0].Message)
		}
		return body, nil
	}
	return nil, fmt.Errorf("after retries: %w", lastErr)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func loadNetscapeCookies(path string) (header string, csrf string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	var parts []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		name := fields[5]
		value := fields[6]
		parts = append(parts, name+"="+value)
		if name == "csrftoken" {
			csrf = value
		}
	}
	return strings.Join(parts, "; "), csrf, nil
}

func mergeAndWriteCSV(path string, fetched []problemRow) error {
	byID := make(map[int]problemRow)

	// Existing rows (never remove)
	if raw, err := os.ReadFile(path); err == nil && len(raw) > 0 {
		r := csv.NewReader(strings.NewReader(string(raw)))
		records, err := r.ReadAll()
		if err == nil && len(records) > 1 {
			for _, rec := range records[1:] {
				if len(rec) < 6 {
					continue
				}
				id, err := strconv.Atoi(strings.TrimSpace(rec[0]))
				if err != nil {
					continue
				}
				acc, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(rec[4]), "%"), 64)
				freq, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(rec[5]), "%"), 64)
				byID[id] = problemRow{
					ID:         id,
					URL:        rec[1],
					Title:      rec[2],
					Difficulty: rec[3],
					Acceptance: acc,
					Frequency:  freq,
				}
			}
		}
	}

	// Merge / overwrite from API
	for _, p := range fetched {
		byID[p.ID] = p
	}

	var list []problemRow
	for _, p := range byID {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Frequency != list[j].Frequency {
			return list[i].Frequency > list[j].Frequency
		}
		return list[i].ID < list[j].ID
	})

	tmp := path + ".tmp"
	if err := writeCSVFile(tmp, list); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func writeCSVFile(path string, problems []problemRow) error {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	header := []string{"ID", "URL", "Title", "Difficulty", "Acceptance %", "Frequency %"}
	if err := w.Write(header); err != nil {
		return err
	}
	for _, problem := range problems {
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
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.TrimSuffix(buf.String(), "\n")), 0o644)
}
