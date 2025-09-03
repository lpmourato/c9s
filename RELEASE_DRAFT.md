Release Draft - c9s

Version: draft (use git tag before publishing)

Artifacts:
- bin/c9s-darwin-amd64.zip (macOS Intel)
- bin/c9s-darwin-arm64.zip (macOS Apple Silicon)
- bin/c9s-windows-amd64.zip (Windows 64-bit)

Checksums (SHA-256):

```
$(cat bin/checksums.txt)
```

Suggested release notes:

- Show timestamps in local timezone in log and UI views
- Enter key opens logs from service list
- Build: multi-platform binaries included for macOS (amd64/arm64) and Windows (amd64)

Publishing guide:
1. Create a git tag (e.g., `v0.1.0`) and push it: `git tag -a v0.1.0 -m "Release v0.1.0" && git push origin v0.1.0`
2. Create a draft release on GitHub and upload the three zip artifacts from `bin/`.
3. Attach `bin/checksums.txt` to the release assets or paste checksums in release notes.
4. Publish the release when ready.

Upload helper (use `scripts/upload_release.sh`):

```sh
# Usage: ./scripts/upload_release.sh <tag>
# Requires: GH_TOKEN env var or authenticated gh CLI
TAG=$1
gh release create "$TAG" --draft --title "$TAG" -F RELEASE_DRAFT.md
gh release upload "$TAG" bin/*.zip bin/checksums.txt
```
