# Diary: per-operation wiring functions for `service.Fat`

`service.Fat` was a flat struct: three capability fields (`bucket`, `db`, `sender`) that every method
could reach, and one method that reached exactly one of them. This task restructures it so that each
operation declares the capabilities it uses and can reach nothing else.

## Step 1: apply the pattern to `GetUser`

**Author:** claude

### Prompt Context

**Task:** restructure `service.Fat` so that every operation is wired by an exported package-level
function taking `*Fat` plus narrow private interfaces, with the public method reduced to a delegate
that panics when unwired; keep `NewFat` for global configuration only; add a wiring-only `Setup`;
make `servicetest.NewFat` construction-only; prove the wiring with an internal test.

**Interpretation:** mirror how HTTP handlers are already registered in `/http` — a function per
thing, taking its dependencies as point-of-use interfaces — and apply it to the service layer.

**Inferred intent:** this template is where new projects start, so it should start with the
dependency discipline already in place rather than with a flat struct that has to be untangled later.

### What I did

Rewrote `/service/fat.go`. `Fat` now holds a `log` field and one private func field per operation
(`getUser`). `NewFatOptions` shrank to `Log`, defaulting to `slog.DiscardHandler`. `Setup(f, db)`
wires every operation and is the only place inside `service` where a concrete dependency type
appears. `GetUser(f, db userGetter)` sets the field with a method value, and `(*Fat).GetUser`
delegates, panicking with `service: GetUser not wired; call service.GetUser or service.Setup` when
the field is nil.

Deleted the `bucket` and `sender` fields. Neither was ever read: `NewFat` stored them and nothing
else touched them.

