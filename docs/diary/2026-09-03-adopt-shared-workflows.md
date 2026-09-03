# Diary: Adopt shared GitHub Actions workflows from `maragudk/workflows`

Replace the app's copy-pasted CI jobs with calls to the reusable workflows in `maragudk/workflows`, following the starter templates in `maragudk/.github`. The app is a template: it keeps scaffolding (the versitygw S3 service container, SQLite build tags) that downstream apps may need even though nothing here uses them yet. Reusable workflows cannot receive `services:` from the caller, so the shared `test.yml` and `compatibility.yml` grow the scaffolding themselves: the SQLite build tags always, and a versitygw service behind an `s3` boolean input. Done one repo at a time: `workflows` first, then this app, then the starter templates in `maragudk/.github`.

## Step 1: Survey and requirements

**Author:** main

### Prompt Context

**Verbatim prompt:** "I've set up shared workflows in ../workflows, also available at maragudk/workflows, see examples in maragudk/.github . Can we adopt them here?" Followed by: "Okay, but remember this is an app template, so apps can use the features we don't, and we should have the scaffolding for that."

**Interpretation:** Convert this repo's four workflows in `/.github/workflows/` to use the reusable workflows where possible, without removing template scaffolding.

**Inferred intent:** Stop CI drift across the ~24 repos in the org by having one source of truth for job definitions, while keeping this template a complete starting point for apps that do use S3 and FTS5.

### What I did

Had an Explore agent read `/Users/maragubot/Developer/workflows`, `/Users/maragubot/Developer/.github/workflow-templates/`, and this repo's `/.github/`, `/Makefile`, `/go.mod`, and diary. Verified by grep that no `*_test.go` references `s3`, `bucket`, `versitygw`, `7072`, or the versitygw access-key env vars. Wrote requirements and handed them to a builder.

### Why

The shared workflows have zero inputs and zero secrets by design, so the only question was which jobs can be called as-is. Two constraints rule out `test` and `compatibility`: reusable workflows cannot take `services:` (needed for versitygw) and there is no input for `-tags sqlite_fts5,sqlite_math_functions`. Markus decided the versitygw container stays because the template must scaffold S3 for downstream apps, and that the shared workflows should carry that scaffolding rather than every repo keeping an inline job. The build tags are hardcoded upstream (unmatched tags are a no-op for repos without go-sqlite3); the service is gated behind an `s3` input so repos without S3 do not pull a container per run.

### What worked

The upstream repo's `docs/decisions.md` and diary already named the `services:` limitation as the known reason repos keep an inline test job, which made the trade-off easy to put in front of Markus. A first plan (hybrid: two inline jobs, four callers) was rejected in favour of changing upstream, since the point of the shared repo is to have no copy-pasted jobs at all.

### What didn't work

Nothing failed at this stage.

### What I learned

The `context: .` in `/.github/workflows/cd.yml` exists so `go build` inside Docker sees `.git` and stamps `vcs.revision`, which becomes `service.version` in OTel (see `/docs/diary/2026-08-20-otel-service-version-unknown.md`). The shared `cd.yml` also uses `context: .`, so calling it preserves that behaviour, but the explanatory comment leaves this repo. This entry is now where that rationale lives.

### What was tricky

Deciding whether the versitygw service is dead weight. It is unused by tests today but is deliberate template scaffolding, so it stays. The upstream decision log says "an input is added when a second repository needs it"; the template plus every app stamped from it counts as that second repository.

### What warrants review

That the inline `test` job mirrors upstream `test.yml` except where it must diverge, and that check names in any branch protection rule on `maragudk/app` are updated (they change from `Lint` to `lint / lint`, etc.).

### Future work

Consider `modernc.org/sqlite` to make the build tags unnecessary altogether.

## Step 2: Convert the four workflows to pure callers

**Author:** app-builder

### Prompt Context

**Verbatim prompt:** "I've set up shared workflows in ../workflows, also available at maragudk/workflows, see examples in maragudk/.github . Can we adopt them here?"

**Interpretation:** Now that the shared `test.yml` and `compatibility.yml` carry the SQLite build tags and an `s3` input, rewrite `/.github/workflows/{ci,cd,compatibility,security}.yml` as callers that follow the starter templates exactly, with `s3: true` on the two workflows that need the versitygw container. Nothing inline should survive.

**Inferred intent:** Make this template repo the worked example of the shared workflows, so an app stamped from it inherits CI that stays current without anyone copying job definitions around.

### What I did

Branched `adopt-shared-workflows` off `main` and rewrote all four workflow files from the starter templates, substituting `main` for `$default-branch`. `/.github/workflows/ci.yml` calls `lint.yml`, `test.yml` (with `s3: true`) and `build.yml`, and gains the top-level `permissions: contents: read` the template has and the old file lacked. `/.github/workflows/compatibility.yml` calls the shared `compatibility.yml` with `s3: true`. `/.github/workflows/security.yml` and `/.github/workflows/cd.yml` are pure callers with no inputs. Every call is pinned to `@main`. The net change is 24 lines in place of 175.

