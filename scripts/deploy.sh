#!/bin/bash

# Script to deploy the project using Docker Compose

# Step 1: Bring down the current containers
echo "Stopping and removing current containers..."
docker-compose down

# Step 2: Build the Docker images
echo "Building Docker images..."
docker-compose build

# Step 3: Bring the containers back up in detached mode
echo "Starting containers..."
docker-compose up -d

echo "Deployment complete!"
