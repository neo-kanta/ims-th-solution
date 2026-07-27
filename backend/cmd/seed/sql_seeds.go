package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// referenceManifestName is the checked-in file, located AT the canonical seed
// root, that enumerates every production reference/catalog seed by its path
// relative to that root. It is the authoritative, declarative source of which
// seed files are safe to run in production.
//
// Classification is fail-closed by design: when demo seeding is not explicitly
// allowed (production, or any run without the verified dev/test opt-in), a seed
// file runs ONLY if it is listed in this manifest. Every unlisted *.sql file —
// including anything under demo/ or zz_demo/, a newly added file, or a renamed
// directory — is treated as demo and skipped. Adding a demo seed and forgetting
// to touch the manifest therefore fails safe (skipped in production), never
// leaks a demo grant.
//
// This deliberately replaces the previous path-text classification (matching a
// directory named "demo"/"zz_demo" anywhere in the path). Path-text matching was
// unsafe: a symlink/junction into a demo directory, a renamed demo directory, or
// (worst) a repository checked out under a parent directory literally named
// "demo" would all misclassify — the last silently reclassifying EVERY reference
// seed as demo, so production would seed nothing and still exit 0. A manifest
// keyed to paths relative to the resolved canonical root is immune to all of
// these because it does not infer intent from directory names on disk.
const referenceManifestName = "reference_manifest.txt"

// runSQLSeeds walks dir recursively, finds all *.sql files, and executes them in
// lexicographic order using the provided executor. Subdirectory contents follow
// top-level files because path strings sort before deeper paths at the same
// prefix.
//
// demoAllowed gates whether demo (non-reference) seeds run, and must already
// reflect a verified, explicit opt-in (see resolveDemoSeeding in main.go); this
// function does not itself decide whether demo seeding is permitted:
//
//   - When true (verified development/test opt-in), every *.sql file — reference
//     and demo alike — executes in lexicographic order. This preserves
//     development/test/CI behavior exactly and does NOT require a manifest.
//   - When false (the default, and the required production behavior), only files
//     listed in the reference manifest at the canonical seed root execute. Demo
//     files are skipped. Reference files still run and are committed. Demo
//     directories or files merely existing on disk must never by themselves
//     cause a non-zero exit.
//
// Fail-closed contract for the demoAllowed=false path: the reference manifest
// MUST be present and readable at the resolved seed root and MUST list at least
// one entry. If it is missing, unreadable, or empty, runSQLSeeds returns an
// error instead of guessing. This is deliberate — it turns two dangerous silent
// outcomes into a loud, non-zero failure:
//
//   - SEEDS_PATH pointed at a demo subdirectory (no manifest lives there), which
//     would otherwise present demo files at the root and run them; and
//   - a missing/misconfigured manifest, which would otherwise cause production to
//     seed nothing (or everything) with no signal.
func runSQLSeeds(ctx context.Context, exec catalogExecutor, dir string, demoAllowed bool) error {
	if dir == "" {
		return fmt.Errorf("empty seeds directory")
	}

	// Resolve symlinks/junctions on the seed root so classification anchors to
	// the real canonical directory, then walk that resolved path so every file's
	// path relative to it is stable. Falling back to Abs+Clean keeps behavior
	// sane if the path cannot be symlink-resolved (e.g. it does not yet exist),
	// in which case the subsequent walk surfaces the real error.
	rootDir := resolveSeedsRoot(dir)

	files, err := collectSQLSeedFiles(rootDir)
	if err != nil {
		return err
	}

	// Development/test opt-in: run everything, no manifest needed. A zero-file
	// directory is a benign no-op ONLY here — production must never silently
	// succeed on an empty/misconfigured seed directory (handled below).
	if demoAllowed {
		if len(files) == 0 {
			slog.Info("No SQL seed files found", "dir", rootDir)
			return nil
		}
		for _, file := range files {
			if err := execSeedFile(ctx, exec, file); err != nil {
				return err
			}
		}
		return nil
	}

	// Production / no opt-in: fail-closed manifest classification.
	referenceSet, err := loadReferenceManifest(rootDir)
	if err != nil {
		return err
	}
	// Fail closed BEFORE any success path: every reference seed the manifest
	// requires MUST exist on disk. A missing required reference seed — or an
	// empty/misconfigured seed directory that contains none of them — is a hard
	// error in production, never a silent success or a mere warning.
	if err := validateReferenceFilesPresent(rootDir, referenceSet); err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no SQL seed files found under %q, but the reference manifest requires production seeds; refusing to seed", rootDir)
	}

	var skippedDemoFiles int
	for _, file := range files {
		isReference, relErr := isReferenceFile(rootDir, file, referenceSet)
		if relErr != nil {
			return relErr
		}
		if isReference {
			if err := execSeedFile(ctx, exec, file); err != nil {
				return err
			}
			continue
		}

		// Not listed as a reference seed. Decide quiet-skip vs loud-error:
		//   - if it lives in a recognized demo location (demo/ or zz_demo/
		//     relative to the seed root), it is expected demo data -> skip
		//     quietly; this is the normal production outcome.
		//   - otherwise it is an UNLISTED, non-demo seed -> the manifest is out
		//     of sync with the tree (e.g. a new production reference seed was
		//     added without updating the manifest, or a demo directory was
		//     renamed). Fail loudly rather than silently skip a file that might
		//     be required production data.
		//
		// IMPORTANT: this demo-location check is NOT the security classifier and
		// is NOT the rejected path-text design. The manifest alone decides what
		// RUNS; a file is executed only if it is listed. This check only decides
		// whether an UN-run file is skipped silently or reported as a sync error,
		// so a renamed/relocated demo directory can never cause demo data to run
		// — at worst it turns into a loud, non-zero failure.
		inDemoLoc, demoErr := isUnderKnownDemoDir(rootDir, file)
		if demoErr != nil {
			return demoErr
		}
		if inDemoLoc {
			skippedDemoFiles++
			slog.Warn("Skipping demo seed file outside development/test", "file", file)
			continue
		}
		return fmt.Errorf(
			"seed file %q is not listed in %s and is not in a recognized demo location (demo/ or zz_demo/): refusing to seed. If this is a production reference seed, add it to %s; if it is demo data, place it under a demo/ directory",
			file, referenceManifestName, referenceManifestName,
		)
	}

	if skippedDemoFiles > 0 {
		slog.Info("Demo seed files skipped (not development/test)", "count", skippedDemoFiles)
	}

	return nil
}

