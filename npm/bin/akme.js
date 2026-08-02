#!/usr/bin/env node
"use strict";

const { spawnSync } = require("node:child_process");
const { ensureBinary } = require("../lib/binary");
const { installCLI } = require("../lib/install");

async function main() {
  const args = process.argv.slice(2);

  // Bare `npx akme-cli` / `akme` with no args → greeting + install/update only.
  // Never forward empty argv to the Go binary (that prints usage + may self-update).
  if (args.length === 0) {
    await installCLI();
    return;
  }

  const binary = await ensureBinary();
  // npm/bun is the update channel for this install path; skip binary self-update.
  const env = { ...process.env };
  if (env.AKME_NO_UPDATE === undefined && env.NX_NO_UPDATE === undefined) {
    env.AKME_NO_UPDATE = "1";
  }

  const result = spawnSync(binary, args, {
    stdio: "inherit",
    env,
  });

  if (result.error) {
    console.error(result.error.message);
    process.exit(1);
  }
  process.exit(result.status === null ? 1 : result.status);
}

main().catch((err) => {
  console.error(err && err.message ? err.message : err);
  process.exit(1);
});
