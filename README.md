# Air Social

## Requirements
- Docker (includes Docker Compose in modern versions)
- (Optional) GNU Make — for shorter commands

## Commands

| Description | Make command | Docker Compose command |
|-------------|--------------|-------------------------|
| Start all services | `make up` | `docker compose up -d` |
| Stop all services | `make down` | `docker compose down` |
| Restart stack | `make restart` | `docker compose down && docker compose up -d` |
| Rebuild images (no cache) | `make rebuild` | `docker compose build --no-cache && docker compose up -d` |
| Show running services | `make ps` | `docker compose ps` |
| Tail app logs | `make logs` | `docker compose logs -f app` |
| Shell into app container | `make sh-app` | `docker compose exec app sh` |
| Shell into Postgres | `make sh-db` | `docker compose exec db sh` |
| Shell into Redis | `make sh-redis` | `docker compose exec redis sh` |
| Build binary for Air reload | `make air-build` | `go build -o ./tmp/main -buildvcs=false .` |

## Notes
- The **app** container runs using **Air + volume mount** → automatic rebuild & reload on code changes.
- **Postgres** and **Redis** use **named volumes**, so data persists across restarts.
- If Make is not installed, you can use the Docker Compose equivalents.
