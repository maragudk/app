#!/bin/sh
# Starts and stops this checkout's versitygw, the local S3 server for the tests.
# Usage: ./versitygw.sh up <binary> <port>
#        ./versitygw.sh down
set -eu

dir=.versitygw-test
log=versitygw-test.log

case "${1:-}" in
up)
  binary=$2
  port=$3
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
  echo "Usage: $0 up <binary> <port> | down" >&2
  exit 2
  ;;
esac
