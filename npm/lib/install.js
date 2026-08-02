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
  const result = await installOrUpdate();
  printGreeting(stream, env);
  if (result.updated && env.AKME_NPM_QUIET !== "1") {
    stream.write(`  updated CLI → ${result.version}\n\n`);
  }
  return result;
}

module.exports = { installCLI };
