"use strict";

// Best-effort: npx / offline / CI without network should not fail install.
// The bin shim downloads on first run if vendor/ is empty.
const { ensureBinary } = require("../lib/binary");

ensureBinary().catch((err) => {
  const msg = err && err.message ? err.message : String(err);
  console.warn(`[akme] binary download deferred: ${msg}`);
});