Before committing I ran the shared workflows' exact test invocation locally, since it is stricter than what this repo ran before:

```
go mod tidy && git diff --exit-code go.mod go.sum
go build -tags sqlite_fts5,sqlite_math_functions,sqlite_foreign_keys ./...
go test -race -shuffle on -tags sqlite_fts5,sqlite_math_functions,sqlite_foreign_keys ./...
```

All three passed. Then pushed the branch, opened PR #65, and dispatched `compatibility.yml` on the branch with `gh workflow run compatibility.yml --ref adopt-shared-workflows`.

I also inspected branch protection on the repository, as asked, without changing anything. `gh api repos/maragudk/app/branches/main/protection` returns `404 Not Found`, so there is no classic protection rule. The rules live in a repository ruleset instead (`gh api repos/maragudk/app/rulesets`): one active ruleset named `main`, targeting `~DEFAULT_BRANCH`, whose `required_status_checks` rule lists exactly four contexts -- `Test`, `Lint`, `Build amd64` and `Build arm64`.

### Why

The starter templates are the contract. Deviating from them here would defeat the point of having them, so the only intended differences are the `$default-branch` substitution, the two `s3: true` inputs, and uncommenting the template's `build` job because this repo has a `/Dockerfile`. Running the stricter test command locally first was cheap insurance: the shared `test.yml` adds `-race`, a third build tag `sqlite_foreign_keys`, and a `go mod tidy` cleanliness check, none of which this repo's CI had ever exercised. Discovering a tidy failure from a CI log would have been slower than discovering it from a shell.

### What worked

Everything, on the first attempt. `Security` went green in under a minute. The local dry run of the stricter test command meant the new `go mod tidy` gate and `-race` were known-good before the branch was even pushed.

Diffing each new file against its template was a useful check on my own work:

```
diff /Users/maragubot/Developer/.github/workflow-templates/ci.yml .github/workflows/ci.yml
```

The output was exactly the three intended differences and nothing else, which is a stronger statement than reading the file and judging it looks right.

### What didn't work

Nothing failed. No error messages to record for this step.

### What I learned

Branch protection on this repository is a ruleset, not a classic protection rule, so `gh api repos/OWNER/REPO/branches/main/protection` returning 404 means "look at rulesets", not "nothing is protected". That 404 is easy to misread as an all-clear.

The shared `lint.yml` keeps the `if: github.triggering_actor != 'dependabot[bot]'` guard on its own job, not on the caller's. The behaviour is unchanged -- Dependabot pull requests still skip linting -- but the skip now shows up one level down, on `lint / lint`.

The shared `cd.yml` uses `password: ${{ github.token }}` where the old file used `${{ secrets.GITHUB_TOKEN }}`. Those are the same token by two names, so the GHCR login is unaffected.

### What was tricky

The check names. Because a called workflow's jobs are reported as `<caller job> / <called job>`, all four names in the ruleset change at once: `Lint` becomes `lint / lint`, `Test` becomes `test / test`, and `Build amd64`/`Build arm64` become `build / build-amd64`/`build / build-arm64`. A required check whose name no longer exists never reports, so the ruleset will hold every pull request in a permanently pending state until the four contexts are renamed. This has to be done at, or just before, merge -- and it is deliberately not part of this branch, since the ruleset is repository configuration rather than code.

The other subtlety is that `cd.yml` cannot be exercised from a branch: it only triggers on push to `main`. It is verified by reading it against the file it replaces -- same `context: .`, same `platforms: linux/amd64`, same `latest` and `${{ github.sha }}` tags, same GHCR registry, same buildx cache. The only difference is the step order (the shared workflow logs in before setting up buildx), which does not matter.

### What warrants review

The four required status check names in the `main` ruleset, which is the one thing that can silently block merges. Beyond that, the three intended deviations from the starter templates: `s3: true` in two places and the uncommented `build` job.

Worth a moment's thought is whether `issues: write` in `/.github/workflows/security.yml` is acceptable. The shared `security.yml` opens a `govulncheck` issue when the scan fails and closes it when the scan recovers, which needs the permission. The issue steps are guarded on `github.ref_name == 'main'`, so a pull request cannot make the workflow file anything, and the token is read-only on fork pull requests regardless.

### Future work

The old `test` job ran without `-race`; the shared one enables it. If the suite ever grows slow enough that this hurts, the fix belongs upstream as an input rather than as a local divergence.

The versitygw container is still started for a repository whose tests do not use it. That is deliberate scaffolding, but the day a test does connect to `http://localhost:7072`, someone should check whether the service needs a health check -- a service container reports ready as soon as it starts, not as soon as it serves requests.
