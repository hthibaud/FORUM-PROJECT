#!/bin/sh

# 1. Gestion de la persistance et de l'emplacement des certificats
mkdir -p /app/certs

if [ ! -f "/app/certs/cert.pem" ] || [ ! -f "/app/certs/key.pem" ]; then
    echo "Certificats TLS manquants. Génération dans le volume persistant..."
    openssl req -x509 -newkey rsa:4096 -keyout "/app/certs/key.pem" -out "/app/certs/cert.pem" -sha256 -days 365 -nodes -subj '/CN=localhost' -addext "subjectAltName = DNS:localhost,IP:127.0.0.1"
fi

# Création des liens symboliques à la racine de l'application (requis par router.Start)
ln -sf /app/certs/cert.pem /app/cert.pem
ln -sf /app/certs/key.pem /app/key.pem

# 2. Sécurité pour openenv.Init() qui applique un os.Exit(1) si le fichier .env est absent
if [ ! -f "/app/.env" ]; then
    echo "Fichier .env introuvable. Copie de .env.exemple..."
    cp /app/.env.exemple /app/.env
fi

# Exécution de la commande principale (le serveur Go)
exec "$@"