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

// writeSeedFile writes a seed file (creating parent dirs) under root at the
// given forward-slash relative path.
func writeSeedFile(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

// writeReferenceManifest writes reference_manifest.txt at root listing the given
// relative reference paths.
func writeReferenceManifest(t *testing.T, root string, refs ...string) {
	t.Helper()
	body := "# test manifest\n" + strings.Join(refs, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(root, referenceManifestName), []byte(body), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

func TestCollectSQLSeedFiles_Sorted(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	files := map[string]string{
		"002_b.sql":               "SELECT 'b';",
		"001_a.sql":               "SELECT 'a';",
		"investment/01_z.sql":     "SELECT 'z';",
		"investment/02_x.sql":     "SELECT 'x';",
		"ignored/notes.txt":       "ignored",
		"ignored/sub/note.txt":    "ignored",
		"investment/sub/03_y.sql": "SELECT 'y';",
	}
	for rel, body := range files {
		writeSeedFile(t, dir, rel, body)
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
	if err := runSQLSeeds(context.Background(), rec, dir, true); err != nil {
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
		"01_first.sql":          "FIRST_SQL;",
		"02_second.sql":         "SECOND_SQL;",
		"investment/01_inv.sql": "INV_SQL;",
	}
	for rel, body := range files {
		writeSeedFile(t, dir, rel, body)
	}

	rec := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec, dir, true); err != nil {
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

// seedFixtureWithDemo lays out a seeds root containing one reference file, one
// investment/ reference file, and one file under each demo directory (demo/,
// zz_demo/). It also writes a reference manifest listing only the two reference
// files. It returns the directory and the exact SQL bodies used so tests can
// assert on execution content without relying on path-string matching.
func seedFixtureWithDemo(t *testing.T) (dir string, referenceSQL, investmentSQL, demoSQL, zzDemoSQL string) {
	t.Helper()
	dir = t.TempDir()

	referenceSQL = "REFERENCE_SQL;"
	investmentSQL = "INVESTMENT_REFERENCE_SQL;"
	demoSQL = "DEMO_SQL;"
	zzDemoSQL = "ZZ_DEMO_SQL;"

	writeSeedFile(t, dir, "001_reference.sql", referenceSQL)
	writeSeedFile(t, dir, "investment/01_ref.sql", investmentSQL)
	writeSeedFile(t, dir, "demo/001_demo.sql", demoSQL)
	writeSeedFile(t, dir, "zz_demo/01_zz_demo.sql", zzDemoSQL)
	writeReferenceManifest(t, dir, "001_reference.sql", "investment/01_ref.sql")
	return dir, referenceSQL, investmentSQL, demoSQL, zzDemoSQL
}

// TestRunSQLSeeds_ProductionSkipsDemoAndSucceeds proves that reference seeds
// (top-level and investment/) execute and are committed, demo/ and zz_demo/
// files are never executed, and the overall call returns nil — a production (or
// any non-opted-in) run must succeed even though demo directories exist on disk.
func TestRunSQLSeeds_ProductionSkipsDemoAndSucceeds(t *testing.T) {
	t.Parallel()
	dir, referenceSQL, investmentSQL, demoSQL, zzDemoSQL := seedFixtureWithDemo(t)

	rec := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec, dir, false); err != nil {
		t.Fatalf("expected runSQLSeeds to succeed when demoAllowed=false and demo files exist, got: %v", err)
	}

	executed := map[string]bool{}
	for _, sql := range rec.executed {
		executed[sql] = true
	}
	if !executed[referenceSQL] {
		t.Errorf("expected reference seed %q to execute", referenceSQL)
	}
	if !executed[investmentSQL] {
		t.Errorf("expected investment/ reference seed %q to execute", investmentSQL)
	}
	if executed[demoSQL] {
		t.Errorf("demo/ seed %q must NOT execute when demoAllowed=false", demoSQL)
	}
	if executed[zzDemoSQL] {
		t.Errorf("zz_demo/ seed %q must NOT execute when demoAllowed=false", zzDemoSQL)
	}
	if len(rec.executed) != 2 {
		t.Errorf("expected exactly 2 executions (reference only), got %d: %v", len(rec.executed), rec.executed)
	}
}

// TestRunSQLSeeds_DevelopmentAndTestAllowDemo proves demo/dev parity: with
// demoAllowed=true (verified dev/test opt-in), every file — reference and demo —
// executes, no manifest is required, and no error is returned.
func TestRunSQLSeeds_DevelopmentAndTestAllowDemo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	referenceSQL := "REFERENCE_SQL;"
	demoSQL := "DEMO_SQL;"
	zzDemoSQL := "ZZ_DEMO_SQL;"
	writeSeedFile(t, dir, "001_reference.sql", referenceSQL)
	writeSeedFile(t, dir, "demo/001_demo.sql", demoSQL)
	writeSeedFile(t, dir, "zz_demo/01_zz_demo.sql", zzDemoSQL)
	// Deliberately NO manifest — demoAllowed=true must not need one.

	rec := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec, dir, true); err != nil {
		t.Fatalf("runSQLSeeds: unexpected error with demoAllowed=true: %v", err)
	}

	executed := map[string]bool{}
	for _, sql := range rec.executed {
		executed[sql] = true
	}
	for name, sql := range map[string]string{
		"reference": referenceSQL,
		"demo":      demoSQL,
		"zz_demo":   zzDemoSQL,
	} {
		if !executed[sql] {
			t.Errorf("expected %s seed %q to execute when demoAllowed=true", name, sql)
		}
	}
	if len(rec.executed) != 3 {
		t.Errorf("expected exactly 3 executions, got %d: %v", len(rec.executed), rec.executed)
	}
}

func TestRunSQLSeeds_NoDemoFilesProductionSucceeds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSeedFile(t, dir, "001_reference.sql", "REFERENCE_ONLY;")
	writeReferenceManifest(t, dir, "001_reference.sql")

	rec := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec, dir, false); err != nil {
		t.Fatalf("expected no error when no demo files exist, got: %v", err)
	}
	if len(rec.executed) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(rec.executed))
	}
}

