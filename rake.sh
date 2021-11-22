git filter-branch --force --index-filter \
'git rm --cached --ignore-unmatch files/25e45f05-8bc2-4967-bbac-772f09392dd3@thumb.tiff' \
--prune-empty --tag-name-filter cat -- --all