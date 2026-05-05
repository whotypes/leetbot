package analytics

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const advisoryLockKey int64 = 8738421

type Source string

const (
	SourceWeb     Source = "web"
	SourceDiscord Source = "discord"
)

type Recorder struct {
	db *sql.DB
	ch chan insertJob
	wg sync.WaitGroup
}

type insertJob struct {
	source      Source
	companySlug string
	timeframe   string
}

func New(ctx context.Context, dsn string) (*Recorder, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, nil
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	ch := make(chan insertJob, 512)
	r := &Recorder{db: db, ch: ch}
	r.wg.Add(1)
	go r.worker()

	return r, nil
}

func NewFromEnv(ctx context.Context) (*Recorder, error) {
	return New(ctx, os.Getenv("DATABASE_URL"))
}

func migrate(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, advisoryLockKey); err != nil {
		return err
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS analytics_events (
			id BIGSERIAL PRIMARY KEY,
			occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			source TEXT NOT NULL CHECK (source IN ('web','discord')),
			event_type TEXT NOT NULL,
			company_slug TEXT NOT NULL,
			timeframe TEXT,
			meta JSONB
		)`,
		`ALTER TABLE analytics_events ADD COLUMN IF NOT EXISTS dedupe TEXT`,
		`CREATE UNIQUE INDEX IF NOT EXISTS analytics_events_dedupe_uidx ON analytics_events (dedupe)`,
		`CREATE INDEX IF NOT EXISTS analytics_events_occurred_at_idx ON analytics_events (occurred_at DESC)`,
		`CREATE INDEX IF NOT EXISTS analytics_events_company_slug_idx ON analytics_events (company_slug)`,
	}
	for _, s := range stmts {
		if _, err := tx.ExecContext(ctx, s); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Migrate ensures analytics_events exists and applies additive schema updates (dedupe for backfill).
func Migrate(ctx context.Context, db *sql.DB) error {
	return migrate(ctx, db)
}

func (r *Recorder) RecordProblemsFetch(_ context.Context, source Source, companySlug, timeframe string) {
	if r == nil || r.db == nil {
		return
	}
	companySlug = strings.ToLower(strings.TrimSpace(companySlug))
	if companySlug == "" {
		return
	}
	timeframe = strings.TrimSpace(timeframe)
	job := insertJob{source: source, companySlug: companySlug, timeframe: timeframe}
	select {
	case r.ch <- job:
	default:
		log.Printf("analytics: event queue full, dropping problems_fetch for %s", companySlug)
	}
}

func (r *Recorder) worker() {
	defer r.wg.Done()
	const q = `INSERT INTO analytics_events (source, event_type, company_slug, timeframe) VALUES ($1, $2, $3, $4)`
	for job := range r.ch {
		runCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		var tf interface{}
		if job.timeframe != "" {
			tf = job.timeframe
		}
		_, err := r.db.ExecContext(runCtx, q, string(job.source), "problems_fetch", job.companySlug, tf)
		cancel()
		if err != nil {
			log.Printf("analytics: insert failed: %v", err)
		}
	}
}
