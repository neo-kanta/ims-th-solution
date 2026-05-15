# Runbook — Secrets handling

How the IMS backend obtains and rotates the credentials it needs at runtime.

## Inventory

| Variable                  | Type                | Where it lives                        | Rotation cadence |
| ------------------------- | ------------------- | ------------------------------------- | ---------------- |
| `APP_JWT_SECRET`          | 32+ random bytes    | environment / secret manager          | quarterly        |
| `APP_JWT_SECRET_PREVIOUS` | 32+ random bytes    | environment / secret manager          | rotation overlap |
| `MFA_ENCRYPTION_KEY`      | 64-char hex (AES-256) | environment / secret manager       | yearly           |
| `DB_PASSWORD`             | DB user password    | environment / secret manager          | quarterly        |
| `REDIS_PASSWORD`          | Redis password      | environment / secret manager          | quarterly        |
| `ALPHA_VANTAGE_API_KEY`   | provider key        | environment / secret manager          | on incident      |

## Source of truth (production)

Production runs on a single Oracle Cloud VM today. The credentials live in:

- `/etc/ims/secrets.env` — root-owned, mode `0600`, sourced by the systemd unit at start-up.
- A second copy in the deploy team's password manager (Bitwarden vault `IMS-PROD`) so the file can be re-created if the VM is rebuilt.

There is **no** runtime secrets manager wired into the application yet. Adding HashiCorp Vault or Oracle Cloud Vault is tracked in `docs/adr/00X-secrets-management.md` (TBD) — when that lands, `Load()` in `platform/config/config.go` should resolve secrets through a small abstraction and the env-file path drops to a fallback.

### Who has access

| Role                     | Can read prod secrets | Can rotate |
| ------------------------ | --------------------- | ---------- |
| Engineering Manager      | yes                   | yes        |
| Senior Backend Lead      | yes                   | yes        |
| On-call SRE              | yes (break-glass)     | yes        |
| Other engineers          | no                    | no         |

Access is gated on Bitwarden group membership; IT removes group membership when an engineer rolls off the team.

## Rotation procedure

### `APP_JWT_SECRET`

1. Generate a new secret on a workstation:
   ```bash
   openssl rand -base64 48
   ```
2. Move the **current** `APP_JWT_SECRET` into `APP_JWT_SECRET_PREVIOUS` and bump `APP_JWT_KEY_ID_PREVIOUS` to the previous `APP_JWT_KEY_ID`. Set the new secret + key id.
3. Apply via systemd reload (no downtime):
   ```bash
   sudo systemctl restart ims-backend.service
   ```
4. Wait for the longest active session to expire — `SESSION_ABSOLUTE_LIFE` (default 24h) — then drop `APP_JWT_SECRET_PREVIOUS` (set it to empty) and restart again.
5. Audit: every JWT minted under the old key continues to verify until the rotation window closes; new tokens use the new key. Log lines tagged `jwt.key_id` show which key signed which token.

### `MFA_ENCRYPTION_KEY`

This key encrypts TOTP seeds at rest. Rotation is a **migration** (every row in `iam__mfa` must be re-encrypted) and is therefore not done casually.

1. Plan a maintenance window.
2. Generate the new key (64 hex chars, 32 bytes):
   ```bash
   openssl rand -hex 32
   ```
3. Run the future re-encryption job (`cmd/iam-mfa-rekey`, not yet built) which:
   - Reads each row with the old key.
   - Re-encrypts with the new key.
   - Writes back with version=2.
4. Update `MFA_ENCRYPTION_KEY` and restart the backend.
5. Audit: a successful run logs one `iam.mfa.rekey_complete` line per user.

### `DB_PASSWORD`

1. Create a new password in PostgreSQL **without** dropping the old user:
   ```sql
   ALTER USER ims_app PASSWORD '<new>';
   ```
2. Update `DB_PASSWORD` in the secrets file and restart. There IS a brief connection-pool churn — schedule it during a low-traffic window.

### `ALPHA_VANTAGE_API_KEY`

1. Create a new key in the Alpha Vantage console.
2. Update `ALPHA_VANTAGE_API_KEY` and restart.
3. Revoke the old key in the console.

## Audit trail

Every successful login, MFA enrolment, and admin action emits an `audit__events` row with `actor_id` and `metadata`. Secret rotations themselves are NOT audited inside the application — they're tracked in the deploy team's change log (a Notion page; see Engineering Manager).

## Local development

Defaults in `platform/config/config.go` permit empty `APP_JWT_SECRET`, `MFA_ENCRYPTION_KEY`, and `ALPHA_VANTAGE_API_KEY` only when `APP_ENV=development`. The dev `APP_JWT_SECRET` falls back to `dev-secret-change-in-production`; treat it as compromised by definition. Never re-use a dev secret in any other environment.
