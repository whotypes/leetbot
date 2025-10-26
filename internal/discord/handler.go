package discord

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lithammer/fuzzysearch/fuzzy"
	paginator "github.com/topi314/dgo-paginator"
	"github.com/whotypes/leetbot/internal/data"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	s1Lower := strings.ToLower(s1)
	s2Lower := strings.ToLower(s2)

	if s1Lower == s2Lower {
		return 0
	}

	if len(s1Lower) == 0 {
		return len(s2Lower)
	}
	if len(s2Lower) == 0 {
		return len(s1Lower)
	}

	// create a 2D array for dynamic programming
	matrix := make([][]int, len(s1Lower)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2Lower)+1)
	}

	// initialize first column and row
	for i := 0; i <= len(s1Lower); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2Lower); j++ {
		matrix[0][j] = j
	}

	// fill in the rest of the matrix
	for i := 1; i <= len(s1Lower); i++ {
		for j := 1; j <= len(s2Lower); j++ {
			cost := 0
			if s1Lower[i-1] != s2Lower[j-1] {
				cost = 1
			}

			deletion := matrix[i-1][j] + 1
			insertion := matrix[i][j-1] + 1
			substitution := matrix[i-1][j-1] + cost

			min := deletion
			if insertion < min {
				min = insertion
			}
			if substitution < min {
				min = substitution
			}
			matrix[i][j] = min
		}
	}

	return matrix[len(s1Lower)][len(s2Lower)]
}

// calculateMatchConfidence returns a confidence score between 0 and 1
func calculateMatchConfidence(input, target string) float64 {
	if strings.EqualFold(input, target) {
		return 1.0
	}

	distance := levenshteinDistance(input, target)
	maxLen := len(input)
	if len(target) > maxLen {
		maxLen = len(target)
	}

	if maxLen == 0 {
		return 0.0
	}

	// confidence is inverse of normalized distance
	confidence := 1.0 - float64(distance)/float64(maxLen)
	if confidence < 0 {
		confidence = 0
	}

	return confidence
}

// companyAliases maps alternative names to canonical company slugs
var companyAliases = map[string]string{
	"meta":     "facebook",
	"fb":       "facebook",
	"alphabet": "google",
	"amzn":     "amazon",
	"msft":     "microsoft",
	"aapl":     "apple",
	"nflx":     "netflix",
}

// getCompanyAlias checks if the input matches a known alias
func getCompanyAlias(input string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	if alias, ok := companyAliases[normalized]; ok {
		return alias, true
	}
	return "", false
}

// SlashCommandHandlers maps command names to their handler methods
// we use this to dispatch slash commands to the appropriate handler
var SlashCommandHandlers = map[string]string{
	"problems": "problems",
	"help":     "help",
	"process":  "process",
	"stats":    "stats",
}

func HandleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, problemsData *data.ProblemsByCompany) {
	data := i.ApplicationCommandData()

	// handle autocomplete for commands that use company autocomplete
	if data.Name != "problems" && data.Name != "process" && data.Name != "stats" {
		return
	}

	var choices []*discordgo.ApplicationCommandOptionChoice
	var currentInput string

	for _, option := range data.Options {
		if option.Name == "company" && option.Focused {
			currentInput = option.StringValue()
			choices = getCompanyAutocompleteChoices(currentInput, problemsData)
			break
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		fmt.Printf("Error responding to autocomplete: %v\n", err)
	}
}

func getStageChoices() []*discordgo.ApplicationCommandOptionChoice {
	stages := data.GetAvailableStages()
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(stages))

	for i, stage := range stages {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  string(stage),
			Value: string(stage),
		}
	}

	return choices
}

func GetSlashCommands(problemsData *data.ProblemsByCompany) []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "problems",
			Description: "Show popular coding interview problems by company",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "company",
					Description:  "Company name (start typing to search)",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "timeframe",
					Description: "Time period (optional)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "All Time",
							Value: "all",
						},
						{
							Name:  "Last 30 Days",
							Value: "thirty-days",
						},
						{
							Name:  "Last 3 Months",
							Value: "three-months",
						},
						{
							Name:  "Last 6 Months",
							Value: "six-months",
						},
						{
							Name:  "More than 6 Months",
							Value: "more-than-six-months",
						},
					},
				},
			},
		},
		{
			Name:        "help",
			Description: "Show available Leetbot commands and usage",
		},
		{
			Name:        "process",
			Description: "Track a new interview process",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "company",
					Description:  "Company you're interviewing with (start typing to search)",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "stage",
					Description: "Interview stage",
					Required:    true,
					Choices:     getStageChoices(),
				},
			},
		},
		{
			Name:        "stats",
			Description: "View interview process statistics",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "company",
					Description:  "Company name (start typing to search)",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
	}
}

func formatCompanyName(company string) string {
	words := strings.Split(company, "-")
	caser := cases.Title(language.English)
	for i, word := range words {
		words[i] = caser.String(word)
	}
	return strings.Join(words, " ")
}

func getDifficultyIndicator(difficulty string) string {
	switch strings.ToLower(difficulty) {
	case "easy":
		return "🟢"
	case "medium":
		return "🟡"
	case "hard":
		return "🔴"
	default:
		return ""
	}
}

func findCompanyByFuzzySearch(input string, problemsData *data.ProblemsByCompany) (string, bool) {
	if input == "" {
		return "", false
	}

	companies := problemsData.GetAvailableCompanies()
	if len(companies) == 0 {
		return "", false
	}

	// check for aliases first (e.g., Meta -> Facebook)
	if alias, ok := getCompanyAlias(input); ok {
		// verify the alias exists in our company list
		for _, company := range companies {
			if company == alias {
				return alias, true
			}
		}
	}

	normalizedInput := strings.ToLower(input)
	normalizedInput = strings.ReplaceAll(normalizedInput, " ", "-")

	// exact match
	for _, company := range companies {
		if company == normalizedInput {
			return company, true
		}
	}

	// contains match
	for _, company := range companies {
		if strings.Contains(company, normalizedInput) {
			return company, true
		}
	}

	// fuzzy match with confidence scoring
	bestMatch, confidence := findBestCompanyMatch(input, companies)
	if confidence > 0.7 { // high confidence threshold for auto-match
		return bestMatch, true
	}

	return "", false
}

// findBestCompanyMatch finds the best matching company with confidence score
func findBestCompanyMatch(input string, companies []string) (string, float64) {
	if len(companies) == 0 {
		return "", 0.0
	}

	normalizedInput := strings.ToLower(input)
	normalizedInput = strings.ReplaceAll(normalizedInput, " ", "-")

	var bestMatch string
	var bestConfidence float64

	// check against company slugs
	for _, company := range companies {
		confidence := calculateMatchConfidence(normalizedInput, company)
		if confidence > bestConfidence {
			bestConfidence = confidence
			bestMatch = company
		}
	}

	// also check against display names
	for _, company := range companies {
		displayName := formatCompanyName(company)
		displayNameNormalized := strings.ToLower(strings.ReplaceAll(displayName, " ", "-"))
		confidence := calculateMatchConfidence(normalizedInput, displayNameNormalized)
		if confidence > bestConfidence {
			bestConfidence = confidence
			bestMatch = company
		}
	}

	return bestMatch, bestConfidence
}

