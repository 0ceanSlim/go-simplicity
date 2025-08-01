# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (needed for go modules)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o simgo cmd/simgo/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/simgo .

# Copy examples
COPY --from=builder /app/examples ./examples

# Make binary executable
RUN chmod +x ./simgo

# Switch to non-root user
USER appuser

# Set entrypoint
ENTRYPOINT ["./simgo"]

# Default command (show help)
CMD ["--help"]

# Metadata
LABEL maintainer="oceanslim@gmx.com"
LABEL description="Go to Simplicity transpiler"
LABEL version="0.1.0"