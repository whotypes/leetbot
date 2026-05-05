package analytics

import (
	"context"
	"testing"
)

func TestNewEmptyDSN(t *testing.T) {
	t.Parallel()
	r, err := New(context.Background(), "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if r != nil {
		t.Fatalf("expected nil recorder for empty dsn")
	}
}

func TestRecordNilRecorder(t *testing.T) {
	t.Parallel()
	var r *Recorder
	r.RecordProblemsFetch(context.Background(), SourceWeb, "google", "all")
}
