# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o bidprentjes-api

# Final stage
FROM alpine:3.18 AS final

# Add non root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bidprentjes-api .

# Copy templates
COPY --from=builder /app/templates ./templates

# Set ownership
RUN chown -R appuser:appuser /app

# Use non root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./bidprentjes-api"] 