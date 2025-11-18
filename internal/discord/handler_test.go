package discord

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/whotypes/leetbot/internal/data"
)

func createTestProblemsData() *data.ProblemsByCompany {
	testData := map[string]map[string][]data.Problem{
		"airbnb": {
			"all": []data.Problem{
				{
					ID:         1,
					URL:        "https://leetcode.com/problems/two-sum",
					Title:      "Two Sum",
					Difficulty: "Easy",
					Acceptance: 55.9,
					Frequency:  100.0,
				},
				{
					ID:         2,
					URL:        "https://leetcode.com/problems/add-two-numbers",
					Title:      "Add Two Numbers",
					Difficulty: "Medium",
					Acceptance: 46.4,
					Frequency:  75.0,
				},
			},
			"thirty-days": []data.Problem{
				{
					ID:         68,
					URL:        "https://leetcode.com/problems/text-justification",
					Title:      "Text Justification",
					Difficulty: "Hard",
					Acceptance: 48.4,
					Frequency:  100.0,
				},
			},
		},
		"amazon": {
			"all": []data.Problem{
				{
					ID:         1,
					URL:        "https://leetcode.com/problems/two-sum",
					Title:      "Two Sum",
					Difficulty: "Easy",
					Acceptance: 55.9,
					Frequency:  100.0,
				},
			},
		},
	}

	return data.NewTestProblemsByCompany(testData)
}

func TestNewHandler(t *testing.T) {
	problemsData := createTestProblemsData()
	handler := NewHandler(problemsData, "!")

	if handler.problemsData == nil {
		t.Error("NewHandler() should set problemsData")
	}

	if handler.prefix != "!" {
		t.Errorf("NewHandler() prefix = %v, want %v", handler.prefix, "!")
	}
}

func TestNormalizeTimeframe(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	tests := []struct {
		input    string
		expected string
	}{
		{"30", "thirty-days"},
		{"30d", "thirty-days"},
		{"30days", "thirty-days"},
		{"thirty", "thirty-days"},
		{"thirtydays", "thirty-days"},
		{"90", "three-months"},
		{"3mo", "three-months"},
		{"3months", "three-months"},
		{"three", "three-months"},
		{"threemonths", "three-months"},
		{"180", "six-months"},
		{"6mo", "six-months"},
		{"6months", "six-months"},
		{"six", "six-months"},
		{"sixmonths", "six-months"},
		{">6mo", "more-than-six-months"},
		{"all", "all"},
		{"alltime", "all"},
		{"everything", "all"},
		{"ALL", "all"},
		{"Thirty-Days", "thirty-days"},
		{"invalid", "all"},
	}

	for _, tt := range tests {
		result := handler.NormalizeTimeframe(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeTimeframe(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatTimeframeDisplay(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	tests := []struct {
		input    string
		expected string
	}{
		{"all", "all"},
		{"thirty-days", "last 30 days"},
		{"three-months", "last 3 months"},
		{"six-months", "last 6 months"},
		{"more-than-six-months", "more than 6 months"},
		{"invalid", "invalid"},
	}

	for _, tt := range tests {
		result := handler.formatTimeframeDisplay(tt.input)
		if result != tt.expected {
			t.Errorf("formatTimeframeDisplay(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatProblemsResponse(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	problems := []data.Problem{
		{
			ID:         1,
			URL:        "https://leetcode.com/problems/two-sum",
			Title:      "Two Sum",
			Difficulty: "Easy",
			Acceptance: 55.9,
			Frequency:  100.0,
		},
		{
			ID:         2,
			URL:        "https://leetcode.com/problems/add-two-numbers",
			Title:      "Add Two Numbers",
			Difficulty: "Medium",
			Acceptance: 46.4,
			Frequency:  75.0,
		},
	}

	result := handler.formatProblemsResponse("airbnb", "all", problems)

	if !contains(result, "Most Popular Problems for Airbnb (all):") {
		t.Error("formatProblemsResponse() should contain title")
	}

	if !contains(result, "Two Sum (100%)") {
		t.Error("formatProblemsResponse() should contain first problem")
	}

	if !contains(result, "Add Two Numbers (75%)") {
		t.Error("formatProblemsResponse() should contain second problem")
	}

	if len(result) > 2000 {
		t.Errorf("formatProblemsResponse() result too long: %d characters", len(result))
	}
}

func TestFormatProblemsResponse_Empty(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	result := handler.formatProblemsResponse("airbnb", "all", []data.Problem{})

	if !contains(result, "No problems found") {
		t.Error("formatProblemsResponse() should handle empty problems list")
	}
}

func TestHandleMessage(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content:   "!problems airbnb",
			ChannelID: "channel123",
		},
	}

	handler.HandleMessage(session, message)
}

func TestHandleMessage_BotMessage(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "bot123",
				Bot: true,
			},
			Content: "!problems airbnb",
		},
	}

	handler.HandleMessage(session, message)
}

func TestHandleMessage_WrongPrefix(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content: "problems airbnb",
		},
	}

	handler.HandleMessage(session, message)
}

