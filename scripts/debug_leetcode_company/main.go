// One-off debug: print raw LeetCode GraphQL responses for a company slug.
//
//	go run ./scripts/debug_leetcode_company google
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	graphqlURL = "https://leetcode.com/graphql"
	cookieFile = "leetcode_cookies_netscape.txt"
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: go run ./scripts/debug_leetcode_company <slug>")
		os.Exit(1)
	}
	slug := strings.TrimSpace(os.Args[1])
	root, _ := os.Getwd()
	cookieHeader, csrf, err := loadNetscapeCookies(filepath.Join(root, cookieFile))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cookies: %v\n", err)
		os.Exit(1)
	}
	client := &http.Client{Timeout: 120 * time.Second}
	ref := "https://leetcode.com/company/" + slug + "/"

	fmt.Println("=== globalData (session / premium) ===")
	globalBody, err := post(client, cookieHeader, csrf, "globalData", "https://leetcode.com/", `query globalData {
  userStatus {
    isSignedIn
    isPremium
    username
    premiumExpiredAt
  }
}`, map[string]any{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "globalData error: %v\n", err)
	} else {
		printJSON(globalBody)
	}
	fmt.Println()

	detailQuery := `query favoriteDetailV2ForCompany($favoriteSlug: String!) {
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
	fmt.Println("=== favoriteDetailV2ForCompany ===")
	detailBody, err := post(client, cookieHeader, csrf, "favoriteDetailV2ForCompany", ref, detailQuery, map[string]string{"favoriteSlug": slug})
	if err != nil {
		fmt.Fprintf(os.Stderr, "detail error: %v\n", err)
		os.Exit(1)
	}
	printJSON(detailBody)

	var parsed struct {
		Data *struct {
			Fav *struct {
				Slug string `json:"slug"`
				GFI  *struct {
					Default string `json:"defaultFavoriteSlug"`
					Cats    []struct {
						CategoryName string `json:"categoryName"`
						FavoriteSlug string `json:"favoriteSlug"`
						DisplayName  string `json:"displayName"`
					} `json:"categoriesToSlugs"`
				} `json:"generatedFavoritesInfo"`
			} `json:"favoriteDetailV2"`
		} `json:"data"`
		Errors json.RawMessage `json:"errors"`
	}
	_ = json.Unmarshal(detailBody, &parsed)
	if parsed.Errors != nil {
		fmt.Println("graphql errors:", string(parsed.Errors))
	}
	if parsed.Data == nil || parsed.Data.Fav == nil || parsed.Data.Fav.GFI == nil {
		fmt.Println("no categories in parsed detail")
		os.Exit(0)
	}
	fmt.Printf("\ndefaultFavoriteSlug: %q\n", parsed.Data.Fav.GFI.Default)
	fmt.Printf("categories: %d\n\n", len(parsed.Data.Fav.GFI.Cats))

	listQuery := `query favoriteQuestionList($favoriteSlug: String!, $filtersV2: QuestionFilterInput, $searchKeyword: String, $sortBy: QuestionSortByInput, $limit: Int, $skip: Int, $version: String = "v2") {
  favoriteQuestionList(
    favoriteSlug: $favoriteSlug
    filtersV2: $filtersV2
    searchKeyword: $searchKeyword
    sortBy: $sortBy
    limit: $limit
    skip: $skip
    version: $version
  ) {
    questions { questionFrontendId title }
    totalLength
    hasMore
  }
}`
	filtersV2 := map[string]any{
		"filterCombineType": "ALL",
		"statusFilter":      map[string]any{"questionStatuses": []any{}, "operator": "IS"},
	}

	for _, cat := range parsed.Data.Fav.GFI.Cats {
		favSlug := cat.FavoriteSlug
		fmt.Printf("=== favoriteQuestionList category=%q favoriteSlug=%q ===\n", cat.CategoryName, favSlug)
		vars := map[string]any{
			"skip":          0,
			"limit":         5,
			"favoriteSlug":  favSlug,
			"filtersV2":     filtersV2,
			"searchKeyword": "",
			"sortBy":        map[string]string{"sortField": "CUSTOM", "sortOrder": "ASCENDING"},
		}
		body, err := post(client, cookieHeader, csrf, "favoriteQuestionList", ref, listQuery, vars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  list error: %v\n", err)
			continue
		}
		printJSON(body)
		fmt.Println()
	}

	if len(parsed.Data.Fav.GFI.Cats) > 0 {
		tryVariants(client, cookieHeader, csrf, ref, listQuery, parsed.Data.Fav.GFI.Cats[0].FavoriteSlug)
	}
}

func tryVariants(client *http.Client, cookieHeader, csrf, ref, listQuery, favSlug string) {
	fmt.Println("=== variant probes (first category slug) ===")
	fullFilters := map[string]any{}
	_ = json.Unmarshal([]byte(`{
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
}`), &fullFilters)

	cases := []struct {
		name string
		vars map[string]any
	}{
		{"no filtersV2", map[string]any{
			"skip": 0, "limit": 5, "favoriteSlug": favSlug,
			"searchKeyword": "",
			"sortBy":        map[string]string{"sortField": "CUSTOM", "sortOrder": "ASCENDING"},
		}},
		{"full filtersV2 (scraper)", map[string]any{
			"skip": 0, "limit": 5, "favoriteSlug": favSlug,
			"filtersV2": fullFilters, "searchKeyword": "",
			"sortBy": map[string]string{"sortField": "CUSTOM", "sortOrder": "ASCENDING"},
		}},
		{"limit 100", map[string]any{
			"skip": 0, "limit": 100, "favoriteSlug": favSlug,
			"filtersV2": fullFilters, "searchKeyword": "",
			"sortBy": map[string]string{"sortField": "CUSTOM", "sortOrder": "ASCENDING"},
		}},
	}
	for _, c := range cases {
		fmt.Printf("--- %s ---\n", c.name)
		body, err := post(client, cookieHeader, csrf, "favoriteQuestionList", ref, listQuery, c.vars)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		var r struct {
			Data struct {
				List struct {
					Total     int `json:"totalLength"`
					Questions []struct {
						ID string `json:"questionFrontendId"`
					} `json:"questions"`
				} `json:"favoriteQuestionList"`
			} `json:"data"`
			Errors json.RawMessage `json:"errors"`
		}
		_ = json.Unmarshal(body, &r)
		fmt.Printf("totalLength=%d questions=%d errors=%s\n", r.Data.List.Total, len(r.Data.List.Questions), string(r.Errors))
	}
}

func printJSON(raw []byte) {
	var pretty json.RawMessage
	if err := json.Unmarshal(raw, &pretty); err != nil {
		fmt.Println(string(raw))
		return
	}
	out, _ := json.MarshalIndent(pretty, "", "  ")
	fmt.Println(string(out))
}

func post(client *http.Client, cookieHeader, csrf, opName, referer, query string, variables any) ([]byte, error) {
	payload, _ := json.Marshal(map[string]any{
		"query":         query,
		"variables":     variables,
		"operationName": opName,
	})
	req, err := http.NewRequest(http.MethodPost, graphqlURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookieHeader)
	req.Header.Set("x-csrftoken", csrf)
	req.Header.Set("x-operation-name", opName)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", "https://leetcode.com")
	req.Header.Set("Referer", referer)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	fmt.Printf("(HTTP %d, %d bytes)\n", resp.StatusCode, len(body))
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 400))
	}
	return body, nil
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
		name, value := fields[5], fields[6]
		parts = append(parts, name+"="+value)
		if name == "csrftoken" {
			csrf = value
		}
	}
	return strings.Join(parts, "; "), csrf, nil
}
