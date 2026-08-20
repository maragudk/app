# Diary: Fix `service.version` always reporting `unknown` in OTel telemetry

Every span the app sends to Honeycomb carries `service.version=unknown`. The goal is to find
out why and make the deployed binary report the git commit it was built from.

## Step 1: Diagnose the missing version

**Author:** main

### Prompt Context

**Verbatim prompt:** In OTel telemetry, the service.version is always unknown (see e.g. c6 in Honeycomb). Figure out why. My hunch is the way we build the container image in CD, and the way that interacts with git versioning.

**Interpretation:** Trace where `service.version` comes from, find the point where it degrades to
`unknown`, and confirm or reject the hypothesis that the CD container build is responsible.

**Inferred intent:** Restore a usable `service.version` so telemetry can be correlated with
deploys, and understand the mechanism well enough that the fix is trustworthy rather than a guess.

### What I did

Traced the value from the span backwards. Nothing in this repo sets `service.version`; the whole
OTel setup lives in `maragu.dev/glue`. In `/app/app.go` glue passes
`otelconfig.WithServiceVersion(getVersion())`, and `getVersion()` reads the Go build info:

```go
func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return "unknown"
}
```

So `unknown` means the binary carries no `vcs.revision` stamp. I verified the Go behaviour
directly with a throwaway module in `/tmp/vcstest`: built with no `.git` present, the binary had
zero `vcs.*` settings; after `git init` plus a commit, the same source produced `vcs.revision`,
`vcs.time`, and `vcs.modified`. Go emits no warning in the first case, it just omits the settings.

Then I checked what the CD build actually receives, via `gh run view 31003897070 --log`:

```
docker buildx build ... https://github.com/maragudk/app.git#f58635ea418c494cc3096b8390f07732caa2d14e
#1 [internal] load git source https://github.com/maragudk/app.git#f58635ea...
```

`.github/workflows/cd.yml` passes no `context:` to `docker/build-push-action`, so it defaults to
the **Git context**. BuildKit clones the repo itself and strips the `.git` directory unless
`BUILDKIT_CONTEXT_KEEP_GIT_DIR=1` is set. The `COPY . ./` in `/Dockerfile` therefore copies a
working tree with no git metadata, `go build` stamps nothing, and the app reports `unknown`.

Finally I confirmed the chosen fix would work: a `git clone --depth 1` checked out at a detached
HEAD — exactly what `actions/checkout` produces — still stamps `vcs.revision` correctly.

I also filed maragudk/glue#192 for a related gap found along the way.

### Why

The hypothesis needed to be confirmed at each link rather than assumed, because "unknown" is a
plausible-looking value that could equally have come from a misconfigured exporter, a missing
env var, or Honeycomb-side column handling. Establishing that Go silently omits the stamp, and
that CD demonstrably uses the Git context, makes the fix a consequence of evidence instead of a
plausible story.

### What worked

Working backwards from the reported value to its source was fast, because the fallback string
`"unknown"` appears in exactly one place in the dependency tree. Reproducing the Go stamping
behaviour in a five-line throwaway module took seconds and turned the central claim from
"documented behaviour" into "observed behaviour".

### What didn't work

I tried to confirm the diagnosis against the actually-shipped binary by pulling the published
image and inspecting it with `go version -m`:

```
timeout 180 docker pull ghcr.io/maragudk/app:latest
pull exit: 127
```

Exit code 127 is "command not found" — there is no Docker in this environment. That leaves the
one link in the chain I could not observe directly, so the proof is deferred to the next deploy.

### What I learned

Go's VCS stamping fails open and silent. A build with no `.git` in the tree produces a working
binary with no version information and no diagnostic of any kind, which is why this survived
unnoticed for months. `actions/checkout` in `cd.yml` was doing nothing at all: with the default
Git context, `build-push-action` never reads the runner workspace.

The shallow, detached-HEAD checkout that `actions/checkout` produces by default is enough for
stamping — no `fetch-depth: 0` needed.

