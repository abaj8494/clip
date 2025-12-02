# Rebrand Status: wiki.abaj.ai → clip.abaj.ai

## ✅ Completed

1. **GitHub Repository** - Renamed from `abaj8494/wiki` to `abaj8494/clip`
2. **Directory Structure** - Moved from `/var/www/wiki` to `/var/www/clip`
3. **Code Updates** - All references updated in:
   - docker-compose.yml (service name: wiki → clip)
   - Dockerfile (binary name: wiki → clip)
   - deploy.sh (paths and service references)
   - index.html (page title)
   - README.md (branding)
4. **Git Remote** - Updated to git@github.com:abaj8494/clip.git
5. **Nginx Configuration** - Created `/usr/local/openresty/nginx/conf/conf.d/07-clip.conf`
6. **Old Config Removed** - Deleted 07-wiki.conf

## 🔄 Remaining Steps

### 1. Update DNS
Point `clip.abaj.ai` A record to your server IP

### 2. Generate SSL Certificates
```bash
certbot certonly --nginx -d clip.abaj.ai
```

### 3. Update Nginx SSL Paths
Edit `/usr/local/openresty/nginx/conf/conf.d/07-clip.conf` and update:
```nginx
ssl_certificate /etc/letsencrypt/live/clip.abaj.ai/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/clip.abaj.ai/privkey.pem;
```

### 4. Reload Nginx
```bash
systemctl reload openresty
```

### 5. Rebuild and Restart the Application
```bash
cd /var/www/clip/backend
./deploy.sh
```

Or manually:
```bash
docker rm -f backend-wiki-1
docker volume rm backend_wiki-data backend_wiki-pages 2>/dev/null || true
cd /var/www/clip/backend
docker build -t clip-app .
docker run -d --name clip -p 21313:21313 \
  -v /var/www/clip/persistence:/app/persistence \
  --restart unless-stopped clip-app
```

### Quick Setup Script
Run: `bash /var/www/clip/complete-rebrand.sh`

## Notes
- The old wiki container (backend-wiki-1) has been stopped but not removed
- Old wiki nginx config has been deleted
- All git commits have been pushed to the renamed repository