// validateReferenceFilesPresent hard-fails unless every reference-manifest entry
// resolves to a REGULAR file that physically lives under the canonical seed root.
// This is both a manifest-drift guard (a missing required production seed is a
// hard error, never a silent skip) AND a symlink-escape guard: os.ReadFile (the
// execution path) follows symlinks, so a listed entry that is a symlink — or that
// sits under a symlinked/junctioned parent directory — could redirect production
// execution to demo or arbitrary SQL outside the seed root. Each entry must
// therefore (a) exist, (b) be a regular file (rejects symlinks/dirs/devices via
// Lstat, which does not follow the final component), and (c) have a fully
// symlink-resolved path that remains at or below the resolved canonical root
// (rejects a parent directory junction/symlink that points outside). Problems are
// reported together, sorted, so the error is deterministic.
func validateReferenceFilesPresent(rootDir string, referenceSet map[string]bool) error {
	var problems []string
	for entry := range referenceSet {
		abs := filepath.Join(rootDir, filepath.FromSlash(entry))
		info, err := os.Lstat(abs)
		if err != nil {
			problems = append(problems, entry+" (missing)")
			continue
		}
		if !info.Mode().IsRegular() {
			// Symlink, directory, device, etc. A reference seed must be a plain
			// file so it cannot redirect execution outside the seed root.
			problems = append(problems, entry+" (not a regular file)")
			continue
		}
		resolved, err := filepath.EvalSymlinks(abs)
		if err != nil {
			problems = append(problems, entry+" (unresolvable path)")
			continue
		}
		if !isWithinRoot(rootDir, resolved) {
			problems = append(problems, entry+" (resolves outside the seed root)")
			continue
		}
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf(
		"reference manifest %s has %d invalid entr(y/ies): %s; refusing to seed",
		referenceManifestName, len(problems), strings.Join(problems, ", "),
	)
}

// isWithinRoot reports whether resolved (an absolute, symlink-resolved path) is
// the root itself or a descendant of it. rootDir must already be absolute and
// symlink-resolved (see resolveSeedsRoot). Used to reject any reference seed
// whose real location escapes the canonical seed root.
func isWithinRoot(rootDir, resolved string) bool {
	rel, err := filepath.Rel(rootDir, resolved)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, "../"))
}

// knownDemoDirNames lists directory names that mark a seed file as *expected*
// demo data for the purpose of quiet-skip-vs-loud-error ONLY (see runSQLSeeds).
// This is deliberately NOT the security classifier: the reference manifest alone
// decides which files RUN. A file under one of these directories that is not in
// the manifest is skipped quietly; a file NOT under one of them and not in the
// manifest is a hard error. Renaming a demo directory therefore cannot leak demo
// data into production — it only converts a quiet skip into a loud failure.
var knownDemoDirNames = map[string]bool{
	"demo":    true,
	"zz_demo": true,
}

// isUnderKnownDemoDir reports whether filePath (under rootDir) has any path
// segment, relative to the seed root, matching a known demo directory name.
func isUnderKnownDemoDir(rootDir, filePath string) (bool, error) {
	rel, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		return false, fmt.Errorf("computing seed path %q relative to root %q: %w", filePath, rootDir, err)
	}
	for _, segment := range strings.Split(filepath.ToSlash(filepath.Clean(rel)), "/") {
		if knownDemoDirNames[strings.ToLower(segment)] {
			return true, nil
		}
	}
	return false, nil
}

