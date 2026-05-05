package analytics

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	cscdFastScanMaxRecords = 150000
	fastScanSkippedSamples = 60
)

func findModuleRoot() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func cscdExportPath() string {
	if p := os.Getenv("LEETBOT_CSCD_EXPORT"); p != "" {
		return p
	}
	if root, ok := findModuleRoot(); ok {
		for _, name := range []string{"cscd.csv", "cscd.html"} {
			p := filepath.Join(root, name)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return filepath.Join("..", "..", "cscd.csv")
}

func truncateForLog(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes]) + "…"
}

// TestCscdExportFitsMetabaseAnalyticsShape scans the first chunk of the CSV export (if present).
func TestCscdExportFitsMetabaseAnalyticsShape(t *testing.T) {
	path := cscdExportPath()
	f, err := os.Open(path)
	if err != nil {
		t.Skipf("no export at %s (%v); set LEETBOT_CSCD_EXPORT or add cscd.csv at repo root", path, err)
	}
	defer func() { _ = f.Close() }()

	res, err := ValidateCscdExport(f, CscdValidateOptions{
		MaxRecords:            cscdFastScanMaxRecords,
		CollectSkippedSamples: fastScanSkippedSamples,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.DateErrors > 0 {
		t.Fatalf("%d rows failed date parse (sample raw: %q)", res.DateErrors, res.FirstDateErrorRaw)
	}
	if res.ProblemsValidated == 0 {
		t.Fatalf("no valid !problems <company> rows in first %d csv reads; widen max records or check export", cscdFastScanMaxRecords)
	}
	t.Logf("ok: scanned %d csv records, %d !problems rows map to analytics_events; skipped %d with no company token",
		res.RecordsRead, res.ProblemsValidated, res.ProblemsSkippedNoCompany)
	for i, s := range res.SkippedNoCompanySamples {
		t.Logf("skipped no company (sample %d of %d shown, %d total in scan): %s",
			i+1, len(res.SkippedNoCompanySamples), res.ProblemsSkippedNoCompany, truncateForLog(s, 320))
	}
}

// TestCscdExportFullFileValidation streams the entire export file. Skips if no file (e.g. CI).
// Set LEETBOT_CSCD_FULL_VALIDATION=1 to enable (slow; streams the whole CSV once).
func TestCscdExportFullFileValidation(t *testing.T) {
	if os.Getenv("LEETBOT_CSCD_FULL_VALIDATION") != "1" {
		t.Skip("set LEETBOT_CSCD_FULL_VALIDATION=1 to stream the entire export")
	}
	path := cscdExportPath()
	f, err := os.Open(path)
	if err != nil {
		t.Skipf("no export at %s (%v); set LEETBOT_CSCD_EXPORT or add cscd.csv at repo root", path, err)
	}
	defer func() { _ = f.Close() }()

	res, err := ValidateCscdExport(f, CscdValidateOptions{
		MaxRecords:            0,
		CollectSkippedSamples: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.DateErrors > 0 {
		t.Fatalf("full pass: %d date parse errors (first raw: %q)", res.DateErrors, res.FirstDateErrorRaw)
	}
	t.Logf("full pass ok: csv_reads=%d problems_validated=%d skipped_no_company=%d empty_company_token=%d",
		res.RecordsRead, res.ProblemsValidated, res.ProblemsSkippedNoCompany, res.EmptyCompanyToken)
}
