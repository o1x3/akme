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

const DEFAULT_REPO = "o1x3/nx";

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

function binaryPath(version = packageVersion()) {
  const override =
    process.env.AKME_BINARY ||
    process.env.AKME_NX_BINARY ||
    process.env.NX_BINARY;
  if (override) {
    return path.resolve(override);
  }
  return path.join(packageRoot(), "vendor", version, "akme");
}

function releaseBase(version) {
  const tag = version.startsWith("v") ? version : `v${version}`;
  return `https://github.com/${repo()}/releases/download/${tag}`;
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

  const version = packageVersion();
  const dest = binaryPath(version);
  if (fs.existsSync(dest) && fs.statSync(dest).size > 0) {
    return dest;
  }

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

module.exports = {
  binaryPath,
  ensureBinary,
  expectedChecksum,
  packageVersion,
  releaseBase,
  repo,
  resolvePlatform,
};
