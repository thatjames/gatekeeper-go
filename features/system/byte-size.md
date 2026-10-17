---
type: Feature
title: Byte size helper
description: util.ByteSize formats a byte count as B, K, M or G; tested but not called by any code (system info has its own copy).
status: stable
code:
  - internal/util/bytes.go
  - internal/util/bytes_test.go
tags: [util]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Behavior
- `ByteSize(n)`: `0B`, then one decimal with `.0` trimmed, units up to G.

# Decisions
- *open*: not called anywhere; `system_linux.go` has its own `byteSize`.
  Found after the grilling; the dead code plan asks whether to drop it or
  use it from `system`.
