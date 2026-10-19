# Offline Rummikub assistant

Implement the personal assistant described in the owner's CODEX_HANDOFF.md
(2026-10-19). The final workflow is two separate local photos (table, then rack),
visible uncertainty and corrections, confirmed-position validation, legal move
search, and exact final-table/remaining-rack rendering on a phone. No runtime
server, inference API, photo upload, or remote solver is permitted.

- Edition 4: two numbered copies and two jokers (106 tiles). Edition 6: three
  numbered copies and four jokers (160 tiles); both have four colors.
- New position commands require explicit `--edition=4|6`.
- Opening is a rack-only 30 points and does not manipulate the existing table.
  `--opening-required=false` skips this phase, independent of player `--opened`.
  Drawing ends the turn. Opening state never changes table validity.
- Maximize rack tiles played; allow unlimited legal table rearrangement. Stable
  enumeration is the initial engineering tie-break, with no strategic claim.
- Conserve identified physical table tiles; consume exact rack occurrences.
  Jokers retain identity independently of their roles. Verify every returned move.
- Distinguish optimal, feasible/incomplete, proven no move, incomplete/no result,
  invalid or unresolved input, and unsupported rule coverage. A resource limit is
  never proof of no move. Search must support cancellation and bounded resources.
- Render every final table and rack occurrence, rack additions, and joker roles.
- Preserve Go core and `go run .`. Structured JSON is a development interface,
  not a substitute for recognition or the phone app.

Rules reference: booklet D-2600-1236-0041 (120819), linked in docs/rules.md.
PR #9's selected profile includes joker replacement and illustrated releases by
restructuring. Recheck the actual booklet before claiming full transition-rule
coverage. Do not silently mix profiles or invent another house rule.

Acceptance needs real representative photos and a physical-phone end-to-end run
with networking disabled and all assets local. Synthetic tests, desktop runs,
WebAssembly builds, and mock predictions alone do not establish that acceptance.
Do not commit or upload private photos. Model/runtime licenses, reproducible
training/export, actual device/OS performance, and ambiguity correction are part
of the remaining acceptance work. Device/model/latency targets are not fixed.
