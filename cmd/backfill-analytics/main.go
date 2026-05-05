package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/whotypes/leetbot/internal/analytics"
	"github.com/whotypes/leetbot/internal/data"
	"github.com/whotypes/leetbot/internal/discord"
)

func main() {
	csvPath := flag.String("csv", "cscd.csv", "path to Discord channel export CSV")
	botPrefix := flag.String("prefix", "!", "bot prefix in the export (BOT_PREFIX)")
	apply := flag.Bool("apply", false, "write to DATABASE_URL; default is dry run")
	skipAuthorCSV := flag.String("skip-author-ids", "", "comma-separated AuthorID values to skip (bot messages)")
	flag.Parse()

	ctx := context.Background()
	_ = godotenv.Load()

	skipIDs := parseSkipSet(*skipAuthorCSV)
	if env := os.Getenv("LEETBOT_BACKFILL_SKIP_AUTHOR_IDS"); env != "" && len(skipIDs) == 0 {
		skipIDs = parseSkipSet(env)
	}

	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if *apply && dsn == "" {
		log.Fatal("DATABASE_URL is required when -apply is set")
	}

	fmt.Fprintf(os.Stderr, "backfill: loading embedded problem data...\n")
	problemsData, err := data.LoadAllProblems()
	if err != nil {
		log.Fatalf("load problems data: %v", err)
	}

	fmt.Fprintf(os.Stderr, "backfill: opening %s...\n", *csvPath)
	f, err := os.Open(*csvPath)
	if err != nil {
		log.Fatalf("open csv: %v", err)
	}
	defer func() { _ = f.Close() }()

	var db *sql.DB
	if *apply {
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			log.Fatalf("db: %v", err)
		}
		defer func() { _ = db.Close() }()
		if err := db.PingContext(ctx); err != nil {
			log.Fatalf("db ping: %v", err)
		}
		if err := analytics.Migrate(ctx, db); err != nil {
			log.Fatalf("migrate: %v", err)
		}
	}

	if !*apply {
		fmt.Fprintf(os.Stderr, "backfill: dry run (no DB writes). Scanning CSV (progress every 25k logical rows)...\n")
	} else {
		fmt.Fprintf(os.Stderr, "backfill: writing to database. Scanning CSV...\n")
	}

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.ReuseRecord = true

	var (
		scanLines      int
		inserted       int
		deDupSkipped   int
		parseSkipped   int
		resolveSkipped int
		timeSkipped    int
		authorSkipped  int
		wouldInsert    int
	)

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("csv read: %v", err)
		}
		scanLines++
		if scanLines%25000 == 0 && scanLines > 0 {
			fmt.Fprintf(os.Stderr, "backfill: scanned %d csv logical rows...\n", scanLines)
		}
		if len(rec) < 4 {
			continue
		}
		if scanLines == 1 && rec[0] == "AuthorID" {
			continue
		}

		authorID := strings.TrimSpace(rec[0])
		if skipIDs[authorID] {
			authorSkipped++
			continue
		}

		ts, err := analytics.ParseDiscordExportTimestamp(rec[2])
		if err != nil {
			timeSkipped++
			continue
		}

		content := rec[3]
		args, slash, ok := discord.ParseExportContentForProblems(*botPrefix, content)
		if !ok {
			parseSkipped++
			continue
		}

		companySlug, timeframe, ok := discord.ResolveProblemsAnalytics(problemsData, args, true)
		if !ok {
			resolveSkipped++
			continue
		}

		sum := sha256.Sum256([]byte(authorID + "|" + strings.TrimSpace(rec[2]) + "|" + content))
		dedupe := hex.EncodeToString(sum[:])
		meta := analytics.BackfillMeta{Source: "discord_export", Slash: slash}

		if !*apply {
			wouldInsert++
			continue
		}

		insertedNow, err := analytics.InsertProblemsFetchBackfill(ctx, db, ts, companySlug, timeframe, dedupe, meta)
		if err != nil {
			log.Fatalf("insert: %v", err)
		}
		if insertedNow {
			inserted++
		} else {
			deDupSkipped++
		}
	}

	fmt.Printf("csv logical records scanned: %d\n", scanLines)
	fmt.Printf("skipped (author filter): %d\n", authorSkipped)
	fmt.Printf("skipped (bad timestamp): %d\n", timeSkipped)
	fmt.Printf("skipped (not problems command): %d\n", parseSkipped)
	fmt.Printf("skipped (unresolved company or no problem data): %d\n", resolveSkipped)
	if !*apply {
		fmt.Printf("dry run — would insert: %d\n", wouldInsert)
		fmt.Println("set -apply to write (requires DATABASE_URL)")
		return
	}
	fmt.Printf("inserted: %d\n", inserted)
	fmt.Printf("skipped duplicate dedupe: %d\n", deDupSkipped)
}

func parseSkipSet(s string) map[string]bool {
	out := make(map[string]bool)
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out[p] = true
		}
	}
	return out
}
