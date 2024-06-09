# Start a new stage from scratch
FROM alpine:3.20.0

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the builder stage
COPY bin/shortcut /app/shortcut

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["/app/shortcut"]