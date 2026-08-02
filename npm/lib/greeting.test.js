"use strict";

const assert = require("node:assert/strict");
const { describe, it } = require("node:test");
const {
  greetingText,
  printGreeting,
  shouldPrintGreeting,
} = require("./greeting");

describe("greeting", () => {
  it("embeds ANSI escapes and block characters", () => {
    const text = greetingText();
    assert.match(text, /\u001b\[38;2;255;85;85m/);
    assert.match(text, /\u001b\[48;2;170;0;0m/);
    assert.match(text, /██/);
    assert.match(text, /\u001b\[0m/);
  });

  it("skips when quiet or CI unless forced", () => {
    assert.equal(shouldPrintGreeting({ AKME_NPM_QUIET: "1" }), false);
    assert.equal(shouldPrintGreeting({ CI: "1" }), false);
    assert.equal(shouldPrintGreeting({ CI: "1", AKME_NPM_GREETING: "1" }), true);
    assert.equal(shouldPrintGreeting({}), true);
  });

  it("printGreeting writes art when allowed", () => {
    let out = "";
    const stream = { write(chunk) { out += chunk; } };
    assert.equal(printGreeting(stream, {}), true);
    assert.match(out, /akme ready/);
    assert.match(out, /████/);
  });
});
