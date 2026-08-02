"use strict";

// ANSI wordmark (ESC restored from terminal dump). Printed on install only.
const ESC = "\u001b";

const LINES = [
  `${ESC}[0;37;40m             ${ESC}[0;97;40m██ ${ESC}[0;91;40m██ ${ESC}[0;37;40m     ${ESC}[0;91;40m▄▄ ${ESC}[0;37;40m                             ${ESC}[0m`,
  `${ESC}[0;37;40m  ${ESC}[0;91;40m▄▄██▀▀▀███▀████ ${ESC}[0;37;40m   ${ESC}[0;91;40m▄█▀ ${ESC}[0;37;40m    ${ESC}[0;97;40m██ ${ESC}[0;91;40m██▀▀██▀███▄▄ ${ESC}[0;37;40m   ${ESC}[0;91;40m▄▄██▀▀████ ${ESC}[0m`,
  `${ESC}[0;91;40m▐ ${ESC}[0;97;40m██ ${ESC}[0;91;41m▓ ${ESC}[0;37;40m    ${ESC}[0;31;40m▐ ${ESC}[0;91;41m▓▓▓ ${ESC}[0;37;40m  ${ESC}[0;31;40m▐ ${ESC}[0;91;41m▓▓▓ ${ESC}[0;91;40m▀▀███▄▄ ${ESC}[0;37;40m  ${ESC}[0;91;41m▓▓▓▓ ${ESC}[0;37;40m   ${ESC}[0;91;41m▓▓ ${ESC}[0;37;40m   ${ESC}[0;91;41m▓▓▓▓ ${ESC}[0;91;40m▌▐ ${ESC}[0;97;40m██ ${ESC}[0;91;41m▓ ${ESC}[0;91;40m▄▄▄ ${ESC}[0;91;41m▓▓▓▓ ${ESC}[0m`,
  `${ESC}[0;91;41m▒▒▒▒ ${ESC}[0;37;40m    ${ESC}[0;31;40m▐ ${ESC}[0;91;41m▒▒▒ ${ESC}[0;31;40m▌▐ ${ESC}[0;91;41m▒▒▒ ${ESC}[0;37;40m    ${ESC}[0;91;41m▓▓▓▓ ${ESC}[0;91;40m▌ ${ESC}[0;31;40m▐ ${ESC}[0;91;41m▒▒▒ ${ESC}[0;37;40m   ${ESC}[0;91;41m▒▒ ${ESC}[0;37;40m   ${ESC}[0;91;41m▒▒▒▒▒▒▒▒▒ ${ESC}[0;37;40m    ${ESC}[0;31;40m▄▄▄▄ ${ESC}[0m`,
  `${ESC}[0;91;41m░░░░ ${ESC}[0;31;40m▌ ${ESC}[0;37;40m   ${ESC}[0;31;40m▐ ${ESC}[0;91;41m░░░ ${ESC}[0;31;40m▌▐ ${ESC}[0;91;41m░░░ ${ESC}[0;37;40m    ${ESC}[0;91;41m▒▒▒▒▒ ${ESC}[0;31;40m▐ ${ESC}[0;91;41m░░░ ${ESC}[0;37;40m   ${ESC}[0;91;41m░░ ${ESC}[0;37;40m   ${ESC}[0;91;41m░░░░░░░░░ ${ESC}[0;37;40m    ${ESC}[0;91;41m░░░░ ${ESC}[0m`,
  `${ESC}[0;37;40m  ${ESC}[0;31;40m▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀ ${ESC}[0;37;40m    ${ESC}[0;91;41m░░░░░ ${ESC}[0;31;40m▀▀▀▀ ${ESC}[0;37;40m       ${ESC}[0;31;40m████▌ ${ESC}[0;37;40m  ${ESC}[0;31;40m▀▀▀▀▀▀████ ${ESC}[0m`,
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
  return `${LINES.join("\n")}\n`;
}

function printGreeting(stream = process.stderr, env = process.env) {
  if (!shouldPrintGreeting(env)) return false;
  stream.write(greetingText());
  stream.write("\n  akme ready — try: akme help\n\n");
  return true;
}

module.exports = {
  greetingText,
  printGreeting,
  shouldPrintGreeting,
};
