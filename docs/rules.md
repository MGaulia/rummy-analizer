# Rummikub validation

The target game is standard Rummikub. Rule reference:
[official English instructions](https://rummikub.com/wp-content/uploads/2019/12/2600-English.pdf).

The `check` package validates **one group of numbered tiles**:

- A group has three or four tiles of the same number, with distinct colors.
- A run has three to thirteen distinct consecutive numbers of one color.
- Numbers are 1–13. Runs cannot wrap: 12–13–1 and 13–1–2 are invalid.

Input order does not matter: the API checks whether the supplied tiles can form
one group/run, without rearranging or mutating the caller's slice. A future photo
validator must separately check the displayed order if that is required.

`check.IsCombination` handles all supported lengths. The four functions in
`combinations` retain their names and exact three/four-tile length requirements,
but use the same rule implementation. Validation takes O(n) time and constant
space, without heap allocations or generated tables.

Empty or short groups, invalid values, and duplicate tiles within a group are
rejected. Unlike the old generated tables, all color orderings work for groups,
and wrapping runs are rejected. The old `Card` type, number constant names, and
grey/black/green/purple color identifiers remain compatible with existing callers;
these four color labels are project identifiers, not a mapping of the physical
set's printed colors. Photo recognition will need an explicit color mapping.

## Scope and next steps

This is not a full table validator or a move solver. Jokers are not yet modeled;
the API must not be treated as a validator for joker-containing groups.

The standard set has two copies of each numbered tile and two jokers. A future
table validator must check both groups and total inventory. Duplicate physical
copies may appear in separate groups and must not be collapsed into one tile.

For move recommendations, track whether the player has completed the 30-point
initial meld. That opening uses rack tiles only. Later turns may rearrange table
tiles, but must leave every tile in a valid group/run. Joker retrieval and reuse
also require explicit rules in the solver.

Photo recognition should produce grouped tiles and flag uncertain readings for
correction. Then validate the table before searching with the player's rack.
Define "best move" explicitly (for example, most rack tiles played). Search must
distinguish a proven lack of moves from a time limit or unsupported input.
