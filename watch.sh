#!/bin/bash

set -euo pipefail

SERVER_PORT=$(grep "^SERVER_ADDRESS=" .env | cut -d: -f2)

kill $(lsof -ti:${SERVER_PORT}) 2>/dev/null || true

echo "" >app.log

go tool air 2>&1 | \
  awk '
    /^#/ || /level=INFO msg="Starting app"/ {
      system("> app.log")
      system("printf \"\\033[2J\\033[H\" >&2")
    }
    { print | "tee -a app.log" }
  '
