# Diary: Hashed dev ports

Each checkout, including worktrees, should get its own consistent ports for the app's HTTP server, the dev versitygw (S3 and web UI), and the test versitygw. The port comes from a hash of the working directory plus a name, so parallel worktrees don't collide, and `SERVER_ADDRESS`, `BASE_URL`, and `AWS_ENDPOINT_URL` can come out of `.env` and `.env.example`. This builds on the test-port hashing from `/docs/diary/2026-11-25-remove-docker-deps.md`.

## Step 1: Agree on scope

**Author:** main

### Prompt Context

**Verbatim prompt:** I really like this pattern of picking a consistent port based on a hash of something. Let's do it for the dev versitygw as well? And the HTTP server port? We can make it depend on the path it's started in, so it works if e.g. you start the server in a worktree
Follow-up answers: "Keep it simple and copy a bit. hash(pwd+name) -> port should be very simple and can live both in Go code and bash or whatever. Don't hardcode. The goal is to be able to remove it from .env/.env.example and have agents tell where their started app is running with no collisions", and for discovery: "Read the logs".
**Interpretation:** A tiny `port(pwd, name)` function, duplicated in Go (the app) and shell (`/versitygw.sh`, `/Makefile`), gives each service its port. No hardcoded port numbers, and no port settings in `.env`. Agents read the app's URL from `app.log`.
**Inferred intent:** Agents in parallel worktrees can each run `make watch` and `make test` without port collisions or per-worktree config.

### What I did

Agreed the requirements and wrote the brief for a builder.

### Why

Worktrees copy `.env` through `/.worktreeinclude`, so any port in `.env` is shared by every worktree. Deriving the port from the path is the only thing that differs per checkout without manual setup.

### What worked

Markus gave a clear direction: simple, duplicated, no hardcoding.

### What didn't work

Nothing failed in this step.

### What I learned

`maragu.dev/env`'s `Load` overwrites existing environment variables, so a port left in `.env` would beat the derived one. Removing them from `.env` is required, not just tidy.

### What was tricky

Go and shell have to compute exactly the same number, so the hash has to be something both can do identically without extra tools.

### What warrants review

Production must not change: if anything in production relies on the app's default port, it has to be set explicitly there.

### Future work

None yet.

## Step 2: Derive every dev and test port from the checkout's path

**Author:** hashed-ports

### Prompt Context