// findCompanyWithSuggestion attempts to find a company by fuzzy search and returns suggestions if not found
// uses confidence thresholds:
// - confidence > 0.8 or distance <= 2: auto-correct
// - confidence 0.6-0.8: suggest with "Did you mean?"
// - confidence < 0.6: show error with top suggestions
func findCompanyWithSuggestion(input string, problemsData *data.ProblemsByCompany) (company string, found bool, suggestions []string) {
	// first try the standard fuzzy search (handles exact matches, aliases, etc.)
	company, found = findCompanyByFuzzySearch(input, problemsData)
	if found {
		return company, true, nil
	}

	// get all available companies
	companies := problemsData.GetAvailableCompanies()
	if len(companies) == 0 {
		return "", false, nil
	}

	// find best matches with confidence scores
	type scoredMatch struct {
		company    string
		confidence float64
		distance   int
	}

	var matches []scoredMatch
	normalizedInput := strings.ToLower(input)
	normalizedInput = strings.ReplaceAll(normalizedInput, " ", "-")

	for _, c := range companies {
		// check slug
		slugConfidence := calculateMatchConfidence(normalizedInput, c)
		slugDistance := levenshteinDistance(normalizedInput, c)

		// check display name
		displayName := formatCompanyName(c)
		displayNameNormalized := strings.ToLower(strings.ReplaceAll(displayName, " ", "-"))
		displayConfidence := calculateMatchConfidence(normalizedInput, displayNameNormalized)
		displayDistance := levenshteinDistance(normalizedInput, displayNameNormalized)

		// use the better of the two
		confidence := slugConfidence
		distance := slugDistance
		if displayConfidence > confidence {
			confidence = displayConfidence
			distance = displayDistance
		}

		matches = append(matches, scoredMatch{
			company:    c,
			confidence: confidence,
			distance:   distance,
		})
	}

	// sort matches by confidence (descending) and distance (ascending)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			// prioritize higher confidence, then lower distance
			if matches[j].confidence > matches[i].confidence ||
				(matches[j].confidence == matches[i].confidence && matches[j].distance < matches[i].distance) {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// check best match against thresholds
	if len(matches) > 0 {
		bestMatch := matches[0]

		// check if this is an ambiguous case (multiple matches with similar confidence)
		// or if input looks like a stock ticker/abbreviation (3-4 characters, all caps or mixed)
		ambiguousThreshold := 0.2 // matches within 20% confidence are considered ambiguous
		ambiguousMatches := 1 // count the best match
		isLikelyTicker := len(input) <= 5 && len(input) >= 2 && strings.ContainsAny(strings.ToUpper(input), "ABCDEFGHIJKLMNOPQRSTUVWXYZ")

		// count how many matches are within ambiguous threshold
		for i := 1; i < len(matches); i++ {
			confidenceDiff := bestMatch.confidence - matches[i].confidence
			if confidenceDiff <= ambiguousThreshold {
				ambiguousMatches++
				if ambiguousMatches >= 3 { // we have enough for ambiguity
					break
				}
			} else {
				break // confidence gap is too large
			}
		}

		// if ambiguous and best match confidence is reasonable, show multiple options
		// also show multiple options for likely stock tickers even if confidence gap is larger
		if (ambiguousMatches >= 2 && bestMatch.confidence >= 0.3) ||
			(isLikelyTicker && ambiguousMatches >= 2 && bestMatch.confidence >= 0.2) {
			maxSuggestions := 3
			for i := 0; i < len(matches) && i < maxSuggestions; i++ {
				if isLikelyTicker {
					// for tickers, include matches with reasonable confidence (not just ambiguous threshold)
					if matches[i].confidence >= 0.2 {
						suggestions = append(suggestions, matches[i].company)
					}
				} else {
					confidenceDiff := bestMatch.confidence - matches[i].confidence
					if confidenceDiff <= ambiguousThreshold {
						suggestions = append(suggestions, matches[i].company)
					}
				}
			}
			return "", false, suggestions
		}

		// high confidence (>0.8) or very close (distance <= 2): auto-correct
		// but don't auto-correct if input looks like a ticker unless exact match or very high confidence
		if (bestMatch.confidence > 0.8 || (bestMatch.distance <= 2 && !isLikelyTicker)) ||
			(bestMatch.distance == 0) || // exact matches always auto-correct
			(isLikelyTicker && bestMatch.confidence > 0.9) { // very high confidence tickers
			return bestMatch.company, true, nil
		}

		// medium confidence (0.6-0.8): suggest options
		if bestMatch.confidence >= 0.6 {
			// for likely tickers, always show multiple options if there are other reasonable matches
			if isLikelyTicker {
				maxSuggestions := 3
				for i := 0; i < len(matches) && i < maxSuggestions; i++ {
					if matches[i].confidence >= 0.2 {
						suggestions = append(suggestions, matches[i].company)
					}
				}
				return "", false, suggestions
			}

			// otherwise suggest single option with a couple more if reasonable
			suggestions = append(suggestions, bestMatch.company)
			for i := 1; i < len(matches) && len(suggestions) < 3; i++ {
				if matches[i].confidence >= 0.5 {
					suggestions = append(suggestions, matches[i].company)
				}
			}
			return "", false, suggestions
		}

		// low confidence: provide top 3 suggestions
		maxSuggestions := 3
		for i := 0; i < len(matches) && i < maxSuggestions; i++ {
			suggestions = append(suggestions, matches[i].company)
		}
	}

	return "", false, suggestions
}

// validCommands lists all valid Leetbot commands
var validCommands = []string{"problems", "help", "process", "stats", "shutdown", "startup"}

// findCommandWithSuggestion attempts to match a command and returns suggestions if it's a typo
// returns: (correctCommand, isValidCommand, didYouMeanSuggestion)
func findCommandWithSuggestion(input string) (string, bool, string) {
	input = strings.ToLower(strings.TrimSpace(input))

	// check if it's a valid command
	for _, cmd := range validCommands {
		if input == cmd {
			return input, true, ""
		}
	}

	// not a valid command, check for typos
	var bestMatch string
	var bestConfidence float64
	var bestDistance int

	for _, cmd := range validCommands {
		confidence := calculateMatchConfidence(input, cmd)
		distance := levenshteinDistance(input, cmd)

		if confidence > bestConfidence || (confidence == bestConfidence && distance < bestDistance) {
			bestConfidence = confidence
			bestDistance = distance
			bestMatch = cmd
		}
	}

	// if very close match (distance <= 2 or confidence > 0.6), suggest it
	if bestDistance <= 2 || bestConfidence > 0.6 {
		return "", false, bestMatch
	}

	return "", false, ""
}

// normalizeStage converts user input to canonical stage format
func normalizeStage(input string) (string, bool) {
	stageCanonical := map[string]string{
		"apply":  "Apply",
		"reject": "Reject",
		"oa":     "OA",
		"phone":  "Phone",
		"onsite": "Onsite",
		"offer":  "Offer",
	}
	lower := strings.ToLower(strings.TrimSpace(input))
	if canonical, ok := stageCanonical[lower]; ok {
		return canonical, true
	}
	return "", false
}

// findStageWithSuggestion attempts to normalize stage and returns suggestions if invalid
func findStageWithSuggestion(input string) (stage string, found bool, suggestions []string) {
	stage, found = normalizeStage(input)
	if found {
		return stage, true, nil
	}

	// we didn't find a valid stage, let's suggest similar ones
	validStages := []string{"Apply", "Reject", "OA", "Phone", "Onsite", "Offer"}
	matches := fuzzy.RankFindNormalizedFold(input, validStages)

	// collect top 3 suggestions
	maxSuggestions := 3
	for i := 0; i < len(matches) && i < maxSuggestions; i++ {
		suggestions = append(suggestions, matches[i].Target)
	}

	return "", false, suggestions
}

func getCompanyAutocompleteChoices(input string, problemsData *data.ProblemsByCompany) []*discordgo.ApplicationCommandOptionChoice {
	companies := problemsData.GetAvailableCompanies()
	var choices []*discordgo.ApplicationCommandOptionChoice

	if input == "" {
		popularCompanies := []string{
			"amazon", "google", "facebook", "microsoft", "apple", "netflix",
			"uber", "meta", "tesla", "nvidia", "openai", "anthropic",
		}
		for _, company := range popularCompanies {
			if problemsData.CompanyExists(company) {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  formatCompanyName(company),
					Value: company,
				})
			}
		}

		companiesSorted := make([]string, len(companies))
		copy(companiesSorted, companies)
		for i := 0; i < len(companiesSorted)-1; i++ {
			for j := i + 1; j < len(companiesSorted); j++ {
				if companiesSorted[i] > companiesSorted[j] {
					companiesSorted[i], companiesSorted[j] = companiesSorted[j], companiesSorted[i]
				}
			}
		}

		for _, company := range companiesSorted {
			isPopular := false
			for _, pop := range popularCompanies {
				if company == pop {
					isPopular = true
					break
				}
			}
			if !isPopular {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  formatCompanyName(company),
					Value: company,
				})
				if len(choices) >= 25 {
					break
				}
			}
		}
		return choices
	}

	normalizedInput := strings.ToLower(input)
	normalizedInput = strings.ReplaceAll(normalizedInput, " ", "-")
	type companyMatch struct {
		slug        string
		displayName string
	}
	var companyList []companyMatch
	for _, company := range companies {
		companyList = append(companyList, companyMatch{
			slug:        company,
			displayName: formatCompanyName(company),
		})
	}

	var displayNames []string
	for _, cm := range companyList {
		displayNames = append(displayNames, cm.displayName)
	}
	matches := fuzzy.RankFindNormalizedFold(input, displayNames)

	for i, match := range matches {
		if i >= 25 {
			break
		}
		if match.OriginalIndex >= 0 && match.OriginalIndex < len(companyList) {
			cm := companyList[match.OriginalIndex]
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  cm.displayName,
				Value: cm.slug,
			})
		}
	}

	if len(choices) == 0 {
		slugMatches := fuzzy.RankFindNormalizedFold(normalizedInput, companies)
		for i, match := range slugMatches {
			if i >= 25 {
				break
			}
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  formatCompanyName(match.Target),
				Value: match.Target,
			})
		}
	}

	return choices
}

