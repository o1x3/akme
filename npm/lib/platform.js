"use strict";

/**
 * Map Node's process.platform / process.arch to akme GitHub release asset names.
 * Release archives are: akme_<os>_<arch>.tar.gz (darwin|linux × amd64|arm64).
 */
function resolvePlatform(platform = process.platform, arch = process.arch) {
  let os;
  switch (platform) {
    case "darwin":
      os = "darwin";
      break;
    case "linux":
      os = "linux";
      break;
    default:
      throw new Error(
        `akme: unsupported platform ${platform}/${arch} (supported: darwin, linux)`,
      );
  }

  let goarch;
  switch (arch) {
    case "x64":
    case "amd64":
      goarch = "amd64";
      break;
    case "arm64":
      goarch = "arm64";
      break;
    default:
      throw new Error(
        `akme: unsupported architecture ${platform}/${arch} (supported: amd64, arm64)`,
      );
  }

  const archiveName = `akme_${os}_${goarch}.tar.gz`;
  return { os, arch: goarch, archiveName };
}

module.exports = { resolvePlatform };
