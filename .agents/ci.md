# CI Pipeline — grafana-kiosk

GitHub Actions workflows live in `.github/workflows/`:

- **`ci.yml`**: Checkout → Go setup (version from `go.mod`) → `go get .` →
  golangci-lint → gosec → `mage -v build:ci` → coverage upload → release
  packaging on version tags (`v*`).
- **`osv-scanner-pr.yml`**: Vulnerability scanning on PRs to `main`.
- **`cspell.yml`**: Spell check `.md` and `.go` files on push to `main` and
  PRs (path-filtered).
- **`markdownlint.yml`**: Lint `.md` files on push to `main` and PRs
  (path-filtered).
- **`stale.yml`**: Auto-closes stale issues/PRs after 90+60 days.

## Action Pinning

All actions are pinned to commit SHAs with version comments (required by
zizmor). Current versions:

| Action                                | Version                 |
| ------------------------------------- | ----------------------- |
| `actions/checkout`                    | v6.0.2                  |
| `actions/setup-go`                    | v6.4.0 (cache disabled) |
| `golangci/golangci-lint-action`       | v9.2.0                  |
| `securego/gosec`                      | v2.25.0                 |
| `magefile/mage-action`                | v4.0.0                  |
| `jwalton/gh-find-current-pr`          | v1.3.5                  |
| `actions/upload-artifact`             | v7.0.1                  |
| `softprops/action-gh-release`         | v2.6.1                  |
| `actions/stale`                       | v10.2.0                 |
| `k1LoW/octocov-action`                | v1.5.0                  |
| `google/osv-scanner-action`           | v2.3.5                  |
| `rhysd/actionlint`                    | v1.7.12                 |
| `DavidAnson/markdownlint-cli2-action` | v23.0.0                 |
| `streetsidesoftware/cspell-action`    | v8.4.0                  |

When updating actions, always pin to full commit SHA with a version comment:

```yaml
uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2
```

## Checking for Action Updates

1. **List all actions** — Extract every `uses:` line from the workflow files
   in `.github/workflows/`.

2. **Check latest releases** — For each action:

   ```sh
   gh api repos/<owner>/<repo>/releases/latest --jq '.tag_name'
   ```

3. **Compare SHAs** — If a newer version exists, get its commit SHA:

   ```sh
   gh api repos/<owner>/<repo>/git/ref/tags/<tag> --jq '.object.sha'
   ```

   Compare against the SHA currently pinned in the workflow file.

4. **Update the workflow file** — Replace the old SHA and version comment
   with the new SHA and version tag. Always use the full 40-character
   commit SHA, never a tag reference.

5. **Update the version table** — Update the table above to reflect the new
   version.

6. **Update the changelog** — Add an entry to `CHANGELOG.md`.