func TestHandleMessage_UnknownCommand(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content: "!unknown command",
		},
	}

	handler.HandleMessage(session, message)
}

func TestHandleMessage_ProblemsNoArgs(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content:   "!problems",
			ChannelID: "channel123",
		},
	}

	handler.HandleMessage(session, message)
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || (len(s) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFindCompanyByFuzzySearch(t *testing.T) {
	testData := map[string]map[string][]data.Problem{
		"palantir-technologies": {
			"all": []data.Problem{
				{ID: 1, Title: "Test", Difficulty: "Easy", Frequency: 100.0},
			},
		},
		"capital-one": {
			"all": []data.Problem{
				{ID: 2, Title: "Test", Difficulty: "Medium", Frequency: 90.0},
			},
		},
		"goldman-sachs": {
			"all": []data.Problem{
				{ID: 3, Title: "Test", Difficulty: "Hard", Frequency: 85.0},
			},
		},
		"jane-street": {
			"all": []data.Problem{
				{ID: 4, Title: "Test", Difficulty: "Medium", Frequency: 80.0},
			},
		},
	}
	problemsData := data.NewTestProblemsByCompany(testData)

	tests := []struct {
		input    string
		expected string
		found    bool
	}{
		{"palantir", "palantir-technologies", true},
		{"palantir-technologies", "palantir-technologies", true},
		{"Palantir Technologies", "palantir-technologies", true},
		{"capital one", "capital-one", true},
		{"Capital One", "capital-one", true},
		{"capital-one", "capital-one", true},
		{"goldman", "goldman-sachs", true},
		{"jane street", "jane-street", true},
		{"nonexistent-company", "", false},
	}

	for _, tt := range tests {
		company, found := findCompanyByFuzzySearch(tt.input, problemsData)
		if found != tt.found {
			t.Errorf("findCompanyByFuzzySearch(%q) found = %v, want %v", tt.input, found, tt.found)
		}
		if found && company != tt.expected {
			t.Errorf("findCompanyByFuzzySearch(%q) = %q, want %q", tt.input, company, tt.expected)
		}
	}
}

func TestGetCompanyAutocompleteChoices(t *testing.T) {
	testData := map[string]map[string][]data.Problem{
		"airbnb": {
			"all": []data.Problem{{ID: 1, Title: "Test", Difficulty: "Easy", Frequency: 100.0}},
		},
		"amazon": {
			"all": []data.Problem{{ID: 2, Title: "Test", Difficulty: "Medium", Frequency: 90.0}},
		},
		"apple": {
			"all": []data.Problem{{ID: 3, Title: "Test", Difficulty: "Hard", Frequency: 85.0}},
		},
	}
	problemsData := data.NewTestProblemsByCompany(testData)

	choices := getCompanyAutocompleteChoices("", problemsData)
	if len(choices) == 0 {
		t.Error("getCompanyAutocompleteChoices with empty input should return choices")
	}
	choices = getCompanyAutocompleteChoices("air", problemsData)
	if len(choices) == 0 {
		t.Error("getCompanyAutocompleteChoices('air') should return choices")
	}

	foundAirbnb := false
	for _, choice := range choices {
		if choice.Value == "airbnb" {
			foundAirbnb = true
			break
		}
	}
	if !foundAirbnb {
		t.Error("getCompanyAutocompleteChoices('air') should include airbnb")
	}
	if len(choices) > 25 {
		t.Errorf("getCompanyAutocompleteChoices should return max 25 choices, got %d", len(choices))
	}
}

func TestIsTimeframeKeyword(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	tests := []struct {
		input    string
		expected bool
	}{
		{"30d", true},
		{"3mo", true},
		{"6mo", true},
		{">6mo", true},
		{"all", true},
		{"thirty-days", true},
		{"three-months", true},
		{"six-months", true},
		{"more-than-six-months", true},
		{"company", false},
		{"amazon", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := handler.isTimeframeKeyword(tt.input)
		if result != tt.expected {
			t.Errorf("isTimeframeKeyword(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestFormatAvailableTimeframesSuggestion(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	availableTimeframes := []string{"all", "six-months"}
	result := handler.formatAvailableTimeframesSuggestion("starbucks", "thirty-days", availableTimeframes)

	if !contains(result, "No data found for Starbucks") {
		t.Error("Suggestion should mention no data found for company")
	}
	if !contains(result, "Available timeframes for Starbucks:") {
		t.Error("Suggestion should list available timeframes")
	}
	if !contains(result, "all") {
		t.Error("Suggestion should include 'all' timeframe")
	}
	if !contains(result, "6mo") {
		t.Error("Suggestion should include short alias for six-months")
	}
	if !contains(result, "!problems starbucks") {
		t.Error("Suggestion should include example command")
	}
}

func TestGetTimeframeShortAlias(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	tests := []struct {
		input    string
		expected string
	}{
		{"thirty-days", "30d"},
		{"three-months", "3mo"},
		{"six-months", "6mo"},
		{"more-than-six-months", ">6mo"},
		{"all", "all"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		result := handler.getTimeframeShortAlias(tt.input)
		if result != tt.expected {
			t.Errorf("getTimeframeShortAlias(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatProblemsResponse_LongList(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	var longList []data.Problem
	for i := 1; i <= 100; i++ {
		longList = append(longList, data.Problem{
			ID:         i,
			URL:        "https://leetcode.com/problems/test",
			Title:      "Test Problem",
			Difficulty: "Easy",
			Acceptance: 50.0,
			Frequency:  100.0,
		})
	}

	result := handler.formatProblemsResponse("test-company", "all", longList)

	if len(result) > 2000 {
		t.Errorf("formatProblemsResponse() length = %d, exceeds Discord limit of 2000", len(result))
	}
}

func TestHandleMessage_ProblemsWithTimeframe(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content:   "!problems airbnb 30d",
			ChannelID: "channel123",
		},
	}

	handler.HandleMessage(session, message)
}

func TestHandleMessage_UnknownCompany(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content:   "!problems unknowncompany",
			ChannelID: "channel123",
		},
	}

	handler.HandleMessage(session, message)
}

func TestNormalizeTimeframe_EdgeCases(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	tests := []struct {
		input    string
		expected string
	}{
		{"  30d  ", "thirty-days"},

		{"AlL", "all"},
		{"ThIrTy-DaYs", "thirty-days"},

		{"", "all"},

		{"thirty days", "thirty-days"},
		{"three months", "three-months"},
	}

	for _, tt := range tests {
		result := handler.NormalizeTimeframe(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeTimeframe(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatProblemsResponse_DifferentDifficulties(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	problems := []data.Problem{
		{ID: 1, Title: "Easy Problem", Difficulty: "Easy", Frequency: 100.0},
		{ID: 2, Title: "Medium Problem", Difficulty: "Medium", Frequency: 90.0},
		{ID: 3, Title: "Hard Problem", Difficulty: "Hard", Frequency: 80.0},
	}

	result := handler.formatProblemsResponse("test-company", "all", problems)

	if !contains(result, "🟢") {
		t.Error("formatProblemsResponse() should contain 🟢 for Easy")
	}
	if !contains(result, "🟡") {
		t.Error("formatProblemsResponse() should contain 🟡 for Medium")
	}
	if !contains(result, "🔴") {
		t.Error("formatProblemsResponse() should contain 🔴 for Hard")
	}
}

// Test fuzzy matching with confidence thresholds
func TestFindCompanyWithSuggestion(t *testing.T) {
	testData := map[string]map[string][]data.Problem{
		"google": {
			"all": []data.Problem{{ID: 1, Title: "Test", Difficulty: "Easy", Frequency: 100.0}},
		},
		"facebook": {
			"all": []data.Problem{{ID: 2, Title: "Test", Difficulty: "Medium", Frequency: 90.0}},
		},
		"amazon": {
			"all": []data.Problem{{ID: 3, Title: "Test", Difficulty: "Hard", Frequency: 85.0}},
		},
		"microsoft": {
			"all": []data.Problem{{ID: 4, Title: "Test", Difficulty: "Medium", Frequency: 80.0}},
		},
		"apple": {
			"all": []data.Problem{{ID: 5, Title: "Test", Difficulty: "Easy", Frequency: 75.0}},
		},
		"dropbox": {
			"all": []data.Problem{{ID: 6, Title: "Test", Difficulty: "Medium", Frequency: 70.0}},
		},
		"box": {
			"all": []data.Problem{{ID: 7, Title: "Test", Difficulty: "Hard", Frequency: 65.0}},
		},
		"jane-street": {
			"all": []data.Problem{{ID: 8, Title: "Test", Difficulty: "Hard", Frequency: 60.0}},
		},
		"jump-trading": {
			"all": []data.Problem{{ID: 9, Title: "Test", Difficulty: "Hard", Frequency: 55.0}},
		},
		"the-trade-desk": {
			"all": []data.Problem{{ID: 10, Title: "Test", Difficulty: "Medium", Frequency: 50.0}},
		},
		"amd": {
			"all": []data.Problem{{ID: 11, Title: "AMD Problem", Difficulty: "Hard", Frequency: 45.0}},
		},
		"td": {
			"all": []data.Problem{{ID: 12, Title: "TD Problem", Difficulty: "Easy", Frequency: 40.0}},
		},
	}
	problemsData := data.NewTestProblemsByCompany(testData)

	tests := []struct {
		name              string
		input             string
		expectedFound     bool
		expectedCompany   string
		expectSuggestions bool
	}{
		// High confidence auto-corrections
		{"exact match", "google", true, "google", false},
		{"exact match with spaces", "jane street", true, "jane-street", false},
		{"exact match with hyphens", "jane-street", true, "jane-street", false},
		{"close typo", "googl", true, "google", false},
		{"close typo 2", "amazn", true, "amazon", false},
		{"close typo 3", "microsft", true, "microsoft", false},

		// Company aliases
		{"meta alias", "meta", true, "facebook", false},
		{"fb alias", "fb", true, "facebook", false},
		{"alphabet alias", "alphabet", true, "google", false},

		// Medium confidence suggestions (these actually auto-correct due to high confidence)
		{"medium confidence", "goog", true, "google", false},       // auto-corrects due to high confidence
		{"medium confidence 2", "amaz", true, "amazon", false},     // auto-corrects due to high confidence
		{"medium confidence 3", "micro", true, "microsoft", false}, // auto-corrects due to high confidence

		// Ambiguous cases
		{"dropbox vs box - dropbox", "drop", true, "dropbox", false},          // auto-corrects to dropbox
		{"dropbox vs box - box", "box", true, "box", false},                   // exact match to box
		{"dropbox vs box - dropbox exact", "dropbox", true, "dropbox", false}, // exact match to dropbox

		// Multi-word companies (these auto-correct due to high confidence)
		{"jane street partial", "jane", true, "jane-street", false},        // auto-corrects to jane-street
		{"jump trading partial", "jump", true, "jump-trading", false},      // auto-corrects to jump-trading
		{"the trade desk partial", "trade", true, "the-trade-desk", false}, // auto-corrects to the-trade-desk

		// Low confidence - should get suggestions
		{"low confidence", "xyz", false, "", true},
		{"very different", "completely-different", false, "", true},

		// Test "zon" -> "amazon" case
		{"zon to amazon", "zon", true, "amazon", false}, // auto-corrects to amazon

		// Test "ttd" -> "the-trade-desk" case (ambiguous - should show multiple options)
		{"ttd ambiguous", "ttd", false, "", true}, // should suggest multiple options including the-trade-desk

		// Test stock ticker behavior
		{"amd exact", "amd", true, "amd", false},            // exact match
		{"AMD uppercase", "AMD", true, "amd", false},        // case insensitive exact match
		{"meta uppercase", "META", true, "facebook", false}, // alias match
		{"fb lowercase", "fb", true, "facebook", false},     // alias match
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			company, found, suggestions := findCompanyWithSuggestion(tt.input, problemsData)

			if found != tt.expectedFound {
				t.Errorf("findCompanyWithSuggestion(%q) found = %v, want %v", tt.input, found, tt.expectedFound)
			}

			if found && company != tt.expectedCompany {
				t.Errorf("findCompanyWithSuggestion(%q) company = %q, want %q", tt.input, company, tt.expectedCompany)
			}

			if tt.expectSuggestions && len(suggestions) == 0 {
				t.Errorf("findCompanyWithSuggestion(%q) expected suggestions but got none", tt.input)
			}

			if !tt.expectSuggestions && len(suggestions) > 0 {
				t.Errorf("findCompanyWithSuggestion(%q) got suggestions %v but expected none", tt.input, suggestions)
			}
		})
	}
}

// Test command fuzzy matching
func TestFindCommandWithSuggestion(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedValid    bool
		expectedCmd      string
		expectSuggestion bool
	}{
		// Valid commands
		{"exact problems", "problems", true, "problems", false},
		{"exact help", "help", true, "help", false},

		// Close typos - should suggest
		{"proces typo", "proces", false, "", true},
		{"problms typo", "problms", false, "", true},
		{"hel typo", "hel", false, "", true},
		{"proces typo 2", "proces", false, "", true},

		// Medium typos - should suggest
		{"probl typo", "probl", false, "", true},
		{"he typo", "he", false, "", true},

		// Too different - no suggestion
		{"completely different", "xyz", false, "", false},
		{"random", "random", false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, isValid, suggestion := findCommandWithSuggestion(tt.input)

			if isValid != tt.expectedValid {
				t.Errorf("findCommandWithSuggestion(%q) isValid = %v, want %v", tt.input, isValid, tt.expectedValid)
			}

			if isValid && cmd != tt.expectedCmd {
				t.Errorf("findCommandWithSuggestion(%q) cmd = %q, want %q", tt.input, cmd, tt.expectedCmd)
			}

			if tt.expectSuggestion && suggestion == "" {
				t.Errorf("findCommandWithSuggestion(%q) expected suggestion but got empty", tt.input)
			}

			if !tt.expectSuggestion && suggestion != "" {
				t.Errorf("findCommandWithSuggestion(%q) got suggestion %q but expected none", tt.input, suggestion)
			}
		})
	}
}

// Test Levenshtein distance calculation
func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"a", "a", 0},
		{"a", "b", 1},
		{"ab", "ba", 2},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"google", "googl", 1},
		{"amazon", "amazn", 1},
		{"microsoft", "microsft", 1},
		{"facebook", "facebok", 1},
		{"dropbox", "drop", 3},
		{"box", "dropbox", 4},
		{"jane street", "jane-street", 1},
		{"jump trading", "jump-trading", 1},
	}

	for _, tt := range tests {
		result := levenshteinDistance(tt.s1, tt.s2)
		if result != tt.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, result, tt.expected)
		}
	}
}

