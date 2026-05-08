package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

type seedRecorder struct {
	executed []string
}

func (s *seedRecorder) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	s.executed = append(s.executed, sql)
	return pgconn.CommandTag{}, nil
}

func TestCollectSQLSeedFiles_Sorted(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	files := map[string]string{
		"002_b.sql":              "SELECT 'b';",
		"001_a.sql":              "SELECT 'a';",
		"investment/01_z.sql":    "SELECT 'z';",
		"investment/02_x.sql":    "SELECT 'x';",
		"ignored/notes.txt":      "ignored",
		"ignored/sub/note.txt":   "ignored",
		"investment/sub/03_y.sql": "SELECT 'y';",
	}
	for rel, body := range files {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %q: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %q: %v", path, err)
		}
	}

	got, err := collectSQLSeedFiles(dir)
	if err != nil {
		t.Fatalf("collectSQLSeedFiles: %v", err)
	}

	wantSuffixes := []string{
		"001_a.sql",
		"002_b.sql",
		filepath.Join("investment", "01_z.sql"),
		filepath.Join("investment", "02_x.sql"),
		filepath.Join("investment", "sub", "03_y.sql"),
	}
	if len(got) != len(wantSuffixes) {
		t.Fatalf("expected %d files, got %d (%v)", len(wantSuffixes), len(got), got)
	}
	for i, want := range wantSuffixes {
		if !strings.HasSuffix(got[i], want) {
			t.Errorf("position %d: got %q, want suffix %q", i, got[i], want)
		}
	}
}

func TestRunSQLSeeds_NoFilesIsNotAnError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rec := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec, dir); err != nil {
		t.Fatalf("runSQLSeeds returned error on empty dir: %v", err)
	}
	if len(rec.executed) != 0 {
		t.Errorf("expected no executions, got %d", len(rec.executed))
	}
}

func TestRunSQLSeeds_ExecutesAllInOrder(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	files := map[string]string{
		"01_first.sql":           "FIRST_SQL;",
		"02_second.sql":          "SECOND_SQL;",
		"investment/01_inv.sql":  "INV_SQL;",
	}
	for rel, body := range files {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %q: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %q: %v", path, err)
		}
	}

	rec := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec, dir); err != nil {
		t.Fatalf("runSQLSeeds: %v", err)
	}
	if len(rec.executed) != 3 {
		t.Fatalf("expected 3 executions, got %d", len(rec.executed))
	}
	wantOrder := []string{"FIRST_SQL;", "SECOND_SQL;", "INV_SQL;"}
	for i, want := range wantOrder {
		if rec.executed[i] != want {
			t.Errorf("execution %d: got %q, want %q", i, rec.executed[i], want)
		}
	}
}