type RestartRequest struct {
	ChannelID string
	Success   bool
	Message   string
}

type Handler struct {
	problemsData   *data.ProblemsByCompany
	processStorage data.Storage
	prefix         string
	reconnectChan  chan RestartRequest
	disabled       bool
	session        *discordgo.Session
	sessionMutex   sync.RWMutex
}

func NewHandler(problemsData *data.ProblemsByCompany, prefix string) *Handler {
	return &Handler{
		problemsData: problemsData,
		prefix:       prefix,
	}
}

func (h *Handler) SetReconnectChannel(ch chan RestartRequest) {
	h.reconnectChan = ch
}

func (h *Handler) SetProcessStorage(storage data.Storage) {
	h.processStorage = storage
}

func (h *Handler) SetSession(session *discordgo.Session) {
	h.sessionMutex.Lock()
	defer h.sessionMutex.Unlock()
	h.session = session
}

func (h *Handler) GetSession() *discordgo.Session {
	h.sessionMutex.RLock()
	defer h.sessionMutex.RUnlock()
	return h.session
}

// HandleSlashCommand routes slash commands to appropriate handlers
func (h *Handler) HandleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName := i.ApplicationCommandData().Name

	// if bot is disabled, only allow help command
	if h.disabled && commandName != "help" {
		// silently ignore all other slash commands when disabled
		return
	}

	switch commandName {
	case "problems":
		h.handleProblemsSlash(s, i)
	case "help":
		h.handleHelpSlash(s, i)
	case "process":
		h.handleProcessSlash(s, i)
	case "stats":
		h.handleStatsSlash(s, i)
	default:
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Unknown command: %s", commandName),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
	}
}

