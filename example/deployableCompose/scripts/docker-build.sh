#!/bin/bash
set -e

cd /tmp/$IMAGE

# Cleanup.
sudo rm -rf bin

# Bake bin/* into the resulting image.
sudo docker-compose build --no-cache
# We'll push the app service
sudo docker-compose push app
