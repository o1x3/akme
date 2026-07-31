# @o1x3/akme

npm distribution of [akme](https://github.com/o1x3/akme), a personal developer CLI.

```sh
npx @o1x3/akme token -i
npx @o1x3/akme git stat .
npm install -g @o1x3/akme
bunx @o1x3/akme help
```

After a global install the `akme` binary is on your PATH. Downloads the matching GitHub release binary (`akme_<os>_<arch>.tar.gz`), verifies `checksums.txt`, and runs it. Self-update inside the binary is disabled for npm installs; update via npm/`npx` instead.

```sh
AKME_BINARY=/path/to/akme npx @o1x3/akme version
```
