# Module: compliance

This module is the reserved backend boundary for compliance and IRG-related rules.

## Intended Scope

- blacklist and whitelist checks
- product and instrument restrictions
- ratio and mandate validation
- compliance hooks around investment decisions and execution

## Current Status

This module is scaffolded only.

It currently provides folder structure and placeholder wiring points, but no real domain logic, persistence, or routes.

## Notes

- keep cross-cutting compliance decisions here once implemented
- avoid mixing future compliance rules directly into `investment` handlers
- update this README when the first real rule set lands