// TestRunSQLSeeds_ProductionMissingManifestFailsClosed proves the core Fix A
// safety property: with demoAllowed=false and NO manifest at the seed root, the
// seeder hard-errors rather than guessing — this is what turns the previous
// "production silently seeds nothing (or everything)" outcomes into a loud
// failure.
func TestRunSQLSeeds_ProductionMissingManifestFailsClosed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSeedFile(t, dir, "001_reference.sql", "REFERENCE;")
	writeSeedFile(t, dir, "demo/001_demo.sql", "DEMO;")
	// No manifest written.

	rec := &seedRecorder{}
	err := runSQLSeeds(context.Background(), rec, dir, false)
	if err == nil {
		t.Fatal("expected a hard error when demoAllowed=false and no reference manifest exists")
	}
	if len(rec.executed) != 0 {
		t.Errorf("no seed must execute when the manifest is missing, got: %v", rec.executed)
	}
}

// TestRunSQLSeeds_ProductionEmptyManifestFailsClosed proves that a manifest with
// no real entries (only comments/blanks) is rejected — it would otherwise skip
// every seed in production silently.
func TestRunSQLSeeds_ProductionEmptyManifestFailsClosed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSeedFile(t, dir, "001_reference.sql", "REFERENCE;")
	if err := os.WriteFile(filepath.Join(dir, referenceManifestName), []byte("# only a comment\n\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	rec := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec, dir, false); err == nil {
		t.Fatal("expected a hard error for an empty reference manifest")
	}
	if len(rec.executed) != 0 {
		t.Errorf("no seed must execute for an empty manifest, got: %v", rec.executed)
	}
}

