# Photos Feature Documentation

## Overview

The `/photos` endpoint provides a photo gallery interface with admin capabilities for managing images.

## Features

### 1. **Image Gallery**
- Clean, responsive grid layout similar to photos-example.html
- Lazy loading for optimal performance
- Search/filter functionality

### 2. **Download All**
- Downloads all images as a single ZIP file
- Uses JSZip library for client-side compression
- Preserves original filenames

### 3. **Enhanced Preview**
- Click any image to open full-screen preview
- **Close button (✕)** in top-left corner
- **Download button** in top-right corner
- Click outside image to close
- Smooth transitions

### 4. **Admin Mode**
- **Admin button** triggers login modal
- **Credentials:**
  - Username: `aj`
  - Password: `red`
- Enables selection and deletion of images
- Checkboxes appear in admin mode
- Multi-select with visual feedback (red border)
- **Delete Selected** button for batch deletion
- **Exit Admin** button to return to normal mode

## API Endpoints

### GET `/photos`
Serves the photos gallery HTML page

### GET `/api/photos/list`
Returns JSON array of all images in the photos directory

**Response:**
```json
{
  "images": ["image1.jpg", "image2.png", ...]
}
```

### POST `/api/photos/delete`
Deletes selected images (requires authentication)

**Request:**
```json
{
  "username": "aj",
  "password": "red",
  "files": ["image1.jpg", "image2.png"]
}
```

**Response:**
- 200 OK: Files deleted successfully
- 401 Unauthorized: Invalid credentials
- 400 Bad Request: Invalid request format

### GET `/photos/{filename}`
Serves individual photo files

## Setup

### 1. Add Photos
Place images in the `/app/photos/` directory (inside container) or configure volume mount:

```yaml
volumes:
  - /path/to/your/photos:/app/photos
```

### 2. Supported Formats
- JPG/JPEG
- PNG
- GIF
- WebP

### 3. File Structure
```
/var/www/clip/
├── backend/
│   ├── photos.html      # Gallery interface
│   ├── wiki.go          # Backend with photos endpoints
│   └── photos/          # Photos directory (auto-created)
```

## Security

- Admin authentication required for deletion
- Path traversal protection
- File type validation
- CORS enabled for cross-origin requests

## Usage

1. **Access the gallery:**
   - Visit `https://clip.abaj.ai/photos` (or `http://localhost:21313/photos`)

2. **View photos:**
   - Browse the grid
   - Click any image for full-screen preview
   - Use search box to filter

3. **Download:**
   - Click "Download All" to get a ZIP of all images
   - Or click individual image → Download button in preview

4. **Admin mode:**
   - Click "Admin" button
   - Enter credentials (aj/red)
   - Select images with checkboxes
   - Click "Delete Selected"
   - Click "Exit Admin" when done

## Technologies

- **Frontend:**
  - Vanilla JavaScript (no framework dependencies)
  - JSZip for ZIP file creation
  - CSS Grid for responsive layout

- **Backend:**
  - Go standard library
  - JSON API
  - File system operations

## Notes

- Photos directory is created automatically on startup
- Images are served directly from the filesystem
- No database required
- Supports large image collections with lazy loading

