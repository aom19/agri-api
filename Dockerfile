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

# Certificate SSL + timezone data, plus un utilizator fără privilegii: serverul nu are
# nevoie de root. UID fix, ca permisiunile pe volume să fie previzibile.
RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S -g 10001 app \
    && adduser -S -D -H -u 10001 -G app app

WORKDIR /app

# Copiază binarul compilat din stage-ul de build (rămâne al lui root, deci read-only pentru app)
COPY --from=builder /app/server .

# Copiază migratiile (necesare dacă rulezi migrate în container)
COPY --from=builder /app/migrations ./migrations

# Singurul director în care scrie aplicația (poze de profil)
RUN mkdir -p uploads/avatars && chown -R app:app uploads

USER app

EXPOSE 8080

CMD ["./server"]