func (h *Handler) handleProcessMessageCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
    if len(args) < 2 {
        h.sendErrorMessage(s, m.ChannelID, "Please specify a company and stage. Usage: !process <company> <Apply|Reject|OA|Phone|Onsite|Offer>")
        return
    }

    if h.processStorage == nil {
        h.sendErrorMessage(s, m.ChannelID, "Process tracking is not configured on Leetbot.")
        return
    }

    // the stage is always the last argument, everything else is the company name
    // this handles multi-word companies like "jump trading" or "jane street"
    stageInput := args[len(args)-1]
    companyInput := strings.Join(args[:len(args)-1], " ")

    // validate and fuzzy match company
    company, companyFound, companySuggestions := findCompanyWithSuggestion(companyInput, h.problemsData)
    if !companyFound {
        var errorMsg strings.Builder
        errorMsg.WriteString(fmt.Sprintf("Could not find company matching '%s'.", companyInput))
        if len(companySuggestions) > 0 {
            errorMsg.WriteString("\n\nDid you mean:")
            for _, suggestion := range companySuggestions {
                errorMsg.WriteString(fmt.Sprintf("\n• %s", formatCompanyName(suggestion)))
            }
        }
        h.sendErrorMessage(s, m.ChannelID, errorMsg.String())
        return
    }

    // validate and normalize stage
    stage, stageFound, stageSuggestions := findStageWithSuggestion(stageInput)
    if !stageFound {
        var errorMsg strings.Builder
        errorMsg.WriteString(fmt.Sprintf("Invalid stage '%s'.", stageInput))
        if len(stageSuggestions) > 0 {
            errorMsg.WriteString("\n\nDid you mean:")
            for _, suggestion := range stageSuggestions {
                errorMsg.WriteString(fmt.Sprintf("\n• %s", suggestion))
            }
        } else {
            errorMsg.WriteString("\n\nValid stages: Apply, Reject, OA, Phone, Onsite, Offer")
        }
        h.sendErrorMessage(s, m.ChannelID, errorMsg.String())
        return
    }

    process := data.Process{
        Company:   company,
        Stage:     stage,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

	ctx := context.Background()
	_, err := h.processStorage.AddProcess(ctx, process)
	if err != nil {
		h.sendErrorMessage(s, m.ChannelID, fmt.Errorf("failed to add process: %w", err).Error())
		return
	}

	// add checkmark reaction to user's message
	err = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
	if err != nil {
		fmt.Printf("Failed to add reaction: %v\n", err)
	}

	// if it's an offer, send congrats message
	if stage == "Offer" {
		h.sendMessage(s, m.ChannelID, "Congrats! 🎉")
	}

	// for text commands, we just react with checkmark to avoid clutter
	// slash commands will show stats as ephemeral embeds
}

func (h *Handler) handleStatsCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		h.sendErrorMessage(s, m.ChannelID, "Please specify a company. Usage: !stats <company>")
		return
	}

    if h.processStorage == nil {
        h.sendErrorMessage(s, m.ChannelID, "Process tracking is not configured on Leetbot.")
        return
    }

	// join all args to support multi-word company names like "jump trading"
	companyInput := strings.Join(args, " ")

	// validate and fuzzy match company
	company, companyFound, companySuggestions := findCompanyWithSuggestion(companyInput, h.problemsData)
	if !companyFound {
		var errorMsg strings.Builder
		errorMsg.WriteString(fmt.Sprintf("Could not find company matching '%s'.", companyInput))
		if len(companySuggestions) > 0 {
			errorMsg.WriteString("\n\nDid you mean:")
			for _, suggestion := range companySuggestions {
				errorMsg.WriteString(fmt.Sprintf("\n• %s", formatCompanyName(suggestion)))
			}
		}
		h.sendErrorMessage(s, m.ChannelID, errorMsg.String())
		return
	}

	h.sendProcessStats(s, m.ChannelID, company, false)
}

// sendProcessStats sends stats for a specific company across all stages
func (h *Handler) sendProcessStats(s *discordgo.Session, channelID, company string, isEmbed bool) {
	ctx := context.Background()

	// get all processes for this company
	allProcesses, err := h.processStorage.GetProcessesByCompany(ctx, company)
	if err != nil {
		h.sendErrorMessage(s, channelID, fmt.Sprintf("Failed to retrieve stats: %v", err))
		return
	}

	// build stats content
	content := h.buildCompanyStatsContent(company, channelID, allProcesses)

	if isEmbed {
		// send as embed
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Stats for %s", formatCompanyName(company)),
			Description: content,
			Color:       0x5865F2, // discord blurple
		}
		_, err := s.ChannelMessageSendEmbed(channelID, embed)
		if err != nil {
			fmt.Printf("Error sending embed: %v\n", err)
		}
	} else {
		// send as regular message with company name header
		header := fmt.Sprintf("**Stats for %s**\n", formatCompanyName(company))
		h.sendMessage(s, channelID, header+content)
	}
}

// sendProcessStatsForInteraction sends stats as a response to an interaction
func (h *Handler) sendProcessStatsForInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, company string) {
	ctx := context.Background()

	// we collect all data first, then respond once
	// this avoids the double acknowledgment error
	var responseContent string
	var responseEmbed *discordgo.MessageEmbed

	// get all processes for this company
	allProcesses, err := h.processStorage.GetProcessesByCompany(ctx, company)
	if err != nil {
		responseContent = fmt.Sprintf("Failed to retrieve stats: %v", err)
	} else {
		// build stats content
		content := h.buildCompanyStatsContent(company, i.ChannelID, allProcesses)

		// send as embed
		responseEmbed = &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Stats for %s", formatCompanyName(company)),
			Description: content,
			Color:       0x5865F2, // discord blurple
		}
	}

	// single response path - either error or success
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}

	if responseEmbed != nil {
		response.Data.Embeds = []*discordgo.MessageEmbed{responseEmbed}
	} else {
		response.Data.Content = responseContent
	}

	err = s.InteractionRespond(i.Interaction, response)
	if err != nil {
		fmt.Printf("Error responding to interaction: %v\n", err)
	}
}

// sendProcessStatsForInteractionWithSuccessHeader sends stats as a response to an interaction with success header
func (h *Handler) sendProcessStatsForInteractionWithSuccessHeader(s *discordgo.Session, i *discordgo.InteractionCreate, company string) {
	ctx := context.Background()

	// we collect all data first, then respond once
	// this avoids the double acknowledgment error
	var responseContent string
	var responseEmbed *discordgo.MessageEmbed

	// get all processes for this company
	allProcesses, err := h.processStorage.GetProcessesByCompany(ctx, company)
	if err != nil {
		responseContent = fmt.Sprintf("Failed to retrieve stats: %v", err)
	} else {
		// build stats content
		content := h.buildCompanyStatsContent(company, i.ChannelID, allProcesses)

		// send as embed with success header
		responseEmbed = &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Process Successfully Logged for %s", formatCompanyName(company)),
			Description: content,
			Color:       0x00FF00, // green color for success
		}
	}

	// single response path - either error or success
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}

	if responseEmbed != nil {
		response.Data.Embeds = []*discordgo.MessageEmbed{responseEmbed}
	} else {
		response.Data.Content = responseContent
	}

	err = s.InteractionRespond(i.Interaction, response)
	if err != nil {
		fmt.Printf("Error responding to interaction: %v\n", err)
	}
}


// buildCompanyStatsContent builds the stats content for a company across all stages
func (h *Handler) buildCompanyStatsContent(company, channelID string, processes []data.Process) string {
	var content strings.Builder

	// count processes by stage
	stageCounts := make(map[string]int)
	for _, p := range processes {
		stageCounts[p.Stage]++
	}

	// format the output according to the specified format with inline monospace
	content.WriteString(fmt.Sprintf("(<#%s>)\n", channelID))
	content.WriteString("`-----------------`\n")
	content.WriteString("`Stage       Count`\n")
	content.WriteString("`-----------------`\n")

	// define stage order and display names
	stageOrder := []string{"Apply", "OA", "Phone", "Onsite", "Offer"}
	stageDisplayNames := map[string]string{
		"Apply":  "Apply",
		"OA":     "OA",
		"Phone":  "Phone",
		"Onsite": "Final", // map Onsite to Final as per the format
		"Offer":  "Offer",
	}

	for _, stage := range stageOrder {
		count := stageCounts[stage]
		displayName := stageDisplayNames[stage]
		content.WriteString(fmt.Sprintf("`%-10s    %3d`\n", displayName, count))
	}

	return content.String()
}

