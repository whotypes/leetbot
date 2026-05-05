package discord

import "strings"

// NormalizeProblemsTimeframe matches Handler.NormalizeTimeframe for prefix and slash commands.
func NormalizeProblemsTimeframe(timeframe string) string {
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

// IsProblemsTimeframeKeyword matches Handler.isTimeframeKeyword.
func IsProblemsTimeframeKeyword(s string) bool {
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
