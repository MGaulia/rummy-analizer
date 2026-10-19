# Rummikub validation

The target game is standard Rummikub. Rule reference:
[official English instructions](https://rummikub.com/wp-content/uploads/2019/12/2600-English.pdf).

The `internal/game` meld analyzer validates **one collection of numbered tiles**:

- A group has three or four tiles of the same number, with distinct colors.
- A run has three to thirteen distinct consecutive numbers of one color.
- Numbers are 1–13. Runs cannot wrap: 12–13–1 and 13–1–2 are invalid.

Input order does not matter: the API checks whether the supplied tiles can form
one group/run, without rearranging or mutating the caller's slice. A future photo
validator must separately check the displayed order if that is required.

`game.AnalyzeMeld` handles all supported lengths and returns a meld kind or a
structured issue. `CanFormMeld`, `CanFormGroup`, and `CanFormRun` share that
implementation. Validation takes O(n) time and constant space without heap
allocations or generated tables. The old fixed-size wrappers have been removed.

Empty or short collections, invalid values, and duplicate tiles within a meld
are rejected. All color orderings work for groups; wrapping runs are rejected.
Canonical colors are black, blue, orange, and red. Legacy grey/green/purple labels
are not physical mappings; text adapters require explicit aliases. Well-formed
jokers are represented explicitly and return an unsupported status rather than
being treated as numbered tiles or ordinary invalid melds.

## Scope and next steps

This is not a full table validator or a move solver. Jokers can be represented
and counted, but their meld/retrieval rules remain unsupported.

The standard set has two copies of each numbered tile and two jokers. A future
table validator must check both groups and total known inventory. The current
`Inventory` enforces copy counts but not table layout or turn legality. Duplicate physical
copies may appear in separate groups and must not be collapsed into one tile.

For move recommendations, track whether the player has completed the 30-point
initial meld. That opening uses rack tiles only. Later turns may rearrange table
tiles, but must leave every tile in a valid group/run. Joker retrieval and reuse
also require explicit rules in the solver.

Photo recognition should produce grouped tiles and flag uncertain readings for
correction. Then validate the table before searching with the player's rack.
Define "best move" explicitly (for example, most rack tiles played). Search must
distinguish a proven lack of moves from a time limit or unsupported input.
