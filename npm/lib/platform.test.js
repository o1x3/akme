"use strict";

const { describe, it } = require("node:test");
const assert = require("node:assert/strict");
const { resolvePlatform } = require("./platform");
const { expectedChecksum } = require("./binary");

describe("resolvePlatform", () => {
  it("maps darwin/arm64", () => {
    assert.deepEqual(resolvePlatform("darwin", "arm64"), {
      os: "darwin",
      arch: "arm64",
      archiveName: "nx_darwin_arm64.tar.gz",
    });
  });

  it("maps linux/x64 to amd64", () => {
    assert.deepEqual(resolvePlatform("linux", "x64"), {
      os: "linux",
      arch: "amd64",
      archiveName: "nx_linux_amd64.tar.gz",
    });
  });

  it("rejects windows", () => {
    assert.throws(() => resolvePlatform("win32", "x64"), /unsupported platform/);
  });
});

describe("expectedChecksum", () => {
  it("parses checksums.txt lines", () => {
    const text = [
      "aaa  nx_linux_amd64.tar.gz",
      "bbb *nx_darwin_arm64.tar.gz",
      "",
    ].join("\n");
    assert.equal(expectedChecksum(text, "nx_darwin_arm64.tar.gz"), "bbb");
    assert.equal(expectedChecksum(text, "nx_linux_amd64.tar.gz"), "aaa");
    assert.equal(expectedChecksum(text, "missing.tar.gz"), "");
  });
});
