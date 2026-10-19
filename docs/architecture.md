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
A full 106-tile set is representable. A partial inventory is valid too: it need
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

Representation errors and unsupported jokers are checked in input order before
size and arrangement checks. The result contains the first issue, not an exhaustive
list. A well-formed joker yields `unsupported_joker`, not an assertion that a joker
meld is impossible. A malformed joker yields `invalid_tile` when encountered.
The CLI renders indices as one-based positions and unsupported as a distinct
status. An exact duplicate in a numbered meld also covers a repeated group color
or a repeated run number; gaps and mixed collections have separate issue codes.

Two different contracts are intentionally **not implemented** yet:

- **Layout/table validation** must preserve group boundaries and decide whether
  displayed run order is valid; it must also check total known inventory.
- **Move validation** must compare an original state and proposed result, conserve
  table tiles, consume only available rack tiles, and account for turn context
  such as the initial meld and joker reuse. A valid final table alone is not enough.

Do not add placeholder APIs that report success for those unimplemented checks.
The CLI is still one growing collection, not a table editor.

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

Start table validation with manually supplied structured tiles, then move search,
then a photo adapter feeding the same representation. Keep uncertain image
readings separate from confirmed tile values. Define the search objective and
report whether a result is optimal, merely feasible, incomplete/timed out, or a
proven absence of moves. None of those search statuses exist in this meld-only API.
