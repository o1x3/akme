"use strict";

/**
 * Publish akme-cli to npm after GoReleaser (GitHub release owns the binaries).
 *
 * Usage:
 *   node npm/scripts/publish-release.js --version 0.6.1 [--dry-run]
 */

const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

function parseArgs(argv) {
  const out = { version: "", dryRun: false };
  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === "--version") {
      out.version = argv[++i] || "";
    } else if (arg === "--dist") {
      // Accepted for backwards compatibility; binaries live on GitHub releases.
      i += 1;
    } else if (arg === "--dry-run") {
      out.dryRun = true;
    } else if (arg === "--help" || arg === "-h") {
      out.help = true;
    } else {
      throw new Error(`unknown argument: ${arg}`);
    }
  }
  return out;
}

function npmRoot() {
  return path.join(__dirname, "..");
}

function stripAuthToken() {
  // setup-node writes //registry.npmjs.org/:_authToken=${NODE_AUTH_TOKEN}.
  // Unset/empty auth blocks OIDC Trusted Publishing; strip it.
  const npmrc =
    process.env.NPM_CONFIG_USERCONFIG || path.join(os.homedir(), ".npmrc");
  if (fs.existsSync(npmrc)) {
    const filtered = fs
      .readFileSync(npmrc, "utf8")
      .split(/\r?\n/)
      .filter((line) => !line.includes("_authToken"))
      .join("\n");
    fs.writeFileSync(npmrc, filtered);
  }
  delete process.env.NODE_AUTH_TOKEN;
}

function syncMainPackage(version) {
  const pkgPath = path.join(npmRoot(), "package.json");
  const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf8"));
  pkg.version = version;
  delete pkg.optionalDependencies;
  fs.writeFileSync(pkgPath, `${JSON.stringify(pkg, null, 2)}\n`);
}

function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help) {
    console.log(
      "Usage: node npm/scripts/publish-release.js --version <semver> [--dry-run]",
    );
    process.exit(0);
  }
  if (!/^\d+\.\d+\.\d+$/.test(args.version)) {
    throw new Error(
      `--version must be major.minor.patch (got ${args.version || "(empty)"})`,
    );
  }

  stripAuthToken();
  syncMainPackage(args.version);

  if (args.dryRun) {
    console.log(`dry-run: would publish akme-cli@${args.version}`);
    return;
  }

  console.log(`publishing akme-cli@${args.version}`);
  execFileSync("npm", ["publish", "--access", "public"], {
    cwd: npmRoot(),
    stdio: "inherit",
    env: process.env,
  });
}

main();
