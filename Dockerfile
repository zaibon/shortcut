# Use the official Golang image to create a build artifact
FROM golang:1.24-alpine AS builder

# Define build arguments for platform
ARG TARGETPLATFORM
ARG BUILDPLATFORM

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Install just
# RUN apk add just

# Build the Go app for the target platform
# Extract target OS and architecture from TARGETPLATFORM (e.g., linux/amd64)
ARG TARGETOS TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o ./bin/shortcut -v ./cmd/*.go

# Start a new stage from scratch
FROM alpine:latest
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bin/shortcut .

USER appuser

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./shortcut"]
