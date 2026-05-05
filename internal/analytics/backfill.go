package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

// BackfillMeta is stored in analytics_events.meta for imported rows (compact JSON for dashboards).
type BackfillMeta struct {
	Source string `json:"source"`
	Slash  bool   `json:"slash,omitempty"`
}

// InsertProblemsFetchBackfill inserts one historical row with explicit occurred_at and dedupe key.
// inserted is false when ON CONFLICT (dedupe) skipped a duplicate.
func InsertProblemsFetchBackfill(ctx context.Context, db *sql.DB, occurredAt time.Time, companySlug, timeframe, dedupe string, meta BackfillMeta) (inserted bool, err error) {
	companySlug = strings.ToLower(strings.TrimSpace(companySlug))
	if companySlug == "" || dedupe == "" {
		return false, nil
	}
	timeframe = strings.TrimSpace(timeframe)

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return false, err
	}

	var tf interface{}
	if timeframe != "" {
		tf = timeframe
	}

	res, err := db.ExecContext(ctx, `
		INSERT INTO analytics_events (occurred_at, source, event_type, company_slug, timeframe, dedupe, meta)
		VALUES ($1, 'discord', 'problems_fetch', $2, $3, $4, $5::jsonb)
		ON CONFLICT (dedupe) DO NOTHING
	`, occurredAt, companySlug, tf, dedupe, metaJSON)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
