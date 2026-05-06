package contract

// PermissionDefinition is the canonical declaration of a single function
// permission code owned by a module. cmd/seed iterates each module's
// PermissionCatalog and upserts these rows into permissions_function_definitions
// so the runtime FK on permissions_function_rights can refuse grants for
// unknown codes.
type PermissionDefinition struct {
	Code        string
	Name        string
	Description string
}

// PermissionCatalog is implemented by every module that owns one or more
// function permission codes. The seeder consumes the slice returned by
// Permissions() and tags each row with Module().
//
// Modules with no codes today implement the interface returning an empty
// slice — that keeps the seeder's module list explicit while letting Phase 1+
// fill the slice without changing the wiring.
type PermissionCatalog interface {
	Module() string
	Permissions() []PermissionDefinition
}
