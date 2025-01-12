# ---------------------------------------
# 1) Build stage
# ---------------------------------------
    FROM golang:1.23-alpine AS builder

    WORKDIR /app
    
    # Copy Go modules first
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy remaining source
    COPY . .
    
    # Build the Go binary
    RUN go build -o server main.go
    
    # ---------------------------------------
    # 2) Final stage
    # ---------------------------------------
    FROM alpine:3.18
    
    # Install ffmpeg
    RUN apk add --no-cache ffmpeg
    
    # Create user
    RUN addgroup -S appgroup && adduser -S appuser -G appgroup
    
    WORKDIR /app
    COPY --from=builder /app/server /app/server
    
    # COPY .env ./
    
    RUN chown -R appuser:appgroup /app
    USER appuser
    
    EXPOSE 8080
    
    ENTRYPOINT ["./server"]
    