// TestRunSQLSeeds_SeedsPathAtSubdirWithoutManifestFailsClosed proves the F1-class
// fix under the manifest design: pointing SEEDS_PATH directly at a demo
// subdirectory (which has no manifest) hard-errors in production instead of
// presenting the demo files as if they were root-level reference data. With the
// dev/test opt-in, the same directory still runs its files.
func TestRunSQLSeeds_SeedsPathAtSubdirWithoutManifestFailsClosed(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	demoRoot := filepath.Join(root, "demo")
	demoSQL := "DEMO_ROOT_SQL;"
	if err := os.MkdirAll(demoRoot, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(demoRoot, "001_demo.sql"), []byte(demoSQL), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// SEEDS_PATH points directly AT the demo directory in production: no manifest
	// lives there, so this must fail closed rather than run the demo file.
	rec := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec, demoRoot, false); err == nil {
		t.Fatal("expected a hard error when SEEDS_PATH points at a manifest-less subdirectory in production")
	}
	if len(rec.executed) != 0 {
		t.Errorf("no seed must execute, got: %v", rec.executed)
	}

	// The same directory with demo genuinely allowed still runs its files.
	rec2 := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec2, demoRoot, true); err != nil {
		t.Fatalf("runSQLSeeds with demoAllowed=true: %v", err)
	}
	if len(rec2.executed) != 1 || rec2.executed[0] != demoSQL {
		t.Errorf("expected the demo file to execute when demoAllowed=true, got: %v", rec2.executed)
	}
}

// TestRunSQLSeeds_ParentDirNamedDemoIsUnaffected proves the worst path-text
// bypass is gone: a seed root whose ancestor directory is literally named "demo"
// must NOT cause reference seeds to be misclassified as demo. Manifest
// classification is relative to the seed root, so absolute-path segments are
// irrelevant.
func TestRunSQLSeeds_ParentDirNamedDemoIsUnaffected(t *testing.T) {
	t.Parallel()
	// Seed root lives under a parent directory named "demo".
	parent := filepath.Join(t.TempDir(), "demo")
	root := filepath.Join(parent, "seeds")
	referenceSQL := "REFERENCE_UNDER_DEMO_PARENT;"
	demoSQL := "ACTUAL_DEMO;"
	writeSeedFile(t, root, "001_reference.sql", referenceSQL)
	writeSeedFile(t, root, "demo/001_demo.sql", demoSQL)
	writeReferenceManifest(t, root, "001_reference.sql")

	rec := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec, root, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rec.executed) != 1 || rec.executed[0] != referenceSQL {
		t.Errorf("expected only the reference seed to run despite the 'demo' parent dir, got: %v", rec.executed)
	}
}

// TestRunSQLSeeds_RenamedDemoDirHardErrors proves the renamed-directory bypass is
// closed LOUDLY: an unlisted file under an arbitrarily named directory (not a
// recognized demo location) is a hard error, not a silent skip. It never runs
// (no demo leak) and the operator is told the manifest/tree are out of sync.
func TestRunSQLSeeds_RenamedDemoDirHardErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	referenceSQL := "REFERENCE;"
	renamedDemoSQL := "RENAMED_DEMO;"
	writeSeedFile(t, dir, "001_reference.sql", referenceSQL)
	writeSeedFile(t, dir, "samples/001_ben_grant.sql", renamedDemoSQL) // demo content, non-demo dir name
	writeReferenceManifest(t, dir, "001_reference.sql")

	rec := &seedRecorder{}
	err := runSQLSeeds(context.Background(), rec, dir, false)
	if err == nil {
		t.Fatal("expected a hard error for an unlisted file outside a recognized demo location")
	}
	for _, sql := range rec.executed {
		if sql == renamedDemoSQL {
			t.Fatalf("the unlisted samples/ file must NEVER execute, got: %v", rec.executed)
		}
	}
}

