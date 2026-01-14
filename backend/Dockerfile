# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o subdomain-finder cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS requests (needed for crt.sh)
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/subdomain-finder .

EXPOSE 8080

CMD ["./subdomain-finder"]
