"use strict";

// Best-effort fallback when optionalDependencies were skipped
// (--omit=optional / --no-optional). Prefer the platform package from npm;
// only hit GitHub releases when that binary is missing.
const { ensureBinary, resolveOptionalBinary } = require("../lib/binary");

if (resolveOptionalBinary()) {
  process.exit(0);
}

ensureBinary().catch((err) => {
  const msg = err && err.message ? err.message : String(err);
  console.warn(`[akme] binary download deferred: ${msg}`);
});
