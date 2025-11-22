# Air Social

## Requirements
- Docker (includes Docker Compose)
- GNU Make (recommended for shorter commands)

---

## Makefile Commands (Summary)

| Description | Make command | Docker Compose command |
|-------------|--------------|-------------------------|
| Start all services | `make up` | `docker compose up -d` |
| Stop all services | `make down` | `docker compose down` |
| Restart stack | `make restart` | `docker compose down && docker compose up -d` |
| Rebuild images (no cache) | `make rebuild` | `docker compose build --no-cache && docker compose up -d` |
| Show running services | `make ps` | `docker compose ps` |
| Tail app logs | `make logs` | `docker compose logs -f app` |
| Shell into app | `make sh-app` | `docker compose exec app sh` |
| Shell into Postgres | `make sh-db` | `docker compose exec db sh` |
| Shell into Redis | `make sh-redis` | `docker compose exec redis sh` |
| Build binary for Air reload | `make air-build` | `go build -o ./tmp/main -buildvcs=false .` |

---

## Migration Commands

| Description | Make command | Notes |
|------------|--------------|-------|
| Show current version | `make migrate-version` | Prints applied version or “no migrations yet” |
| Create new migration | `make migrate-create name=add_users` | Generates `xxx_add_users.up.sql` & `.down.sql` |
| Apply migrations | `make migrate-up` | Applies al* pending migrations |
| Apply N migrations | `make migrate-up n=1` | Runs only the next n migrations |
| Rollback migrations | `make migrate-down` | Rolls back all applied migrations |
| Rollback N migrations | `make migrate-down n=1` | Rolls back n last migrations |
| Force version | `make migrate-force version=12` | Fixes dirty state; does NOT change schema |

---

## Notes
- The app service runs with Air + volume mount, so source code changes reload automatically.
- Postgres and Redis use named volumes → persistent between restarts.
- Migration commands run inside the app container, using the shared migrations folder.
- If Make is not installed, you can use equivalent Docker Compose commands.
