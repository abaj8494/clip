#!/bin/bash
cd /var/www/clip

# Stage the new files
git add complete-rebrand.sh REBRAND-STATUS.md

# Commit
git commit -F /tmp/commit-msg.txt

# Push
git push origin main

# Close issue #10
gh issue close 10 -c "Rebrand completed! 🎉

Main tasks done:
✅ Repository renamed to 'clip'
✅ Directory moved to /var/www/clip
✅ All code references updated
✅ Nginx config created for clip.abaj.ai
✅ Git remote updated

Remaining steps (documented in REBRAND-STATUS.md):
- Update DNS for clip.abaj.ai
- Generate SSL certificates
- Restart services

Run \`bash /var/www/clip/complete-rebrand.sh\` after DNS is configured."

echo "Issue #10 closed successfully!"