// TestRunSQLSeeds_UnlistedNonDemoFileHardErrors proves the manifest-sync guard:
// a NEW top-level production reference seed added without updating the manifest
// (e.g. a seed landing from a parallel branch) hard-errors at seed time — in
// CI/staging — instead of being silently skipped and missing from production.
func TestRunSQLSeeds_UnlistedNonDemoFileHardErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSeedFile(t, dir, "001_reference.sql", "REFERENCE;")
	writeSeedFile(t, dir, "099_new_process_seed.sql", "NEW_REFERENCE;") // not in manifest, not demo
	writeSeedFile(t, dir, "demo/001_demo.sql", "DEMO;")                 // demo-located: still quiet-skip
	writeReferenceManifest(t, dir, "001_reference.sql")

	rec := &seedRecorder{}
	err := runSQLSeeds(context.Background(), rec, dir, false)
	if err == nil {
		t.Fatal("expected a hard error for an unlisted, non-demo-located seed file")
	}
	if !strings.Contains(err.Error(), "099_new_process_seed.sql") {
		t.Errorf("error should name the offending file; got: %v", err)
	}
}

// TestRunSQLSeeds_ProductionZeroFilesHardErrors proves production seeding fails
// loudly on an empty/misconfigured seed directory (no *.sql files) rather than
// silently succeeding.
func TestRunSQLSeeds_ProductionZeroFilesHardErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeReferenceManifest(t, dir, "001_reference.sql") // manifest present, but no *.sql on disk

	rec := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec, dir, false); err == nil {
		t.Fatal("expected a hard error when production seeding finds zero SQL files")
	}
	if len(rec.executed) != 0 {
		t.Errorf("no seed must execute, got: %v", rec.executed)
	}
}

// TestRunSQLSeeds_ProductionMissingListedReferenceHardErrors proves that a
// manifest entry naming a reference seed absent from disk is a HARD ERROR (not a
// warning) — production must never silently skip required reference data.
func TestRunSQLSeeds_ProductionMissingListedReferenceHardErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSeedFile(t, dir, "001_reference.sql", "REFERENCE;")
	writeReferenceManifest(t, dir, "001_reference.sql", "002_missing.sql") // 002 absent

	rec := &seedRecorder{}
	err := runSQLSeeds(context.Background(), rec, dir, false)
	if err == nil {
		t.Fatal("expected a hard error when a manifest-listed reference seed is absent")
	}
	if !strings.Contains(err.Error(), "002_missing.sql") {
		t.Errorf("error should name the missing seed; got: %v", err)
	}
	if len(rec.executed) != 0 {
		t.Errorf("no seed must execute when a required reference is missing, got: %v", rec.executed)
	}
}

// TestValidateReferenceFilesPresent_RejectsNonRegularFile proves a manifest
// entry that is not a regular file (here: a directory) is rejected. This is the
// cross-platform half of the symlink-escape guard and needs no symlink privilege.
func TestValidateReferenceFilesPresent_RejectsNonRegularFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "001_reference.sql"), 0o755); err != nil { // a DIR named like a seed
		t.Fatalf("mkdir: %v", err)
	}
	err := validateReferenceFilesPresent(dir, map[string]bool{"001_reference.sql": true})
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("expected a not-a-regular-file rejection, got: %v", err)
	}
}

