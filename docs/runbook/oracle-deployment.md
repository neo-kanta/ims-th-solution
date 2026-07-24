# Runbook — Oracle Cloud backend deployment

Real, on-host layout for the production backend VM, captured 2026-07-22 while
diagnosing a stuck deploy. Supersedes the systemd/`/etc/ims/secrets.env`
description in [secrets.md](secrets.md) for this host — that description does
not match reality (see "Known issues" below).

## Host

- Hostname: `ims-vnic` (Oracle Cloud VM)
- SSH user: `ubuntu`, working directory `/home/ubuntu`
- No systemd units are used for the backend (`systemctl list-units --type=service | grep -i ims` returns nothing). Deployment is pure Docker, driven entirely by `.github/workflows/deploy-backend-oracle.yml` over SSH (`appleboy/ssh-action`).

## Files in `/home/ubuntu`

| File | Owner | Mode | Purpose |
| --- | --- | --- | --- |
| `ims-backend.env` | ubuntu:ubuntu | 664 | The runtime env file passed to every backend container via `--env-file`. This is `$ORACLE_BACKEND_ENV_FILE`. |
| `ims-backend.env.backup.20260717092928` | root:root | 644 | A one-off backup taken 2026-07-17, the same day the current incident started. Someone was already investigating. |
| `ims-th-solution-backend.tar` | ubuntu:ubuntu | — | Stale build artifact from April; unrelated to the current image-based deploy flow. Safe to remove once confirmed unused. |

## Docker layout

- **Network**: `ims-net` (bridge) — this is `$ORACLE_DOCKER_NETWORK`.
- **Containers**:
  - `postgres` — image `postgres:16`, host port `5437 -> 5432`, up continuously for 3 months. This is the persistent database; it is never touched by the deploy workflow itself.
  - `ims-backend` / `ims-backend-new` — the app container. The deploy workflow always stages the new container as `<name>-new`, health-checks it, then renames it to the live name only after `curl http://127.0.0.1:$HOST_PORT/health` succeeds. `<name>` is `$ORACLE_BACKEND_CONTAINER_NAME`.
- **Images**: tagged `ghcr.io/neo-kanta/ims-th-backend:<git-sha>` plus a rolling `:latest`. Several stale SHA tags accumulate on the host over time (the workflow's final `docker image prune -f` only removes dangling/untagged layers, not old tags) — periodic manual `docker image prune -a` is a reasonable housekeeping task.
- **DB connection from the host**: `sudo docker exec -it postgres psql -U <DB_USER> -d <DB_NAME>` — no need for a throwaway network-attached container since `postgres` is a named, addressable container.

## Deploy workflow steps (for reference)

1. `test-build-push` job: `go test`, `go build`, then build+push the Docker image to GHCR.
2. `deploy` job, over SSH:
   - Pull the new image.
   - Run a one-shot `<name>-migrate` container (`--rm`, `--entrypoint /app/migrate ... up`) as a migration preflight.
   - Remove any leftover `<name>-new` / `<name>` containers.
   - Start the new image detached as `<name>-new`.
   - `sleep 10`, then curl the health endpoint.
   - Rename `<name>-new` -> `<name>` only on success; `trap ... ERR` cleans up and dumps logs on any failure.

## Known issues (open as of 2026-07-22)

1. **Dirty migration** — `schema_migrations` is stuck at version `20260615000003`, blocking every deploy since at least 2026-07-17. See migration-fix runbook (chat/PR) for the resolution steps.
2. **Production has effectively been down for ~5 days.** `ims-backend-new` (image `a9c84684d684...`, created 2026-07-17) has been crash-looping every ~60s since creation: its entrypoint runs `migrate up` on every boot, hits the dirty-database guard, and exits 1; `--restart unless-stopped` just keeps retrying. No container is currently running under the plain `ims-backend` name — the last deploy never passed its health check and rename step.
3. **`APP_ENV=development` in the production env file.** `/home/ubuntu/ims-backend.env` has `APP_ENV=development`, `APP_JWT_SECRET=dev-secret-change-in-production` (the publicly-documented example default), and dev-style DB naming (`ims_dev` / `ims_app`). Practical effects:
   - `EnsureSecureBootstrap`'s check for default admin/admin2 credentials only runs outside `development`, so it's silently skipped on this host.
   - `config.go`'s requirement that `MFA_ENCRYPTION_KEY` / `ALPHA_VANTAGE_API_KEY` be set outside development is also skipped.
   - The JWT signing secret is a known, publicly-documented value.
   **Recommended follow-up** (not blocking the migration fix): set `APP_ENV=production` (or a real `staging` label), and rotate `APP_JWT_SECRET` and `DB_PASSWORD` to unique generated values — see [secrets.md](secrets.md) rotation procedures.
4. **Env file permissions are wider than intended.** `ims-backend.env` is `ubuntu:ubuntu 664` (group- and world-readable on the host), looser than the root-owned `0600` posture [secrets.md](secrets.md) describes. Tighten to `600` once secrets are rotated.
5. **[secrets.md](secrets.md) is stale for this host** — it describes a systemd unit reading `/etc/ims/secrets.env`; neither exists. Needs a correction pass once the above is resolved.

## Quick reference — variables used by the deploy workflow

| Workflow variable | Real value on this host |
| --- | --- |
| `ORACLE_HOST` | `ims-vnic` (Oracle Cloud VM) |
| `ORACLE_USER` | `ubuntu` |
| `ORACLE_DOCKER_NETWORK` | `ims-net` |
| `ORACLE_BACKEND_ENV_FILE` | `/home/ubuntu/ims-backend.env` |
| `ORACLE_BACKEND_CONTAINER_NAME` | (base name whose `-new` / plain forms appear in `docker ps -a`, e.g. `ims-backend`) |
| `ORACLE_BACKEND_HOST_PORT` | not directly observed (container currently down); check the secret or a prior successful `docker ps` port mapping |