// Test confidence calculation
func TestCalculateMatchConfidence(t *testing.T) {
	tests := []struct {
		input    string
		target   string
		expected float64
	}{
		{"", "", 1.0},
		{"a", "a", 1.0},
		{"google", "google", 1.0},
		{"googl", "google", 0.833333},       // 1 char difference out of 6 = 5/6
		{"goog", "google", 0.666667},        // 2 char difference out of 6 = 4/6
		{"goo", "google", 0.5},              // 3 char difference out of 6 = 3/6
		{"xyz", "google", 0.0},              // completely different
		{"amazn", "amazon", 0.833333},       // 1 char difference out of 6 = 5/6
		{"microsft", "microsoft", 0.888889}, // 1 char difference out of 9 = 8/9
		{"ttd", "the-trade-desk", 0.214286}, // 3/14 = 0.214286 (low confidence)
		{"ttd", "td", 0.666667},             // 1/3 = 0.666667 (medium confidence)
		{"ttd", "amd", 0.333333},            // 2/3 = 0.333333 (low-medium confidence)
	}

	for _, tt := range tests {
		result := calculateMatchConfidence(tt.input, tt.target)
		// use approximate equality for floating point comparison
		if result < tt.expected-0.000001 || result > tt.expected+0.000001 {
			t.Errorf("calculateMatchConfidence(%q, %q) = %f, want %f", tt.input, tt.target, result, tt.expected)
		}
	}
}

