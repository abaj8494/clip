# Adding Photos to Your Gallery

## Quick Start

Your photos endpoint is now live at: **https://clip.abaj.ai/photos**

## Adding Photos

Simply copy image files to the photos directory:

```bash
# Copy individual files
cp /path/to/your/image.jpg /var/www/clip/backend/photos/

# Copy multiple files
cp /path/to/photos/*.{jpg,jpeg,png,gif,webp} /var/www/clip/backend/photos/

# Or use rsync for bulk uploads
rsync -av /path/to/photos/ /var/www/clip/backend/photos/
```

## Supported Formats

- `.jpg` / `.jpeg`
- `.png`
- `.gif`
- `.webp`

## Admin Features

To access admin features (delete photos):

1. Visit https://clip.abaj.ai/photos
2. Click "Admin Mode" button
3. Login with:
   - Username: `aj`
   - Password: `red`
4. Select photos and click "Delete Selected"

## Features

✅ Beautiful responsive gallery grid  
✅ Fullscreen image preview  
✅ Download all photos as ZIP  
✅ Individual photo download  
✅ Admin authentication for deletions  
✅ Search/filter functionality

## Troubleshooting

If photos don't appear:
```bash
# Check the directory
ls -la /var/www/clip/backend/photos/

# Check API response
curl https://clip.abaj.ai/api/photos/list

# Restart container if needed
cd /var/www/clip/backend
docker compose restart
```

## Notes

- Photos are automatically persisted on the host system
- No restart needed after adding new photos
- The gallery automatically refreshes when you visit the page

