package database

import "testing"

func TestEnsureSecureBootstrap_AllowsDevelopmentWithoutDB(t *testing.T) {
	if err := EnsureSecureBootstrap(nil, nil, "development"); err != nil {
		t.Fatalf("expected development bootstrap guard to pass, got %v", err)
	}
}
