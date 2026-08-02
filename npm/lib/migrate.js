"use strict";

const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

const PATH_MARKER = "# managed by akme installer";
const LEGACY_PATH_MARKER = "# managed by nx installer";

/** Distinctive substrings present in akme / legacy-nx release binaries. */
const BINARY_MARKERS = [
  Buffer.from("github.com/o1x3/akme"),
  Buffer.from("github.com/o1x3/nx"),
  Buffer.from("akme help"),
  Buffer.from("nx help"),
  Buffer.from("akme git stat"),
  Buffer.from("nx git stat"),
];

function homeDir(env = process.env) {
  return env.HOME || env.USERPROFILE || os.homedir();
}

function defaultInstallDir(env = process.env) {
  if (env.AKME_INSTALL_DIR) return path.resolve(env.AKME_INSTALL_DIR);
  if (env.NX_INSTALL_DIR) return path.resolve(env.NX_INSTALL_DIR);
  return path.join(homeDir(env), ".local", "bin");
}

function resolvePath(filePath) {
  try {
    return fs.realpathSync(filePath);
  } catch {
    return path.resolve(filePath);
  }
}

function which(command, env = process.env) {
  try {
    const out = execFileSync("sh", ["-c", `command -v -- ${JSON.stringify(command)}`], {
      encoding: "utf8",
      env,
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    return out || "";
  } catch {
    return "";
  }
}

/**
 * True when path looks like our CLI (legacy `nx` or current `akme`), not
 * Nrwl's nx or unrelated tools on PATH.
 */
function looksLikeOurCLI(filePath) {
  if (!filePath) return false;
  let st;
  try {
    st = fs.statSync(filePath);
  } catch {
    return false;
  }
  if (!st.isFile()) return false;

  try {
    const out = execFileSync(filePath, ["version"], {
      encoding: "utf8",
      timeout: 3000,
      env: {
        ...process.env,
        AKME_NO_UPDATE: "1",
        NX_NO_UPDATE: "1",
      },
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    // Historical: "nx 0.4.x (…)" · current: "akme 0.6.x (…)"
    if (/^(nx|akme)\s+v?\d+/i.test(out)) return true;
  } catch {
    // fall through to binary scan
  }

  // Skip tiny scripts / npm shims for the byte scan; our release binary is
  // a multi-MB Go build. Version probe above already caught real CLIs.
  if (st.size < 100_000) return false;

  try {
    // Scan a prefix + suffix; markers live in the rodata of the Go binary.
    const fd = fs.openSync(filePath, "r");
    try {
      const size = st.size;
      const chunk = Math.min(size, 4 * 1024 * 1024);
      const buf = Buffer.alloc(chunk);
      fs.readSync(fd, buf, 0, chunk, 0);
      if (BINARY_MARKERS.some((m) => buf.includes(m))) return true;
      if (size > chunk) {
        fs.readSync(fd, buf, 0, chunk, size - chunk);
        if (BINARY_MARKERS.some((m) => buf.includes(m))) return true;
      }
    } finally {
      fs.closeSync(fd);
    }
  } catch {
    return false;
  }
  return false;
}

function candidateNxPaths(env = process.env) {
  const home = homeDir(env);
  const installDir = defaultInstallDir(env);
  const seen = new Set();
  const out = [];

  const add = (p) => {
    if (!p || !path.isAbsolute(p)) return;
    const key = resolvePath(p);
    if (seen.has(key)) return;
    seen.add(key);
    out.push(p);
  };

  add(path.join(installDir, "nx"));
  add(path.join(home, ".local", "bin", "nx"));
  add("/usr/local/bin/nx");

  const onPath = which("nx", env);
  if (onPath) add(onPath);

  return out;
}

function canWriteDir(dir) {
  try {
    fs.mkdirSync(dir, { recursive: true });
    const probe = path.join(dir, `.akme-write-${process.pid}`);
    fs.writeFileSync(probe, "");
    fs.unlinkSync(probe);
    return true;
  } catch {
    return false;
  }
}

function removeFile(candidate) {
  try {
    fs.unlinkSync(candidate);
    return true;
  } catch {
    return false;
  }
}

function installBinary(source, dest) {
  fs.mkdirSync(path.dirname(dest), { recursive: true });
  // Copy then rename so a partial write cannot leave a broken `akme`.
  const tmp = `${dest}.tmp-${process.pid}`;
  fs.copyFileSync(source, tmp);
  fs.chmodSync(tmp, 0o755);
  fs.renameSync(tmp, dest);
}

function refreshPathMarkers(env = process.env) {
  const home = homeDir(env);
  const profiles = [
    path.join(home, ".profile"),
    path.join(home, ".zprofile"),
    path.join(home, ".zshrc"),
    path.join(home, ".bashrc"),
    path.join(home, ".bash_profile"),
    path.join(home, ".config", "fish", "config.fish"),
  ];
  const renamed = [];
  for (const profile of profiles) {
    if (!fs.existsSync(profile)) continue;
    let text;
    try {
      text = fs.readFileSync(profile, "utf8");
    } catch {
      continue;
    }
    if (!text.includes(LEGACY_PATH_MARKER)) continue;
    const next = text.split(LEGACY_PATH_MARKER).join(PATH_MARKER);
    if (next === text) continue;
    try {
      fs.writeFileSync(profile, next);
      renamed.push(profile);
    } catch {
      // best-effort; profile may be root-owned
    }
  }
  return renamed;
}

/**
 * If a legacy `nx` install of this CLI is present, install `akme` beside it
 * (or into AKME_INSTALL_DIR / ~/.local/bin) and remove the old `nx` binary.
 *
 * Skips unrelated `nx` tools (e.g. Nrwl) via version/binary fingerprinting.
 *
 * @param {{ binary: string, env?: NodeJS.ProcessEnv }} options
 * @returns {{
 *   migrated: boolean,
 *   installPath: string,
 *   removed: string[],
 *   markers: string[],
 *   skipped: string[],
 *   messages: string[],
 * }}
 */
function migrateLegacyNx(options = {}) {
  const env = options.env || process.env;
  const sourceBinary = options.binary;
  const messages = [];
  const removed = [];
  const skipped = [];

  if (!sourceBinary || !fs.existsSync(sourceBinary)) {
    return {
      migrated: false,
      installPath: "",
      removed,
      markers: [],
      skipped,
      messages,
    };
  }

  const ours = [];
  for (const candidate of candidateNxPaths(env)) {
    let exists = false;
    try {
      // lstat so broken symlinks still count as candidates to clean up.
      fs.lstatSync(candidate);
      exists = true;
    } catch {
      exists = false;
    }
    if (!exists) continue;

    // Broken symlink to a removed nx: remove without fingerprint.
    let isBrokenLink = false;
    try {
      const st = fs.lstatSync(candidate);
      if (st.isSymbolicLink() && !fs.existsSync(candidate)) {
        isBrokenLink = true;
      }
    } catch {
      // ignore
    }

    if (!isBrokenLink && !looksLikeOurCLI(candidate)) {
      skipped.push(candidate);
      continue;
    }
    ours.push(candidate);
  }

  const markers = refreshPathMarkers(env);
  for (const profile of markers) {
    messages.push(`renamed nx installer PATH marker to akme in ${profile}`);
  }

  if (ours.length === 0) {
    return {
      migrated: false,
      installPath: "",
      removed,
      markers,
      skipped,
      messages,
    };
  }

  // Prefer the bindir of the first legacy binary we found; else default.
  let installDir = path.dirname(ours[0]);
  if (!canWriteDir(installDir)) {
    installDir = defaultInstallDir(env);
  }
  if (!canWriteDir(installDir)) {
    messages.push(
      `warning: could not migrate legacy nx — install dir not writable: ${installDir}`,
    );
    return {
      migrated: false,
      installPath: "",
      removed,
      markers,
      skipped,
      messages,
    };
  }

  const installPath = path.join(installDir, "akme");
  const sourceResolved = resolvePath(sourceBinary);
  const destResolved = resolvePath(installPath);

  // Don't clobber if source somehow is the destination (npx vendor path
  // should never equal ~/.local/bin/akme, but be safe).
  if (sourceResolved !== destResolved) {
    try {
      installBinary(sourceBinary, installPath);
      messages.push(`migrated legacy nx → ${installPath}`);
    } catch (err) {
      messages.push(
        `warning: could not install akme to ${installPath}: ${err.message}`,
      );
      return {
        migrated: false,
        installPath: "",
        removed,
        markers,
        skipped,
        messages,
      };
    }
  }

  for (const candidate of ours) {
    const resolved = resolvePath(candidate);
    if (resolved === destResolved || candidate === installPath) {
      continue;
    }
    if (removeFile(candidate)) {
      removed.push(candidate);
      messages.push(`removed previous nx install at ${candidate}`);
    } else {
      messages.push(
        `warning: could not remove previous nx install at ${candidate}; it may shadow ${installPath} on PATH`,
      );
    }
  }

  // Also drop /usr/local/bin/nx when present and ours (install.sh parity).
  const usrLocal = "/usr/local/bin/nx";
  if (
    !ours.includes(usrLocal) &&
    fs.existsSync(usrLocal) &&
    looksLikeOurCLI(usrLocal)
  ) {
    if (removeFile(usrLocal)) {
      removed.push(usrLocal);
      messages.push(`removed previous nx install at ${usrLocal}`);
    }
  }

  return {
    migrated: true,
    installPath,
    removed,
    markers,
    skipped,
    messages,
  };
}

module.exports = {
  BINARY_MARKERS,
  LEGACY_PATH_MARKER,
  PATH_MARKER,
  candidateNxPaths,
  defaultInstallDir,
  looksLikeOurCLI,
  migrateLegacyNx,
  refreshPathMarkers,
};