### What was tricky

The failure is invisible from inside the repo. Building locally with `make build-docker` works
correctly, because there is no `.dockerignore` and `.git` is present in the local context, so the
bug only manifests in CD. Anyone testing the theory locally would conclude the build was fine.

There is also a security wrinkle specific to the chosen fix. `actions/checkout` writes a
credential into `.git/config`; once `.git` is copied into the `gobuilder` stage, `cache-to:
type=gha,mode=max` exports every intermediate layer to the Actions cache, carrying that token
along. It never reaches the published image — the runner stage copies only `/bin/app` — and the
token is short-lived and repo-scoped on a public repo, so the real risk is negligible. Setting
`persist-credentials: false` removes it from the picture entirely, which is the reason that
option is part of the fix rather than an optional extra.

### What warrants review

The change is confined to `/.github/workflows/cd.yml`. Two things to check: that `context: .` is
present and the comment explains why (it exists so `.git` reaches `go build` for stamping), and
that `persist-credentials: false` is set on the checkout step. Validation is the next deploy —
`service.version` in Honeycomb should show a 40-character sha instead of `unknown`.

### Future work

`maragudk/glue#192` asks for the version to be logged next to the app name on startup, so a
broken build becomes visible in `app.log` immediately rather than requiring a Honeycomb query.
Once that lands, this repo bumps its glue dependency in the usual update pass.

The repo has no `.dockerignore`, so a local `make build-docker` sends `app.db`, `.worktrees`,
and `cover.out` into the build context. Pre-existing and harmless in CI, but worth tidying.

## Step 2: Implement the fix

**Author:** builder

### Prompt Context

**Verbatim prompt:** Actually, I think option 2 is the more transparent one. Feel free to add a
one-line comment on why we do that.

**Interpretation:** Of the two options identified in Step 1 for keeping `.git` available to
`go build` — `context: .` versus `build-args: BUILDKIT_CONTEXT_KEEP_GIT_DIR=1` — implement the
`context: .` approach, since it makes explicit in the workflow file that the checked-out
workspace (not the Git context) is what feeds the build. Document the reason inline so a future
reader doesn't mistake the line for boilerplate and remove it.

**Inferred intent:** Ship the diagnosed fix as a minimal, self-explanatory change to
`/.github/workflows/cd.yml`, so `service.version` in Honeycomb starts reporting the deployed
commit sha instead of `unknown`, without expanding scope into the Dockerfile, Go code, or the
separately tracked `maragudk/glue#192` logging improvement.

### What I did

Edited `/.github/workflows/cd.yml` in three small ways, all confined to the two existing steps
that Step 1 identified:

1. Added `context: .` to the `docker/build-push-action` step, with a one-line comment directly
   above it: `# Use the checked-out workspace (with .git) so go build can stamp vcs.revision for
   service.version.` This switches the build away from the default Git context (which strips
   `.git`) to the runner's checked-out workspace, making the preceding `actions/checkout` step
   actually load-bearing.
2. Added `persist-credentials: false` to the `actions/checkout@v7` step, per the security
   rationale Step 1 worked out: once `.git` is copied into the `gobuilder` build stage,
   `cache-to: type=gha,mode=max` would otherwise export the checkout's credential to the Actions
   cache along with every other intermediate layer.
3. Left everything else untouched — no `.dockerignore`, no `Dockerfile` changes, no Go changes,
   no `fetch-depth: 0`, no CI assertions.

Validated the result by parsing it with `python3 -c "import yaml; yaml.safe_load(open(...))"`,
which loaded cleanly.

### Why

`context: .` was the explicitly chosen approach (over `BUILDKIT_CONTEXT_KEEP_GIT_DIR=1`) because
it's visible in the workflow diff itself rather than buried in a build arg — a reviewer scanning
the `with:` block sees exactly which context feeds the build. The comment exists because that
line looks, out of context, like an unnecessary default value someone might "simplify" away;
one line is enough to stop that. `persist-credentials: false` closes the credential-in-cache gap
that only exists *because* of the `context: .` change — before this fix, `.git` never reached a
build stage, so the credential never had anywhere to leak to.

