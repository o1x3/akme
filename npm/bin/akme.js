#!/usr/bin/env node
"use strict";

const { spawnSync } = require("node:child_process");
const { ensureBinary } = require("../lib/binary");

async function main() {
  const binary = await ensureBinary();
  // npm/bun is the update channel for this install path; skip nx self-update.
  const env = { ...process.env };
  if (env.NX_NO_UPDATE === undefined) {
    env.NX_NO_UPDATE = "1";
  }

  const result = spawnSync(binary, process.argv.slice(2), {
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