`/servicetest/service.go` is now construction-only: `service.NewFat` with a logger that writes
through `t.Log` (a private `testWriter`, mirroring the convention in glue's `sqlitetest`). Its
`NewFatOption`/`WithSQLiteTestOptions` machinery is gone, because a helper that wires nothing has no
database to pass options to. Tests build their own dependencies and hand them to the wiring function
they exercise.

Added `/service/fat_internal_test.go` with `TestSetup`, and `/service/fat_test.go` covering `GetUser`
against a real SQLite database via the `admin` fixture plus the unwired panic.

In `/cmd/app/main.go`, `service.NewFat` now takes only a logger and is followed by
`service.Setup(svc, db)`. Removing the bucket from `Fat` left `bucket` and `awsConfig` with no
consumer in `main`, so the `aws.LoadDefaultConfig` and `s3.NewBucket` block went with them.

### Why

The gain is declared and enforced dependencies. A wiring function can only give an operation what it
was handed, which is stronger than a struct field that merely happens to be unused today — and this
struct had already collected two of those.

`Setup` exists so `main` stays the composition root and so there is one list of what each operation
depends on. The narrow interfaces are for the operations, not for `Setup`, which is why `Setup` names
concrete types.

### What worked

The mapping was clean. With one operation there is exactly one wiring function, and the trivial case
collapses to a method value: `func GetUser(f *Fat, db userGetter) { f.getUser = db.GetUser }`. The
`http.userGetter` interface in `/http/auth.go` keeps being satisfied, because the delegate method is
still a method — that is the whole reason the pattern uses delegates instead of exported func fields.

I expected `golangci-lint`'s `unused` check to complain about `Fat.log`, since no operation logs yet.
It does not: the composite literal in `NewFat` counts as a use, so `make lint` reports `0 issues.`

### What didn't work

`make test` never got as far as the Go tests:

```
docker compose up -d versitygw-test
Error response from daemon: failed to set up container networking: driver failed programming external
connectivity on endpoint app-versitygw-test-1 ...: Bind for 0.0.0.0:7072 failed: port is already allocated
```

An unrelated container held 7072. That specific collision is environmental, but the dependency itself
is now pointless: no test in this repo touches S3, and none did before either — `s3test` appeared
only in `servicetest.NewFat`, which nothing called. I ran the suite with
`go test -tags sqlite_fts5,sqlite_math_functions -shuffle on ./...`, which passes.

### What I learned

The unused capability fields were not dead weight only in the struct. `bucket` in `/cmd/app/main.go`
existed solely to be handed to `service.NewFat`, so deleting the field orphaned the construction too —
Go then refuses to compile the unused local. A field nothing reads can still hold a whole subsystem's
setup in place.

### What was tricky

Whether to delete the S3 setup from `main` at all. It arrived with the original glue scaffolding and
is a thing a template is expected to demonstrate. Keeping it would have meant either keeping a field
no method reads or passing a bucket to `Setup` that no wiring function receives — and `Setup`'s
parameter list is supposed to be the honest list of what the operations need. I deleted it: about a
dozen lines and two imports, which an operation that needs a bucket brings back along with its wiring
function.

### What warrants review

The removal of `aws.LoadDefaultConfig` and `s3.NewBucket` from `/cmd/app/main.go` is the judgement
call, and the easiest thing to disagree with.

Worth a second look too: `servicetest.NewFat` losing its options, since that is a small API break for
anyone already building on the template; and whether `Fat.log` should exist before any operation logs
through it.

### Future work

An unwired operation is a runtime panic, not a compile error, and `glue`'s HTTP router installs no
recovery middleware — so an unwired operation reached from a global middleware such as
`http.AddUserToContext` breaks requests rather than the boot. `TestSetup` is the guard.

## Step 2: self-review

**Author:** claude

### Prompt Context

**Task:** self-review the change with two competing reviewers and address what they find.

**Interpretation:** run the review, act on what both reviewers found independently plus anything
serious either found alone, and record the calls I did not act on.

**Inferred intent:** catch the things a single pass misses, particularly where the pattern's claims
outrun what the code enforces.

### What I did

Both reviewers landed on the same four substantive points, and I fixed all four.

`TestSetup` asserted `f.getUser != nil` by hand, which cannot catch the failure it exists to catch: an
operation that gains a wiring function but never a line in `Setup`. It now reflects over `*Fat`,
asserts every field of func kind is non-nil, and names the field in the failure message. It also
asserts it found at least one, so the loop cannot pass vacuously.

The doc comments looked outward at their callers ("wire the ones under test", "which is what a
running app wants"), which the house style forbids, and `Setup`'s comment called itself the
composition root when `main` is. Rewritten.

The `Fat` doc claimed "an operation cannot reach a capability it did not declare" without saying why
that holds — a wiring function does receive `*Fat`, so a closure can reach its fields. What makes the
claim true is that `Fat` holds no capabilities, so the doc now says so. It also documents the
lifecycle both reviewers worried about: the func fields are written by the wiring functions and read
afterwards, so a wired `Fat` is safe for concurrent use and rewiring one in use is not.

`TestNewFat` tested `(*Fat).GetUser` panicking rather than anything about `NewFat`. The panic subtest
moved to `TestFat_GetUser`, and `TestNewFat` moved into the internal test file where it can assert
what `NewFat` actually promises: a nil `Log` gets `slog.DiscardHandler`. I also dropped the
nonexistent-user subtest, which duplicated `/sqlite/auth_test.go` through a pass-through method.

### Why

The reflection version of `TestSetup` is the one that earns the comment above it. Everything else was
a doc or a test saying something the code did not.

### What worked

Verifying the new guard by breaking `Setup` on purpose:

```
--- FAIL: TestSetup/wires_every_operation
    fat_internal_test.go:35: operation getUser is not wired
```

`reflect.Value.IsNil` reads unexported fields without `Interface()`, so the test needs no export and
no build tag.

### What didn't work

Nothing failed in this step. `make lint` stays at `0 issues.` and the suite passes.

### What was tricky

Both reviewers wanted the S3 fallout finished: `make test` and CI still start a versitygw container
for a suite with no S3 tests, `.env.example` still documents `S3_BUCKET_NAME` and `S3_PATH_STYLE`
that nothing reads, and `go mod tidy` would now drop the aws-sdk indirect requirements. They are
right that the half-state is the least coherent of the options. I still left all of it alone,
including `go mod tidy`, because tidying bakes into the dependency graph the decision this change is
asking to have made — whether the template ships S3 support at all — and that is not the same
question as how the service is wired.

Three suggestions I did not take, on purpose: renaming the wiring function to `WireGetUser` to avoid
reading like a getter (it is named after what it wires, exactly as the handler registrations in
`/http` are); giving `Setup` an options struct (positional is what keeps it a wiring list); and
replacing the real-database test with a stub `userGetter` (the repo prefers real dependencies, and
the surviving test doubles as the template's example of wiring an operation under test).

### What warrants review

Same as step 1, plus: `Fat.log` still has no reader, which both reviewers noted is the same shape as
the fields this change deleted. I kept it because it is ambient state the pattern is defined around
rather than a capability, and the first operation that logs will use it — but it is a fair thing to
strike.

### Future work

If the S3 scaffolding is meant to stay in the template, it needs an operation that uses it; if not,
the compose services, CI service containers, Makefile targets and environment variables should go
together with a `go mod tidy`.

## Step 3: keep unused capabilities plumbed through `Setup`

**Author:** claude

### Prompt Context

**Task:** in this repo the sender stays even though no operation uses it, and so does any other
dependency in the same position — the bucket, on inspection. Rather than sitting on `Fat`, an unused
capability belongs in the composition path: a `Setup` parameter, with a note that it awaits the first
operation to wire it.

**Interpretation:** reverse step 1's deletion of the S3 setup from `/cmd/app/main.go`, and route both
capabilities through `Setup` instead of dropping them.

**Inferred intent:** this is a starter template, so the plumbing is the product. A project that
starts here and writes its first email-sending operation should find the sender already flowing.

### What I did

`Setup` is now `Setup(f *Fat, bucket *s3.Bucket, db *sqlite.Database, sender *postmark.Sender)`, with
a doc sentence saying that a capability no operation wires yet is a parameter all the same, and
naming the two that are waiting. `/cmd/app/main.go` builds the bucket again — the
`aws.LoadDefaultConfig` and `s3.NewBucket` block is restored exactly as it was — and passes both to
`Setup`. `TestSetup` passes three nils.

Checked the bucket before restoring it: on `main` it reached `service.NewFat` and nothing else, so it
is in exactly the sender's position. The sender also feeds `jobs.Register`, so it was never at risk
of disappearing from `main`.

### Why

`Fat` still holds no capabilities, so the guarantee the pattern is for is untouched: an operation can
still only reach what its wiring function was handed. What changes is where an unwired capability
waits — in the composition path rather than in the struct, where it costs nothing and is one wiring
function away from use.

### What worked

Unused function parameters are legal Go and `golangci-lint` has nothing to say about them: `0 issues.`
The diff of `/cmd/app/main.go` against `main` is now three lines, all of them the wiring change.

Everything step 2 recorded as unfinished S3 fallout is moot. The bucket is constructed again, so the
compose services, CI service container, Makefile targets and `.env.example` entries all mean
something, and `go mod tidy` has nothing to remove.

### What didn't work

Nothing failed here.

### What I learned

"Delete what nothing uses" is a repo-dependent instinct. In an application it is right, and it was
right in the project this pattern came from. In a starter template the unused plumbing is a large
part of what the reader came for, and deleting it optimises for the wrong repository.

### What was tricky

Finding the placement that keeps the pattern honest. On `Fat` the capability would be reachable by
every operation, which is the thing the change exists to prevent. As a `Setup` parameter it is
reachable by nothing until a wiring function asks for it. The cost is that `service` now imports
`postmark` and `s3` for a signature alone.

### What warrants review

`Setup`'s parameter list is where unused capabilities accumulate, and the doc sentence naming which
ones are waiting has to be pruned as they get wired. Worth watching that it stays accurate.

### Future work

The first operation that sends an email wires the sender through its own wiring function, and the
sentence in `Setup`'s doc loses a name. Same for the bucket.

## Step 4: stock the tracer alongside the logger

**Author:** claude

### Prompt Context

**Task:** give `Fat` a `tracer trace.Tracer` field, initialized in `NewFat` with
`otel.Tracer("app/service")`, ready for the first operation that traces.

**Interpretation:** the same "plumbing is the product" reasoning as the bucket and the sender, applied
to ambient state rather than to a capability.

**Inferred intent:** two stocked-and-waiting fields make the policy legible. One on its own reads as
an oversight, which is how the reviewers in step 2 read `Fat.log`.

### What I did

Added `tracer trace.Tracer` to `Fat`, set in `NewFat` from `otel.Tracer("app/service")`, matching the
`otel.Tracer("app/http")` in `/http/auth.go`. It is not a `NewFatOptions` field: the tracer is the
package's own, which the options doc now says. `TestNewFat` gained a subtest asserting `NewFat`
stocks it, which also means both ambient fields have a reader.

### Why

An operation that wants a span should find a tracer on `f` rather than reaching for `otel` itself, so
that the instrumentation name stays one decision for the package instead of one per operation.

### What worked

Nothing surprising: `go vet`, `make lint` at `0 issues.`, suite green. `unused` leaves the field alone
for the same reason it leaves `log` alone — the composite literal in `NewFat` counts as a use — and
the new subtest reads it besides.

### What didn't work

Nothing failed here.

### What I learned

The two ambient fields answer the objection raised against the first one. A single unused field looks
like something left behind; a stocked logger and a stocked tracer look like what they are, which is a
template deciding what every operation gets for free.

### What was tricky

Only the wording. Saying in the `Fat` doc that these are "waiting for their first operation" would go
stale the moment one uses them, so the doc just names what `Fat` carries, and the awaiting-first-use
status lives here and in the pull request. `Setup`'s parameter note is different: unused parameters
need the explanation on the spot, or a reader files them as a mistake.

### What warrants review

Whether `otel.Tracer("app/service")` is the instrumentation name wanted long term. It mirrors
`app/http`, so the two are at least consistent.

### Future work

The first operation to open a span uses `f.tracer`, and the first to log uses `f.log`.

## Step 5: make the wire-once rule enforced instead of documented

**Author:** claude

### Prompt Context

**Task:** review feedback on the `Fat` doc's lifecycle sentence — rewiring an already wired service
should panic rather than be described as a thing not to do. Guard in each wiring function, trim the
doc to match, and test the panic and its message.

**Interpretation:** the doc sentence added in step 2 was a rule with nothing behind it; turn it into
code and let the doc state the consequence.

**Inferred intent:** a rule a reader has to remember is weaker than one the program refuses to break,
and this one is cheap to enforce.

### What I did

`GetUser` now checks its field before setting it and panics with `service: GetUser already wired`.
The `Fat` doc keeps the concurrency consequence — fields written once by the wiring functions, read
after that, so a wired `Fat` is safe for concurrent use — and the clause about rewiring became a
statement that it panics. Added a subtest asserting the panic and its exact message by wiring
`GetUser` twice on one `Fat`.

### Why

The two panics bracket the lifecycle from both ends. The delegate's panic says an operation was
called before it was wired; this one says it was wired twice. Between them, an operation runs against
exactly one set of capabilities, chosen once, which is the property the whole change is for.

### What worked

Verified the test earns its place by deleting the guard and running it:

```
--- FAIL: TestFat_GetUser/panics_when_the_operation_is_wired_twice
    fat_test.go:47: expected a panic
```

The subtest sits next to the unwired-method panic, so the two ends of the lifecycle read together.

### What didn't work

Nothing failed here.

### What I learned

Writing the guard made the reason for it sharper than the doc sentence had been. The question is not
really whether rewiring races — it is that a second wiring silently replaces the first, so an
operation would run against capabilities nobody at the call site chose.

### What was tricky

Only where the subtest belongs. It exercises the package-level `service.GetUser` rather than the
method, which argues for a `TestGetUser` of its own, but splitting the two panics across two test
functions hides that they are one lifecycle. It went next to the unwired panic.

### What warrants review

Every future wiring function needs the same three lines, and nothing enforces that — the same
hand-maintenance problem `TestSetup` had before it became reflective. A helper along the lines of
`mustNotBeWired(f.getUser, "GetUser")` would at least make an omission visible as a missing call.

### Future work

If the wiring functions multiply, fold the guard and its message into one helper, so a new operation
cannot pick up half the lifecycle and skip the other half.
