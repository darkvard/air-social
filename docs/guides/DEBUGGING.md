# Debugging Guide (VSCode)

This guide explains how to debug the Go Application directly on your host machine (VSCode) while running dependent services (DB, Redis, Nginx, etc.) in Docker.

## 1. VSCode Configuration (`launch.json`)

Create or update the file `.vscode/launch.json` in the project root. This tells VSCode how to launch the Go debugger and load the local environment variables.

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Go App",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/api",
      "cwd": "${workspaceFolder}",
      "env": {
        "APP_ENV": "debug"
      },
      "envFile": "${workspaceFolder}/.env.local",
      "args": []
    }
  ]
}
```

## 2. Environment Configuration (`.env.local`)

Create a file named `.env.local` in the project root.
**Important:** All hostnames must be set to `localhost` because the App is running on your machine, not inside the Docker network.

```ini
# App
APP_ENV=debug
APP_PORT=8080
APP_NAME=air-social
APP_DOMAIN=localhost
APP_PROTOCOL=http
APP_BASIC_AUTH_USERNAME=admin
APP_BASIC_AUTH_PASSWORD=password

# Database
DB_USER=postgres
DB_PASS=postgres
DB_NAME=air_social
DB_HOST=localhost
DB_PORT=5432
DB_SSL_MODE=disable
DB_DSN=postgres://${DB_USER}:${DB_PASS}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0
REDIS_PASS=

# JWT
JWT_SECRET=my_secret_key
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=7d
JWT_AUD=air-social
JWT_ISS=air-social-api

# Mailtrap
MAILTRAP_HOST=sandbox.smtp.mailtrap.io
MAILTRAP_PORT=587
MAILTRAP_USERNAME=
MAILTRAP_PASSWORD=
MAILTRAP_FROM_ADDRESS=no-reply@airsocial.com
MAILTRAP_FROM_NAME="Air Social"

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_UI_PORT=15672
RABBITMQ_USER=admin
RABBITMQ_PASS=password
RABBITMQ_URL=amqp://${RABBITMQ_USER}:${RABBITMQ_PASS}@localhost:${RABBITMQ_PORT}/

# MinIO
MINIO_API_PORT=9000
MINIO_CONSOLE_PORT=9001
MINIO_ENDPOINT=localhost:9000
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=password
MINIO_BUCKET_PUBLIC=air-social-media-public
MINIO_BUCKET_PRIVATE=air-social-media-private
MINIO_USE_SSL=false
```

## 3. Debug Workflow

**Hybrid Setup:** Infrastructure (DB, Redis, Nginx, MQ) runs in Docker. The App runs on your Host (IDE).

### Commands

1.  **Start Infrastructure:**

    ```bash
    make debug
    ```

2.  **Run App:**
    Start the Go App in your IDE (Debug Mode).

3.  **Stop Environment:**
    ```bash
    make down
    ```

## 4. Troubleshooting

- **Nginx 502 Bad Gateway:**
  - _Cause:_ Host firewall blocks Docker from accessing IDE port 8080.
  - _Fix (Ubuntu):_ `sudo ufw allow 8080/tcp`

- **Connection Refused (DB/Redis/MQ...):**
  - Ensure `.env.local` uses `localhost` or `127.0.0.1` (not service names).
  - Check if services are running: `make ps`.
