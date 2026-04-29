# Getting Started

Setup the environment and run the full application stack using Docker.

## 1. Configuration

Create a `.env` file in the root directory. This configuration is optimized for the **Docker Internal Network**, allowing services to communicate via hostnames.

```ini
# Server
SERVER_ENV=dev
SERVER_PORT=8080
SERVER_NAME=air-social
SERVER_DOMAIN=localhost
SERVER_PROTOCOL=http
SERVER_READ_TIMEOUT=5s
SERVER_WRITE_TIMEOUT=10s
SERVER_BASIC_AUTH_USERNAME=admin
SERVER_BASIC_AUTH_PASSWORD=password

# Database
DB_USER=postgres
DB_PASS=postgres
DB_NAME=air_social
DB_HOST=postgresdb
DB_PORT=5432
DB_SSL_MODE=disable
DB_MAX_IDLE=5
DB_MAX_OPEN=10
DB_MAX_LIFETIME=1h
DB_MAX_IDLE_TIME=15m
DB_DSN=postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}

# Redis
REDIS_HOST=redis
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
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_UI_PORT=15672
RABBITMQ_USER=admin
RABBITMQ_PASS=password
RABBITMQ_URL=amqp://${RABBITMQ_USER}:${RABBITMQ_PASS}@rabbitmq:${RABBITMQ_PORT}/

# MinIO
MINIO_API_PORT=9000
MINIO_CONSOLE_PORT=9001
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=password
MINIO_ENDPOINT=minio:9000
MINIO_BUCKET_PUBLIC=air-social-media-public
MINIO_BUCKET_PRIVATE=air-social-media-private
MINIO_USE_SSL=false

# MongoDB
MONGO_USER=admin
MONGO_PASS=password
MONGO_URI=mongodb://${MONGO_USER}:${MONGO_PASS}@mongodb:27017
MONGO_DB=air_social_chat
MONGO_CONNECT_TIMEOUT=10s
MONGO_MAX_POOL=100
MONGO_MIN_POOL=5
MONGO_MAX_IDLE_TIME=10m
MONGO_HEARTBEAT_INTERVAL=5s
MONGO_SERVER_SELECTION_TIMEOUT=5s
```

## 2. Setup

Install dev tools and build Docker images:

```bash
make setup
```

## 3. Run

```bash
make setup   # Install tools, build images, start stack, apply migrations
make seed    # (optional) Seed database with dummy data — run make up after to restart the full stack
```

Once the stack is up, the API is live at `http://localhost/air-social/api/v1/swagger/index.html`.

### Other useful commands

```bash
make down         # Stop all services
make rebuild      # Rebuild images after code changes
make seed         # Re-seed database (wipes existing data)
```

## 3. Verification

After the system is up, you can access the services via Nginx (Port 80):

| Service           | URL                                                   | Credentials          |
| :---------------- | :---------------------------------------------------- | :------------------- |
| **API Swagger**   | http://localhost/air-social/api/v1/swagger/index.html | N/A                  |
| **Health Check**  | http://localhost/air-social/api/v1/health             | `admin` / `password` |
| **RabbitMQ UI**   | http://localhost/rabbitmq/                            | `admin` / `password` |
| **MinIO Console** | http://localhost/storage-admin/                       | `admin` / `password` |
