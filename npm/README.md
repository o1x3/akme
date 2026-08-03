# akme-cli

<p align="center">
  <a href="https://www.npmjs.com/package/akme-cli"><img src="https://img.shields.io/npm/v/akme-cli?style=for-the-badge&logo=npm&logoColor=white&color=CB3837" alt="npm version"></a>
  <a href="https://www.npmjs.com/package/akme-cli"><img src="https://img.shields.io/npm/dm/akme-cli?style=for-the-badge&logo=npm&logoColor=white&color=CB3837" alt="npm downloads"></a>
  <a href="https://github.com/o1x3/akme/releases/latest"><img src="https://img.shields.io/github/v/release/o1x3/akme?style=for-the-badge&logo=github&logoColor=white&color=181717" alt="GitHub release"></a>
</p>

<p align="center">
  <a href="https://github.com/o1x3/akme/actions/workflows/release.yml"><img src="https://img.shields.io/github/actions/workflow/status/o1x3/akme/release.yml?branch=main&style=for-the-badge&logo=githubactions&logoColor=white&label=CI" alt="CI"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge" alt="MIT License"></a>
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey?style=for-the-badge" alt="macOS and Linux">
</p>

<p align="center"><b>akme</b></p>

<p align="center">personal developer CLI<br>git contribution graph · branch stats · coding-agent token usage · self-update</p>

npm wrapper for the [akme](https://github.com/o1x3/akme) binary. Install downloads the matching GitHub release archive with `checksums.txt` verification.

## Install

```sh
npm install -g akme-cli
npx akme-cli
```

Both install or **update** the native CLI to the latest GitHub release and print the greeting (they do not run a command). A legacy `nx` install of this CLI is migrated to `akme` on your PATH automatically. Then:

```sh
akme help
akme token -i
```

`AKME_NPM_QUIET=1` skips the greeting. Run a command: `npx akme-cli <cmd>` / `bunx akme-cli <cmd>`.

## Quick start

```sh
akme help
akme git .
akme git . -i
akme git stat .
akme token
akme token -i
```

## Notes

Binary self-update inside the Go binary is off for npm installs (`AKME_NO_UPDATE=1`). Use bare `npx akme-cli` / `npm update -g akme-cli` to refresh.

```sh
AKME_BINARY=/path/to/akme npx akme-cli version
```

macOS and Linux only (`amd64` / `arm64`). Full docs: [github.com/o1x3/akme](https://github.com/o1x3/akme#readme).