func (h *Handler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Update the handler's session reference
	h.SetSession(s)

	if m.Author.Bot {
		return
	}

	if !strings.HasPrefix(m.Content, h.prefix) {
		return
	}
	content := strings.TrimPrefix(m.Content, h.prefix)
	content = strings.TrimSpace(content)

	if content == "" {
		return
	}

	parts := strings.Fields(content)
	command := strings.ToLower(parts[0])
	args := parts[1:]

	// check if command is valid or a typo
	correctCommand, isValid, suggestion := findCommandWithSuggestion(command)

	if !isValid {
		// command is not valid
		if suggestion != "" {
			// we have a suggestion - reconstruct the command with args
			var exampleCommand strings.Builder
			exampleCommand.WriteString(h.prefix)
			exampleCommand.WriteString(suggestion)
			if len(args) > 0 {
				exampleCommand.WriteString(" ")
				exampleCommand.WriteString(strings.Join(args, " "))
			}

			h.sendErrorMessage(s, m.ChannelID,
				fmt.Sprintf("Unknown command '%s%s'. Did you mean `%s`?",
					h.prefix, command, exampleCommand.String()))
		} else {
			h.sendErrorMessage(s, m.ChannelID,
				fmt.Sprintf("Unknown command '%s'. Use `%shelp` for available commands.",
					command, h.prefix))
		}
		return
	}

	// use the validated command
	command = correctCommand

	// check if Leetbot is disabled (but allow shutdown, startup, and help commands)
	if h.disabled {
		// only allow shutdown, startup, and help commands when disabled
		if command != "shutdown" && command != "startup" && command != "help" {
			return // silently ignore all other commands
		}
	}

	switch command {
	case "problems":
		h.handleProblemsCommand(s, m, args)
	case "help":
		h.handleHelpCommand(s, m)
	case "process":
		h.handleProcessMessageCommand(s, m, args)
	case "stats":
		h.handleStatsCommand(s, m, args)
	case "shutdown":
		h.handleShutdownMessage(s, m, args)
	case "startup":
		h.handleStartupMessage(s, m, args)
	default:
		h.sendErrorMessage(s, m.ChannelID, fmt.Sprintf("Unknown command '%s'. Use `%shelp` for available commands.", command, h.prefix))
	}
}

func (h *Handler) handleProblemsCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		h.sendErrorMessage(s, m.ChannelID, "Please specify a company. Usage: !problems <company> [timeframe]")
		return
	}

	companyInput := args[0]
	timeframeArg := ""

	if len(args) >= 2 {
		lastArg := strings.ToLower(args[len(args)-1])
		if h.isTimeframeKeyword(lastArg) {
			timeframeArg = lastArg
			if len(args) > 2 {
				companyInput = strings.Join(args[:len(args)-1], " ")
			}
		} else {
			companyInput = strings.Join(args, " ")
		}
	}

	// use enhanced fuzzy matching with suggestions
	company, companyFound, companySuggestions := findCompanyWithSuggestion(companyInput, h.problemsData)
	if !companyFound {
		var errorMsg strings.Builder
		errorMsg.WriteString(fmt.Sprintf("Could not find company matching '%s'.", companyInput))
		if len(companySuggestions) > 0 {
			errorMsg.WriteString("\n\nDid you mean:")
			for _, suggestion := range companySuggestions {
				errorMsg.WriteString(fmt.Sprintf("\n• %s", formatCompanyName(suggestion)))
			}
		}
		h.sendErrorMessage(s, m.ChannelID, errorMsg.String())
		return
	}

	var problems []data.Problem
	var timeframe string

	if timeframeArg != "" {

		timeframe = h.NormalizeTimeframe(timeframeArg)
		problems = h.problemsData.GetProblems(company, timeframe)
	} else {

		problems, timeframe = h.problemsData.GetProblemsWithPriority(company)
	}

	if problems == nil {

		availableTimeframes := h.problemsData.GetAvailableTimeframes(company)
		if len(availableTimeframes) > 0 && timeframeArg != "" {

			suggestion := h.formatAvailableTimeframesSuggestion(company, timeframe, availableTimeframes)
			h.sendMessage(s, m.ChannelID, suggestion)
		} else {

			h.sendMessage(s, m.ChannelID, fmt.Sprintf("No data found for company '%s'", formatCompanyName(company)))
		}
		return
	}

	if shouldUsePagination(len(problems)) {
		err := sendPaginatedProblemsMessage(s, m.ChannelID, company, timeframe, problems)
		if err != nil {
			fmt.Printf("Error sending paginated message: %v\n", err)

			response := h.formatProblemsResponse(company, timeframe, problems)
			h.sendMessage(s, m.ChannelID, response)
		}
		return
	}

	response := h.formatProblemsResponse(company, timeframe, problems)
	h.sendMessage(s, m.ChannelID, response)
}

func (h *Handler) isTimeframeKeyword(s string) bool {
	s = strings.ToLower(s)
	timeframeKeywords := []string{
		"all", "alltime", "everything",
		"30", "30d", "30days", "thirty", "thirtydays",
		"90", "3mo", "90days", "3months", "three", "threemonths",
		"180", "6mo", "180days", "6months", "six", "sixmonths",
		">6mo", "more-than-six-months",
		"thirty-days", "three-months", "six-months", "more-than-six-months",
	}

	for _, keyword := range timeframeKeywords {
		if s == keyword || strings.Contains(s, keyword) {
			return true
		}
	}
	return false
}

func (h *Handler) NormalizeTimeframe(timeframe string) string {
	timeframe = strings.ToLower(strings.TrimSpace(timeframe))
	timeframe = strings.ReplaceAll(timeframe, " ", "-")
	switch timeframe {
	case "30", "30d", "90d", "30days", "30-days", "thirty", "thirtydays", "thirty-days":
		return "thirty-days"
	case "90", "3mo", "90days", "90-days", "three", "threemonths", "three-months", "3months", "3-months":
		return "three-months"
	case "180", "6mo", "180days", "180-days", "six", "sixmonths", "six-months", "6months", "6-months":
		return "six-months"
	case ">6mo", ">6months", "more-than-six-months", "more-than-6-months", "morethan6months":
		return "more-than-six-months"
	case "all", "alltime", "all-time", "everything", "":
		return "all"
	default:
		for _, tf := range []string{"all", "thirty-days", "three-months", "six-months", "more-than-six-months"} {
			if timeframe == tf {
				return tf
			}
		}
		return "all"
	}
}

