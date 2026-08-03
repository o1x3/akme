"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { describe, it, beforeEach, afterEach } = require("node:test");

const {
  LEGACY_PATH_MARKER,
  PATH_MARKER,
  looksLikeOurCLI,
  migrateLegacyNx,
  refreshPathMarkers,
} = require("./migrate");

function makeTmp() {
  return fs.mkdtempSync(path.join(os.tmpdir(), "akme-migrate-"));
}

function writeExec(filePath, body) {
  fs.mkdirSync(path.dirname(filePath), { recursive: true });
  fs.writeFileSync(filePath, body, { mode: 0o755 });
}

function fakeOurNx(filePath, name = "nx") {
  writeExec(
    filePath,
    `#!/bin/sh\nif [ "$1" = "version" ]; then echo "${name} 0.4.5 (abc, 2026-01-01)"; exit 0; fi\necho "usage"\n`,
  );
}

function fakeAkme(filePath) {
  writeExec(
    filePath,
    `#!/bin/sh\nif [ "$1" = "version" ]; then echo "akme 0.6.5 (abc, 2026-01-01)"; exit 0; fi\necho "usage"\n`,
  );
}

function fakeNrwlNx(filePath) {
  writeExec(
    filePath,
    `#!/bin/sh\nif [ "$1" = "--version" ] || [ "$1" = "version" ]; then echo "Nx Version:\\n- Local: 19.0.0"; exit 0; fi\necho "nrwl"\n`,
  );
}

describe("looksLikeOurCLI", () => {
  let dir;
  beforeEach(() => {
    dir = makeTmp();
  });
  afterEach(() => {
    fs.rmSync(dir, { recursive: true, force: true });
  });

  it("accepts legacy nx version output", () => {
    const p = path.join(dir, "nx");
    fakeOurNx(p);
    assert.equal(looksLikeOurCLI(p), true);
  });

  it("accepts current akme version output", () => {
    const p = path.join(dir, "nx");
    fakeAkme(p);
    assert.equal(looksLikeOurCLI(p), true);
  });

  it("rejects Nrwl nx", () => {
    const p = path.join(dir, "nx");
    fakeNrwlNx(p);
    assert.equal(looksLikeOurCLI(p), false);
  });

  it("accepts binary with embedded marker when version fails", () => {
    const p = path.join(dir, "nx");
    // Non-executable-looking payload that still contains a marker and is large.
    const buf = Buffer.alloc(120_000, 0);
    Buffer.from("github.com/o1x3/akme").copy(buf, 1000);
    fs.writeFileSync(p, buf, { mode: 0o644 });
    assert.equal(looksLikeOurCLI(p), true);
  });
});

describe("migrateLegacyNx", () => {
  let home;
  let vendor;
  beforeEach(() => {
    home = makeTmp();
    vendor = makeTmp();
  });
  afterEach(() => {
    fs.rmSync(home, { recursive: true, force: true });
    fs.rmSync(vendor, { recursive: true, force: true });
  });

  it("no-ops when no nx is present", () => {
    const source = path.join(vendor, "akme");
    fakeAkme(source);
    const env = { HOME: home, PATH: "/usr/bin:/bin" };
    const result = migrateLegacyNx({ binary: source, env });
    assert.equal(result.migrated, false);
    assert.equal(result.removed.length, 0);
    assert.equal(fs.existsSync(path.join(home, ".local", "bin", "akme")), false);
  });

  it("installs akme and removes legacy nx from ~/.local/bin", () => {
    const bindir = path.join(home, ".local", "bin");
    const nxPath = path.join(bindir, "nx");
    fakeOurNx(nxPath);

    const source = path.join(vendor, "akme");
    fakeAkme(source);

    const env = {
      HOME: home,
      PATH: `${bindir}:/usr/bin:/bin`,
    };
    const result = migrateLegacyNx({ binary: source, env });

    assert.equal(result.migrated, true);
    assert.equal(result.installPath, path.join(bindir, "akme"));
    assert.ok(result.removed.includes(nxPath));
    assert.equal(fs.existsSync(nxPath), false);
    assert.equal(fs.existsSync(result.installPath), true);
    assert.match(fs.readFileSync(result.installPath, "utf8"), /akme 0\.6\.5/);
    assert.ok(
      result.messages.some((m) => m.includes("migrated legacy nx")),
    );
  });

  it("does not remove Nrwl nx", () => {
    const bindir = path.join(home, ".local", "bin");
    const nxPath = path.join(bindir, "nx");
    fakeNrwlNx(nxPath);

    const source = path.join(vendor, "akme");
    fakeAkme(source);

    const env = {
      HOME: home,
      PATH: `${bindir}:/usr/bin:/bin`,
    };
    const result = migrateLegacyNx({ binary: source, env });

    assert.equal(result.migrated, false);
    assert.equal(fs.existsSync(nxPath), true);
    assert.ok(result.skipped.includes(nxPath));
  });

  it("refreshes legacy PATH markers even when nx binary is already gone", () => {
    const profile = path.join(home, ".zshrc");
    fs.writeFileSync(
      profile,
      `\n${LEGACY_PATH_MARKER}\nexport PATH="$HOME/.local/bin:$PATH"\n`,
    );

    const source = path.join(vendor, "akme");
    fakeAkme(source);

    const env = { HOME: home, PATH: "/usr/bin:/bin", SHELL: "/bin/zsh" };
    const result = migrateLegacyNx({ binary: source, env });

    assert.equal(result.migrated, false);
    assert.ok(result.markers.includes(profile));
    const text = fs.readFileSync(profile, "utf8");
    assert.ok(text.includes(PATH_MARKER));
    assert.equal(text.includes(LEGACY_PATH_MARKER), false);
  });

  it("honors AKME_INSTALL_DIR for the akme destination", () => {
    const custom = path.join(home, "bin");
    const nxPath = path.join(custom, "nx");
    fakeOurNx(nxPath);

    const source = path.join(vendor, "akme");
    fakeAkme(source);

    const env = {
      HOME: home,
      AKME_INSTALL_DIR: custom,
      PATH: `${custom}:/usr/bin:/bin`,
    };
    const result = migrateLegacyNx({ binary: source, env });

    assert.equal(result.migrated, true);
    assert.equal(result.installPath, path.join(custom, "akme"));
    assert.equal(fs.existsSync(nxPath), false);
    assert.equal(fs.existsSync(result.installPath), true);
  });
});

describe("refreshPathMarkers", () => {
  let home;
  beforeEach(() => {
    home = makeTmp();
  });
  afterEach(() => {
    fs.rmSync(home, { recursive: true, force: true });
  });

  it("rewrites the legacy installer comment", () => {
    const profile = path.join(home, ".profile");
    fs.writeFileSync(profile, `${LEGACY_PATH_MARKER}\n`);
    const renamed = refreshPathMarkers({ HOME: home });
    assert.deepEqual(renamed, [profile]);
    assert.equal(fs.readFileSync(profile, "utf8"), `${PATH_MARKER}\n`);
  });
});
