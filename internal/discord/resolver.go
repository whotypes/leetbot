package discord

import (
	"strings"

	"github.com/whotypes/leetbot/internal/data"
)

// ResolveProblemsAnalytics returns the canonical company slug and resolved timeframe when the same
// logic as a successful prefix problems command would produce data. skipExternalAPI matches
// backfill: no company-enrich HTTP calls, only embedded data + fuzzy matching.
func ResolveProblemsAnalytics(pbc *data.ProblemsByCompany, args []string, skipExternalAPI bool) (companySlug, timeframe string, ok bool) {
	if len(args) == 0 {
		return "", "", false
	}

	companyInput, timeframeArg := parseProblemsCommandArgs(args, IsProblemsTimeframeKeyword)
	cleaned := cleanCompanyInput(companyInput)
	company, companyFound, _ := findCompanyWithSuggestion(cleaned, pbc, skipExternalAPI)
	if !companyFound {
		return "", "", false
	}

	var problems []data.Problem
	var tf string
	if timeframeArg != "" {
		tf = NormalizeProblemsTimeframe(timeframeArg)
		problems = pbc.GetProblems(company, tf)
	} else {
		problems, tf = pbc.GetProblemsWithPriority(company)
	}
	if problems == nil {
		return "", "", false
	}
	return company, tf, true
}

// ParseExportContentForProblems extracts problems command args from a Discord channel export line.
// prefix is the bot prefix (e.g. "!"). Only messages that are valid prefix "problems" or "/problems" (slash) are accepted.
func ParseExportContentForProblems(prefix, content string) (args []string, slash bool, ok bool) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, false, false
	}

	lowerP := strings.ToLower(strings.TrimSpace(prefix))
	if lowerP == "" {
		lowerP = "!"
	}

	// slash: /problems ... (options or freeform)
	lowerC := strings.ToLower(content)
	if idx := strings.Index(lowerC, "/problems"); idx >= 0 {
		rest := strings.TrimSpace(content[idx+len("/problems"):])
		if rest == "" {
			return nil, true, false
		}
		args = parseSlashProblemsRest(rest)
		return args, true, len(args) > 0
	}

	// prefix: <prefix>problems ... (command name must be exactly "problems", same as live handler)
	if !strings.HasPrefix(content, lowerP) {
		return nil, false, false
	}
	trimmed := strings.TrimSpace(content[len(lowerP):])
	if trimmed == "" {
		return nil, false, false
	}
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return nil, false, false
	}
	if strings.ToLower(parts[0]) != "problems" {
		return nil, false, false
	}
	args = parts[1:]
	return args, false, len(args) > 0
}

func parseSlashProblemsRest(rest string) []string {
	rest = strings.TrimSpace(rest)
	lower := strings.ToLower(rest)
	companyKey := "company:"
	tfKey := "timeframe:"

	if idx := strings.Index(lower, companyKey); idx >= 0 {
		afterCompany := strings.TrimSpace(rest[idx+len(companyKey):])
		tfIdx := strings.Index(strings.ToLower(afterCompany), tfKey)
		var companyPart string
		var tfPart string
		if tfIdx >= 0 {
			companyPart = strings.TrimSpace(strings.Trim(afterCompany[:tfIdx], `"'`))
			tfPart = strings.TrimSpace(strings.Trim(afterCompany[tfIdx+len(tfKey):], `"'`))
		} else {
			companyPart = strings.TrimSpace(strings.Trim(afterCompany, `"'`))
		}

		var args []string
		if companyPart != "" {
			args = strings.Fields(companyPart)
		}
		if tfPart != "" {
			args = append(args, strings.Fields(tfPart)...)
		}
		return args
	}

	return strings.Fields(rest)
}