func (h *Handler) formatProblemsResponse(company, timeframe string, problems []data.Problem) string {
	if len(problems) == 0 {
		return fmt.Sprintf("No problems found for %s (%s)", formatCompanyName(company), h.formatTimeframeDisplay(timeframe))
	}

	displayTimeframe := h.formatTimeframeDisplay(timeframe)

	title := fmt.Sprintf("Most Popular Problems for %s (%s):", formatCompanyName(company), displayTimeframe)

	var message strings.Builder
	message.WriteString(title + "\n")
	maxProblems := 20
	if len(problems) < maxProblems {
		maxProblems = len(problems)
	}

	for i := 0; i < maxProblems; i++ {
		problem := problems[i]
		difficultyIndicator := getDifficultyIndicator(problem.Difficulty)
		problemLine := fmt.Sprintf("%s %s (%.0f%%): %s\n",
			difficultyIndicator, problem.Title, problem.Frequency, problem.URL)
		message.WriteString(problemLine)
	}

	return message.String()
}

func (h *Handler) formatTimeframeDisplay(timeframe string) string {
	switch timeframe {
	case "all":
		return "all"
	case "thirty-days":
		return "last 30 days"
	case "three-months":
		return "last 3 months"
	case "six-months":
		return "last 6 months"
	case "more-than-six-months":
		return "more than 6 months"
	default:
		return strings.ToLower(strings.ReplaceAll(timeframe, "-", " "))
	}
}



func (h *Handler) formatAvailableTimeframesSuggestion(company, requestedTimeframe string, availableTimeframes []string) string {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("No data found for %s (%s).\n\n",
		formatCompanyName(company),
		h.formatTimeframeDisplay(requestedTimeframe)))

	message.WriteString(fmt.Sprintf("Available timeframes for %s:\n", formatCompanyName(company)))

	priorityOrder := map[string]int{
		"thirty-days":          1,
		"three-months":         2,
		"six-months":           3,
		"more-than-six-months": 4,
		"all":                  5,
	}

	type timeframeWithPriority struct {
		name     string
		priority int
	}

	var sortedTimeframes []timeframeWithPriority
	for _, tf := range availableTimeframes {
		priority := priorityOrder[tf]
		if priority == 0 {
			priority = 999
		}
		sortedTimeframes = append(sortedTimeframes, timeframeWithPriority{name: tf, priority: priority})
	}

	for i := 0; i < len(sortedTimeframes)-1; i++ {
		for j := i + 1; j < len(sortedTimeframes); j++ {
			if sortedTimeframes[i].priority > sortedTimeframes[j].priority {
				sortedTimeframes[i], sortedTimeframes[j] = sortedTimeframes[j], sortedTimeframes[i]
			}
		}
	}

	for _, tf := range sortedTimeframes {

		shortAlias := h.getTimeframeShortAlias(tf.name)
		message.WriteString(fmt.Sprintf("• **%s** (%s)\n", shortAlias, h.formatTimeframeDisplay(tf.name)))
	}

	message.WriteString(fmt.Sprintf("\nTry: `%sproblems %s <timeframe>`", h.prefix, company))

	return message.String()
}

func (h *Handler) getTimeframeShortAlias(timeframe string) string {
	switch timeframe {
	case "thirty-days":
		return "30d"
	case "three-months":
		return "3mo"
	case "six-months":
		return "6mo"
	case "more-than-six-months":
		return ">6mo"
	case "all":
		return "all"
	default:
		return timeframe
	}
}

func (h *Handler) sendMessage(s *discordgo.Session, channelID, message string) {
	session := h.GetSession()
	if session == nil {
		fmt.Printf("[TEST] Would send to %s: %s\n", channelID, message)
		return
	}

	if session.Token == "" {
		fmt.Printf("[TEST] Would send to %s: %s\n", channelID, message)
		return
	}

	// Update session reference to the latest one
	if s != nil && s != session {
		h.SetSession(s)
		session = s
	}

	_, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: message,
		Flags:   discordgo.MessageFlagsSuppressEmbeds,
	})
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}

func (h *Handler) createHelpPaginator(isAdmin bool) *paginator.Paginator {
	return &paginator.Paginator{
		PageFunc: func(page int, embed *discordgo.MessageEmbed) {
			switch page {
			case 0:
				// Page 1: Basic Commands
				embed.Title = "Basic Commands"
				embed.Color = 0x5865F2
				embed.Description = fmt.Sprintf(`**Text Commands (prefix: %s):**
• **%sproblems <company> [timeframe]** - Show interview problems
• **%sprocess <company> <stage>** - Track interview stage
• **%sstats <company>** - View process statistics

**Slash Commands:**
• **/problems** - Show interview problems (with dropdown options)
• **/process** - Track new interview process
• **/stats** - View interview process statistics
• **/help** - Show this help message`, h.prefix, h.prefix, h.prefix, h.prefix)

				if isAdmin {
					embed.Description += fmt.Sprintf(`

**Admin Commands:**
• **%sshutdown [indef]** - Shutdown Leetbot (admin only)
• **%sstartup** - Restart Leetbot or re-enable if disabled (admin only)`, h.prefix, h.prefix)
				}

				embed.Footer = &discordgo.MessageEmbedFooter{
					Text: "Page 1/4 • Use the buttons below to navigate",
				}

			case 1:
				// Page 2: Problems Command Usage
				embed.Title = "Problems Command & Timeframe Options"
				embed.Color = 0x5865F2
				embed.Description = `**Problems Command Usage:**
• *company*: Company name (e.g., airbnb, amazon, google)
• *timeframe*: Optional timeframe filter (if not specified, uses smart priority system)

**Timeframe Options:**
• **all** - All time
• **30d** or **thirty-days** - Last 30 days
• **3mo** or **three-months** - Last 3 months
• **6mo** or **six-months** - Last 6 months
• **>6mo** or **more-than-six-months** - More than 6 months ago

**Smart Priority System:**
When no timeframe is specified, Leetbot automatically tries:
1. Last 30 days (most recent)
2. Last 3 months (if 30d has no data)
3. Last 6 months (if 3mo has no data)
4. More than 6 months (if 6mo has no data)
5. All time (fallback)`

				embed.Footer = &discordgo.MessageEmbedFooter{
					Text: "Page 2/4 • Use the buttons below to navigate",
				}

			case 2:
				// Page 3: Process Tracking
				embed.Title = "Process Tracking & Examples"
				embed.Color = 0x5865F2
				embed.Description = `**Process Tracking:**
• **Stages:** Apply, Reject, OA, Phone, Onsite, Offer
• Track your interview progress with **` + h.prefix + `process**
• View statistics with **` + h.prefix + `stats**

**Examples:**
• ` + h.prefix + `problems airbnb (uses smart priority)
• ` + h.prefix + `problems amazon 30d (forces 30 days)
• ` + h.prefix + `problems google 3mo (forces 3 months)
• ` + h.prefix + `process google apply (track application)
• ` + h.prefix + `stats google (view statistics)
• /problems company:airbnb (uses smart priority)
• /problems company:amazon timeframe:thirty-days
• /stats company:google`

				embed.Footer = &discordgo.MessageEmbedFooter{
					Text: "Page 3/4 • Use the buttons below to navigate",
				}

			case 3:
				// Page 4: Additional Info
				embed.Title = "Additional Information & Need Help?"
				embed.Color = 0x5865F2
				embed.Description = `**Supported Companies:**
• Use the slash command dropdown to see all available companies!

**Notes:**
• Problems are sorted by interview frequency (most popular first)
• Leetbot uses smart priority system for better results
• All commands support both text and slash command formats

**Need Help?**
• Use the pagination buttons to navigate through this help
• Try the slash commands for better user experience
• Check the dropdown options for available companies`

				embed.Footer = &discordgo.MessageEmbedFooter{
					Text: "Page 4/4 • Use the buttons below to navigate",
				}
			}

			embed.Timestamp = time.Now().Format(time.RFC3339)
		},
		MaxPages:        4,
		ExpiryLastUsage: true,
		Expiry:          time.Now().Add(10 * time.Minute),
	}
}

