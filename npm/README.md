# @o1x3/akme

<p align="center">
  <a href="https://www.npmjs.com/package/@o1x3/akme"><img src="https://img.shields.io/npm/v/@o1x3/akme?style=for-the-badge&logo=npm&logoColor=white&color=CB3837" alt="npm version"></a>
  <a href="https://www.npmjs.com/package/@o1x3/akme"><img src="https://img.shields.io/npm/dm/@o1x3/akme?style=for-the-badge&logo=npm&logoColor=white&color=CB3837" alt="npm downloads"></a>
  <a href="https://github.com/o1x3/akme/releases/latest"><img src="https://img.shields.io/github/v/release/o1x3/akme?style=for-the-badge&logo=github&logoColor=white&color=181717" alt="GitHub release"></a>
</p>

<p align="center">
  <a href="https://github.com/o1x3/akme/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/o1x3/akme/ci.yml?branch=main&style=for-the-badge&logo=githubactions&logoColor=white&label=CI" alt="CI"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge" alt="MIT License"></a>
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey?style=for-the-badge" alt="macOS and Linux">
</p>

<p align="center"><b>akme</b></p>

<p align="center">personal developer CLI<br>git branch stats · coding-agent token usage · self-update</p>

npm wrapper for the [akme](https://github.com/o1x3/akme) binary. On install (or first run) it downloads the matching GitHub release (`akme_<os>_<arch>.tar.gz`), checks `checksums.txt`, and runs it.

## Install

```sh
npx @o1x3/akme token -i
npm install -g @o1x3/akme
bunx @o1x3/akme help
```

Global install puts `akme` on your PATH.

## Quick start

```sh
akme help
akme git stat .
akme token
akme token -i
```

```sh
akme help git
akme help token
akme help token harness
```

## Notes

Binary self-update is off for npm installs. Update with the package manager instead:

```sh
npm update -g @o1x3/akme
```

Point at a local binary (skips the download):

```sh
AKME_BINARY=/path/to/akme npx @o1x3/akme version
```

macOS and Linux only (`amd64` / `arm64`). Full docs: [github.com/o1x3/akme](https://github.com/o1x3/akme#readme).
