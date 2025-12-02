#!/bin/bash
cd /var/www/clip

# Stage the new files
git add backend/photos.html backend/wiki.go PHOTOS-FEATURE.md

# Commit
git commit -m "Add /photos endpoint with admin features

Created comprehensive photo gallery with:
- Clean grid layout based on photos-example.html style
- Download all button (creates ZIP of all images)
- Enhanced preview system:
  - Click image to open full-screen preview
  - Close button (✕) in top-left
  - Download button in top-right
  - Click outside to close
- Admin mode with authentication:
  - Username: aj, Password: red
  - Multi-select with checkboxes
  - Batch delete functionality
  - Visual selection feedback

Backend additions:
- GET /photos - serves gallery page
- GET /api/photos/list - returns image list
- POST /api/photos/delete - authenticated deletion
- GET /photos/{filename} - serves individual photos
- Automatic photos directory creation
- Security: path traversal protection, auth required for delete

Fixes #11"

# Push
git push origin main

# Close issue #11
gh issue close 11 -c "Photos feature complete! 🎉📸

All requested features implemented:

✅ Gallery with photos-example.html style
✅ Download all button (creates ZIP)
✅ Admin authentication (user: aj, pass: red)
✅ Image selection and deletion in admin mode
✅ Enhanced preview with:
   - Click to focus/fullscreen
   - Download button (top right)
   - Close button (top left)

Access at: https://clip.abaj.ai/photos

See PHOTOS-FEATURE.md for complete documentation."

echo "Issue #11 closed successfully!"

