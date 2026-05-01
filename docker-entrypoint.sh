#!/bin/sh
set -eu

if [ "${1:-}" = "web" ]; then
    shift
    exec /app/bin/issue2mdweb "$@"
fi

exec /app/bin/issue2md "$@"
