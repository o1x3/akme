"use strict";

const { installOrUpdate } = require("./binary");
const { printGreeting } = require("./greeting");
const { migrateLegacyNx } = require("./migrate");

/**
 * Shared path for `npm install -g akme-cli` (postinstall) and bare
 * `npx akme-cli` (no args): install or update the native CLI from GitHub,
 * greet, migrate a legacy `nx` install when present. Does not spawn `akme`.
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
      stream.write(`  installed CLI ${result.version}\n`);
    } else if (result.action === "update") {
      stream.write(
        `  updated CLI ${result.previous} → ${result.version}\n`,
      );
    } else {
      stream.write(`  CLI ${result.version} already up to date\n`);
    }
  }

  // People who still have the pre-rename `nx` binary: install `akme` onto
  // their PATH bindir and remove the old `nx` (same idea as install.sh).
  const migration = migrateLegacyNx({ binary: result.binary, env });
  if (!quiet) {
    for (const line of migration.messages) {
      stream.write(`  ${line}\n`);
    }
    stream.write("\n");
  }

  return { ...result, migration };
}

module.exports = { installCLI };
