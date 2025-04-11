# Build stage
FROM golang:1.21-alpine AS builder

# Install git and SSL certificates
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

# Install SSL certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy templates and static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Create necessary directories
RUN mkdir -p /app/data

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8000

# Expose port
EXPOSE 8000

# Run the application
CMD ["./main"]
