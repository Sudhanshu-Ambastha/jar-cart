#!/bin/bash

# 1. Create a temporary branch from the current state
git checkout --orphan latest_branch

# 2. Add all files to the staging area
git add -A

# 3. Commit the current state as a single "Initial Commit"
git commit -m "feat: initial release of jar-cart (v1.0)"

# 4. Delete the old branch
git branch -D main

# 5. Rename the current branch to main
git branch -m main

# 6. Force push to overwrite the remote history
git push -f origin main

echo "✨ History squashed successfully!"