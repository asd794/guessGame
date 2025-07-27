# Dockerfile.build
FROM golang:1.24.5-bookworm

WORKDIR /app
COPY . .

# 靜態編譯，適用 Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app-linux main.go
