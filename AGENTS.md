# app

This is a web app for

## Access

You can access the app in a browser using playwright at http://localhost:8080. Assume the server is already running and is automatically reloaded on code changes. You can take screenshots (`browser_screen_capture`) and page snapshots (`browser_snapshot`). Prefer snapshots when you don't need an image, since it takes up less tokens in your context. The server log is in `app.log`. Always check it for errors after code changes before accessing the app.

You can access the SQLite database directly using `sqlite3 app.db`. ALWAYS run `pragma foreign_keys = 1;` as the first statement after connecting.
