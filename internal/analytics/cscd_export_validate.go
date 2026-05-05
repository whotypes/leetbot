package analytics

import (
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"time"
)

const cscdProblemsPrefix = "!problems"

// CscdValidateOptions configures streaming validation of a Discord chat CSV export.
type CscdValidateOptions struct {
	// MaxRecords is the maximum number of data rows to read after the header. 0 means read until EOF.
	MaxRecords int
	// CollectSkippedSamples, if > 0, keeps up to this many distinct skipped "!problems" bodies (no company token).
	CollectSkippedSamples int
}

// CscdValidateResult holds counters from a validation pass.
type CscdValidateResult struct {
	RecordsRead              int
	ProblemsValidated        int
	ProblemsSkippedNoCompany int
	EmptyCompanyToken        int
	DateErrors               int
	FirstDateErrorRaw        string
	SkippedNoCompanySamples  []string
}

// ParseDiscordExportTimestamp parses timestamp strings from Discord CSV exports for TIMESTAMPTZ compatibility.
func ParseDiscordExportTimestamp(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("empty timestamp")
	}
	return time.Parse(time.RFC3339Nano, s)
}

func fitsProblemsFetchShape(companySlug string) bool {
	return strings.TrimSpace(companySlug) != ""
}

// ValidateCscdExport streams a Discord-export-style CSV (AuthorID, Author, Date, Content, …) and checks that
// dates parse and !problems rows map to analytics_events-friendly fields. Does not use the database.
func ValidateCscdExport(r io.Reader, opt CscdValidateOptions) (CscdValidateResult, error) {
	cr := csv.NewReader(r)
	cr.ReuseRecord = true
	cr.FieldsPerRecord = -1

	var res CscdValidateResult
	maxSamples := opt.CollectSkippedSamples
	if maxSamples < 0 {
		maxSamples = 0
	}

	n := 0
	for opt.MaxRecords == 0 || n < opt.MaxRecords {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return res, err
		}
		n++
		if len(rec) < 4 {
			return res, errors.New("record has fewer than 4 columns")
		}
		if n == 1 && rec[0] == "AuthorID" {
			continue
		}

		_, err = ParseDiscordExportTimestamp(rec[2])
		if err != nil {
			res.DateErrors++
			if res.FirstDateErrorRaw == "" {
				res.FirstDateErrorRaw = rec[2]
			}
			continue
		}

		body := strings.TrimSpace(rec[3])
		if !strings.HasPrefix(body, cscdProblemsPrefix) {
			continue
		}

		fields := strings.Fields(body)
		if len(fields) < 2 {
			res.ProblemsSkippedNoCompany++
			if len(res.SkippedNoCompanySamples) < maxSamples {
				res.SkippedNoCompanySamples = append(res.SkippedNoCompanySamples, strings.Clone(body))
			}
			continue
		}

		companyToken := strings.ToLower(strings.TrimSpace(fields[1]))
		if !fitsProblemsFetchShape(companyToken) {
			res.EmptyCompanyToken++
			continue
		}

		res.ProblemsValidated++
	}

	res.RecordsRead = n
	return res, nil
}
