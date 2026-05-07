// Package e2e holds end-to-end integration tests that boot the full module
// graph (audit + iam + compliance + workflow + investment + market_data)
// against a live Postgres test database and exercise the wired chi router
// via httptest.NewServer.
//
// Tests skip cleanly when no live database is available, matching the
// project pattern used by position_repository_concurrency_test.go. To run:
//
//	IMS_TEST_DATABASE_DSN="host=localhost port=5437 ..." \
//	    go test -tags e2e -count=1 -v ./tests/e2e/...
//
// The default DSN points at the dev database. Tests assume `make migrate-up`
// + `make seed` have been run so reference data (asset classes, countries,
// fund categories, etc.) and the seeded admin user exist.
package e2e
