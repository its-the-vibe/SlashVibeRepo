# Build stage
FROM golang:1.25.5-alpine AS builder

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the binary with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o slashviberepo .

# Runtime stage using scratch
FROM scratch

# Copy the binary from builder
COPY --from=builder /app/slashviberepo /slashviberepo

# Copy SSL certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Run the binary
ENTRYPOINT ["/slashviberepo"]
