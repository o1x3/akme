"use strict";

const { installOrUpdate } = require("./binary");
const { printGreeting } = require("./greeting");

/**
 * Shared path for `npm install -g akme-cli` (postinstall) and bare
 * `npx akme-cli` (no args): install or update the native CLI from GitHub,
 * greet, done. Does not spawn `akme`.
 */
async function installCLI(options = {}) {
  const stream = options.stream || process.stderr;
  const env = options.env || process.env;
  // Art first, then install/update (never spawn the Go CLI here).
  printGreeting(stream, env);
  if (env.AKME_NPM_QUIET !== "1") {
    stream.write("  installing / updating CLI…\n");
  }
  const result = await installOrUpdate();
  if (env.AKME_NPM_QUIET !== "1") {
    if (result.updated) {
      stream.write(`  updated CLI → ${result.version}\n\n`);
    } else {
      stream.write(`  CLI ${result.version} ready\n\n`);
    }
  }
  return result;
}

module.exports = { installCLI };
