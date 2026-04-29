# Release Process — grafana-kiosk

Uses [semantic versioning](https://semver.org/) with tags in the form
`vX.X.X`. The CI workflow in `.github/workflows/ci.yml` handles building,
packaging, and publishing releases automatically when a version tag is
pushed.

## Determining the New Version

1. **Get the current version** from git tags:

   ```sh
   git tag --list 'v*' --sort=-v:refname | head -1
   ```

2. **Choose the version bump** — default is patch:

   | Bump  | When to use                            | Example            |
   | ----- | -------------------------------------- | ------------------ |
   | Patch | Bug fixes, dependency updates, docs    | v1.0.10 → v1.0.11  |
   | Minor | New features, non-breaking changes     | v1.0.10 → v1.1.0   |
   | Major | Breaking changes to CLI, config, login | v1.0.10 → v2.0.0   |

## Pre-release Checklist

1. **Switch to `main` and pull latest changes**:

   ```sh
   git checkout main
   git pull origin main
   ```

2. **Merge any feature branches** that should be included in the release.

3. **Verify the changelog** — `CHANGELOG.md` must have a section header
   matching the new version number (without the `v` prefix). Remove the
   `[Unreleased]` header entirely — do not leave it as an empty section.
   Cross-reference all entries against the commits since the last tag:

   ```sh
   git log <last-tag>..HEAD --oneline
   ```

   Every user-facing change must have a corresponding changelog entry.

4. **Run the full verification suite**:

   ```sh
   mage -v                                       # build
   mage -v test:default                           # tests
   golangci-lint run --timeout 5m ./pkg/...       # lint
   gosec ./...                                    # security scan
   ```

5. **Lint the changelog**:

   ```sh
   npx markdownlint-cli2 CHANGELOG.md
   ```

## Tagging and Pushing

Create the tag on `main` and push both the branch and the tag:

```sh
git tag vX.X.X
git push origin main
git push origin vX.X.X
```

## What CI Does Automatically

Pushing a `v*` tag triggers the CI workflow which:

1. Runs lint, security scan, and tests.
2. Builds all 9 OS/arch binaries via `mage -v build:ci`.
3. Packages binaries into a flat directory with platform-specific names
   (e.g., `grafana-kiosk.linux.amd64`, `grafana-kiosk.darwin.arm64`).
4. Creates `.zip` and `.tar.gz` archives.
5. Publishes a GitHub **prerelease** with auto-generated release notes
   and the archives attached.

## Post-release

1. **Verify the release** — Go to the
   [GitHub Releases](https://github.com/grafana/grafana-kiosk/releases) page
   and confirm the artifacts are attached and downloadable.
2. **Promote the release** — Edit the release on GitHub and uncheck
   "Set as a pre-release" to make it a full release.
