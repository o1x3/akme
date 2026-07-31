"use strict";

/**
 * Stage + publish platform packages (with embedded binaries) then the meta
 * package. Used by .github/workflows/release.yml after GoReleaser.
 *
 * Usage:
 *   node npm/scripts/publish-release.js --version 0.6.0 --dist dist
 *
 * Expects GoReleaser archives in --dist:
 *   akme_darwin_arm64.tar.gz, akme_darwin_amd64.tar.gz,
 *   akme_linux_arm64.tar.gz,  akme_linux_amd64.tar.gz
 */

const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

const { allPlatforms } = require("../lib/platform");

function parseArgs(argv) {
  const out = { version: "", dist: "dist", dryRun: false };
  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === "--version") {
      out.version = argv[++i] || "";
    } else if (arg === "--dist") {
      out.dist = argv[++i] || "dist";
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

function repoRoot() {
  return path.join(npmRoot(), "..");
}

function platformDir(nodeOs, nodeArch) {
  return path.join(npmRoot(), "platforms", `${nodeOs}-${nodeArch}`);
}

function stripAuthToken() {
  // setup-node writes //registry.npmjs.org/:_authToken=${NODE_AUTH_TOKEN}.
  // Unset/empty auth blocks OIDC Trusted Publishing; strip it.
  const npmrc = process.env.NPM_CONFIG_USERCONFIG || path.join(os.homedir(), ".npmrc");
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

function runNpm(args, cwd) {
  execFileSync("npm", args, {
    cwd,
    stdio: "inherit",
    env: process.env,
  });
}

function stagePlatform(platform, version, distDir) {
  const archive = path.join(distDir, platform.archiveName);
  if (!fs.existsSync(archive)) {
    throw new Error(`missing release archive: ${archive}`);
  }

  const dir = platformDir(platform.nodeOs, platform.nodeArch);
  const pkgPath = path.join(dir, "package.json");
  const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf8"));
  if (pkg.name !== platform.npmPackage) {
    throw new Error(
      `platform package name mismatch in ${pkgPath}: got ${pkg.name}, want ${platform.npmPackage}`,
    );
  }
  pkg.version = version;
  fs.writeFileSync(pkgPath, `${JSON.stringify(pkg, null, 2)}\n`);

  const binDir = path.join(dir, "bin");
  fs.rmSync(binDir, { recursive: true, force: true });
  fs.mkdirSync(binDir, { recursive: true });

  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), "akme-stage-"));
  try {
    execFileSync("tar", ["-xzf", archive, "-C", tmp, "akme"], {
      stdio: ["ignore", "ignore", "pipe"],
    });
    const extracted = path.join(tmp, "akme");
    if (!fs.existsSync(extracted)) {
      throw new Error(`${platform.archiveName} did not contain akme binary`);
    }
    const dest = path.join(binDir, "akme");
    fs.copyFileSync(extracted, dest);
    fs.chmodSync(dest, 0o755);
  } finally {
    fs.rmSync(tmp, { recursive: true, force: true });
  }

  return dir;
}

function syncMainPackage(version) {
  const pkgPath = path.join(npmRoot(), "package.json");
  const pkg = JSON.parse(fs.readFileSync(pkgPath, "utf8"));
  pkg.version = version;

  const platforms = allPlatforms();
  const optional = {};
  for (const p of platforms) {
    optional[p.npmPackage] = version;
  }
  pkg.optionalDependencies = optional;
  fs.writeFileSync(pkgPath, `${JSON.stringify(pkg, null, 2)}\n`);
}

function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help) {
    console.log(
      "Usage: node npm/scripts/publish-release.js --version <semver> [--dist dist] [--dry-run]",
    );
    process.exit(0);
  }
  if (!/^\d+\.\d+\.\d+$/.test(args.version)) {
    throw new Error(`--version must be major.minor.patch (got ${args.version || "(empty)"})`);
  }

  const distDir = path.isAbsolute(args.dist)
    ? args.dist
    : path.join(repoRoot(), args.dist);
  if (!fs.existsSync(distDir)) {
    throw new Error(`dist directory not found: ${distDir}`);
  }

  stripAuthToken();
  syncMainPackage(args.version);

  const platforms = allPlatforms();
  const staged = [];
  for (const platform of platforms) {
    const dir = stagePlatform(platform, args.version, distDir);
    staged.push({ platform, dir });
    console.log(`staged ${platform.npmPackage}@${args.version}`);
  }

  if (args.dryRun) {
    console.log("dry-run: skipping npm publish");
    return;
  }

  // Platform packages first so the meta package's optionalDependencies resolve.
  for (const { platform, dir } of staged) {
    console.log(`publishing ${platform.npmPackage}@${args.version}`);
    runNpm(["publish", "--access", "public"], dir);
  }

  console.log(`publishing @o1x3/akme@${args.version}`);
  runNpm(["publish", "--access", "public"], npmRoot());
}

main();
