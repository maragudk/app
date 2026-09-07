# Diary: Drop the Compatibility workflow

Remove `/.github/workflows/compatibility.yml`, the nightly caller of the shared `maragudk/workflows` compatibility workflow. Scope is this repository only: the shared workflow and the starter template in `maragudk/.github` stay for other repositories.

## Step 1: Requirements

**Author:** main

### Prompt Context

**Verbatim prompt:** "Let's drop the compatibility workflow"

**Interpretation:** Delete the Compatibility GitHub Actions workflow from this repository.

**Inferred intent:** Stop the nightly `go get -u -t ./...` run for the template. It has been green every night, and the maintenance signal it gives is not worth a scheduled job per repository stamped from the template.

### What I did

Read the four files in `/.github/workflows/`, the previous diary entry `/docs/diary/2026-09-03-adopt-shared-workflows.md`, and the starter template at `maragudk/.github/workflow-templates/compatibility.yml`. Grepped the repository for `compatibility`: the only hits outside the workflow itself are historical diary entries, which stay as they are. Listed recent runs with `gh run list --workflow=compatibility.yml`: all green, nightly on `main`. Checked for open issues mentioning compatibility: none.

Asked Markus how far "drop" reaches. Answer: this repository only.

### Why

The workflow is a 13-line caller with no other references, so the only real question was scope. Removing the shared workflow or the starter template affects other repositories and was explicitly ruled out.

### What worked

The previous diary entry made the shape of the workflow files obvious without any exploration.

### What didn't work

Nothing failed at this stage.

### What I learned

The `/docs/diary/2026-09-02-datastar-1.0.3-upgrade.md` entry cites the nightly `go get -u -t ./...` as the reason for pinning the Datastar bundle version in a test. That test still has value without the nightly job -- a manual `go get -u` can move the pins apart just as well -- so nothing there needs to change.

### What was tricky

Nothing. The change is a single file deletion.

### What warrants review

That only `/.github/workflows/compatibility.yml` is removed and the other three workflows are untouched. The `main` ruleset does not list any compatibility check, so branch protection is unaffected.

### Future work

None implied.

## Step 2: Delete the workflow

**Author:** app-builder

### Prompt Context

**Verbatim prompt:** "Let's drop the compatibility workflow"

**Interpretation:** Carry out the deletion scoped by Step 1: remove `/.github/workflows/compatibility.yml` from this repository only, on a branch, and open a PR.

**Inferred intent:** Turn the reviewed decision into a mergeable change with a clean diff and a diary trail.

### What I did

Created the branch `drop-compatibility-workflow` off `main` and ran `git rm .github/workflows/compatibility.yml`. Confirmed `/.github/workflows/` now contains exactly `cd.yml`, `ci.yml`, and `security.yml`. Appended this step to the diary. Committed the deletion and the diary together, pushed the branch, and opened a PR against `main`.

### Why

Step 1 already settled the scope question (this repository only) and confirmed there were no other references to clean up, so the build step is the deletion itself plus the paperwork around it.

### What worked

Nothing surprising -- `git rm` on a single file is about as low-risk as a change gets, and the previous diary entry meant no re-investigation was needed.

### What didn't work

Nothing failed. There are no tests to run for a workflow YAML deletion; validation is `git diff main --stat` showing exactly the two files and the untouched workflows still triggering on the PR.

### What I learned

Nothing new -- this step confirmed what Step 1 already established.

### What was tricky

Nothing. Single-file deletion.

### What warrants review

That `git diff main --stat` shows only `.github/workflows/compatibility.yml` (deleted) and `docs/diary/2026-09-07-drop-compatibility-workflow.md` (added), and that `ci.yml`, `cd.yml`, and `security.yml` are byte-for-byte untouched.

### Future work

None implied.
