# README Policy — grafana-kiosk

## Review Trigger

**When any code changes are made, check if `README.md` needs to be updated.**
This includes changes to CLI flags, environment variables, configuration
options, default values, login methods, build targets, or any user-facing
behavior. If the change does not affect user-facing behavior (tests,
internal refactoring, CI-only changes), no README update is needed.

## Updating Usage Documentation

When CLI flags or environment variables change, update the README to match:

1. **Fix source first** — Ensure the `env-description` tags in
   `pkg/kiosk/config.go` and the flag definitions in
   `pkg/cmd/grafana-kiosk/main.go` are accurate and consistent with each
   other. For example, if `-login-method` lists `aws` as an option, the
   corresponding `KIOSK_LOGIN_METHOD` env-description must also list `aws`.

2. **Build the binary** — Run `mage -v` to produce a fresh local binary.

3. **Capture the help output** — Run
   `./bin/<os>_<arch>/grafana-kiosk --help` and capture the full output.

4. **Update the CLI flags section** — Replace the code block under the
   `## Usage` heading in `README.md` with the flags portion of the `--help`
   output (everything from the first flag up to the `Environment variables:`
   line).

5. **Update the environment variables section** — Replace the code block
   under the "Environment variables can be set..." paragraph in `README.md`
   with the environment variables portion of the `--help` output.

6. **Lint the README** — Run `npx markdownlint-cli2 README.md` and fix any
   issues before committing.

7. **Run tests** — Run `mage -v test:default` to confirm no regressions.
