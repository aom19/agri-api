# ─── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Instalează dependențe de sistem necesare pentru compilare
RUN apk --no-cache add ca-certificates tzdata git

WORKDIR /app

# Copiază fișierele de dependențe mai întâi (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copiază restul codului
COPY . .

# Compilează binarul (CGO dezactivat pentru imagine minimală)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server ./cmd/api

# ─── Stage 2: Runtime ─────────────────────────────────────────────────────────
FROM alpine:3.20

# Certificate SSL + timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copiază binarul compilat din stage-ul de build
COPY --from=builder /app/server .

# Copiază migratiile (necesare dacă rulezi migrate în container)
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./server"]
