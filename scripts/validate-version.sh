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
  if [ "$npm_name" != "@o1x3/akme" ]; then
    echo "npm/package.json name must be @o1x3/akme (got ${npm_name})" >&2
    exit 1
  fi
  if [ "$npm_version" != "$version" ]; then
    echo "npm/package.json version (${npm_version}) must match VERSION (${version})" >&2
    exit 1
  fi

  node -e '
    const fs = require("fs");
    const path = require("path");
    const { allPlatforms } = require("./npm/lib/platform");
    const version = fs.readFileSync("VERSION", "utf8").trim();
    const main = require("./npm/package.json");
    const optional = main.optionalDependencies || {};
    const platforms = allPlatforms();
    for (const p of platforms) {
      if (optional[p.npmPackage] !== version) {
        console.error(
          `npm/package.json optionalDependencies[${p.npmPackage}] must be ${version} (got ${optional[p.npmPackage]})`,
        );
        process.exit(1);
      }
      const pkgPath = path.join("npm", "platforms", `${p.nodeOs}-${p.nodeArch}`, "package.json");
      const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf8"));
      if (pkg.name !== p.npmPackage) {
        console.error(`${pkgPath} name must be ${p.npmPackage} (got ${pkg.name})`);
        process.exit(1);
      }
      // Templates may stay at 0.0.0; publish-release.js stamps VERSION at release.
      if (pkg.version !== version && pkg.version !== "0.0.0") {
        console.error(`${pkgPath} version must be ${version} or 0.0.0 (got ${pkg.version})`);
        process.exit(1);
      }
      if (!pkg.os || !pkg.cpu) {
        console.error(`${pkgPath} must set os and cpu for optionalDependency gating`);
        process.exit(1);
      }
    }
    for (const name of Object.keys(optional)) {
      if (!platforms.some((p) => p.npmPackage === name)) {
        console.error(`unexpected optionalDependency ${name}`);
        process.exit(1);
      }
    }
  '
fi
