#!/bin/bash

set -euo pipefail

kill $(lsof -ti:8080) 2>/dev/null || true
kill $(lsof -ti:8081) 2>/dev/null || true

echo "" >app.log

go tool air 2>&1 | \
  awk '
    /^#/ || /level=INFO msg="Starting app"/ {
      system("> app.log")
      system("printf \"\\033[2J\\033[H\" >&2")
    }
    { print | "tee -a app.log" }
  '
