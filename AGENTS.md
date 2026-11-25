## Specific instructions for this project

This is a web app for

### Running

Start the app with `make watch`. This watches for file changes and automatically rebuilds/restarts, so it doesn't need to be restarted manually. Logs are written to `app.log`.

`make watch` also runs versitygw, a local S3 server, natively on port 7070 (web UI on 7071), with data in `.versitygw/` and logs in `versitygw.log`. It's built with `go install` into `bin/` on first use, at the version pinned in the `Makefile`. `make test` runs `make test-up` first, which starts a separate versitygw for the tests in the background, unless it's already running. The start and stop logic lives in `versitygw.sh` (`up` and `down`), which `make test-up` and `make test-down` call. Each checkout gets its own test server, so tests in parallel worktrees don't share one. Its port is fixed per checkout, derived from the checkout's path, and `make test-up` prints it when it starts the server; if that port is taken, `make test-up` says so, and you set `VERSITYGW_TEST_PORT` to a free one in that checkout's `.env` (not the main checkout's, since new worktrees copy `.env`). Its state lives in `.versitygw-test/` (data and pid) and its logs in `versitygw-test.log`, and `make test` points the tests at it through `AWS_ENDPOINT_URL`, so don't set that in `.env.test`. To run `go test` directly, run `make test-up` first and set `AWS_ENDPOINT_URL` the way `make -n test | tail -1` shows. The server keeps running after the tests; `make test-down` stops it and deletes `.versitygw-test/`. After changing the versitygw version in the `Makefile`, run `make test-down`, so the next `make test` starts the new version. No Docker needed. To run several dev servers side by side, set `VERSITYGW_PORT` and `VERSITYGW_WEBUI_PORT` in `.env`, and point `AWS_ENDPOINT_URL` in `.env` at the new `VERSITYGW_PORT`.

### Access

You can access the app in a browser using your "playwright-cli" skill. The URL and port are configured in the `.env` file - check `BASE_URL` for the full URL or `SERVER_ADDRESS` for just the port.
You can access the SQLite database directly using `sqlite3 app.db`. ALWAYS run `pragma foreign_keys = 1;` as the first statement after connecting.
