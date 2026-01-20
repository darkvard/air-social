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

## 3. Docker Override (`docker-compose.override.yml`)

Create a file named `docker-compose.override.yml`.
This file exposes ports to your host machine and configures Nginx to talk back to your host (where the Debugger is running).

```yaml
services:
  nginx:
    ports:
      - "80:80"
    environment:
      - NGINX_TIMEOUT=3600s  
    extra_hosts:
      - "app:host-gateway"
      - "minio:host-gateway"
      - "rabbitmq:host-gateway"

  db:
    ports:
      - "5432:5432"

  redis:
    ports:
      - "6379:6379"

  rabbitmq:
    ports:
      - "5672:5672"
      - "15672:15672"

  minio:
    ports:
      - "9000:9000"
      - "9001:9001"
```

## 4. How to Debug

### Step 1: Start Infrastructure
Run the following command to start all services (DB, Redis, Nginx...) **EXCEPT** the App service.

```bash
make debug
```

### Step 2: Start App (Debugger)
In VSCode, ensure **"Debug Go App"** is selected and press **F5**.
The app will start on port `8080`.
### Step 3: Stop
When you are done, run:

```bash
make down
```

## 5. Troubleshooting

### Firewall Issues
If Nginx (running in Docker) cannot connect to your App (running on Host), you might see a **502 Bad Gateway** error.

This is often due to the Firewall blocking the connection from the Docker Network to your Host's port 8080.

**Fix:** Allow traffic on port 8080.
```bash
# Ubuntu (UFW)
sudo ufw allow 8080/tcp
sudo ufw reload
```

## 6. Switching Back to Normal Mode

When running the app normally (inside Docker), you **MUST disable** `docker-compose.override.yml`, otherwise Nginx causes a **502 Bad Gateway**.

### The Reason (Host Gateway)
The override file uses `extra_hosts: host-gateway`, forcing Nginx to route traffic to your **Host Machine** (where the Debugger was), instead of the **Docker Container**.
* **Debug Mode:** Nginx -> Host Machine (VSCode) ✅
* **Normal Mode (with override active):** Nginx -> Host Machine (Empty) ❌ -> **502 Error**

### Solution
**Option 1: Comment Out**
Comment out all lines in `docker-compose.override.yml` to disable the `host-gateway` routing.

**Option 2: Delete File**
```bash
rm docker-compose.override.yml