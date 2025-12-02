#!/bin/bash

# Script to complete the rebrand from wiki.abaj.ai to clip.abaj.ai
# Run this script after DNS has been updated

set -e

echo "=== Completing clip.abaj.ai rebrand ==="

# Step 1: Test nginx configuration
echo "Testing nginx configuration..."
/usr/local/openresty/nginx/sbin/nginx -t

# Step 2: Reload nginx
echo "Reloading nginx..."
systemctl reload openresty || /usr/local/openresty/nginx/sbin/nginx -s reload

# Step 3: Remove old Docker container and volumes
echo "Removing old Docker resources..."
docker rm -f backend-wiki-1 2>/dev/null || true
docker volume rm backend_wiki-data backend_wiki-pages 2>/dev/null || true

# Step 4: Build and start the new clip container
echo "Building clip container..."
cd /var/www/clip/backend
docker build -t clip-app .

echo "Starting clip container..."
docker run -d \
  --name clip \
  -p 21313:21313 \
  -v /var/www/clip/persistence:/app/persistence \
  --restart unless-stopped \
  clip-app

echo ""
echo "=== Rebrand Complete! ==="
echo ""
echo "Next steps:"
echo "1. Update DNS to point clip.abaj.ai to this server"
echo "2. Generate SSL certificates:"
echo "   certbot certonly --nginx -d clip.abaj.ai"
echo "3. Update nginx config to use new certificates:"
echo "   Edit /usr/local/openresty/nginx/conf/conf.d/07-clip.conf"
echo "   Update ssl_certificate paths to /etc/letsencrypt/live/clip.abaj.ai/"
echo "4. Reload nginx: systemctl reload openresty"
echo ""
echo "Service is now running at http://localhost:21313"
echo "Once DNS and SSL are configured, it will be available at https://clip.abaj.ai"

