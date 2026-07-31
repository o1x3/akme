"use strict";

/**
 * Map Node's process.platform / process.arch to akme release + npm package ids.
 * GitHub archives: akme_<os>_<goarch>.tar.gz (darwin|linux × amd64|arm64).
 * npm platform packages: @o1x3/akme-<nodeOs>-<nodeArch> (x64|arm64).
 */
const PLATFORMS = [
  {
    nodeOs: "darwin",
    nodeArch: "arm64",
    goos: "darwin",
    goarch: "arm64",
    npmPackage: "@o1x3/akme-darwin-arm64",
  },
  {
    nodeOs: "darwin",
    nodeArch: "x64",
    goos: "darwin",
    goarch: "amd64",
    npmPackage: "@o1x3/akme-darwin-x64",
  },
  {
    nodeOs: "linux",
    nodeArch: "arm64",
    goos: "linux",
    goarch: "arm64",
    npmPackage: "@o1x3/akme-linux-arm64",
  },
  {
    nodeOs: "linux",
    nodeArch: "x64",
    goos: "linux",
    goarch: "amd64",
    npmPackage: "@o1x3/akme-linux-x64",
  },
];

function normalizeArch(arch) {
  if (arch === "amd64") return "x64";
  return arch;
}

function resolvePlatform(platform = process.platform, arch = process.arch) {
  const nodeArch = normalizeArch(arch);
  const match = PLATFORMS.find(
    (p) => p.nodeOs === platform && p.nodeArch === nodeArch,
  );
  if (!match) {
    if (platform !== "darwin" && platform !== "linux") {
      throw new Error(
        `akme: unsupported platform ${platform}/${arch} (supported: darwin, linux)`,
      );
    }
    throw new Error(
      `akme: unsupported architecture ${platform}/${arch} (supported: amd64, arm64)`,
    );
  }

  return {
    os: match.goos,
    arch: match.goarch,
    nodeOs: match.nodeOs,
    nodeArch: match.nodeArch,
    npmPackage: match.npmPackage,
    archiveName: `akme_${match.goos}_${match.goarch}.tar.gz`,
    binarySubpath: "bin/akme",
  };
}

function allPlatforms() {
  return PLATFORMS.map((p) => ({
    ...p,
    archiveName: `akme_${p.goos}_${p.goarch}.tar.gz`,
    binarySubpath: "bin/akme",
  }));
}

module.exports = { PLATFORMS, allPlatforms, resolvePlatform };
