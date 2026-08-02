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
  const quiet = env.AKME_NPM_QUIET === "1";

  // Art first, then install/update (never spawn the Go CLI here).
  printGreeting(stream, env);

  const result = await installOrUpdate({
    onStatus(status) {
      if (quiet) return;
      if (status.action === "install") {
        stream.write(`  installing CLI ${status.version}…\n`);
      } else if (status.action === "update") {
        stream.write(
          `  updating CLI ${status.previous} → ${status.version}…\n`,
        );
      }
    },
  });

  if (!quiet) {
    if (result.action === "install") {
      stream.write(`  installed CLI ${result.version}\n\n`);
    } else if (result.action === "update") {
      stream.write(
        `  updated CLI ${result.previous} → ${result.version}\n\n`,
      );
    } else {
      stream.write(`  CLI ${result.version} already up to date\n\n`);
    }
  }

  return result;
}

module.exports = { installCLI };
