# akme

npm distribution of [akme](https://github.com/o1x3/akme) — a personal developer CLI.

```sh
npx akme token -i
npx akme git stat .
npm install -g akme
bunx akme help
```

Downloads the matching GitHub release binary (`akme_<os>_<arch>.tar.gz`), verifies `checksums.txt`, and runs it. Self-update inside the binary is disabled for npm installs — update via npm/`npx` instead.

```sh
AKME_BINARY=/path/to/akme npx akme version
```
