#!/usr/bin/env node
"use strict";

/**
 * Smoke-test optionalDependency resolution without publishing.
 * Stages a fake platform package next to @o1x3/akme layout and checks
 * resolveOptionalBinary finds it.
 */

const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const assert = require("node:assert/strict");
const { resolvePlatform } = require("../lib/platform");

function main() {
  const { npmPackage, binarySubpath, nodeOs, nodeArch } = resolvePlatform();
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "akme-resolve-"));
  try {
    // Mimic npm hoisting: node_modules/@o1x3/akme + sibling platform package.
    const meta = path.join(root, "node_modules", "@o1x3", "akme");
    // Scoped package lives at node_modules/@o1x3/akme-<os>-<arch>.
    const plat = path.join(root, "node_modules", npmPackage);
    fs.mkdirSync(path.join(meta, "lib"), { recursive: true });
    fs.mkdirSync(path.dirname(path.join(plat, binarySubpath)), {
      recursive: true,
    });
    fs.writeFileSync(path.join(plat, binarySubpath), "#!/bin/sh\necho ok\n");
    fs.chmodSync(path.join(plat, binarySubpath), 0o755);

    // Copy resolver into the fake package root so paths: [packageRoot()] works.
    fs.copyFileSync(
      path.join(__dirname, "..", "lib", "platform.js"),
      path.join(meta, "lib", "platform.js"),
    );
    fs.copyFileSync(
      path.join(__dirname, "..", "lib", "binary.js"),
      path.join(meta, "lib", "binary.js"),
    );
    fs.writeFileSync(
      path.join(meta, "package.json"),
      JSON.stringify({ name: "@o1x3/akme", version: "0.0.0-test" }),
    );

    const { resolveOptionalBinary } = require(path.join(meta, "lib", "binary.js"));
    const resolved = resolveOptionalBinary();
    assert.ok(resolved, `expected optional binary for ${nodeOs}/${nodeArch}`);
    assert.equal(path.basename(resolved), "akme");
    assert.ok(fs.existsSync(resolved));
    console.log(`ok: resolved ${npmPackage} -> ${resolved}`);
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
}

main();
