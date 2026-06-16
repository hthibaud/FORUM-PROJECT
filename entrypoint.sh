#!/bin/sh

# 1. Manage persistence and the location of certificates
mkdir -p /app/certs

if [ ! -f "/app/certs/cert.pem" ] || [ ! -f "/app/certs/key.pem" ]; then
    echo "Certificats TLS manquants. Génération dans le volume persistant..."
    openssl req -x509 -newkey rsa:4096 -keyout "/app/certs/key.pem" -out "/app/certs/cert.pem" -sha256 -days 365 -nodes -subj '/CN=localhost' -addext "subjectAltName = DNS:localhost,IP:127.0.0.1"
fi

# Create symbolic links at the application root (required by router.Start)
ln -sf /app/certs/cert.pem /app/cert.pem
ln -sf /app/certs/key.pem /app/key.pem

# 2. Safety for openenv.Init() which calls os.Exit(1) if the .env file is missing
if [ ! -f "/app/.env" ]; then
    echo "Fichier .env introuvable. Copie de .env.exemple..."
    cp /app/.env.exemple /app/.env
fi

# Execute the main command (the Go server)
exec "$@"