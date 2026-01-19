# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o main .

# Run stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy templates and static files
COPY templates ./templates
COPY static ./static

# Create directory for database
RUN mkdir -p /data

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8080

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
