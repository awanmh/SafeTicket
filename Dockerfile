FROM golang:1.25.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o safeticket cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/safeticket .
COPY .env .env

EXPOSE 8080

CMD ["./safeticket"]
