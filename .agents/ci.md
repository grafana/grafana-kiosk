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
| `actions/checkout`                    | v7.0.1                  |
| `actions/setup-go`                    | v7.0.0 (cache enabled)  |
| `golangci/golangci-lint-action`       | v9.3.0                  |
| `securego/gosec`                      | v2.29.0                 |
| `magefile/mage-action`                | v4.0.0                  |
| `jwalton/gh-find-current-pr`          | v1.3.5                  |
| `actions/upload-artifact`             | v7.0.1                  |
| `softprops/action-gh-release`         | v3.0.3                  |
| `actions/stale`                       | v11.0.0                 |
| `k1LoW/octocov-action`                | v1.5.2                  |
| `google/osv-scanner-action`           | v2.5.1                  |
| `rhysd/actionlint`                    | v1.7.12                 |
| `DavidAnson/markdownlint-cli2-action` | v24.2.0                 |
| `streetsidesoftware/cspell-action`    | v9.1.0                  |

When updating actions, always pin to full commit SHA with a version comment:

```yaml
uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
```

## Pinned Tool Versions

Pinning the action is not enough when the action installs a tool — the tool
version must be pinned too, or CI silently drifts from local. These are
duplicated across files and must be changed together:

| Tool            | CI                                    | Local       |
| --------------- | ------------------------------------- | ----------- |
| `golangci-lint` | `version:` on `golangci-lint-action`  | `mise.toml` |
| `gosec`         | action tag (`securego/gosec`)         | `mise.toml` |
| `mage`          | `version:` on `mage-action`           | `go.mod`    |

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

6. **Update the changelog** — Only when the change affects the released
   artifact. A bump to a build or scanning action (checkout, gosec,
   actionlint, cspell) changes nothing a person running the kiosk binary can
   observe, so it gets no entry. Bumps that alter the shipped binary or the
   release packaging — a Go toolchain change, or `mage-action` building with
   a different mage — do get one.
