#!/usr/bin/env sh
set -eu

fail=0

say() {
  printf '%s\n' "$*"
}

run() {
  say "==> $*"
  if "$@"; then
    return 0
  fi
  fail=1
  return 0
}

check_gofmt() {
  files="$(gofmt -l .)"
  if [ -n "$files" ]; then
    say "gofmt needed:"
    say "$files"
    fail=1
  fi
}

check_gofmt
run go test ./...
run sh -n scripts/format.sh
run sh -n scripts/install.sh
run sh -n scripts/check.sh
run sh -n scripts/validate-version.sh
run scripts/validate-version.sh
run sh -n .githooks/pre-commit

if command -v node >/dev/null 2>&1; then
  run node --check npm/bin/akme.js
  run node --check npm/lib/binary.js
  run node --check npm/lib/platform.js
  run node --check npm/scripts/postinstall.js
  run node --test npm/lib/*.test.js
else
  say "==> skip npm package checks (node not installed)"
fi

exit "$fail"
