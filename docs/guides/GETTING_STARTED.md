# Getting Started

Setup the environment and run the full application stack using Docker.

## 1. Configuration

Create a `.env` file in the root directory. This configuration is optimized for the **Docker Internal Network**, allowing services to communicate via hostnames.

```ini
# App
APP_ENV=production
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
DB_HOST=db
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
```

## 2. Build & Run

Use Makefile commands to manage the application lifecycle.

### First Run (Build Images)
For the first setup, you need to build the Go application image from the Dockerfile.

```bash
make rebuild
```

### Start Application
Once images are built, use this command to start all containers (App, DB, Redis, Queue, Storage).

```bash
make up
```

### Stop Application
Stop and remove all containers and networks.

```bash
make down
```

## 3. Verification

After the system is up, you can access the services via Nginx (Port 80):

| Service | URL | Credentials |
| :--- | :--- | :--- |
| **API Swagger** | http://localhost/air-social/api/v1/swagger/index.html | N/A |
| **RabbitMQ UI** | http://localhost/rabbitmq/ | `admin` / `password` |
| **MinIO Console** | http://localhost/storage-admin/ | `admin` / `password` |