func (h *Handler) handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// if bot is disabled, send short offline message
	if h.disabled {
		h.sendMessage(s, m.ChannelID, "Leetbot is currently offline. Please try again later.")
		return
	}

	// check if user is admin
	isAdmin := m.Author.ID == "700444827287945316"

	// create help paginator
	pg := h.createHelpPaginator(isAdmin)

	// send paginated help
	err := PaginatorManager.CreateMessage(s, m.ChannelID, pg)
	if err != nil {
		fmt.Printf("Error creating help paginator: %v\n", err)
		// fallback to simple message
		h.sendMessage(s, m.ChannelID, "Error displaying help. Please try again.")
	}
}

func (h *Handler) handleProblemsSlash(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	companyOpt, ok := optionMap["company"]
	if !ok {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Company is required!",
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}
	company := strings.ToLower(companyOpt.StringValue())

	var problems []data.Problem
	var timeframe string

	if timeframeOpt, ok := optionMap["timeframe"]; ok {
		timeframe = timeframeOpt.StringValue()
		problems = h.problemsData.GetProblems(company, timeframe)
	} else {
		problems, timeframe = h.problemsData.GetProblemsWithPriority(company)
	}

	if problems == nil {
		availableTimeframes := h.problemsData.GetAvailableTimeframes(company)
		var responseContent string

		if len(availableTimeframes) > 0 {
			_, specifiedTimeframe := optionMap["timeframe"]
			if specifiedTimeframe {
				responseContent = h.formatAvailableTimeframesSuggestionSlash(company, timeframe, availableTimeframes)
			} else {
				responseContent = fmt.Sprintf("No data found for %s", formatCompanyName(company))
			}
		} else {
			responseContent = fmt.Sprintf("No data found for %s", formatCompanyName(company))
		}

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: responseContent,
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	if shouldUsePagination(len(problems)) {
		err := sendPaginatedProblems(s, i, company, timeframe, problems)
		if err != nil {
			fmt.Printf("Error sending paginated response: %v\n", err)
			// don't try to respond again - the interaction is already acknowledged
		}
		return
	}

	response := h.formatProblemsResponse(company, timeframe, problems)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
			Flags:   discordgo.MessageFlagsSuppressEmbeds,
		},
	})
	if err != nil {
		fmt.Printf("Error responding to interaction: %v\n", err)
	}
}

