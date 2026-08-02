"use strict";

const { execFileSync } = require("node:child_process");
const crypto = require("node:crypto");
const fs = require("node:fs");
const https = require("node:https");
const http = require("node:http");
const os = require("node:os");
const path = require("node:path");
const { pipeline } = require("node:stream/promises");

const { resolvePlatform } = require("./platform");

const DEFAULT_REPO = "o1x3/akme";

function packageRoot() {
  return path.join(__dirname, "..");
}

function packageVersion() {
  const pkg = JSON.parse(
    fs.readFileSync(path.join(packageRoot(), "package.json"), "utf8"),
  );
  return String(pkg.version);
}

function repo() {
  return process.env.AKME_REPO || process.env.NX_REPO || DEFAULT_REPO;
}

function vendorRoot() {
  return path.join(packageRoot(), "vendor");
}

function versionFilePath() {
  return path.join(vendorRoot(), "VERSION");
}

function binaryPath(version) {
  const override =
    process.env.AKME_BINARY ||
    process.env.AKME_NX_BINARY ||
    process.env.NX_BINARY;
  if (override) {
    return path.resolve(override);
  }
  return path.join(vendorRoot(), version, "akme");
}

function releaseBase(version) {
  const tag = version.startsWith("v") ? version : `v${version}`;
  return `https://github.com/${repo()}/releases/download/${tag}`;
}

function stripV(tag) {
  return String(tag || "").replace(/^v/i, "");
}

function parseSemver(version) {
  const parts = stripV(version).split(".").map((p) => Number.parseInt(p, 10));
  if (parts.length !== 3 || parts.some((n) => Number.isNaN(n))) {
    return null;
  }
  return parts;
}

/** True when latest is strictly newer than current (semver). */
function newer(latest, current) {
  const a = parseSemver(latest);
  const b = parseSemver(current);
  if (!a || !b) return stripV(latest) !== stripV(current);
  for (let i = 0; i < 3; i++) {
    if (a[i] > b[i]) return true;
    if (a[i] < b[i]) return false;
  }
  return false;
}

function installedVersion() {
  try {
    const v = fs.readFileSync(versionFilePath(), "utf8").trim();
    if (v) return stripV(v);
  } catch {
    // fall through
  }
  return "";
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const client = url.startsWith("https:") ? https : http;
    const req = client.get(
      url,
      {
        headers: { "User-Agent": "akme-npm" },
      },
      (res) => {
        if (
          res.statusCode >= 300 &&
          res.statusCode < 400 &&
          res.headers.location
        ) {
          res.resume();
          download(res.headers.location, dest).then(resolve, reject);
          return;
        }
        if (res.statusCode !== 200) {
          res.resume();
          reject(
            new Error(`akme: download failed (${res.statusCode}): ${url}`),
          );
          return;
        }
        const out = fs.createWriteStream(dest);
        pipeline(res, out).then(resolve, reject);
      },
    );
    req.on("error", reject);
  });
}

function requestHead(url) {
  return new Promise((resolve, reject) => {
    const client = url.startsWith("https:") ? https : http;
    const req = client.request(
      url,
      {
        method: "HEAD",
        headers: { "User-Agent": "akme-npm" },
      },
      (res) => {
        // Follow redirects manually so we can read the final URL.
        if (
          res.statusCode >= 300 &&
          res.statusCode < 400 &&
          res.headers.location
        ) {
          res.resume();
          const next = new URL(res.headers.location, url).toString();
          requestHead(next).then(resolve, reject);
          return;
        }
        resolve({ statusCode: res.statusCode, url: url, headers: res.headers });
        res.resume();
      },
    );
    req.on("error", reject);
    req.end();
  });
}

