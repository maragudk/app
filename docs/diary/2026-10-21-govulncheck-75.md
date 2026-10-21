# Diary: Fix govulncheck findings from issue 75

Goal: the `Security` workflow failed on govulncheck (issue #75). Apply the smallest dependency bump that makes govulncheck pass.

## Step 1: Requirements and delegation

**Author:** main (lead)

### Prompt Context

**Verbatim prompt:** Fix https://github.com/maragudk/app/issues/75 with minimal dep bumps

**Interpretation:** Issue #75 is the auto-filed "govulncheck found vulnerabilities" issue pointing at a failed `Security` workflow run. Fix it by bumping as few dependencies as possible, as little as possible.

**Inferred intent:** Get the `Security` workflow green again with the smallest reviewable diff, without opportunistic upgrades.

### What I did
Ran `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` in the main checkout at `6a674e9`. It reported five vulnerabilities (exit status 3), all HTTP/2 issues in one module, `golang.org/x/net@v0.58.0`, all fixed in v0.60.0:

- GO-2026-6617, HTTP/2 server crash due to HPACK encoder race
- GO-2026-6612, double flow control refund on HTTP/2 server streams
- GO-2026-6611, excessive CPU consumption from repeated initial window changes
- GO-2026-6610, HTTP/2 transport accepts malformed framing-related headers
- GO-2026-6603, HTTP/2 server memory exhaustion due to Trailer headers

The scan found no other vulnerabilities. In `/go.mod`, `golang.org/x/net` is `// indirect` at v0.58.0; `go mod why -m` shows it arrives through `maragu.dev/glue/app`, `github.com/honeycombio/otel-config-go/otelconfig` and `google.golang.org/grpc`.

Created worktree `govulncheck-75`, started this diary, and spawned one builder with this requirement: bump only `golang.org/x/net` to exactly v0.60.0 (the fix version, not `@latest`), run `go mod tidy`, and accept other module moves only where the module graph forces them. govulncheck must be clean, build, vet and tests must pass, and the diff must touch only `/go.mod` and `/go.sum` (plus this diary).

### Why
"Minimal dep bumps" means pinning to the advisory's fix version rather than `@latest`, and not touching modules govulncheck doesn't flag.

### What worked
N/A yet.

### What didn't work
N/A yet.

### What I learned
Unlike issue #73, where one advisory spanned several OTel exporter modules, all five findings here sit in a single module with a single fix version.

### What was tricky
Nothing so far. The one risk is that x/net v0.60.0 requires newer `golang.org/x/*` siblings (`crypto`, `sys`, `text`), which MVS will drag along.

### What warrants review
The builder's step will say; expect a `/go.mod`/`/go.sum`-only diff.

### Future work
None identified.

## Step 2: Bump `golang.org/x/net` to v0.60.0

**Author:** vuln-builder

### Prompt Context

**Verbatim prompt:** Fix GitHub issue maragudk/app#75 ("govulncheck found vulnerabilities") with minimal dependency bumps. [...] Bump ONLY `golang.org/x/net` to exactly `v0.60.0` (the fix version, NOT `@latest`): `go get golang.org/x/net@v0.60.0` then `go mod tidy`. Accept other module moves only where the module graph forces them (e.g. x/crypto, x/sys, x/text). Do not bump anything else directly.
**Interpretation:** Pin `golang.org/x/net` to the advisory's fix version and let the module graph move only what v0.60.0 requires. Verify with govulncheck, build, vet and tests. Leave the changes uncommitted.
**Inferred intent:** Get the `Security` workflow green with the smallest possible `/go.mod` and `/go.sum` diff.

### What I did
Ran `go get golang.org/x/net@v0.60.0` and then `go mod tidy` in the worktree. `go get` printed these upgrades, and `go mod tidy` changed nothing further:

- `golang.org/x/net` v0.58.0 -> v0.60.0 (the requested bump)
- `golang.org/x/sync` v0.22.0 -> v0.23.0 (pulled in by the module graph)
- `golang.org/x/sys` v0.47.0 -> v0.48.0 (pulled in by the module graph)
- `golang.org/x/text` v0.41.0 -> v0.42.0 (pulled in by the module graph)

No other module moved. `golang.org/x/crypto` is not in `/go.mod` at all. Only `/go.mod` and `/go.sum` changed.

Then ran `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`, `go build ./...`, `go vet ./...` and `go test -shuffle on ./...`.

### Why
All five findings (GO-2026-6617, -6612, -6611, -6610, -6603) sit in `golang.org/x/net@v0.58.0` and are fixed in v0.60.0. Pinning to that exact version avoids opportunistic upgrades.

### What worked
govulncheck printed "No vulnerabilities found." and exited 0. `go build ./...` and `go vet ./...` passed. `go test -shuffle on ./...` passed in `app/html`, `app/jobs`, `app/service`, `app/sqlite` and `app/sqlitetest`; the other packages have no test files. No second round of bumps was needed.

### What didn't work
Nothing failed. One of my compound shell commands was refused by the sandbox because it combined git with other commands, so I split it into plain commands. My first attempt to read govulncheck's exit status via `${PIPESTATUS[0]}` printed nothing because the shell is zsh; I re-ran it without a pipe and got `exit 0`.

### What I learned
`golang.org/x/net` v0.60.0 raises the minimum of `x/sync`, `x/sys` and `x/text` by one minor version each. Of the siblings named in the task, `x/crypto` is not a dependency of this module.

### What was tricky
Nothing. The bump resolved cleanly.

### What warrants review
Read the `/go.mod` diff: four `// indirect` lines changed, all under `golang.org/x/`. Check that the `/go.sum` changes are limited to those four modules. To validate, run `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` and `go test -shuffle on ./...`.

### Future work
None identified.
