# Compilation du binaire
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copie du code source et compilation statique
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o forum-server ./cmd/main.go

# Image d'exécution minimale
FROM alpine:latest

# Installation d'openssl
RUN apk --no-cache add ca-certificates openssl

WORKDIR /app

# Récupération du binaire compilé
COPY --from=builder /app/forum-server .

# Récupération des configurations et des assets indispensables
COPY --from=builder /app/config.json .
COPY --from=builder /app/.env .
COPY --from=builder /app/template/ ./template/
COPY --from=builder /app/static/ ./static/
COPY --from=builder /app/entrypoint.sh .

# Droits d'exécution sur le script d'entrée
RUN chmod +x entrypoint.sh

# Ports HTTP et HTTPS définis dans config.json
EXPOSE 80 443

ENTRYPOINT ["./entrypoint.sh"]
CMD ["./forum-server"]