# Diary: Fix govulncheck findings from issue 81

Goal: the `Security` workflow failed on govulncheck (issue #81). Apply the smallest dependency bump that makes govulncheck pass.

## Step 1: Requirements and delegation

**Author:** main (lead)

### Prompt Context

**Verbatim prompt:** Fix https://github.com/maragudk/app/issues/81 with the minimal possible dep bump.

**Interpretation:** Issue #81 is the auto-filed "govulncheck found vulnerabilities" issue pointing at a failed `Security` workflow run. Fix it by bumping as few dependencies as possible, as little as possible.

**Inferred intent:** Get the `Security` workflow green again with the smallest reviewable diff, without opportunistic upgrades.

### What I did
Ran `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` in the main checkout at `512df2d`. It reported two vulnerabilities, both in the OpenTelemetry Go core modules at v1.45.0 and both fixed in v1.47.0:

- GO-2026-6975, memory allocation amplification in W3C baggage parsing, found in `go.opentelemetry.io/otel@v1.45.0`
- GO-2026-6974, `AttributeValueLengthLimit` bypass via U+FFFD runes, found in `go.opentelemetry.io/otel/sdk@v1.45.0` and `go.opentelemetry.io/otel/trace@v1.45.0`

In `/go.mod`, `otel` and `otel/trace` are direct requirements and `otel/sdk` is indirect. The OTLP exporters sit at v1.44.0/v1.45.0 and `otel/metric` and `otel/sdk/metric` at v1.45.0.

Created worktree `govulncheck-81`, started this diary, and spawned one builder with this requirement: bump only `otel`, `otel/sdk` and `otel/trace` to exactly v1.47.0 (the fix version, not `@latest`), run `go mod tidy`, and accept other module moves only where the module graph forces them. govulncheck must be clean, build, vet and tests must pass, and the diff must touch only `/go.mod` and `/go.sum` (plus this diary).

### Why
"Minimal possible dep bump" means pinning to the advisory's fix version rather than `@latest`, and not touching modules govulncheck doesn't flag.

### What worked
N/A yet.

### What didn't work
N/A yet.

### What I learned
Like issue #73, this advisory set spans several OTel modules, but here they're the core API/SDK modules rather than the exporters.

### What was tricky
Nothing so far. The risk is that OTel v1.47.0 forces `otel/metric` and maybe `otel/sdk/metric` or the exporters along, or that version-skewed exporters fail to build.

### What warrants review
The builder's step will say; expect a `/go.mod`/`/go.sum`-only diff.

### Future work
None identified.

## Step 2: Bump the OpenTelemetry core modules to v1.47.0

**Author:** builder

### Prompt Context

**Verbatim prompt:** Fix GitHub issue maragudk/app#81 ("govulncheck found vulnerabilities") with the minimal possible dependency bump. Bump ONLY `go.opentelemetry.io/otel`, `go.opentelemetry.io/otel/sdk` and `go.opentelemetry.io/otel/trace` to exactly v1.47.0 (the fix version, NOT @latest) ... Accept other module moves only where the module graph forces them. Do not bump anything else directly. Verify: govulncheck reports no vulnerabilities (exit 0), `go build ./...`, `go vet ./...` and `go test -shuffle on ./...` all pass. Diff touches only `/go.mod` and `/go.sum`, plus the diary. Do NOT commit.
**Interpretation:** Pin the three flagged modules to v1.47.0, let the module graph pull in whatever it must, and change nothing else.
**Inferred intent:** A green `Security` workflow with the smallest reviewable `/go.mod` and `/go.sum` diff.

### What I did
First ran `go get go.opentelemetry.io/otel@v1.47.0 go.opentelemetry.io/otel/sdk@v1.47.0 go.opentelemetry.io/otel/trace@v1.47.0`, which failed (see below). Instead I ran `go get go.opentelemetry.io/otel@v1.47.0 go.opentelemetry.io/otel/trace@v1.47.0 go.opentelemetry.io/otel/sdk/resource@v1.47.0`, then `go mod tidy`. The `sdk/resource` package lives in the `otel/sdk` module, so this requires `otel/sdk@v1.47.0` without naming its unusable root package path.

Module changes in `/go.mod`:

- `go.opentelemetry.io/otel` v1.45.0 to v1.47.0 (requested, direct)
- `go.opentelemetry.io/otel/trace` v1.45.0 to v1.47.0 (requested, direct)
- `go.opentelemetry.io/otel/sdk` v1.45.0 to v1.47.0 (requested, indirect)
- `go.opentelemetry.io/otel/metric` v1.45.0 to v1.47.0 (forced: `otel` and `otel/sdk` v1.47.0 require it)
- `go.opentelemetry.io/otel/sdk/metric` v1.45.0 to v1.47.0 (forced: `otel/sdk` v1.47.0 requires it at the same version)
- `go.opentelemetry.io/otel/log` v1.47.0 added as indirect (new requirement of `otel/sdk` v1.47.0)
- `github.com/kr/text` v0.2.0 added as indirect (see below)

Only in `/go.sum`, not in `/go.mod`: `otel/metric/x` v0.67.0 to v0.69.0, `github.com/stretchr/testify` v1.11.1 to v1.12.1, `go.yaml.in/yaml/v3` v3.0.5 added, `github.com/creack/pty` v1.1.9 `/go.mod` hash added, and `go-spew` and `go-difflib` hashes dropped. These are test-dependency graph shifts pulled in by the new OTel versions.

The OTLP exporters (v1.44.0 and v1.45.0), `otel/exporters/*`, and everything else stayed where they were. Version skew with the exporters did not break the build.

### Why
v1.47.0 is the advisory fix version for GO-2026-6975 (`otel`) and GO-2026-6974 (`otel/sdk`, `otel/trace`). `metric`, `sdk/metric` and `log` moved only because the graph requires them.

### What worked
`go build ./...`, `go vet ./...` and `go test -shuffle on ./...` all pass. `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` exits 0 with "No vulnerabilities found." and "Your code is affected by 0 vulnerabilities."

### What didn't work
The requested command failed:

```
$ go get go.opentelemetry.io/otel@v1.47.0 go.opentelemetry.io/otel/sdk@v1.47.0 go.opentelemetry.io/otel/trace@v1.47.0
go: module go.opentelemetry.io/otel@v1.47.0 found, but does not contain package go.opentelemetry.io/otel/sdk
```

`go get go.opentelemetry.io/otel/sdk@v1.47.0` alone fails the same way. The module exists (`go list -m go.opentelemetry.io/otel/sdk@v1.47.0` works), so the failure is in package resolution of the module root path, not in the version.

### What I learned
The govulncheck summary still mentions "2 vulnerabilities in modules you require, but your code doesn't appear to call these". Those are not the two OTel findings. They are other advisories at module level that are not reachable. The exit code is 0, so the workflow passes. I did not investigate them.

### What was tricky
Getting `otel/sdk` required at the exact version without bumping with `@latest`. Naming a package inside the module (`otel/sdk/resource`) worked around it.

`go mod tidy` added `github.com/kr/text v0.2.0` to `/go.mod`. `go mod why -m github.com/kr/text` shows the chain `maragu.dev/redo` -> `gopkg.in/yaml.v3` test -> `gopkg.in/check.v1` -> `github.com/kr/pretty` -> `github.com/kr/text`. It is a test-only transitive dependency that tidy now needs listed, a side effect of the graph shift and not a deliberate bump.

### What warrants review
The diff is `/go.mod` and `/go.sum` only. Check that the `otel*` lines in `/go.mod` match the list above and that no non-OTel module appears in `/go.mod` other than the indirect `github.com/kr/text`.

### Future work
The exporters (`otlptrace*` v1.45.0, `otlpmetric*` v1.44.0) now trail the core at v1.47.0. A later routine bump can align them. It is not needed for this fix.