func (h *Handler) handleHelpSlash(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// if bot is disabled, send short offline message
	if h.disabled {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Leetbot is currently offline. Please try again later.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	// check if user is admin
	isAdmin := i.Member != nil && i.Member.User.ID == "700444827287945316"

	// create help paginator
	pg := h.createHelpPaginator(isAdmin)

	// send paginated help
	err := PaginatorManager.CreateInteraction(s, i.Interaction, pg, false)
	if err != nil {
		fmt.Printf("Error creating help paginator: %v\n", err)
		// fallback to simple message
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error displaying help. Please try again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
	}
}

func (h *Handler) handleProcessSlash(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	companyOpt, ok := optionMap["company"]
	if !ok {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Company is required!",
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	companyInput := strings.ToLower(companyOpt.StringValue())

	// get stage from option (now required)
	stageOpt, ok := optionMap["stage"]
	if !ok {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Stage is required!",
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}
	stage := stageOpt.StringValue()

	// validate and fuzzy match company
	company, companyFound, companySuggestions := findCompanyWithSuggestion(companyInput, h.problemsData)
	if !companyFound {
		var errorMsg strings.Builder
		errorMsg.WriteString(fmt.Sprintf("Could not find company matching '%s'.", companyInput))
		if len(companySuggestions) > 0 {
			errorMsg.WriteString("\n\nDid you mean:")
			for _, suggestion := range companySuggestions {
				errorMsg.WriteString(fmt.Sprintf("\n• %s", formatCompanyName(suggestion)))
			}
		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: errorMsg.String(),
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	// check if process storage is configured
	if h.processStorage == nil {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Process tracking is not configured on Leetbot.",
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	// create the process
	process := data.Process{
		Company:   company,
		Stage:     stage,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx := context.Background()
	_, err := h.processStorage.AddProcess(ctx, process)
	if err != nil {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Failed to add process: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	// if it's an offer, send congrats message
	if stage == "Offer" {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Congrats! 🎉",
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	// otherwise, respond with stats in embed with success header
	h.sendProcessStatsForInteractionWithSuccessHeader(s, i, company)
}


func formatTimeframeDisplay(timeframe string) string {
	switch timeframe {
	case "all":
		return "all"
	case "thirty-days":
		return "last 30 days"
	case "three-months":
		return "last 3 months"
	case "six-months":
		return "last 6 months"
	case "more-than-six-months":
		return "more than 6 months"
	default:
		return strings.ToLower(strings.ReplaceAll(timeframe, "-", " "))
	}
}

func (h *Handler) handleStatsSlash(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	companyOpt, ok := optionMap["company"]
	if !ok {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Company is required!",
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	companyInput := strings.ToLower(companyOpt.StringValue())

	// validate and fuzzy match company
	company, companyFound, companySuggestions := findCompanyWithSuggestion(companyInput, h.problemsData)
	if !companyFound {
		var errorMsg strings.Builder
		errorMsg.WriteString(fmt.Sprintf("Could not find company matching '%s'.", companyInput))
		if len(companySuggestions) > 0 {
			errorMsg.WriteString("\n\nDid you mean:")
			for _, suggestion := range companySuggestions {
				errorMsg.WriteString(fmt.Sprintf("\n• %s", formatCompanyName(suggestion)))
			}
		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: errorMsg.String(),
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	// check if process storage is configured
	if h.processStorage == nil {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Process tracking is not configured on Leetbot.",
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	// send stats as embed
	h.sendProcessStatsForInteraction(s, i, company)
}


func (h *Handler) handleShutdownMessage(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// check if the user is authorized (nyumat's user ID)
	if m.Author.ID != "700444827287945316" {
		h.sendErrorMessage(s, m.ChannelID, "Only the owner of Leetbot can use this command.")
		return
	}

	// check if indefinite shutdown is requested
	if len(args) > 0 && args[0] == "indef" {
		// indefinite shutdown - disable Leetbot but don't exit process
		h.disabled = true

		// unregister all slash commands except help
		h.sendMessage(s, m.ChannelID, "Unregistering commands...")
		err := h.unregisterCommandsExceptHelp(s)
		if err != nil {
			fmt.Printf("Error unregistering commands: %v\n", err)
			h.sendErrorMessage(s, m.ChannelID, fmt.Sprintf("Failed to unregister commands: %v", err))
			h.disabled = false // revert disabled state on error
			return
		}

		// set Leetbot status to invisible
		err = s.UpdateStatusComplex(discordgo.UpdateStatusData{
			Status: "invisible",
		})
		if err != nil {
			fmt.Printf("Error setting Leetbot status to invisible: %v\n", err)
		}

		h.sendMessage(s, m.ChannelID, "Leetbot is now disabled indefinitely. Use `!startup` to re-enable.")
		return
	}

	// regular shutdown - exit the process
	// send confirmation message first
	h.sendMessage(s, m.ChannelID, "Leetbot is now shutting down...")

	// close the session to disconnect from Discord
	// use a goroutine with a small delay to ensure the message is sent
	go func() {
		time.Sleep(100 * time.Millisecond)
		err := s.Close()
		if err != nil {
			fmt.Printf("Error closing Discord session: %v\n", err)
		}
		// exit the program
		os.Exit(0)
	}()
}

func (h *Handler) handleStartupMessage(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// check if the user is authorized (nyumat's user ID)
	if m.Author.ID != "700444827287945316" {
		h.sendErrorMessage(s, m.ChannelID, "Only the owner of Leetbot can use this command.")
		return
	}

	// check if Leetbot is disabled
	if h.disabled {
		// re-register slash commands
		h.sendMessage(s, m.ChannelID, "Re-registering commands...")
		err := h.registerAllCommands(s)
		if err != nil {
			fmt.Printf("Error re-registering commands: %v\n", err)
			h.sendErrorMessage(s, m.ChannelID, fmt.Sprintf("Failed to re-register commands: %v", err))
			return
		}

		// re-enable Leetbot
		h.disabled = false

		// restore normal Leetbot status (online)
		err = s.UpdateStatusComplex(discordgo.UpdateStatusData{
			Status: "online",
		})
		if err != nil {
			fmt.Printf("Error setting Leetbot status to online: %v\n", err)
		}

		h.sendMessage(s, m.ChannelID, "Leetbot is now back online.")
		return
	}

	// Send restart request through the channel
	if h.reconnectChan != nil {
		select {
		case h.reconnectChan <- RestartRequest{
			ChannelID: m.ChannelID,
			Success:   false,
			Message:   "Leetbot is restarting...",
		}:
			// Signal sent successfully
		default:
			// Channel full, send error message
			h.sendErrorMessage(s, m.ChannelID, "Restart already in progress, please wait.")
		}
	} else {
		// No channel configured, send error
		h.sendErrorMessage(s, m.ChannelID, "Restart mechanism not available.")
	}
}


func (h *Handler) formatAvailableTimeframesSuggestionSlash(company, requestedTimeframe string, availableTimeframes []string) string {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("No data found for %s (%s).\n\n",
		formatCompanyName(company),
		formatTimeframeDisplay(requestedTimeframe)))

	message.WriteString(fmt.Sprintf("Available timeframes for %s:\n", formatCompanyName(company)))

	priorityOrder := map[string]int{
		"thirty-days":          1,
		"three-months":         2,
		"six-months":           3,
		"more-than-six-months": 4,
		"all":                  5,
	}

	type timeframeWithPriority struct {
		name     string
		priority int
	}

	var sortedTimeframes []timeframeWithPriority
	for _, tf := range availableTimeframes {
		priority := priorityOrder[tf]
		if priority == 0 {
			priority = 999
		}
		sortedTimeframes = append(sortedTimeframes, timeframeWithPriority{name: tf, priority: priority})
	}

	for i := 0; i < len(sortedTimeframes)-1; i++ {
		for j := i + 1; j < len(sortedTimeframes); j++ {
			if sortedTimeframes[i].priority > sortedTimeframes[j].priority {
				sortedTimeframes[i], sortedTimeframes[j] = sortedTimeframes[j], sortedTimeframes[i]
			}
		}
	}

	for _, tf := range sortedTimeframes {
		message.WriteString(fmt.Sprintf("• **%s** (%s)\n", tf.name, formatTimeframeDisplay(tf.name)))
	}

	message.WriteString(fmt.Sprintf("\nTry: `/problems company:%s timeframe:<option>`", company))

	return message.String()
}

func (h *Handler) sendErrorMessage(s *discordgo.Session, channelID, message string) {

	h.sendMessage(s, channelID, message)
}

// unregisterCommandsExceptHelp removes all slash commands except the help command
func (h *Handler) unregisterCommandsExceptHelp(s *discordgo.Session) error {
	// get all currently registered commands
	registeredCommands, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		return fmt.Errorf("failed to get registered commands: %w", err)
	}

	// delete all commands except help
	for _, cmd := range registeredCommands {
		if cmd.Name != "help" {
			fmt.Printf("Unregistering command: /%s\n", cmd.Name)
			err := s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
			if err != nil {
				return fmt.Errorf("failed to delete command '%s': %w", cmd.Name, err)
			}
		}
	}

	return nil
}

// registerAllCommands registers all slash commands
func (h *Handler) registerAllCommands(s *discordgo.Session) error {
	commands := GetSlashCommands(h.problemsData)

	// get currently registered commands to avoid duplicates
	registeredCommands, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		return fmt.Errorf("failed to get registered commands: %w", err)
	}

	// create a map of registered command names
	registeredMap := make(map[string]bool)
	for _, cmd := range registeredCommands {
		registeredMap[cmd.Name] = true
	}

	// register commands that aren't already registered
	for _, cmd := range commands {
		if !registeredMap[cmd.Name] {
			fmt.Printf("Registering command: /%s\n", cmd.Name)
			_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
			if err != nil {
				return fmt.Errorf("failed to create command '%s': %w", cmd.Name, err)
			}
		}
	}

	return nil
}
