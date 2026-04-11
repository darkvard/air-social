# Base Go image for development (includes Go compiler + tools)
FROM golang:1.25-alpine

# Application working directory inside the container
WORKDIR /app

# Alpine does not include these tools by default
# curl: needed to download the migrate binary
# tar: needed to extract the downloaded .tar.gz archive
# make: needed to run Makefile commands (e.g. make air-build used by Air)
RUN apk add --no-cache curl tar make

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

# Install migrate CLI — required because migrations run inside this container
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.0/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate /usr/local/bin/

# Copy module files first — cached until go.mod or go.sum changes
COPY go.* ./
RUN go mod download

# Copy source code last — this layer changes most often
COPY . .

# Dev mode: Air watches for file changes and rebuilds automatically
CMD ["air"]


# -------- OPTIONAL: PRODUCTION (Multi-stage build) --------
# Uncomment to build a minimal production image (~15MB):
#
# FROM golang:1.25-alpine AS builder
# WORKDIR /app
# RUN apk add --no-cache curl tar
# COPY go.* ./
# RUN go mod download
# COPY . .
# RUN CGO_ENABLED=0 go build -buildvcs=false -o server ./cmd/api
#
# FROM alpine:3.22
# WORKDIR /app
# COPY --from=builder /app/server .
# CMD ["./server"]