// Test company aliases
func TestGetCompanyAlias(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		found    bool
	}{
		{"meta", "facebook", true},
		{"Meta", "facebook", true},
		{"META", "facebook", true},
		{"fb", "facebook", true},
		{"FB", "facebook", true},
		{"alphabet", "google", true},
		{"amzn", "amazon", true},
		{"msft", "microsoft", true},
		{"aapl", "apple", true},
		{"nflx", "netflix", true},
		{"google", "", false},
		{"amazon", "", false},
		{"unknown", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		alias, found := getCompanyAlias(tt.input)
		if found != tt.found {
			t.Errorf("getCompanyAlias(%q) found = %v, want %v", tt.input, found, tt.found)
		}
		if found && alias != tt.expected {
			t.Errorf("getCompanyAlias(%q) = %q, want %q", tt.input, alias, tt.expected)
		}
	}
}

// Test ambiguous company matching (dropbox vs box)
func TestAmbiguousCompanyMatching(t *testing.T) {
	testData := map[string]map[string][]data.Problem{
		"dropbox": {
			"all": []data.Problem{{ID: 1, Title: "Dropbox Problem", Difficulty: "Medium", Frequency: 100.0}},
		},
		"box": {
			"all": []data.Problem{{ID: 2, Title: "Box Problem", Difficulty: "Hard", Frequency: 90.0}},
		},
		"drop": {
			"all": []data.Problem{{ID: 3, Title: "Drop Problem", Difficulty: "Easy", Frequency: 80.0}},
		},
	}
	problemsData := data.NewTestProblemsByCompany(testData)

	tests := []struct {
		name              string
		input             string
		expectedFound     bool
		expectedCompany   string
		expectSuggestions bool
	}{
		{"exact dropbox", "dropbox", true, "dropbox", false},
		{"exact box", "box", true, "box", false},
		{"exact drop", "drop", true, "drop", false},

		// Ambiguous cases - should suggest multiple options
		{"ambiguous drop", "drop", true, "drop", false},                // exact match to "drop"
		{"ambiguous box partial", "bo", true, "box", false},            // auto-corrects to "box"
		{"ambiguous dropbox partial", "dropb", true, "dropbox", false}, // auto-corrects to "dropbox"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			company, found, suggestions := findCompanyWithSuggestion(tt.input, problemsData)

			if found != tt.expectedFound {
				t.Errorf("findCompanyWithSuggestion(%q) found = %v, want %v", tt.input, found, tt.expectedFound)
			}

			if found && company != tt.expectedCompany {
				t.Errorf("findCompanyWithSuggestion(%q) company = %q, want %q", tt.input, company, tt.expectedCompany)
			}

			if tt.expectSuggestions && len(suggestions) == 0 {
				t.Errorf("findCompanyWithSuggestion(%q) expected suggestions but got none", tt.input)
			}

			if !tt.expectSuggestions && len(suggestions) > 0 {
				t.Errorf("findCompanyWithSuggestion(%q) got suggestions %v but expected none", tt.input, suggestions)
			}
		})
	}
}

