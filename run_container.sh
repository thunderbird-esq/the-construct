#!/bin/bash

# Sage-68's Container Launcher
# "We are plugging you in."

IMAGE_NAME="matrix-mud"
CONTAINER_NAME="the_construct"

# 1. Ensure Data Directory Exists
# We do this so Docker doesn't create it with root permissions, locking you out.
if [ ! -d "data" ]; then
    echo ">>> Creating data directory..."
    mkdir -p data
fi

# 2. Build the Image
echo ">>> Building the Matrix Image..."
# We use --no-cache to ensure it picks up the latest code changes (like web.go)
docker build --no-cache -t $IMAGE_NAME .

if [ $? -ne 0 ]; then
    echo "!!! Build Failed. Aborting."
    exit 1
fi

# 3. Check for existing container and kill it if running
if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
    echo ">>> Stopping existing Construct instance..."
    docker stop $CONTAINER_NAME
fi

# 4. Launch
echo ">>> Launching The Construct..."
echo "---------------------------------------------------"
echo "   Telnet:  telnet localhost 2323"
echo "   Web HUD: http://localhost:8080"
echo "   Admin:   http://localhost:9090"
echo "---------------------------------------------------"
echo "Press Ctrl+C to disconnect."

docker run --rm -it \
  -p 2323:2323 \
  -p 8080:8080 \
  -p 9090:9090 \
  -v "$(pwd)/data":/root/data \
  --name $CONTAINER_NAME \
  $IMAGE_NAME