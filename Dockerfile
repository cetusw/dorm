FROM node:22-alpine AS frontend-builder

WORKDIR /frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend ./
RUN npm run build


FROM golang:1.25-alpine AS backend-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend-builder /web/app ./web/app

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/dorm ./cmd/main.go
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/dorm-migrate ./cmd/migrate


FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=backend-builder /out/dorm ./dorm
COPY --from=backend-builder /out/dorm-migrate ./dorm-migrate
COPY --from=backend-builder /app/config.json ./config.json
COPY --from=backend-builder /app/data/mysql/migrations ./data/mysql/migrations
COPY --from=backend-builder /app/web ./web

EXPOSE 8080

CMD ["./dorm"]
