FROM golang:1.21-alpine AS builder

WORKDIR /app

# Force rebuild on every deployment
RUN echo "Rebuild timestamp: $(date)"

# Copy source code and dependencies
COPY go.mod go.sum ./
RUN go mod download
COPY wiki.go backup.go ./
COPY edit.html view.html index.html media.html manifest.json ./
COPY icon/ ./icon/

# Build the application
RUN go build -o clip .

# Use a smaller image for the final container
FROM alpine:latest

WORKDIR /app

# Copy the binary and template files
COPY --from=builder /app/clip /app/clip
COPY --from=builder /app/edit.html /app/edit.html
COPY --from=builder /app/view.html /app/view.html
COPY --from=builder /app/index.html /app/index.html
COPY --from=builder /app/media.html /app/media.html
COPY --from=builder /app/manifest.json /app/manifest.json
COPY --from=builder /app/icon/ /app/icon/

# Create directories
RUN mkdir -p /app/files
RUN mkdir -p /app/persistence

# Expose the port
EXPOSE 21313

# Run the application
CMD ["/app/clip"]