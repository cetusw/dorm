FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o ./main ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache tzdata

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/config.json .
COPY --from=builder /app/credentials.json .
COPY --from=builder /app/data/mysql/migrations ./data/mysql/migrations
COPY --from=builder /app/data/mysql/my.cnf /etc/mysql/conf.d/my.cnf

EXPOSE 8080

CMD ["./main"]