function tagFromReleaseURL(raw) {
  const marker = "/releases/tag/";
  const idx = raw.indexOf(marker);
  if (idx < 0) {
    throw new Error(`akme: could not determine latest release from ${raw}`);
  }
  let tag = raw.slice(idx + marker.length);
  const cut = tag.search(/[/?#]/);
  if (cut >= 0) tag = tag.slice(0, cut);
  tag = tag.trim();
  if (!tag) {
    throw new Error(`akme: could not determine latest release from ${raw}`);
  }
  return tag;
}

/**
 * Resolve newest GitHub release tag via releases/latest redirect (not the API).
 */
async function latestReleaseTag() {
  const latestURL = `https://github.com/${repo()}/releases/latest`;
  const res = await requestHead(latestURL);
  if (res.statusCode < 200 || res.statusCode >= 400) {
    throw new Error(`akme: GitHub returned ${res.statusCode} for ${latestURL}`);
  }
  // requestHead follows redirects; final url is in res.url
  return tagFromReleaseURL(res.url);
}

function sha256File(filePath) {
  const hash = crypto.createHash("sha256");
  hash.update(fs.readFileSync(filePath));
  return hash.digest("hex");
}

function expectedChecksum(checksumsText, archiveName) {
  for (const line of checksumsText.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    const parts = trimmed.split(/\s+/);
    if (parts.length < 2) continue;
    const digest = parts[0];
    const name = parts[1].replace(/^\*/, "");
    if (name === archiveName) {
      return digest;
    }
  }
  return "";
}

async function downloadFromGitHub(version, dest) {
  const { archiveName } = resolvePlatform();
  const base = releaseBase(version);
  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), "akme-"));
  const archivePath = path.join(tmp, archiveName);
  const checksumPath = path.join(tmp, "checksums.txt");

  try {
    await download(`${base}/${archiveName}`, archivePath);
    await download(`${base}/checksums.txt`, checksumPath);

    const expected = expectedChecksum(
      fs.readFileSync(checksumPath, "utf8"),
      archiveName,
    );
    if (!expected) {
      throw new Error(`akme: checksums.txt has no entry for ${archiveName}`);
    }
    const actual = sha256File(archivePath);
    if (actual !== expected) {
      throw new Error(
        `akme: checksum mismatch for ${archiveName} (got ${actual}, want ${expected})`,
      );
    }

    execFileSync("tar", ["-xzf", archivePath, "-C", tmp, "akme"], {
      stdio: ["ignore", "ignore", "pipe"],
    });

    const extracted = path.join(tmp, "akme");
    if (!fs.existsSync(extracted)) {
      throw new Error("akme: archive did not contain akme binary");
    }

    fs.mkdirSync(path.dirname(dest), { recursive: true });
    fs.copyFileSync(extracted, dest);
    fs.chmodSync(dest, 0o755);
    return dest;
  } finally {
    fs.rmSync(tmp, { recursive: true, force: true });
  }
}

function writeInstalledVersion(version) {
  fs.mkdirSync(vendorRoot(), { recursive: true });
  fs.writeFileSync(versionFilePath(), `${stripV(version)}\n`);
}

/**
 * Ensure a runnable binary exists (for `npx akme-cli <cmd>`).
 * Uses AKME_BINARY, else vendor copy, else downloads the npm package version.
 */
async function ensureBinary() {
  const override =
    process.env.AKME_BINARY ||
    process.env.AKME_NX_BINARY ||
    process.env.NX_BINARY;
  if (override) {
    const resolved = path.resolve(override);
    if (!fs.existsSync(resolved)) {
      throw new Error(`akme: AKME_BINARY not found: ${resolved}`);
    }
    return resolved;
  }

  const current = installedVersion();
  if (current) {
    const dest = binaryPath(current);
    if (fs.existsSync(dest) && fs.statSync(dest).size > 0) {
      return dest;
    }
  }

  const version = packageVersion();
  const dest = binaryPath(version);
  if (fs.existsSync(dest) && fs.statSync(dest).size > 0) {
    writeInstalledVersion(version);
    return dest;
  }

  await downloadFromGitHub(version, dest);
  writeInstalledVersion(version);
  return dest;
}

/**
 * Install or update to the latest GitHub release (bare `npx akme-cli` /
 * postinstall). Returns { binary, version, previous, action }.
 * action: "install" | "update" | "unchanged"
 *
 * Optional onStatus({ action, version, previous }) fires before download
 * (or immediately for unchanged).
 */
async function installOrUpdate(options = {}) {
  const onStatus = options.onStatus;

  const override =
    process.env.AKME_BINARY ||
    process.env.AKME_NX_BINARY ||
    process.env.NX_BINARY;
  if (override) {
    const resolved = path.resolve(override);
    if (!fs.existsSync(resolved)) {
      throw new Error(`akme: AKME_BINARY not found: ${resolved}`);
    }
    const version = installedVersion() || packageVersion();
    const result = {
      binary: resolved,
      version,
      previous: "",
      action: "unchanged",
    };
    onStatus?.(result);
    return result;
  }

  const latestTag = await latestReleaseTag();
  const latest = stripV(latestTag);
  const current = installedVersion();
  const dest = binaryPath(latest);

  if (
    current &&
    !newer(latest, current) &&
    fs.existsSync(dest) &&
    fs.statSync(dest).size > 0
  ) {
    const result = {
      binary: dest,
      version: current,
      previous: current,
      action: "unchanged",
    };
    onStatus?.(result);
    return result;
  }

  // Same version already on disk (e.g. reinstall).
  if (fs.existsSync(dest) && fs.statSync(dest).size > 0 && current === latest) {
    const result = {
      binary: dest,
      version: latest,
      previous: current,
      action: "unchanged",
    };
    onStatus?.(result);
    return result;
  }

  const action = current && newer(latest, current) ? "update" : "install";
  const pending = {
    binary: dest,
    version: latest,
    previous: current || "",
    action,
  };
  onStatus?.(pending);

  await downloadFromGitHub(latest, dest);
  writeInstalledVersion(latest);
  return {
    ...pending,
    updated: action === "update",
  };
}

module.exports = {
  binaryPath,
  ensureBinary,
  expectedChecksum,
  installOrUpdate,
  installedVersion,
  latestReleaseTag,
  newer,
  packageVersion,
  releaseBase,
  repo,
  resolvePlatform,
  stripV,
};