// Test parsing logic for problems command with trailing text
func TestProblemsCommandParsing(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	tests := []struct {
		name              string
		input             string
		expectedCompany   string
		expectedTimeframe string
	}{
		// Basic cases
		{"basic case", "google", "google", ""},
		{"with timeframe", "amazon 30d", "amazon", "30d"},
		{"multi-word company", "jump trading", "jump trading", ""},

		// Cases with trailing text
		{"trailing text", "google (what are the best problems", "google", ""},
		{"trailing text with timeframe", "amazon 30d (show me the hardest", "amazon", "30d"},
		{"multi-word with timeframe", "jump trading 3mo (any system design", "jump trading", "3mo"},
		{"trailing text only", "facebook (show me everything", "facebook", ""},

		// Cases with job-related words (should be cleaned, but timeframes preserved if present)
		{"new grad", "pure storage new grad swe", "pure storage", ""},
		{"software engineer", "google software engineer", "google", ""},
		{"internship", "facebook internship", "facebook", ""},
		{"senior role", "apple senior software engineer", "apple", ""},

		// Edge cases
		{"timeframe in middle", "google 30d extra stuff", "google", "30d"},
		{"multiple timeframes", "google 30d 3mo extra", "google 30d", "3mo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := strings.Fields(tt.input)

			// Use the same parsing logic as the actual handler
			var companyInput, timeframeArg string
			var timeframeIndex = -1
			maxSearchDistance := 4
			if len(args) < maxSearchDistance {
				maxSearchDistance = len(args)
			}

			for i := 1; i <= maxSearchDistance && i <= len(args); i++ {
				candidateTimeframe := strings.ToLower(args[len(args)-i])
				if handler.isTimeframeKeyword(candidateTimeframe) {
					timeframeIndex = len(args) - i
					timeframeArg = candidateTimeframe
					break
				}
			}

			if timeframeIndex != -1 {
				companyInput = strings.Join(args[:timeframeIndex], " ")
			} else {
				companyInput = strings.Join(args, " ")
			}

			// Test the cleaned company input (this is what the actual handler uses)
			cleanedCompanyInput := cleanCompanyInput(companyInput)
			if cleanedCompanyInput != tt.expectedCompany {
				t.Errorf("Expected cleaned company %q, got %q (from raw: %q)", tt.expectedCompany, cleanedCompanyInput, companyInput)
			}

			if timeframeArg != tt.expectedTimeframe {
				t.Errorf("Expected timeframe %q, got %q", tt.expectedTimeframe, timeframeArg)
			}
		})
	}
}
