# Diary: Fix govulncheck findings from issue 68

Goal: the `Security` workflow failed on govulncheck (issue #68). Upgrade the affected indirect dependencies so govulncheck passes, and open a PR that closes the issue.

## Step 1: Requirements and delegation

**Author:** main (lead)

### Prompt Context

**Verbatim prompt:** See issue 68

**Interpretation:** Issue #68 is the auto-filed "govulncheck found vulnerabilities" issue pointing at a failed `Security` workflow run. The fix is to upgrade the vulnerable dependencies until govulncheck is clean.

**Inferred intent:** Get CI green again with a minimal, reviewable dependency bump, and close the issue.

### What I did
Ran `go run golang.org/x/vuln/cmd/govulncheck@latest -show verbose ./...` in the main checkout. It reported four vulnerabilities, all in indirect dependencies:

- `google.golang.org/grpc@v1.82.1`: GO-2026-6348 (reachable from app code via OTel exporters), GO-2026-6441, GO-2026-6443. Fixed in v1.83.1.
- `golang.org/x/net@v0.55.0`: GO-2026-5942. Fixed in v0.56.0.

Only GO-2026-6348 is reachable from app code, but all four are the same kind of bump, so the requirement is to clear all of them. Created worktree `govulncheck-68`, started this diary, and spawned a builder.

### Why
Routine dependency upgrade; delegated to a single builder because it is self-contained.

### What worked
N/A yet.

### What didn't work
N/A yet.

### What I learned
govulncheck's default output only lists the vulnerabilities reachable from your code; `-show verbose` is needed to see the ones in imported-but-uncalled packages.

### What was tricky
Nothing so far.

### What warrants review
The builder's step will say; expect a `go.mod`/`go.sum`-only diff.

### Future work
None identified.

## Step 2: Bump `grpc`, let `go mod tidy` carry `x/net` along

**Author:** vuln-builder

### Prompt Context

**Verbatim prompt:** Task brief from the lead: upgrade the vulnerable dependencies reported by govulncheck for issue #68 so govulncheck is clean, then open a PR that closes the issue. Bump `google.golang.org/grpc` to at least v1.83.1 and `golang.org/x/net` to at least v0.56.0 via `go get ... @latest` + `go mod tidy`, confirm `govulncheck -show verbose ./...` reports zero vulnerabilities, run `go build ./... && go test ./...`, keep the diff go.mod/go.sum-only if possible, diary, self-review, commit, and open the PR.

Mid-task, the coordinator narrowed the scope: only fix the vulnerability reachable from app code (GO-2026-6348 via `grpc`). Bump `google.golang.org/grpc` to at least v1.83.1 only, and do **not** bump `golang.org/x/net` directly — only let it move if `go mod tidy` pulls it along as a transitive requirement of the `grpc` bump. The bar for success changed to: the default (non-verbose) `govulncheck ./...` output is clean with exit code 0; vulnerabilities in uncalled modules like GO-2026-5942 in `x/net` were allowed to remain.

**Interpretation:** Ship the narrowest possible fix for the CI-breaking, code-reachable vulnerability, without hand-picking an `x/net` version, while still letting the module graph settle wherever `grpc`'s own requirements put it.

**Inferred intent:** Minimize the diff and the review surface — don't force an unrelated dependency bump that isn't needed to unblock CI, even if it happens to also be vulnerable in a way govulncheck doesn't consider blocking.

### What I did
Invoked the `fabrik:go` skill first, per project convention, before touching `go.mod`.

Ran `go get google.golang.org/grpc@latest golang.org/x/net@latest` initially (per the original brief), which bumped `grpc` to v1.83.2 and `x/net` to v0.59.0. After the scope-narrowing message arrived, reverted with `git checkout -- go.mod go.sum` and redid it as `go get google.golang.org/grpc@latest` alone, followed by `go mod tidy`. That left:

- `google.golang.org/grpc`: v1.82.1 -> v1.83.2 (direct ask; satisfies the >= v1.83.1 requirement)
- `golang.org/x/net`: v0.55.0 -> v0.58.0 (pulled in transitively because `grpc`'s own `go.mod` requires a newer `x/net`; this also happens to satisfy the >= v0.56.0 fix version from the original GO-2026-5942 advisory, even though that was no longer required)
- `golang.org/x/sync`: v0.21.0 -> v0.22.0 (transitive)
- `golang.org/x/text`: v0.39.0 -> v0.41.0 (transitive)
- `golang.org/x/sys`: unchanged at v0.47.0

No other `go.mod` requires changed, and everything stayed `// indirect`, matching the pre-bump shape of `go.mod`. `go.sum` also picked up `golang.org/x/crypto` v0.51.0 -> v0.55.0 (hash and `/go.mod` lines only): it isn't listed in `go.mod` because no built package imports it directly, but it's a real transitive requirement of the newer `grpc`/`x/net`, confirmed by `go mod graph`.

Correction to Step 1's advisory note: it says all three `grpc` advisories (GO-2026-6348, GO-2026-6441, GO-2026-6443) are "Fixed in v1.83.1". That's only true for two of them. Per the vuln.go.dev record for GO-2026-6443 (`curl -s https://vuln.go.dev/ID/GO-2026-6443.json`), the affected range is `[1.83.0, 1.83.2)` — v1.83.1 is still vulnerable, and the fix lands only in v1.83.2 (or v1.82.2 on the older branch). The requirement in this step's own prompt context ("at least v1.83.1") undersold that bar; the version actually needed, and the one that landed (v1.83.2), is correct, but a reader relying on "v1.83.1 fixes everything" from Step 1 would ship a build still vulnerable to GO-2026-6443.

Ran `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` (default, non-verbose): "No vulnerabilities found," exit code 0. Also ran the verbose form (`-show verbose ./...`) to check the "modules you require" category as the original brief asked: it also reports "No vulnerabilities found," listing all 74 scanned modules including `golang.org/x/net@v0.58.0` and `google.golang.org/grpc@v1.83.2` — so GO-2026-5942 is fully resolved too, as a side effect, even though the narrowed scope didn't require it.

Ran `go build ./...` (clean, no output) and `go test -shuffle on ./...`: all packages pass (`app/html`, `app/jobs`, `app/service`, `app/sqlite`, `app/sqlitetest` — the rest have no test files).

Confirmed the diff is `go.mod`/`go.sum` only via `git status --short` and `git diff go.mod`.

### Why
The coordinator's scope change treats the govulncheck-reachable vulnerability as the actual CI blocker, and wants the fix to be exactly that — not an opportunistic bump of an unrelated module. Using `go get` on `grpc` alone and letting `go mod tidy` resolve the graph is the mechanical way to get "only what `grpc` needs" without hand-picking `x/net`'s version.

### What worked
`grpc`'s own module requirements already needed a newer `x/net` than v0.55.0, so the transitive pull landed on v0.58.0 — comfortably past the v0.56.0 fix version — without me ever asking for `x/net` directly. This satisfies both the narrowed requirement (code-reachable vuln only) and, incidentally, the original stronger one (verbose output also clean).

### What didn't work
My first pass ran the original brief's `go get grpc@latest x/net@latest` combo before the scope-narrowing message arrived, which bumped `x/net` to v0.59.0 directly (a bigger jump than the transitive pull alone would have caused). I reverted cleanly with `git checkout -- go.mod go.sum` (working tree was otherwise clean, so this was safe) and redid the `go get`/`go mod tidy` sequence with `grpc` only.

### What I learned
`go get <module>@latest` alone can still move other modules' versions when the target's own `go.mod` requires newer transitive dependencies — `go mod tidy` after the `get` surfaces that. It's a useful way to get "the minimum needed by X" without a blanket `go get -u`.

### What was tricky
Nothing beyond the mid-task scope change, which was a clean revert-and-redo since no other files had changed yet.

### What warrants review
The diff is `go.mod`/`go.sum` only: `google.golang.org/grpc` v1.82.1 -> v1.83.2, `golang.org/x/net` v0.55.0 -> v0.58.0 (transitive), `golang.org/x/sync` v0.21.0 -> v0.22.0 (transitive), `golang.org/x/text` v0.39.0 -> v0.41.0 (transitive), plus a `go.sum`-only entry for `golang.org/x/crypto` v0.51.0 -> v0.55.0 (transitive, not in `go.mod`). All remain `// indirect`. `govulncheck` (both default and verbose) reports zero vulnerabilities; `go build ./...`, `go vet ./...`, and `go test -shuffle on ./...` all pass; `go mod verify` and `go mod tidy` (re-run, no-op) confirm the module graph is consistent.

Two competing self-review agents (model `fable`) independently checked this diff. Both confirmed build/vet/test/govulncheck/`go mod verify` are genuinely clean, and both independently caught the same issue: this diary originally said "no other modules moved," missing the `x/crypto` `go.sum` line — fixed above. One reviewer additionally caught that Step 1's "Fixed in v1.83.1" note doesn't hold for GO-2026-6443 (needs v1.83.2) — also corrected above, verified directly against the vuln.go.dev advisory JSON. No serious (correctness/security/architecture) issues were raised by either reviewer.

### Future work
None identified. The `x/net` line item in the original brief is effectively resolved as a side effect, but future dependency bumps to `grpc` should be re-checked against govulncheck since the pin is now dictated by `grpc`'s own requirements rather than a direct ask.
