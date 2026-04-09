# build stage
FROM golang:1.25.8-alpine3.23 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o /bin/seeds ./scripts/seeds.go
RUN go build -o /bin/main ./cmd/main.go

# run stage
FROM alpine:3.23
WORKDIR /app

COPY --from=build /bin/seeds /bin/main /bin/