// TestRunSQLSeeds_ProductionRejectsSymlinkedReference proves a manifest-listed
// reference seed that is a SYMLINK (which os.ReadFile would follow, potentially
// to demo/arbitrary SQL outside the root) is refused in production and never
// executes. Skips where the OS/user cannot create symlinks (e.g. Windows without
// privilege).
func TestRunSQLSeeds_ProductionRejectsSymlinkedReference(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outside := t.TempDir()
	// A malicious/demo target OUTSIDE the canonical seed root.
	target := filepath.Join(outside, "evil.sql")
	if err := os.WriteFile(target, []byte("EVIL_SQL;"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	link := filepath.Join(root, "001_reference.sql")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink not supported on this platform/user: %v", err)
	}
	writeReferenceManifest(t, root, "001_reference.sql")

	rec := &seedRecorder{}
	err := runSQLSeeds(context.Background(), rec, root, false)
	if err == nil {
		t.Fatal("expected a hard error for a symlinked reference seed")
	}
	for _, sql := range rec.executed {
		if sql == "EVIL_SQL;" {
			t.Fatalf("a symlinked reference seed must NEVER execute, got: %v", rec.executed)
		}
	}
}

// TestRunSQLSeeds_ProductionRejectsReferenceUnderSymlinkedParent proves that a
// reference seed which is itself a regular file but sits under a symlinked/
// junctioned parent directory that escapes the seed root is rejected (the
// EvalSymlinks-under-root check). Skips where symlinks are unsupported.
func TestRunSQLSeeds_ProductionRejectsReferenceUnderSymlinkedParent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "x.sql"), []byte("OUTSIDE_SQL;"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// root/sub -> outside (a directory symlink/junction escaping the root).
	if err := os.Symlink(outside, filepath.Join(root, "sub")); err != nil {
		t.Skipf("directory symlink not supported on this platform/user: %v", err)
	}
	writeReferenceManifest(t, root, "sub/x.sql")

	rec := &seedRecorder{}
	if err := runSQLSeeds(context.Background(), rec, root, false); err == nil {
		t.Fatal("expected a hard error for a reference seed under an escaping symlinked parent")
	}
	if len(rec.executed) != 0 {
		t.Errorf("nothing must execute, got: %v", rec.executed)
	}
}

func TestIsWithinRoot(t *testing.T) {
	t.Parallel()
	root := filepath.FromSlash("/seeds")
	cases := []struct {
		path string
		want bool
	}{
		{filepath.FromSlash("/seeds"), true},
		{filepath.FromSlash("/seeds/001.sql"), true},
		{filepath.FromSlash("/seeds/investment/01.sql"), true},
		{filepath.FromSlash("/seeds/../evil.sql"), false},
		{filepath.FromSlash("/etc/evil.sql"), false},
		{filepath.FromSlash("/seedsbad/x.sql"), false},
	}
	for _, tc := range cases {
		if got := isWithinRoot(root, filepath.Clean(tc.path)); got != tc.want {
			t.Errorf("isWithinRoot(%q, %q) = %v, want %v", root, tc.path, got, tc.want)
		}
	}
}

func TestLoadReferenceManifest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	body := strings.Join([]string{
		"# a comment",
		"",
		"001_reference.sql",
		"  investment/01_ref.sql  ", // surrounding whitespace trimmed
		"# another comment",
		"./002_reference.sql", // normalized to 002_reference.sql
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, referenceManifestName), []byte(body), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	set, err := loadReferenceManifest(dir)
	if err != nil {
		t.Fatalf("loadReferenceManifest: %v", err)
	}
	for _, want := range []string{"001_reference.sql", "investment/01_ref.sql", "002_reference.sql"} {
		if !set[want] {
			t.Errorf("expected manifest to contain %q; set=%v", want, set)
		}
	}
	if len(set) != 3 {
		t.Errorf("expected 3 entries, got %d: %v", len(set), set)
	}
}

func TestLoadReferenceManifest_MissingIsError(t *testing.T) {
	t.Parallel()
	if _, err := loadReferenceManifest(t.TempDir()); err == nil {
		t.Fatal("expected an error for a missing manifest")
	}
}

func TestLoadReferenceManifest_RejectsUnsafeEntries(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"../escape.sql", "/abs/path.sql"} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, referenceManifestName), []byte(bad+"\n"), 0o644); err != nil {
			t.Fatalf("write manifest: %v", err)
		}
		if _, err := loadReferenceManifest(dir); err == nil {
			t.Errorf("expected an error for unsafe manifest entry %q", bad)
		}
	}
}

