"use strict";

const { describe, it } = require("node:test");
const assert = require("node:assert/strict");
const { allPlatforms, resolvePlatform } = require("./platform");
const { expectedChecksum } = require("./binary");

describe("resolvePlatform", () => {
  it("maps darwin/arm64", () => {
    assert.deepEqual(resolvePlatform("darwin", "arm64"), {
      os: "darwin",
      arch: "arm64",
      nodeOs: "darwin",
      nodeArch: "arm64",
      npmPackage: "@o1x3/akme-darwin-arm64",
      archiveName: "akme_darwin_arm64.tar.gz",
      binarySubpath: "bin/akme",
    });
  });

  it("maps linux/x64 to amd64 + npm x64 package", () => {
    assert.deepEqual(resolvePlatform("linux", "x64"), {
      os: "linux",
      arch: "amd64",
      nodeOs: "linux",
      nodeArch: "x64",
      npmPackage: "@o1x3/akme-linux-x64",
      archiveName: "akme_linux_amd64.tar.gz",
      binarySubpath: "bin/akme",
    });
  });

  it("maps darwin/amd64 alias to x64 package", () => {
    assert.equal(
      resolvePlatform("darwin", "amd64").npmPackage,
      "@o1x3/akme-darwin-x64",
    );
  });

  it("rejects windows", () => {
    assert.throws(() => resolvePlatform("win32", "x64"), /unsupported platform/);
  });
});

describe("allPlatforms", () => {
  it("covers the four release targets", () => {
    const names = allPlatforms().map((p) => p.npmPackage).sort();
    assert.deepEqual(names, [
      "@o1x3/akme-darwin-arm64",
      "@o1x3/akme-darwin-x64",
      "@o1x3/akme-linux-arm64",
      "@o1x3/akme-linux-x64",
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
