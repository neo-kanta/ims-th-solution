# Settings Module — Frontend

Identity, audit, and admin surface for the IMS Thailand PoC. Wires
exclusively into `backend/internal/iam` and `backend/internal/audit` —
nothing else.

## Layout


```
features/settings/
├── account.types.ts         # Personal-account / MFA payload shapes
├── admin.types.ts           # Admin user list / status / session shapes
├── audit.types.ts           # Audit event shape + filters
├── ui.types.ts              # UI-only types (KPIs, nav items, demo rows)
│
├── components/              # Vue panels (one per nav section)
│   ├── SettingsControlCenter.vue   # Top-level orchestrator (state wiring)
│   ├── SettingsSectionNav.vue      # Left section nav
│   ├── SettingsOverviewPanel.vue   # KPI grid + API capability map
│   ├── SettingsPersonalAccountPanel.vue
│   ├── SettingsUserDirectory.vue   # Admin user list
│   ├── SettingsUserDetailPanel.vue # Selected user detail + sessions
│   ├── SettingsCreateUserForm.vue
│   ├── SettingsGroupsRolesPanel.vue        # Read-only (sample)
│   ├── SettingsFunctionPermissionsPanel.vue # Read-only (sample)
│   ├── SettingsDataPermissionsPanel.vue     # Read-only (sample)
│   ├── SettingsSecurityPolicyPanel.vue      # Read-only (sample)
│   ├── SettingsNotificationsPanel.vue       # Read-only (sample)
│   ├── SettingsAuditLog.vue
│   ├── SettingsConfirmDialog.vue            # Focus-trapped destructive confirm
│   ├── SettingsMfaEnrollDialog.vue          # QR + recovery codes + verify
│   ├── SettingsMfaDisableDialog.vue         # TOTP confirm
│   ├── SettingsPaginationFooter.vue         # Range + per-page + prev/next
│   └── SettingsPasswordToggle.vue           # Eye/eye-off button
│
├── composables/             # Reactive state + side-effect orchestration
│   ├── useSettingsActiveSection.ts # URL `?section=…` ↔ activeSection
│   ├── useSettingsToasts.ts        # Toast queue with auto-dismiss
│   ├── useSettingsConfirm.ts       # Confirm-dialog state machine
│   └── useSettingsUserMetrics.ts   # KPI totals (3-call backend fallback)
│
├── lib/                     # Pure helpers (no Vue reactivity)
│   ├── audit.ts             # Severity classification (CRITICAL/HIGH/MED/LOW)
│   ├── csv.ts               # RFC 4180 quoting + formula-injection guard
│   └── settingsCatalog.ts   # Demo data for not-yet-live panels
│
└── services/                # Typed API clients (use ~/api/openapi)
    ├── accountApi.ts        # /auth/me, /auth/change-password, sessions
    ├── adminApi.ts          # /admin/users/*
    ├── auditApi.ts          # /admin/audit, /admin/audit/export
    └── mfaApi.ts            # /auth/mfa/{enroll,verify,disable}
```


## Conventions

- **Composables own state.** Components receive props + emit events.
  `SettingsControlCenter` is the wiring layer that connects them.
- **Sample-data panels are clearly marked.** Function permissions, data
  permissions, security policy, and notifications render a yellow
  "Sample data — not connected" banner. They will turn live when the
  backend exposes their respective admin endpoints.
- **Cleartext credentials never enter Vue reactivity.** The reset-password
  flow stores the new password in a module-level `let` inside
  `useSettingsConfirm`, never in a `ref`. It is cleared in the `finally`
  branch of `confirm()` and on `cancelConfirm()`.
- **All times displayed in Asia/Bangkok.** Use `useBangkokFormatter()` from
  `~/shared/composables/useBangkokFormatter`.
- **Pagination range strings are i18n'd.** See `common.pagination.{empty,range,pageSize}`.
- **CSV exports sanitize formula injection** (see `lib/csv.ts`) — same
  rules as the backend audit-export sanitizer.

## Backend dependencies for "live" status

| Panel | Status | Backend gap |
|---|---|---|
| Personal account | ✅ Live | — |
| Other accounts (list/create/lock/disable/reset/sessions) | ✅ Live | — |
| MFA self-service | ✅ Live | — |
| Audit log + export | ✅ Live | — |
| Groups / Roles | ⚠️ Read-only (derived from user list) | Group CRUD + member assignment endpoints |
| Function permissions | ⚠️ Sample only | Permission grant CRUD endpoints |
| Data permissions | ⚠️ Sample only | Data scope grant CRUD endpoints |
| Security policy | ⚠️ Sample only | Security policy admin endpoints |
| Notifications | ⚠️ Sample only | Notification subscription API |
| KPI totals | ⚠️ 3-call fallback | Single `/admin/users/stats` endpoint |
