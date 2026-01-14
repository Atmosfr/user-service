# syntax=docker/dockerfile:1
FROM golang:1.25.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /user-service ./cmd/user-service

COPY migrations /app/migrations

FROM alpine:latest

RUN apk add --no-cache netcat-openbsd

WORKDIR /app

COPY --from=builder /user-service /user-service
COPY --from=builder /app/migrations /app/migrations

COPY wait-for-db.sh /wait-for-db.sh
RUN chmod +x /wait-for-db.sh

EXPOSE 8080

ENTRYPOINT ["/wait-for-db.sh"]
CMD ["/user-service"]
