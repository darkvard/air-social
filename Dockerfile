# Base Go image for development (includes Go compiler + tools)
FROM golang:1.25

# Application working directory inside the container
WORKDIR /app

# Install Air tool
RUN go install github.com/air-verse/air@latest

# Install migrate CLI tools
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.0/migrate.linux-amd64.tar.gz | tar xvz && mv migrate /usr/local/bin/


# Copy Go module files first for caching benefits (go.mod, go.sum)
COPY go.* ./
RUN go mod download

# Copy project source code
COPY . .

# -------- DEV MODE (Hot Reload) --------
# Air will monitor file changes (when using volume mount)
# Use this mode with:
#   docker run -v $(PWD):/app image_name
CMD ["air"]


# -------- OPTIONAL: RUN MODE (No Reload) --------
# Uncomment the two lines below to build & run the binary directly:
#
# RUN go build -o server .
# CMD ["./server"]
#
# This mode is used when you don't mount source code
# and want a normal (non–hot-reload) execution.
