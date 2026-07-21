set -e
TARGET_DIR="$1"
COMMIT_MSG="$2"
shift 2
FILES_TO_ADD="$@"

cd "$TARGET_DIR"

git config --unset-all http.https://github.com/.extraheader || true
git config user.name "github-actions[bot]"
git config user.email "github-actions[bot]@users.noreply.github.com"

if ! git symbolic-ref -q HEAD > /dev/null; then
  CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
  if [ "$CURRENT_BRANCH" = "HEAD" ]; then
    git checkout -b main || git checkout -b master
  fi
fi

git add $FILES_TO_ADD
git diff --staged --quiet && echo "No changes to commit" || git commit -m "$COMMIT_MSG"

BRANCH=$(git rev-parse --abbrev-ref HEAD)
git push origin "$BRANCH"