### What worked

The change was mechanical once Step 1 had already done the diagnostic work and named both the
exact lines to touch and the exact comment content — there was no ambiguity left to resolve
during implementation.

### What didn't work

Nothing failed. This was a config-only change with no code path to exercise locally; the only
verification available in this environment was YAML syntax validation.

### What I learned

Nothing new technically — this step was pure execution of a diagnosis and plan that Step 1 had
already fully worked out, including the exact comment wording and the security rationale for
`persist-credentials: false`.

### What was tricky

Nothing was tricky in the implementation itself. The only judgment call was keeping the comment
to a single line as instructed, while still naming both halves of the causal chain (`.git` →
`go build` → `vcs.revision` → `service.version`) so a future reader doesn't need to re-derive
Step 1's investigation to understand why the line matters.

### What warrants review

Confirm in `/.github/workflows/cd.yml` that `context: .` sits under the `docker/build-push-action`
step with its explanatory comment directly above it, and that `persist-credentials: false` sits
under `actions/checkout@v7`. Real validation happens on the next push to `main`: check the
resulting workflow run's build logs for `[build] load build definition from Dockerfile` sourced
from the local context rather than a Git clone, then confirm in Honeycomb that new spans carry a
40-character `service.version` instead of `unknown`.

### Future work

None beyond what Step 1 already named: `maragudk/glue#192` (log the version on startup) and the
missing `.dockerignore` are both out of scope here and tracked separately.

## Step 3: Pick up the `glue` version-logging fix

**Author:** builder

### Prompt Context

**Verbatim prompt:** Follow-up task in the same worktree. Bump `maragu.dev/glue` to pick up commit
`210b9717` ("Log the service version on app startup"), which closes `maragudk/glue#192` from
Step 1's future work. Confirm no call sites need changes, build with
`-tags sqlite_fts5,sqlite_math_functions`, and verify end to end by running the binary and
checking the "Starting app" log line's `version` field. Also check whether Go's VCS stamping
behaves the same from inside this git worktree (where `.git` is a file, not a directory) as it
does from a normal checkout, since that was unverified.

**Interpretation:** A dependency bump plus verification, not new development.

**Inferred intent:** Close out `maragudk/glue#192` now that it's landed upstream, and flag
anything about this repo's worktree-based dev setup that could mislead someone reading
`service.version` locally.

### What I did

`go get maragu.dev/glue@main` first resolved to `v0.0.0-20260819080311-c34a99649e6f`, whose
`app/app.go` still logged `"Starting app"` with no `version` field — the public module proxy's
`@main` cache was stale. `GOPROXY=direct go get maragu.dev/glue@main` resolved to
`v0.0.0-20260820083844-210b9717c337`, the exact commit named in the task, confirmed by grepping
its `app/app.go` for the described `version := getVersion()` / `log.InfoContext(...,"version",
version)` change. `GOPROXY=direct go mod tidy` followed; it also bumped
`github.com/mattn/go-sqlite3` v1.14.34 → v1.14.49 as a transitive change in `glue`'s own `go.mod`.

`go doc maragu.dev/glue/app.Start` confirms the exported signature (`func Start(startCallback
StartFunc)`) is unchanged, and `go build -tags sqlite_fts5,sqlite_math_functions ./...` plus `go
vet` (same tags) both pass with zero edits to `/cmd/app/main.go` or any other call site.

Built the binary to a scratch path and ran it briefly (no `.env` in this worktree, so all
`maragu.dev/env` defaults applied). First log line:

```json
{"msg":"Starting app","name":"App","version":"f58635ea418c494cc3096b8390f07732caa2d14e"}
```

