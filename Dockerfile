FROM node:22-alpine AS frontend-builder

WORKDIR /frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend ./
RUN npm run build


FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend-builder /web/app ./web/app

RUN CGO_ENABLED=0 go build -o ./dorm ./cmd/main.go
RUN CGO_ENABLED=0 go build -o ./dorm-migrate ./cmd/migrate


FROM alpine:latest

RUN apk add --no-cache tzdata

WORKDIR /root/

COPY --from=builder /app/dorm .
COPY --from=builder /app/dorm-migrate .
COPY --from=builder /app/config.json .
COPY --from=builder /app/credentials.json .
COPY --from=builder /app/data/mysql/migrations ./data/mysql/migrations
COPY --from=builder /app/data/mysql/my.cnf /etc/mysql/conf.d/my.cnf
COPY --from=builder /app/web ./web

EXPOSE 8080
CMD ["./dorm"]