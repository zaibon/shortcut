# Stage 1: Build Assets
FROM oven/bun:alpine AS assets-builder
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install
COPY . .
RUN bun run build

# Stage 2: Build Go App
FROM golang:1.24-alpine AS builder

# Define build arguments for platform
ARG TARGETPLATFORM
ARG BUILDPLATFORM

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source
COPY . .

# Copy built assets
COPY --from=assets-builder /app/static/dist ./static/dist

# Convert migrations to sequential order
RUN go tool goose -dir=db/migrations fix

# Build the Go app for the target platform
ARG TARGETOS TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o ./bin/shortcut -v ./cmd/*.go

# Stage 3: Final Image
FROM alpine:latest
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file
COPY --from=builder /app/bin/shortcut .

USER appuser

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./shortcut"]