// resolveDemoSeeding decides whether demo seed files may run, and fails closed
// rather than silently masking a missing APP_ENV. rawAppEnv must be the raw,
// unresolved value of the APP_ENV environment variable (e.g. from
// os.Getenv("APP_ENV")) — NOT a config field that has already substituted a
// default for an unset/empty value, since that would make an unset APP_ENV
// indistinguishable from an explicit "development" and defeat the fail-closed
// contract this function exists to provide.
//
//   - includeDemoSeeds=false (the default: no opt-in) -> demo seeding is disabled
//     and this always succeeds, regardless of rawAppEnv. This is the required
//     production behavior and the required behavior for any run that never asked
//     for demo data.
//   - includeDemoSeeds=true and rawAppEnv is explicitly "development" or "test"
//     (case-insensitive, trimmed) -> demo seeding is enabled.
//   - includeDemoSeeds=true and rawAppEnv is anything else, INCLUDING empty
//     (APP_ENV unset) -> hard error. A missing APP_ENV must never be treated as
//     an implicit "development" for this decision.
func resolveDemoSeeding(rawAppEnv string, includeDemoSeeds bool) (demoAllowed bool, err error) {
	if !includeDemoSeeds {
		return false, nil
	}

	normalized := strings.ToLower(strings.TrimSpace(rawAppEnv))
	if normalized == "development" || normalized == "test" {
		return true, nil
	}

	return false, fmt.Errorf(
		"INCLUDE_DEMO_SEEDS=true requires APP_ENV to be explicitly %q or %q; got APP_ENV=%q",
		"development", "test", rawAppEnv,
	)
}

// execSeedFile reads and executes a single seed file.
func execSeedFile(ctx context.Context, exec catalogExecutor, file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading seed file %q: %w", file, err)
	}
	slog.Info("Executing seed file", "file", file)
	if _, err := exec.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("executing seed file %q: %w", file, err)
	}
	return nil
}

// resolveSeedsRoot returns dir with symlinks resolved and in absolute, cleaned
// form. If symlink resolution fails (for example the directory does not exist),
// it falls back to the absolute, cleaned path so the caller's later walk reports
// the underlying error rather than masking it here.
func resolveSeedsRoot(dir string) string {
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return filepath.Clean(dir)
}

// loadReferenceManifest reads the reference manifest at the root of the seed
// directory and returns the set of reference seed paths (relative to the seed
// root, in forward-slash form). It fails closed: a missing, unreadable, or empty
// manifest is an error, because in the demoAllowed=false path there is no safe
// way to distinguish reference from demo seeds without it.
func loadReferenceManifest(rootDir string) (map[string]bool, error) {
	manifestPath := filepath.Join(rootDir, referenceManifestName)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf(
			"reading reference manifest %q: %w; the manifest is required to classify seeds when demo seeding is disabled — in production SEEDS_PATH must point at the seed root that contains %s (not a subdirectory), and the manifest must list every production reference seed",
			manifestPath, err, referenceManifestName,
		)
	}

	set := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		normalized := filepath.ToSlash(filepath.Clean(filepath.ToSlash(line)))
		// Reject anything that is not a plain relative path under the seed root:
		// current-dir, parent traversal, rooted paths (leading "/" survives
		// Clean+ToSlash on every platform), a Windows drive/UNC absolute path, or
		// a bare "\\" root. This keeps classification anchored to the seed root.
		if normalized == "." ||
			strings.HasPrefix(normalized, "../") ||
			strings.HasPrefix(normalized, "/") ||
			filepath.IsAbs(line) ||
			strings.HasPrefix(line, "\\") {
			return nil, fmt.Errorf(
				"invalid reference manifest entry %q in %q: entries must be seed paths relative to the seed root",
				line, manifestPath,
			)
		}
		set[normalized] = true
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning reference manifest %q: %w", manifestPath, err)
	}
	if len(set) == 0 {
		return nil, fmt.Errorf(
			"reference manifest %q lists no reference seeds; refusing to seed (this would silently skip every seed in production)",
			manifestPath,
		)
	}
	return set, nil
}

// isReferenceFile reports whether filePath (as returned by collectSQLSeedFiles,
// i.e. under rootDir) is a reference seed, by matching its path relative to the
// seed root against the manifest set. Classification is relative to the resolved
// canonical root, so it is unaffected by where on disk the repository lives
// (including a parent directory named "demo").
func isReferenceFile(rootDir, filePath string, referenceSet map[string]bool) (bool, error) {
	rel, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		return false, fmt.Errorf("computing seed path %q relative to root %q: %w", filePath, rootDir, err)
	}
	relSlash := filepath.ToSlash(filepath.Clean(rel))
	if relSlash == "." || strings.HasPrefix(relSlash, "../") {
		// A file that does not live under the seed root should never happen from
		// a WalkDir of that root; treat it as non-reference (skipped) defensively.
		return false, nil
	}
	return referenceSet[relSlash], nil
}

// collectSQLSeedFiles returns every *.sql file under dir in lexicographic order.
// Returned paths are absolute (dir is resolved by resolveSeedsRoot before this
// is called), matching the behaviour of the previous loader.
func collectSQLSeedFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".sql") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking seeds directory %q: %w", dir, err)
	}
	sort.Strings(files)
	return files, nil
}
