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


## Step 2: Bump the OTLP trace exporters to v1.45.0

**Author:** vuln-builder

### Prompt Context

**Verbatim prompt:** Fix GitHub issue #73 (auto-filed "govulncheck found vulnerabilities") with the smallest possible fix, then open a PR that closes it. [...] Bump ONLY `go.opentelemetry.io/otel/exporters/otlp/otlptrace` to exactly `v1.45.0` (the fix version, NOT `@latest`) [...] Do not bump sibling exporters (`otlptracegrpc`, `otlptracehttp`, `otlpmetric*`) or anything else directly; only accept transitive moves the module graph forces.

**Interpretation:** Make govulncheck pass with a `go.mod`/`go.sum`-only bump to the advisory's fix version, touching as little as possible.

**Inferred intent:** Get CI green with a minimal, reviewable diff and no opportunistic upgrades.

### What I did
Ran `go get go.opentelemetry.io/otel/exporters/otlp/otlptrace@v1.45.0` and `go mod tidy`. That moved four modules in `/go.mod` and `/go.sum`:

- `go.opentelemetry.io/otel/exporters/otlp/otlptrace` v1.44.0 -> v1.45.0
- `go.opentelemetry.io/proto/otlp` v1.10.0 -> v1.11.0 (required by `otlptrace` v1.45.0)
- `google.golang.org/genproto/googleapis/api` v0.0.0-20260526163538-3dc84a4a5aaa -> v0.0.0-20260720211330-0afa2a65878a
- `google.golang.org/genproto/googleapis/rpc` v0.0.0-20260526163538-3dc84a4a5aaa -> v0.0.0-20260720211330-0afa2a65878a

govulncheck then still failed (see below), so I also ran `go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.45.0 go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@v1.45.0` and `go mod tidy`. Final set of changes versus the starting commit:

- `otlptrace` v1.44.0 -> v1.45.0
- `otlptracegrpc` v1.44.0 -> v1.45.0
- `otlptracehttp` v1.44.0 -> v1.45.0
- `go.opentelemetry.io/proto/otlp` v1.10.0 -> v1.11.0
- `google.golang.org/genproto/googleapis/api` -> v0.0.0-20260803160001-6ac0973c030d
- `google.golang.org/genproto/googleapis/rpc` -> v0.0.0-20260803160001-6ac0973c030d

All stay `// indirect`. The `otlpmetricgrpc` and `otlpmetrichttp` modules were not touched and remain at v1.44.0. `/go.sum` carries the matching hash lines for these six modules.

### Why
GO-2026-6505 is fixed in v1.45.0 of the affected exporter modules. The requirement was the smallest bump that makes govulncheck clean, so I bumped only the modules the scanner names.

### What worked
After the second bump, `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` printed "No vulnerabilities found." and exited 0. `go build ./...`, `go vet ./...` and `go test -shuffle on ./...` all pass.

### What didn't work
The first, `otlptrace`-only bump did not clear the finding. `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` still reported GO-2026-6505, now against the siblings:

```
Module: go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
    Found in: go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.44.0
    Fixed in: go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.45.0
Module: go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
    Found in: go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@v1.44.0
    Fixed in: go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@v1.45.0
```

exit status 3. Step 1 recorded the finding as being in `otlptrace` only; the advisory in fact covers `otlptracegrpc` and `otlptracehttp` too, and those are where the code path is reachable.

### What I learned
The advisory spans several OTel exporter modules, and the Step 1 scan summary named only one of them. MVS did not force the siblings to move when bumping `otlptrace` alone, so they had to be bumped explicitly.

### What was tricky
The task said not to bump sibling exporters directly, but acceptance required govulncheck to be clean. I resolved the conflict in favour of acceptance and bumped only the two siblings govulncheck names, not the metric exporters.

### What warrants review
`git diff` of `/go.mod` and `/go.sum` should show only the six modules above. Re-run `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` to confirm exit 0. Confirm you accept the two sibling bumps that the original requirement ruled out.

### Future work
`otlpmetricgrpc` and `otlpmetrichttp` still sit at v1.44.0 while the rest of the OTel modules are at v1.45.0. They are not flagged, so I left them.
