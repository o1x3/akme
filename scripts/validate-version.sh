#!/usr/bin/env sh
set -eu

version="$(cat VERSION)"

case "$version" in
  *[!0-9.]* | "" | .* | *. | *..*)
    echo "VERSION must be plain semver like 0.0.1" >&2
    exit 1
    ;;
esac

old_ifs="$IFS"
IFS=.
set -- $version
IFS="$old_ifs"

if [ "$#" -ne 3 ]; then
  echo "VERSION must have major.minor.patch" >&2
  exit 1
fi

for part in "$@"; do
  case "$part" in
    "" | *[!0-9]*)
      echo "VERSION must have numeric major.minor.patch parts" >&2
      exit 1
      ;;
  esac
done

if ! grep -Eq "^##[[:space:]]+v?${version}([[:space:]]|$)" CHANGELOG.md; then
  echo "CHANGELOG.md must include a section for ${version}" >&2
  exit 1
fi

if [ -f npm/package.json ]; then
  npm_name="$(
    node -e 'const p=require("./npm/package.json"); process.stdout.write(String(p.name||""))'
  )"
  npm_version="$(
    node -e 'const p=require("./npm/package.json"); process.stdout.write(String(p.version||""))'
  )"
  if [ "$npm_name" != "akme" ]; then
    echo "npm/package.json name must be akme (got ${npm_name})" >&2
    exit 1
  fi
  if [ "$npm_version" != "$version" ]; then
    echo "npm/package.json version (${npm_version}) must match VERSION (${version})" >&2
    exit 1
  fi
fi
