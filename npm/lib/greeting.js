"use strict";

// Clean foreground-only wordmark (no bg fills — those dump poorly in most terms).
const ESC = "\u001b";
const PINK = `${ESC}[91m`;
const DIM = `${ESC}[2m`;
const RESET = `${ESC}[0m`;

const LINES = [
  "  ████   ██  ██  ███▄   █████",
  " ██  ██  ██ ██   ██ ██  ██   ",
  " ██████  ████    ██  ██ ████ ",
  " ██  ██  ██ ██   ██ ██  ██   ",
  " ██  ██  ██  ██  ███▀   █████",
];

function shouldPrintGreeting(env = process.env) {
  if (env.AKME_NPM_QUIET === "1") return false;
  if (env.AKME_NPM_GREETING === "0") return false;
  if (env.CI === "true" || env.CI === "1") {
    return env.AKME_NPM_GREETING === "1";
  }
  return true;
}

function greetingText() {
  return `${LINES.map((line) => `${PINK}${line}${RESET}`).join("\n")}\n`;
}

function printGreeting(stream = process.stderr, env = process.env) {
  if (!shouldPrintGreeting(env)) return false;
  stream.write("\n");
  stream.write(greetingText());
  stream.write(`\n  ${DIM}akme ready — try: akme help${RESET}\n\n`);
  return true;
}

module.exports = {
  greetingText,
  printGreeting,
  shouldPrintGreeting,
};