**Verbatim prompt:** Goal (Markus's words): "hash(pwd+name) -> port should be very simple and can live both in Go code and bash or whatever. Don't hardcode. The goal is to be able to remove it from .env/.env.example and have agents tell where their started app is running with no collisions." Agents find the app's URL by reading the logs (app.log). (The brief went on to list seven requirements: one tiny port function copied into Go and shell, identical in both, below 32768, with a Go test pinning values from the shell version; the names `app`, `versitygw`, `versitygw-webui` and `versitygw-test`, with no hardcoded ports left in `/versitygw.sh`, `/Makefile` or the dev config; app defaults for `SERVER_ADDRESS`, `BASE_URL` and `AWS_ENDPOINT_URL`, with the URL logged at startup; no change in production, stopping to report if there's no clean dev signal; the three variables out of `/.env.example` and port settings out of `/.env.test.example`, with the "Access" section in `/CLAUDE.md` pointing at `app.log`; `make test` on `port(pwd, "versitygw-test")` without the `VERSITYGW_TEST_PORT` override; and no collision detection or other safeguards.)
**Interpretation:** Replace 8080, 7070, 7071 and the Makefile's test-port formula with one hash function, copied into `/versitygw.sh` and `/cmd/app/port.go`, and drop the port settings from the example env files.
**Inferred intent:** Several agents in several worktrees can each run `make watch` and `make test` at once, with nothing to configure, and find their app's URL in `app.log`.

### What I did

The hash is POSIX `cksum`, which stock macOS and Linux both ship. In `/versitygw.sh`:

```sh
port() {
  printf '%s' "$PWD/$1" | cksum | awk '{ print 20000 + $1 % 10000 }'
}
```

`dev` now runs versitygw on `port versitygw` with the web UI on `port versitygw-webui`. `up` takes no argument and uses `port versitygw-test`. A new `port <name>` subcommand prints a port. The busy-port message no longer mentions `VERSITYGW_TEST_PORT`.

`/cmd/app/port.go` has `port(dir, name string) int`, a hand-written copy of the cksum CRC: polynomial `0x04c11db7`, MSB first, the length appended least significant byte first, then the complement, mapped into 20000 + n % 10000. Go's `hash/crc32` only does the reflected variant, so it can't be used. `/cmd/app/port_internal_test.go` pins eight values I took from the shell pipeline, for `/home/me/app` and `/home/me/app/.claude/worktrees/feature` with all four names. I wrote the test first and saw it fail against a stub that returned 0.

In `/cmd/app/main.go`, `SERVER_ADDRESS` defaults to `:` plus `port(os.Getwd(), "app")`, and `BASE_URL` defaults to `http://localhost:` plus the port part of whatever address is in use, from `net.SplitHostPort`. glue's server already logs `msg="Starting server" ... address=<base URL>`, so the URL shows up in `app.log` without new logging.

`/Makefile` lost `VERSITYGW_TEST_PORT`. `test-up` runs `./versitygw.sh up`, and `test` sets `AWS_ENDPOINT_URL=http://localhost:$$(./versitygw.sh port versitygw-test)`. `/redo.yaml` starts the app with `AWS_ENDPOINT_URL=http://localhost:$(./versitygw.sh port versitygw)` in front of `go run`. redo runs each command through `sh -c` with `eval`, so the command substitution works there. `/Dockerfile` gets `ENV SERVER_ADDRESS=:8080` next to `EXPOSE 8080`. `/.env.example` lost `AWS_ENDPOINT_URL`, `BASE_URL` and `SERVER_ADDRESS`, and `/.env.test.example` lost `BASE_URL`. The Access section in `/AGENTS.md` (which `/CLAUDE.md` links to) now says to find the URL in the "Starting server" line in `app.log`.

### Why

The brief's requirement 3 put the `AWS_ENDPOINT_URL` default in the app, but only if dev and production could be told apart. They can't. Nothing in the repo marks production: glue's `app.Start` loads `.env` and `/run/secrets/env`, and the only hints are defaults like `LOG_JSON=true` and `SECURE_COOKIE=true`, which aren't a dev/prod switch. The AWS SDK reads `AWS_ENDPOINT_URL` from the environment on its own, so if the app filled it in whenever it was unset, a production deploy on real AWS would quietly start calling localhost. The dev runner sets it instead. That keeps production exactly as it was without guessing anything about it, and it's still "Go and bash" as Markus put it.

`SERVER_ADDRESS` does get a Go default, and the Dockerfile pins `:8080` so the container keeps listening where `EXPOSE` says. Deriving `BASE_URL` from the address in use, not from the hash, means a production image with no `BASE_URL` still gets `http://localhost:8080`, the same as before.

### What worked

- `make test` passed. It started the test versitygw on 24929, the value `./versitygw.sh port versitygw-test` gives for this worktree.
- `make lint` reported `0 issues.`
- A throwaway internal test (deleted afterwards) ran the shell pipeline through `sh -c` and compared it with `port` for the main checkout, this worktree, a scratch path, a path with a space and `æøå`, a 300-byte path and a 70000-byte path (two and three length bytes), and the empty string, with all four names. Every value matched.
- I copied `/.env.example` to this worktree's `.env` and ran `go tool redo`. The app listened on 29548, and versitygw on 23846 (S3) and 21710 (web UI), all derived. `app.log` showed `msg="Starting server" component=http.Server address=http://localhost:29548`. curl got 200 from the app and 200 from the web UI. `ps eww` showed `AWS_ENDPOINT_URL=http://localhost:23846` in the app's environment. With the aws CLI I made `s3://bucket`, uploaded a file and listed it. A throwaway `cmd/s3check` (also deleted) used glue's `s3.NewBucket` with the `.env` settings to `Put` and `List` against the dev versitygw, and saw both objects.
- A scratch `git worktree add --detach` in the session scratchpad, with the changed files copied in, derived different ports for all four names: app 20807, versitygw 22738, web UI 21680, test 23475, against 29548, 23846, 21710 and 24929 here. Its app logged `address=http://localhost:20807`. I stopped that app and removed the worktree.

### What didn't work

The sandbox refused two compound commands: "this command is too complex to verify that it stays inside the worktree". One was a Python heredoc editing `/cmd/app/main.go` followed by `git diff`. The other wrote `cmd/s3check/main.go` with a heredoc and ran it in the same command. I split both into separate Edit/Write and Bash calls. A third, appending this diary step through a shell heredoc, was refused by the removal safety check, so I appended it with Edit.

When I stopped the scratch checkout's app, `pkill -f "scratchpad/second"` missed it, because `go run` executes a binary from the Go build cache. `lsof -d cwd` on the leftover pid showed it belonged to the scratch checkout, and I stopped it with `kill`.

Self-review (`/code-review medium`) raised seven findings. I changed nothing because of them. Two are the stale-`.env` problem under What warrants review: an old `.env` overrides the derived endpoint, port and base URL, so the logged URL can point at 8080 while the server listens elsewhere. That needs a one-time edit to each `.env`, not a warning in code. Three are safeguards the brief ruled out: an override for each port, detecting two names in one checkout that hash to the same port, and a usage message for `./versitygw.sh port` with no name. One says a `SERVER_ADDRESS` without a colon, such as `8080`, gives the base URL `http://localhost:`. `net/http` rejects that address anyway ("missing port in address"), so the server fails to start before the URL matters. The last says production outside Docker would get a hashed port. The only deploy artifact in the repo is the Docker image, and that pins `:8080`.

### What I learned

- `cksum` is the one hash both systems ship under the same name and with the same output. `md5`/`md5sum` and `shasum`/`sha256sum` differ. Go's `crc32.MakeTable` can't produce cksum's CRC, because it only takes reflected polynomials.
- Go's `os.Getwd` returns `$PWD` when that names the current directory, so it gives the same string the shell's `$PWD` does, symlinks and all. That's why the app's derived port matched `./versitygw.sh port app`.
- glue's `http.Server` logs the base URL, not the listen address, as `address=` in its "Starting server" line, so the `BASE_URL` default is what agents see in `app.log`.

### What was tricky

The `AWS_ENDPOINT_URL` default is the one place I didn't follow the brief to the letter. See Why.

The old Makefile hashed `pwd | cksum`, which includes the trailing newline. The new function uses `printf '%s'`, so every port has moved, including the test port. A test server already running under the old port keeps running until `make test-down`.

### What warrants review

- The `AWS_ENDPOINT_URL` decision in `/redo.yaml`. A plain `go run ./cmd/app` outside redo won't find the dev S3 unless the variable is set.
- `/Dockerfile`: whether the deploy setup (maragudk/workflows' `cd.yml`, which isn't in this repo) ever relied on `SERVER_ADDRESS` being unset. With the `ENV` line it gets `:8080` either way, and `/run/secrets/env` still wins.
- Existing `.env` files, including Markus's in the main checkout, still have `AWS_ENDPOINT_URL=http://localhost:7070`, `BASE_URL` and `SERVER_ADDRESS`. `env.Load` overwrites the environment, so those values beat the derived ones. `AWS_ENDPOINT_URL=...:7070` would point `make watch` at a port where nothing runs any more. Delete those three lines from each `.env`.
- To validate: `go test ./cmd/app -run TestPort`, then `./versitygw.sh port app` in two checkouts, then `make watch` and `grep "Starting server" app.log`.

### Future work

None planned beyond removing the stale lines from existing `.env` files.

## Step 3: Opt into local endpoint and server port with two flags

**Author:** endpoint-default

### Prompt Context

**Verbatim prompt:** Change decided by Markus: the app itself defaults `AWS_ENDPOINT_URL` too, instead of redo setting it. (Two follow-ups from Markus, relayed by the coordinator mid-task: instead of defaulting `AWS_ENDPOINT_URL` when unset, add a bool `AWS_ENDPOINT_URL_LOCAL` that sets it to the dev versitygw; and a second bool `SERVER_ADDRESS_LOCAL`, so `SERVER_ADDRESS` defaults to `:8080` as before unless the flag is true, with `ENV SERVER_ADDRESS=:8080` removed from `/Dockerfile` again.)
**Interpretation:** Keep production behaviour identical by default. Dev opts in with two separate flags in `.env`, and `/redo.yaml` goes back to a plain `go run`.
**Inferred intent:** Dev and production can be told apart by an explicit setting, not by guessing, so the app can own the derived defaults.

### What I did

In `/cmd/app/main.go`, `os.Getwd` moved up, ahead of the AWS config. When `AWS_ENDPOINT_URL_LOCAL` is true, the app sets `AWS_ENDPOINT_URL` to `http://localhost:<port(cwd, "versitygw")>` with `os.Setenv`, before `aws.LoadDefaultConfig`. `SERVER_ADDRESS` defaults to `:8080`, or to `:<port(cwd, "app")>` when `SERVER_ADDRESS_LOCAL` is true. An explicit `SERVER_ADDRESS` still wins, and `BASE_URL` still derives from the address in use. `/redo.yaml` lost the `AWS_ENDPOINT_URL=...` prefix, and `/Dockerfile` lost `ENV SERVER_ADDRESS=:8080`. `/.env.example` got `AWS_ENDPOINT_URL_LOCAL=true` and `SERVER_ADDRESS_LOCAL=true`. I added both to this worktree's `.env` as well.

### Why

Step 2 left `AWS_ENDPOINT_URL` in `/redo.yaml` because nothing marks production. An explicit flag does that job, and it defaults to false, so a production deploy that sets neither flag behaves as before: no endpoint override and `:8080`, which is why the Dockerfile no longer needs to pin it.

### What worked

- `go build ./...`, `make lint` (`0 issues.`) and `make test` passed.
- With both flags true and no `AWS_ENDPOINT_URL`, `SERVER_ADDRESS` or `BASE_URL` in `.env`, `go tool redo` started the app on 29548 (`app.log`: `msg="Starting server" ... address=http://localhost:29548`), versitygw on 23846 and the web UI on 21710, matching `./versitygw.sh port`. A throwaway log line after `aws.LoadDefaultConfig` (reverted) showed `awsConfig.BaseEndpoint` as `http://localhost:23846`, the dev versitygw. curl gave 200 from the app and the web UI.
- With both flags removed from `.env`, `go run ./cmd/app` logged `address=http://localhost:8080`, and the throwaway line showed an empty `AWS_ENDPOINT_URL` and a nil `BaseEndpoint`.

### What didn't work

The auto-mode classifier denied two commands: `pkill -f "redo"` and a `pkill -f "go run -tags ..."`, both too broad because other agents share the machine. I stopped my processes by pid instead, after checking each one's cwd with `lsof -a -p <pid> -d cwd`. I did not run an S3 put and list this time. The check of the SDK's resolved endpoint stands in for it, and step 2 already did the put and list against the same versitygw.

### What I learned

`env.GetBoolOrDefault` reads the flags after `.env` has loaded, so they work from `.env` like any other setting.

### What was tricky

The flag sets a process environment variable, because that is the one way the AWS SDK's default config loading picks the endpoint up without changing how the bucket is built.

### What warrants review

- Every `.env` needs `AWS_ENDPOINT_URL_LOCAL=true` and `SERVER_ADDRESS_LOCAL=true`, and the old `AWS_ENDPOINT_URL`, `BASE_URL` and `SERVER_ADDRESS` lines removed. Without the flags the app serves on `:8080` with no endpoint override.
- Step 2's descriptions of the `AWS_ENDPOINT_URL` default in `/redo.yaml` and the `/Dockerfile` `ENV` line no longer hold.

### Future work

None planned.

## Step 4: Move `os.Getwd` into `port`

**Author:** inline-getwd

### Prompt Context

**Verbatim prompt:** Review feedback (agreed with Markus): `dir` in main.go is only used with `port`, so move `os.Getwd` into it. Rename the current pure `port(dir, name string) int` to `hashPort(dir, name string) int`, unchanged. The test keeps testing `hashPort` with its fixed paths. New `port(name string) (int, error)` calls `os.Getwd` and `hashPort`, returning the Getwd error. main.go: remove the `dir, err := os.Getwd()` block and its comment. The two call sites call `port("versitygw")` / `port("app")` and wrap the error like the surrounding code does. Getwd now only runs when a flag is on.
**Interpretation:** Make `port` own the working directory lookup, keep the hash pure and tested.
**Inferred intent:** Keep `start` free of a variable that only two flag branches need.

### What I did

In `/cmd/app/port.go` I renamed the pure function to `hashPort` and added `port(name)`, which calls `os.Getwd`, wraps its error with `errors.Wrap`, and returns `hashPort(dir, name)`. In `/cmd/app/main.go` I removed the up-front `os.Getwd` block and call `port("versitygw")` and `port("app")` inside their flag branches, returning the error as is, since `port` already wraps it. In `/cmd/app/port_internal_test.go` I renamed `TestPort` to `TestHashPort`; the fixed paths and expected ports are unchanged.

### Why

`dir` was only used by `port`, and `os.Getwd` ran on every start even with both flags off.

### What worked

`go build ./...`, `make lint` (`0 issues.`) and `make test` passed first time.

### What didn't work

Nothing failed. The sandbox refused my first chained shell command as too complex, so I split it into separate commands.

### What I learned

`err` in `start` is now first declared by `aws.LoadDefaultConfig`, so the existing `:=` there still compiles.

### What was tricky

Nothing; the new call sites declare `err` inside their own branches.

### What warrants review

`port` itself is untested because it only wraps `os.Getwd`. The flag-on paths were not run again; behavior is the same as step 3 apart from where `Getwd` is called.

### Future work

None.
