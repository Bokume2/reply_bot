FROM mcr.microsoft.com/devcontainers/go:1.25-trixie
WORKDIR /app

RUN go install github.com/air-verse/air@v1.64.5

COPY go.mod go.sum ./
RUN go mod download
