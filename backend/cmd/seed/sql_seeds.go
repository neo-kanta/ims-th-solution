package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// runSQLSeeds walks dir recursively, finds all *.sql files and executes them
// in lexicographic order using the provided executor. Subdirectory contents
// follow top-level files because path strings sort before subdirectory paths
// at the same depth.
func runSQLSeeds(ctx context.Context, exec catalogExecutor, dir string) error {
	if dir == "" {
		return fmt.Errorf("empty seeds directory")
	}

	files, err := collectSQLSeedFiles(dir)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		slog.Info("No SQL seed files found", "dir", dir)
		return nil
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading seed file %q: %w", file, err)
		}
		slog.Info("Executing seed file", "file", file)
		if _, err := exec.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("executing seed file %q: %w", file, err)
		}
	}
	return nil
}

// collectSQLSeedFiles returns every *.sql file under dir in lexicographic
// order. Returned paths are absolute relative to the working directory the
// seeder was invoked from, matching the behaviour of the previous loader.
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
