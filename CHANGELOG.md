# Changelog

All notable changes to the clip project are documented here.

---

## [2.0.2] - 2025-12-02

### 🎥 Major Features

#### Media Support (Photos → Media)
- **Renamed**: `/photos` endpoint is now `/media`
- **Video Support**: Upload and view videos (.mp4, .webm, .mov, .avi, .mkv, .m4v)
- **Video Playback**: Full video player with controls in preview mode
- **Type Detection**: API returns media type (image/video) for each file
- **Play Icons**: Videos show ▶ overlay in gallery thumbnails

#### Navigation Enhancements
- **Arrow Buttons**: Previous/Next buttons on focused media
- **Keyboard Support**: Use ←/→ to navigate, ESC to close
- **Seamless Browsing**: Navigate through entire gallery without closing preview
- **Smart Boundaries**: Arrow buttons disable at start/end

#### Mobile Sharing
- **PWA Manifest**: Progressive Web App with Share Target API
- **Share-to Support**: Share files from any mobile app to clip
- **Shared Page**: Files from mobile go to `/view/shared`
- **Multi-file**: Share multiple files at once
- **Text Support**: Share text, URLs, and titles along with files

### 🎨 UI Improvements

#### Color Consistency
- **Green Download**: All download buttons now green (#28a745)
- **Focused Media**: Download button on preview is green
- **Consistent Theme**: Unified color scheme across app

#### Layout Improvements
- **Full-Width Inputs**: View and Edit search boxes now span full card width
- **All Media Display**: Admin home shows all media items (not just 8)
- **Better Spacing**: Improved form layouts

### 🔧 Backend Changes

#### Curl Upload Target
- **Changed**: Curl uploads now go to `/view/curl` page (was `/view/unsorted`)
- **Command**: `curl -F "file=@file.txt" https://clip.abaj.ai/`

#### API Endpoints
- `/api/photos/*` → `/api/media/*`
- `/api/media/list` - Returns array of `{name, type}` objects
- `/api/media/upload` - Accepts images and videos
- `/api/media/delete` - Delete media files
- `/api/share` - Share Target API endpoint (mobile)

#### Storage
- `persistence/photos/` → `persistence/media/`
- All media files persist across container restarts
- Larger upload limit (100MB for videos)

### 📱 Mobile Features

#### Share Target API
On mobile devices, after installing clip as a PWA:
1. Share any file from photos, videos, etc.
2. Select "clip" from share menu
3. Files automatically attach to `/view/shared` page
4. Can share text, URLs, and titles too

#### Installation
On mobile browser (Chrome/Safari):
1. Visit `https://clip.abaj.ai`
2. Tap "Add to Home Screen" or "Install"
3. App installs as PWA
4. Now appears in share menu

---

## [2.0.1] - 2025-12-02

### ✨ Features
- Arrow navigation in photo gallery
- Keyboard shortcuts (arrows, ESC)
- Full-width input boxes
- Show all photos on admin home

### 🎨 UI
- Green download button on focused images
- Consistent color scheme

### 🧹 Cleanup
- Removed temporary documentation files
- Removed helper scripts

---

## [2.0.0] - 2025-12-02

### 🎯 Major Release

#### Data Persistence
- Fixed: All data now persists across restarts
- Pages stored in `/app/persistence/*.txt`
- Files stored in `/app/persistence/files/`
- Photos stored in `/app/persistence/photos/`

#### Markdown Support
- View pages render GitHub Flavored Markdown
- Full styling: headers, code blocks, tables, lists
- Graceful fallback to plain text
- Client-side rendering with marked.js

#### Concurrency Protection
- Mutex-based locking for all page operations
- Thread-safe multi-user editing
- Per-page locks (fine-grained)
- No data corruption from simultaneous edits

#### Admin Authentication
- Global admin sign-on on root page
- Admin state syncs between pages
- Multiple user support (aj/red, kiyo/blue)
- Persistent admin state (localStorage)

#### Photo Gallery
- Upload multiple photos
- Download all as ZIP
- Admin mode for deletions
- Drag and drop file upload
- AJAX uploads (no redirect pages)

#### Curl Support
- Upload files via curl to pages
- Root domain uploads to dedicated page
- Multiple file support

---

## Earlier Versions

### Issues Resolved (#8-#17)
- #8: Multiple file upload support
- #9: Fixed text formatting on paste
- #10: Rebrand from wiki to clip
- #11: Photos endpoint with admin features
- #12: Drag and drop files
- #13: Global admin sign-on
- #14: Curl file uploads
- #15: Photo upload improvements
- #16: Data persistence fixes
- #17: Home link on photos page

---

## Migration Guide

### From 2.0.1 → 2.0.2

**URL Changes:**
- `/photos` → `/media` (automatic redirect recommended)
- `/api/photos/*` → `/api/media/*`

**Storage:**
- Photos directory renamed to media
- Run: `mv persistence/photos persistence/media`

**Code:**
- Update any hardcoded `/photos` URLs to `/media`
- Update API calls to use `/api/media/*`

**Mobile:**
- Reinstall PWA to get Share Target support
- Visit site, tap "Add to Home Screen"

---

## Technical Specifications

### Supported Media Types

**Images:**
- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- WebP (.webp)
- BMP (.bmp)

**Videos:**
- MP4 (.mp4)
- WebM (.webm)
- QuickTime (.mov)
- AVI (.avi)
- Matroska (.mkv)
- M4V (.m4v)

### API Response Format

**Before (v2.0.1):**
```json
{"images": ["file1.jpg", "file2.png"]}
```

**After (v2.0.2):**
```json
{
  "media": [
    {"name": "file1.jpg", "type": "image"},
    {"name": "video.mp4", "type": "video"}
  ]
}
```

### Endpoints

| Endpoint | Purpose |
|----------|---------|
| `/media` | Media gallery page |
| `/media/{filename}` | Serve individual media file |
| `/api/media/list` | List all media (JSON) |
| `/api/media/upload` | Upload media files |
| `/api/media/delete` | Delete media (auth required) |
| `/api/share` | Share Target API (mobile) |
| `/manifest.json` | PWA manifest |

---

## Credits

Developed for efficient note-taking and media sharing.

**Technology Stack:**
- Go 1.21
- Docker & Docker Compose
- Alpine Linux
- Marked.js (Markdown)
- JSZip (Archive creation)
- QR Code Generator
- PWA/Share Target API

---

For more information, see README.md or visit https://clip.abaj.ai

