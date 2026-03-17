## Specific instructions for this project

This is a web app for

### Running

Start the app with `make watch`. This watches for file changes and automatically rebuilds/restarts, so it doesn't need to be restarted manually. Logs are written to `app.log`.

### Access

You can access the app in a browser using your "rodney" skill. The URL and port are configured in the `.env` file - check `BASE_URL` for the full URL or `SERVER_ADDRESS` for just the port.
You can access the SQLite database directly using `sqlite3 app.db`. ALWAYS run `pragma foreign_keys = 1;` as the first statement after connecting.
