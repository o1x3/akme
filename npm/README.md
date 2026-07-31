# akme

npm distribution of [`nx`](https://github.com/o1x3/nx) — a personal developer CLI.

```sh
npx akme token -i
npx akme git stat .
npm install -g akme
bunx akme token -i
```

The package downloads the matching GitHub release binary for your OS/CPU (`darwin`/`linux` × `amd64`/`arm64`), verifies `checksums.txt`, and runs it. Self-update inside the binary is disabled for npm installs; bump/update via npm instead.

Override the binary (dev/testing):

```sh
AKME_NX_BINARY=/path/to/nx npx akme version
```
