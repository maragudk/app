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
