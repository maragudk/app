## Specific instructions for this project

This is a web app for

### Access

You can access the app in a browser using chrome devtools. The URL and port are configured in the `.env` file - check `BASE_URL` for the full URL or `SERVER_ADDRESS` for just the port.
Assume the server is already running and is automatically reloaded on code changes.

You can access the SQLite database directly using `sqlite3 app.db`. ALWAYS run `pragma foreign_keys = 1;` as the first statement after connecting.
