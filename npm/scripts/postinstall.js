"use strict";

// Same path as bare `npx akme-cli`: ensure binary + greeting, never run CLI.
const { installCLI } = require("../lib/install");

installCLI().catch((err) => {
  const msg = err && err.message ? err.message : String(err);
  console.warn(`[akme] binary install deferred: ${msg}`);
});
