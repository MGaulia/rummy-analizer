# Architecture and migration

## Keep the application small

The dependency direction is `main -> internal/cli -> internal/game`.
The game domain imports no CLI, image, network, or persistence code. The CLI owns
text parsing, alias configuration, display formatting, process arguments, and
stream handling. `main` owns process exit status and stderr. There is no generic
repository layer, service hierarchy, dependency-injection framework, or CI setup.

`Tile` is a comparable value. Its zero value is invalid; joker is an explicit
kind, not an invalid numbered tile. Canonical colors are a closed four-color
vocabulary. Input aliases live in the CLI and cannot redefine canonical colors.
Photo color interpretation must be explicit rather than reusing legacy names as
unverified physical mappings.

`Inventory` is a bounded multiset backed by a private fixed array. Copies are
independent, so a future search can branch without sharing mutable counts.
`Consume(tile)` removes one occurrence; `ConsumeAll(tiles)` removes the exact
requested multiplicities atomically. There is no remove-every-match operation.
Both 106-tile and 160-tile editions are representable. A partial inventory is valid too: it need
not include unseen tiles. Combined table+rack counts must obey the same bounds.
Counts do not track where a tile is or which physical copy it is. Add snapshot-
local IDs at a photo or move-explanation boundary when those features exist.

## Validation contracts

`AnalyzeMeld` answers **can this unordered collection form one meld?** It returns
`Group` or `Run` on success. A zero result is not valid. A failure has an issue
code and zero-based input indices (`-1` when no individual tile applies).
Duplicate reports reference both the repeated tile and its first occurrence.
The analyzer does not sort or modify input; it uses O(n) time and constant scratch
space. All boolean conveniences call the same analyzer.

Jokers are now supported. Representation errors precede size and arrangement
checks. A successful result supplies concrete numbered roles, a represented-value
score, and an ambiguity flag. A role does not change a joker's physical tile value.
Boolean group/run checks use the same resolver with a kind filter, since a joker
collection can support both kinds. Analysis uses fixed scratch arrays.

`AnalyzeLayout` preserves adjacency; runs may ascend or descend but cannot be
arbitrarily permuted or wrap. `Piece.ID` identifies a physical occurrence within a
snapshot, including duplicate values. `Piece.As` optionally constrains a table
joker's current role. Rack pieces cannot carry roles. `ValidatePosition` checks
all displayed melds and known inventory together; its report never aliases input.
An empty table is allowed. Empty melds, duplicate IDs, unknown editions, invalid
tiles, and excess copies are rejected. It does not invent unseen opponent tiles.

The four- and six-player editions share meld rules but have different inventory
limits. `Rules.OpeningRequired` is independent of `Turn.Opened`; `DefaultRules`
enables the official opening. Move validation must enforce opening and turn state,
conserve every physical table tile, and consume only identified rack pieces.
Whole-table validity alone does not establish a legal move.

## Migration from the old packages

The old exported packages are removed intentionally for this single-user CLI.
There are no aliases that imply a stable library API. Migration inside this
module is:

| Previous concept | New application API |
|---|---|
| `model.Card`, card/rank constants | `game.Tile`, numeric values and bounds |
| `mapper.StringToCard` | `cli.Parser.ParseTile` |
| `check.IsCombination` | `game.AnalyzeMeld` / `game.CanFormMeld` |
| `check.IsValidSet` / `IsValidSequence` | `game.CanFormGroup` / `CanFormRun` |
| fixed-size `combinations` wrappers | general meld APIs; check exact length at a caller only when needed |
| `repository.Delete` / `RemoveAll` | removed; exact-count inventory consumption for gameplay |

Old regression scenarios were migrated to the new domain vocabulary. A color
substitution in a test preserves its equality/distinctness scenario; it is NOT a
physical-color mapping for users. Legacy CLI labels now fail unless explicitly
configured as aliases. Error strings and result labels changed to reflect the new
semantics. `go run .`, streaming input, EOF behavior, and recovery after malformed
input remain supported.

## Next feature boundary

Use structured table validation for move search,
then a photo adapter feeding the same representation. Keep uncertain image
readings separate from confirmed tile values. Define the search objective and
report whether a result is optimal, merely feasible, incomplete/timed out, or a
proven absence of moves. Search statuses belong to the search layer, not meld analysis.

## Application, rendering, and offline phone adapters

`internal/app` is the shared request/response boundary for CLI and Go/Wasm.
`internal/render` draws only validated identified positions. The game package
still owns rules and depends on neither adapter. `cmd/phonewasm` bridges bounded
requests and cancellation; the phone interface invokes the actual Go engine.
Local script-embedded Wasm bytes avoid WKWebView file-URL fetch/worker problems.
The generated bundle is reproducible and ignored by Git.

Raw photo observations and corrected/confirmed positions remain separate in
`phone/web/photo-state.js`. The iPhone adapter uses native image import/camera
and local Vision proposal code. It is not a verified recognition pipeline or
phone product until real-photo and physical-device acceptance. See
phone/README.md for exact supported checks and remaining platform requirements.
