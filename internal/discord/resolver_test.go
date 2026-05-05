package discord

import (
	"testing"

	"github.com/whotypes/leetbot/internal/data"
)

func TestParseExportContentForProblems_prefix(t *testing.T) {
	args, slash, ok := ParseExportContentForProblems("!", "!problems google")
	if !ok || slash || len(args) != 1 || args[0] != "google" {
		t.Fatalf("got args=%v slash=%v ok=%v", args, slash, ok)
	}
}

func TestParseExportContentForProblems_slashOptions(t *testing.T) {
	args, slash, ok := ParseExportContentForProblems("!", "/problems company:ibm timeframe:90d")
	if !ok || !slash || len(args) < 1 {
		t.Fatalf("got args=%v slash=%v ok=%v", args, slash, ok)
	}
	if args[0] != "ibm" {
		t.Fatalf("company token: %v", args)
	}
}

func TestParseExportContentForProblems_slashFreeform(t *testing.T) {
	args, slash, ok := ParseExportContentForProblems("!", "/problems amazon")
	if !ok || !slash || len(args) != 1 || args[0] != "amazon" {
		t.Fatalf("got args=%v slash=%v ok=%v", args, slash, ok)
	}
}

func TestResolveProblemsAnalytics_offlineGoogle(t *testing.T) {
	pbc, err := data.LoadAllProblems()
	if err != nil {
		t.Fatal(err)
	}
	slug, tf, ok := ResolveProblemsAnalytics(pbc, []string{"google"}, true)
	if !ok || slug == "" {
		t.Fatalf("resolve: slug=%q ok=%v", slug, ok)
	}
	_ = tf
}
