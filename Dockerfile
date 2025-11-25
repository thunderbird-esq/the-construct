# --- Build Stage ---
FROM golang:1.21-alpine AS builder

# Install git (needed for fetching dependencies)
RUN apk add --no-cache git

WORKDIR /app

# 1. Copy EVERYTHING first. 
# We do this so 'go get' and 'go mod tidy' can see the actual code 
# and know which libraries we need.
COPY . .

# 2. Force-fetch the websocket library
RUN go get github.com/gorilla/websocket

# 3. Clean up and download dependencies
RUN go mod tidy
RUN go mod download

# 4. Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o construct .

# --- Runtime Stage ---
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder
COPY --from=builder /app/construct .

# Copy the default data structure
COPY data/ ./data/

# Expose ports
EXPOSE 2323 8080 9090

# Run
CMD ["./construct"]