func TestIsReferenceFile(t *testing.T) {
	t.Parallel()
	root := filepath.FromSlash("/seeds")
	set := map[string]bool{
		"001_reference.sql":     true,
		"investment/01_ref.sql": true,
	}
	cases := []struct {
		name string
		path string
		want bool
	}{
		{"listed top-level", filepath.Join(root, "001_reference.sql"), true},
		{"listed investment", filepath.Join(root, "investment", "01_ref.sql"), true},
		{"unlisted demo file", filepath.Join(root, "demo", "001_demo.sql"), false},
		{"unlisted top-level", filepath.Join(root, "999_new.sql"), false},
	}
	for _, tc := range cases {
		got, err := isReferenceFile(root, tc.path, set)
		if err != nil {
			t.Fatalf("%s: isReferenceFile: %v", tc.name, err)
		}
		if got != tc.want {
			t.Errorf("%s: isReferenceFile(%q) = %v, want %v", tc.name, tc.path, got, tc.want)
		}
	}
}

// TestResolveDemoSeeding_NoOptIn proves the default (no INCLUDE_DEMO_SEEDS) is
// always reference-only and always succeeds, for any APP_ENV.
func TestResolveDemoSeeding_NoOptIn(t *testing.T) {
	t.Parallel()
	for _, rawAppEnv := range []string{"", "production", "development", "test", "staging"} {
		demoAllowed, err := resolveDemoSeeding(rawAppEnv, false)
		if err != nil {
			t.Errorf("APP_ENV=%q, no opt-in: expected no error, got %v", rawAppEnv, err)
		}
		if demoAllowed {
			t.Errorf("APP_ENV=%q, no opt-in: expected demoAllowed=false, got true", rawAppEnv)
		}
	}
}

// TestResolveDemoSeeding_MissingAppEnvFailsClosed proves an unset/empty APP_ENV
// combined with an explicit demo opt-in hard-errors, not silently behaves as
// development.
func TestResolveDemoSeeding_MissingAppEnvFailsClosed(t *testing.T) {
	t.Parallel()
	demoAllowed, err := resolveDemoSeeding("", true)
	if err == nil {
		t.Fatal("expected an error when INCLUDE_DEMO_SEEDS=true and APP_ENV is unset/empty")
	}
	if demoAllowed {
		t.Error("expected demoAllowed=false on error")
	}
}

// TestResolveDemoSeeding_ExplicitOptInDevTest proves demo seeding is enabled only
// for an explicit development/test APP_ENV alongside the opt-in.
func TestResolveDemoSeeding_ExplicitOptInDevTest(t *testing.T) {
	t.Parallel()
	for _, rawAppEnv := range []string{"development", "test", "DEVELOPMENT", " test "} {
		demoAllowed, err := resolveDemoSeeding(rawAppEnv, true)
		if err != nil {
			t.Errorf("APP_ENV=%q with opt-in: unexpected error: %v", rawAppEnv, err)
		}
		if !demoAllowed {
			t.Errorf("APP_ENV=%q with opt-in: expected demoAllowed=true", rawAppEnv)
		}
	}
}

// TestResolveDemoSeeding_ExplicitOptInProductionHardErrors proves the production
// hard-error: requesting demo seeds while APP_ENV is production (or anything
// other than dev/test) fails closed.
func TestResolveDemoSeeding_ExplicitOptInProductionHardErrors(t *testing.T) {
	t.Parallel()
	for _, rawAppEnv := range []string{"production", "staging", "PRODUCTION"} {
		demoAllowed, err := resolveDemoSeeding(rawAppEnv, true)
		if err == nil {
			t.Errorf("APP_ENV=%q with opt-in: expected a hard error, got none", rawAppEnv)
		}
		if demoAllowed {
			t.Errorf("APP_ENV=%q with opt-in: expected demoAllowed=false on error", rawAppEnv)
		}
	}
}
