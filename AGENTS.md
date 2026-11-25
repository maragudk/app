## Specific instructions for this project

This is a web app for

### Running

Start the app with `make watch`. This watches for file changes and automatically rebuilds/restarts, so it doesn't need to be restarted manually. Logs are written to `app.log`.

`make watch` also runs versitygw, a local S3 server, natively with `go run` on port 7070 (web UI on 7071), with data in `.versitygw/` and logs in `versitygw.log`. `make test` starts its own throwaway versitygw on port 7072 and stops it afterwards. No Docker needed.

### Access

You can access the app in a browser using your "playwright-cli" skill. The URL and port are configured in the `.env` file - check `BASE_URL` for the full URL or `SERVER_ADDRESS` for just the port.
You can access the SQLite database directly using `sqlite3 app.db`. ALWAYS run `pragma foreign_keys = 1;` as the first statement after connecting.
