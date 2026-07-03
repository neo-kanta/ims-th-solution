package chat_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// TestChatModuleBoundary enforces correction 5: chat reaches all
// business/financial data only via MCP and never imports another module's
// internal/ package. The only permitted cross-module in-process imports are:
//
//   - backend/pkg/contract — the explicit cross-module contract surface.
//   - backend/internal/audit/domain — the audit Recorder port, the
//     established cross-module audit dependency every business module uses.
//   - backend/platform/* — shared platform code (middleware, config, etc.).
//
// Any other backend/internal/<module>/ import from inside chat fails this
// test. Adding a new allowed exception requires updating allowedInternal
// AND explaining the exception in the chat module's package doc.
func TestChatModuleBoundary(t *testing.T) {
	const (
		repoModule      = "github.com/neo-kanta/ims-th-solution/backend"
		chatPkgPrefix   = repoModule + "/internal/chat"
		internalPkgGlob = repoModule + "/internal/"
		platformPkgGlob = repoModule + "/platform/"
		contractPkgGlob = repoModule + "/pkg/"
	)

	allowedInternal := map[string]bool{
		repoModule + "/internal/audit/domain": true,
	}

	fset := token.NewFileSet()
	chatRoot := "."

	var violations []string

	err := filepath.WalkDir(chatRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// boundary test itself imports go/parser etc. — skip test files so
		// we audit production code only.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			pkg := strings.Trim(imp.Path.Value, `"`)

			// Self-imports inside chat are always fine.
			if strings.HasPrefix(pkg, chatPkgPrefix) {
				continue
			}

			// Allowed: backend/platform/* and backend/pkg/*
			if strings.HasPrefix(pkg, platformPkgGlob) || strings.HasPrefix(pkg, contractPkgGlob) {
				continue
			}

			// Forbidden: backend/internal/<other-module>/* unless explicitly allowed.
			if strings.HasPrefix(pkg, internalPkgGlob) {
				if !allowedInternal[pkg] {
					violations = append(violations,
						path+": forbidden import "+pkg)
				}
				continue
			}

			// Any non-repo import (stdlib, third-party) is fine.
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk chat module: %v", err)
	}

	if len(violations) > 0 {
		t.Errorf("chat module boundary violations:\n  %s", strings.Join(violations, "\n  "))
	}
}
