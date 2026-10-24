---
type: Feature
title: UI DHCP pages
description: Lease pages (active and reserved tables, details, reserve, edit, release) and the DHCP settings view and form.
status: stable
code:
  - internal/web/ui/src/scenes/dhcp/**
  - internal/web/ui/src/lib/dhcp/**
tags: [ui, dhcp]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T17:20:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Behavior
- `#/dhcp/leases`: active and reserved lease tables; a lease opens its
  details; "Add Reserved Lease" posts a reservation; a reservation can be
  edited or released. Each action shows the lists the server answers.
- `#/dhcp/settings`: the options; edit opens a form with an interface picker
  from `/system/interfaces` and a DNS server list, saved with
  `PUT /dhcp/options`; field errors show under their fields.
- A viewer ([OIDC roles](/web/auth-roles.md)) sees no Add Lease, Reserve,
  Save, Release or Edit Settings; a lease's details open read only.

# Interfaces
- [DHCP API](/web/dhcp-api.md).
