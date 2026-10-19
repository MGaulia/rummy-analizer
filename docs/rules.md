# Rules profile and validation

Source: [official English Rummikub booklet, D-2600-1236-0041 120819](https://rummikub.com/wp-content/uploads/2019/12/2600-English.pdf),
read October 19, 2026, including the joker illustrations. This profile does not
silently mix in older, more restrictive joker house rules.

## Selected rules

- Edition 4: 106 tiles, two copies of each numbered value and two jokers.
- Edition 6: 160 tiles, three copies of each numbered value and four jokers.
- Both pools use four colors, numbers 1–13. Groups have three or four distinct
  colors with one number; runs have consecutive numbers of one color, no wrapping.
- Jokers stand for numbered tiles, including during a 30-point rack-only opening.
- Existing table manipulation is available on subsequent turns after opening.
- Later rearrangement may use any table tiles, but must end with legitimate
  melds, all previous table tiles retained, and at least one rack tile played.
- A joker can be freed by replacement or restructuring, including using table
  tiles. It must be reused in a valid new set on the same turn. The rack-tile
  requirement is per turn, not a requirement for every joker's new set.
- Drawing ends the turn. A retained joker has a 30-point end-of-game penalty.

Disabling `OpeningRequired` is an explicit house rule: skip the opening phase
entirely and allow table manipulation immediately. It does not disable other
move rules. Edition selects the pool, not the current number of players.

## Implemented API contracts

`AnalyzeMeld` checks unordered feasibility including multiple jokers. Its
concrete witness preserves input indexing, not displayed order. `CanFormGroup`
and `CanFormRun` can both be true for the same wildcard collection. Analysis
chooses a highest-scoring interpretation and marks alternatives as ambiguous.

`AnalyzeLayout` checks actual adjacency, allowing ascending or descending runs.
That direction convention avoids rejecting a valid run viewed from the other
side of a table; nonconsecutive permutations still fail. Explicit joker roles
are checked, and unspecified roles are resolved. A joker remains a joker in
inventory, whatever numbered role it takes in a meld.

`ValidatePosition` keeps group boundaries, validates snapshot-local physical IDs,
and checks combined known table/rack inventory. A partial inventory is allowed;
unseen tiles are not invented. The 13-tile limit applies to a single meld, not to
the complete table or rack. Results are independent of opening state. Every
reported role is one valid interpretation, not a claim about image confidence.

Table validity is not move legality. A move additionally needs the original
position, selected rules, player opening/draw state, and physical tile accounting.
The current single-collection text CLI remains available through `go run .`.

## Transition checks and search coverage

`ValidateMove` verifies original/final positions, exact physical accounting,
opening/draw state, and joker release/reuse checkpoints. Initial unspecified
roles remain ambiguous until bound; explicit roles cannot reset without release.
Checkpoints support chained release/reuse and preparation rearrangements.

`Search` enumerates arbitrary final table partitions, including rack jokers,
and independently validates candidate transitions. It first seeks a single
release certificate, then a bounded graph of release/reuse checkpoints. It
returns a proven optimum when a verified move reaches an exhaustively computed
relaxed partition upper bound, or the absolute rack-size bound. An exhausted
zero relaxed bound proves no move even with table jokers. A larger unresolved
bound or resource limit produces an incomplete result, never a false proof.

The official PDF is now accessible and all four release illustrations were
visually reviewed and encoded as regressions for both editions. See
[joker transitions and proof boundaries](joker-transitions.md) for the exact
source hash, certificate semantics, independent oracle coverage, resource limits,
and remaining temporary-loose-tile coverage limitation.
