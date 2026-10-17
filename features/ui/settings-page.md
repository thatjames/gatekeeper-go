---
type: Feature
title: UI settings page
description: draft - a settings page in the UI, with backing API and web server changes, in progress on the feature/settings-page branch.
status: draft
code:
  - internal/web/ui/src/lib/settings/**
tags: [ui, settings]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Behavior
- Not on main. `feature/settings-page` (1 commit, 2026-09-14) touches
  `internal/web/web.go`, `internal/web/v1/dns.go`, `types.go` and the UI.

- On main only a stub: `src/lib/settings/settings.js` imports the API and
  nothing uses it.

# Decisions
- *open*: what the page covers; settle from the branch when the plan starts.
