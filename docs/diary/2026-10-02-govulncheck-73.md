# Diary: Fix govulncheck findings from issue 73

Goal: the `Security` workflow failed on govulncheck (issue #73). Apply the smallest dependency bump that makes govulncheck pass, and open a PR that closes the issue.

## Step 1: Requirements and delegation

**Author:** main (lead)

### Prompt Context

**Verbatim prompt:** Fix 73 with the smallest possible fix, then make a PR

**Interpretation:** Issue #73 is the auto-filed "govulncheck found vulnerabilities" issue pointing at a failed `Security` workflow run. Fix it with a minimal dependency bump and open a PR that closes it.

**Inferred intent:** Get CI green again with the smallest reviewable diff, without opportunistic upgrades.

### What I did
Ran `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` in the main checkout. It reported exactly one vulnerability (exit status 3):

- GO-2026-6505, "OpenTelemetry-Go: Exporter config logging may leak endpoint URLs in info logs", in `go.opentelemetry.io/otel/exporters/otlp/otlptrace@v1.44.0`. Fixed in v1.45.0. Reachable from app code via `app.Start` and `otlptrace.New` / `Exporter.ExportSpans`.

The scan reported no other vulnerabilities in imported packages or required modules. In `/go.mod`, `otlptrace` and its sibling exporters (`otlptracegrpc`, `otlptracehttp`, `otlpmetricgrpc`, `otlpmetrichttp`) are all `// indirect` at v1.44.0, while the core `otel`, `otel/trace`, `otel/metric`, and `otel/sdk` modules are already at v1.45.0.

Created worktree `govulncheck-73`, started this diary, and spawned one builder with this requirement: bump only `go.opentelemetry.io/otel/exporters/otlp/otlptrace` to exactly v1.45.0 (the fix version, not `@latest`), run `go mod tidy`, and let other modules move only if the module graph requires it. govulncheck must be clean, build and tests must pass, and the diff must touch only `go.mod`/`go.sum` (plus the diary).

### Why
"Smallest possible fix" means pinning to the fix version rather than `@latest`, and not bumping the sibling exporters, which govulncheck does not flag.

### What worked
N/A yet.

### What didn't work
N/A yet.

### What I learned
OTel-Go's exporter modules version separately from the core modules in this `go.mod`: the core is already at v1.45.0, but the exporters lag at v1.44.0 because they arrive indirectly.

### What was tricky
Nothing so far.

### What warrants review
The builder's step will say; expect a `go.mod`/`go.sum`-only diff.

### Future work
None identified.
