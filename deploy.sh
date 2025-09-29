#!/bin/bash

SERVER_USER="root"
SERVER_IP=
SERVER_DIR="~/usr/www/data"

echo "Building Docker image..."
docker build -t dorm-app .

echo "Saving image to tar..."
docker save dorm-app -o dorm-app.tar

echo "Copying image and files to server..."
scp dorm-app.tar $SERVER_USER@$SERVER_IP:$SERVER_DIR/
scp docker-compose.yml $SERVER_USER@$SERVER_IP:$SERVER_DIR/
scp .env $SERVER_USER@$SERVER_IP:$SERVER_DIR/
scp config.json $SERVER_USER@$SERVER_IP:$SERVER_DIR/
scp credentials.json $SERVER_USER@$SERVER_IP:$SERVER_DIR/
scp -r mysql $SERVER_USER@$SERVER_IP:$SERVER_DIR/
scp -r data $SERVER_USER@$SERVER_IP:$SERVER_DIR/
scp .env $SERVER_USER@$SERVER_IP:$SERVER_DIR/

echo "Deploying on remote server..."
ssh $SERVER_USER@$SERVER_IP << EOF
cd $SERVER_DIR

# Load the image
docker load -i dorm-app.tar

# Remove tar file
rm dorm-app.tar

# Stop and remove old containers (optional)
docker compose down

# Start new containers
docker compose up -d

echo "Deployment done!"
EOF