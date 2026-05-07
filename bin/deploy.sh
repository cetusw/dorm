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
scp -r ./bin $SERVER_USER@$SERVER_IP:$SERVER_DIR/
rm dorm-app.tar

echo "Deploying on remote server..."
ssh $SERVER_USER@$SERVER_IP << EOF
cd $SERVER_DIR

echo "Setting execution permissions for scripts..."
chmod +x ./bin/*

echo "Updating PATH environment variable..."
BIN_PATH="\$HOME/usr/www/data/dorm/bin"
if ! grep -q "export PATH=\$PATH:$BIN_PATH" ~/.bashrc; then
    echo "Adding bin directory to .bashrc..."
    echo '' >> ~/.bashrc
    echo '# Add project custom scripts to PATH' >> ~/.bashrc
    echo "export PATH=\$PATH:$BIN_PATH" >> ~/.bashrc
    echo "PATH updated. Please reconnect via SSH or run 'source ~/.bashrc' to apply changes."
else
    echo "PATH is already configured in .bashrc."
fi

# Load the image
docker load -i dorm-app.tar

# Remove tar file
rm dorm-app.tar

# Stop and remove old containers (optional)
docker compose down

# Start database, apply migrations, then start the application
docker compose up -d db
docker compose run --rm migrate
docker compose up -d app adminer

echo "Deployment done!"
EOF
