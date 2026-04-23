FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/bin/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /app

COPY --from=builder /app/bin/server .
COPY migrations ./migrations

COPY migrations/run-migrations.sh /app/run-migrations.sh
RUN chmod +x /app/run-migrations.sh

COPY docs ./docs

EXPOSE 8085

CMD ["./server"]