A real 40-char sha, not `unknown` — but it's the wrong sha: `f58635e...` is `main`'s HEAD from
before Step 2's commit, not this worktree's actual HEAD (`bd84f9d...`, confirmed with `git
rev-parse HEAD`). A `-a` rebuild produced the identical stamp, ruling out build-cache staleness.
The worktree's own git metadata explains it: `cat .git` shows `gitdir:
.../app/.git/worktrees/otel-service-version`, whose `HEAD` file correctly reads `ref:
refs/heads/worktree-otel-service-version` — but `git rev-parse --git-common-dir` resolves to
`/Users/maragubot/Developer/app/.git`, and *that* directory's `HEAD` reads `ref: refs/heads/main`.
Go's VCS stamping picked up the common dir's HEAD, not the worktree-specific one. I couldn't
build from `/Users/maragubot/Developer/app` directly to compare — this session's sandbox refused
a `git -C /Users/maragubot/Developer/app ...` command outright — but the ref-file evidence alone
already isolates the cause, and since this project nests worktrees inside the main checkout's own
tree (`.claude/worktrees/<name>`), it's consistent with Go's repo-root detection not special-casing
a `.git` *file* and falling through to the nearest ancestor `.git` *directory* it can find.

Past that log line the run continued normally (DB connected, migrations found, job runner
started) then failed on `listen tcp :8080: bind: address already in use` — expected, since `make
watch` is presumably already running elsewhere on this machine, and unrelated to this change.
Cleaned up the `app.db*` files the run created (gitignored, but tidied anyway).

Did not run `make test`: it needs `docker compose up versitygw-test`, and there is no Docker in
this environment.

### Why

The value here was confirming the bump is inert at the call-site level and getting a real
observed log line rather than trusting the diff description. The worktree check turned up a
subtler failure mode than the one anticipated (`unknown`): a plausible, real sha that's simply
the wrong commit — worth recording precisely because a naive "is it `unknown`" check wouldn't
catch it.

### What worked

`GOPROXY=direct` cleanly bypassed the stale proxy cache with no other workarounds needed. The
worktree's own git metadata was sufficient to explain the stamping discrepancy without a
side-by-side build.

### What didn't work

The first `go get @main` silently resolved to a version missing the target commit — no error, it
just wasn't current, caught only by grepping the fetched source. `make test` wasn't runnable at
all: no Docker in this environment (consistent with Step 1's `docker pull` failure), so I'm
reporting that rather than claiming any test result.

### What I learned

The public Go proxy's `@main` resolution can lag the real branch tip; `GOPROXY=direct` is the
fix when freshness matters more than caching. Separately: in this repo's worktrees-nested-inside-
the-checkout layout, a binary built inside a worktree stamps `vcs.revision` from the *main*
checkout's HEAD, not the worktree's own branch tip. That's irrelevant to Step 2's CD fix — CD
never builds inside a nested worktree — but it means `service.version` from a worktree-built
binary shouldn't be trusted as evidence of which commit is running; `git rev-parse HEAD` stays
accurate, only Go's own VCS auto-detection is fooled.

### What was tricky

Ruling out stale build cache vs. genuine misresolution took an extra `-a` rebuild to be sure. The
sandbox boundary blocked the most direct confirmation (building the same source from
`/Users/maragubot/Developer/app` to diff the stamps), so the explanation rests on git-metadata
inspection rather than a side-by-side build.

### What warrants review

Confirm `go.mod`/`go.sum` pin `maragu.dev/glue v0.0.0-20260820083844-210b9717c337` and
`github.com/mattn/go-sqlite3 v1.14.49`, and that `/cmd/app/main.go` is untouched. The
worktree-stamping finding is worth an independent check from `/Users/maragubot/Developer/app`
directly — I'd expect it to stamp its own HEAD correctly there, matching what CD's checkout
produces, but couldn't verify it from this sandboxed session.

### Future work

None proposed as code — this is a local-dev-only observation with no effect on the shipped CD
fix. If it's worth guarding against, it'd mean documenting that `service.version` from a
worktree-built binary isn't trustworthy for manual verification.
