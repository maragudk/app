#!/bin/sh
# Builds and runs versitygw, the local S3 server.
# Usage: ./versitygw.sh dev         runs it for development in the foreground, started by make watch
#        ./versitygw.sh up <port>   starts this checkout's server for the tests in the background
#        ./versitygw.sh down        stops it
set -eu

# Built into bin/ with go install, so it stays out of go.mod. The version is part of the path, so changing it rebuilds.
# After changing it, run make test-down, so the next make test starts the new version.
version=v1.11.0
binary=bin/versitygw-$version/versitygw

dir=.versitygw-test
log=versitygw-test.log

build() {
  if [ ! -x "$binary" ]; then
    GOBIN="$PWD/bin/versitygw-$version" go install github.com/versity/versitygw/cmd/versitygw@$version
  fi
}

case "${1:-}" in
dev)
  build
  mkdir -p .versitygw
  exec "$binary" --access access --secret secretsecret --port :7070 --webui :7071 --webui-no-tls posix .versitygw
  ;;
up)
  port=$2
  url=http://localhost:$port

  # Already running for this checkout.
  if kill -0 "$(cat $dir/pid 2>/dev/null)" 2>/dev/null && curl -fs --max-time 1 -o /dev/null "$url/health"; then
    exit 0
  fi

  # Something else answers on the port.
  if curl -s --max-time 1 -o /dev/null "$url"; then
    echo "Port $port is in use, set VERSITYGW_TEST_PORT in .env to a free port" >&2
    exit 1
  fi

  build
  echo "Starting versitygw on port $port for the tests"
  mkdir -p $dir/data
  "$binary" --access access --secret secretsecret --port ":$port" --health /health posix $dir/data >$log 2>&1 &
  echo $! >$dir/pid

  # Wait up to about 10 seconds for it to answer.
  for _ in $(seq 20); do
    if curl -fs --max-time 1 -o /dev/null "$url/health"; then
      exit 0
    fi
    sleep 0.5
  done
  echo "The test S3 server didn't start, see $log:" >&2
  cat $log >&2
  exit 1
  ;;
down)
  kill "$(cat $dir/pid 2>/dev/null)" 2>/dev/null || true
  rm -rf $dir
  ;;
*)
  echo "Usage: $0 dev | up <port> | down" >&2
  exit 2
  ;;
esac
