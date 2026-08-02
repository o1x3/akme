"use strict";

const { describe, it } = require("node:test");
const assert = require("node:assert/strict");
const { allPlatforms, resolvePlatform } = require("./platform");
const { expectedChecksum, newer } = require("./binary");

describe("resolvePlatform", () => {
  it("maps darwin/arm64", () => {
    assert.deepEqual(resolvePlatform("darwin", "arm64"), {
      os: "darwin",
      arch: "arm64",
      nodeOs: "darwin",
      nodeArch: "arm64",
      archiveName: "akme_darwin_arm64.tar.gz",
    });
  });

  it("maps linux/x64 to amd64 archive", () => {
    assert.deepEqual(resolvePlatform("linux", "x64"), {
      os: "linux",
      arch: "amd64",
      nodeOs: "linux",
      nodeArch: "x64",
      archiveName: "akme_linux_amd64.tar.gz",
    });
  });

  it("maps darwin/amd64 alias", () => {
    assert.equal(resolvePlatform("darwin", "amd64").arch, "amd64");
  });

  it("rejects windows", () => {
    assert.throws(() => resolvePlatform("win32", "x64"), /unsupported platform/);
  });
});

describe("allPlatforms", () => {
  it("covers the four release targets", () => {
    const names = allPlatforms()
      .map((p) => p.archiveName)
      .sort();
    assert.deepEqual(names, [
      "akme_darwin_amd64.tar.gz",
      "akme_darwin_arm64.tar.gz",
      "akme_linux_amd64.tar.gz",
      "akme_linux_arm64.tar.gz",
    ]);
  });
});

describe("expectedChecksum", () => {
  it("parses checksums.txt lines", () => {
    const text = [
      "aaa  akme_linux_amd64.tar.gz",
      "bbb *akme_darwin_arm64.tar.gz",
      "",
    ].join("\n");
    assert.equal(expectedChecksum(text, "akme_darwin_arm64.tar.gz"), "bbb");
    assert.equal(expectedChecksum(text, "akme_linux_amd64.tar.gz"), "aaa");
    assert.equal(expectedChecksum(text, "missing.tar.gz"), "");
  });
});

describe("newer", () => {
  it("compares semver", () => {
    assert.equal(newer("0.6.1", "0.6.0"), true);
    assert.equal(newer("v0.7.0", "0.6.9"), true);
    assert.equal(newer("0.6.0", "0.6.0"), false);
    assert.equal(newer("0.5.9", "0.6.0"), false);
  });
});
