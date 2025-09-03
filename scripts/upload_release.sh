#!/usr/bin/env bash
set -euo pipefail

if [ -z "${1-}" ]; then
  echo "Usage: $0 <tag>"
  exit 2
fi

TAG=$1

# Create draft release from RELEASE_DRAFT.md
if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required. Install from https://cli.github.com/"
  exit 1
fi

# Create release as draft
gh release create "$TAG" --draft --title "$TAG" -F RELEASE_DRAFT.md

# Upload artifacts
gh release upload "$TAG" bin/*.zip bin/checksums.txt

echo "Draft release $TAG created and assets uploaded."
