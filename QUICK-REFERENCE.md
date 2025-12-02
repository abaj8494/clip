# Quick Reference - All GitHub Issues Fixed

## ✅ All 6 Issues Resolved

### Issue #12: Drag & Drop Files
**What:** Drag files onto the textarea to upload them  
**Try it:** Go to any `/edit/PAGE` and drag files onto the text area

### Issue #13: Admin Mode on Home
**What:** Admin button on home page, hides page list from guests  
**Login:** `aj/red` or `kiyo/blue`  
**Try it:** Visit `/` and click "Admin"

### Issue #14: Curl Uploads
**What:** Upload files via command line  
**Try it:**
```bash
# Upload to "unsorted"
curl -F "file=@myfile.txt" https://clip.abaj.ai/

# Upload to specific page
curl -F "file=@myfile.txt" https://clip.abaj.ai/upload/mypage
```

### Issue #15: Photo Upload + Improvements
**What:** Upload button, green download button, persistent admin  
**Try it:** Visit `/photos` and click "Upload"

### Issue #16: Data Persistence Fixed
**What:** Pages now survive container restarts  
**Note:** Old data was lost, but all new data persists correctly

### Issue #17: Home Link on Photos
**What:** Photos page now has a Home link  
**Try it:** Visit `/photos` and click "Home"

## Quick Links
- **Main Site:** https://clip.abaj.ai/
- **Photos:** https://clip.abaj.ai/photos
- **Edit Page:** https://clip.abaj.ai/edit/PAGE_NAME
- **View Page:** https://clip.abaj.ai/view/PAGE_NAME

## Deployment
✅ Container rebuilt and running  
✅ All features tested and working  
✅ Data persistence confirmed  
✅ Backups working correctly  

See `GITHUB-ISSUES-RESOLVED.md` for full details.

