# Joker transitions and proof boundaries

## Verified source

Official English booklet D-2600-1236-0041 (120819):
https://rummikub.com/wp-content/uploads/2019/12/2600-English.pdf

Downloaded after environment publication on 2026-10-19. SHA-256:
`88a6f8b14c119b19b641dcfd13352f9eca3364f1b1a4219459b0f8f1b506ab00`.
Pages 3 and 4 were rendered and visually inspected, including all four joker
release illustrations. Page 1 confirms rack-only 30-point opening, joker score
as represented, manipulation on subsequent turns, and drawing ending the turn.
The XP pool is 160 tiles with four jokers; the standard pool is 106 with two.

The joker may be replaced with any tiles keeping its set legitimate, including
from the table. The three-tile group example permits either missing color or
both. The split-run example releases the joker using both endpoint additions;
the endpoint example shifts a run; the fourth example distributes black 1 and
black 2 into other table groups. Retrieval requires same-turn reuse in a new set
and at least one rack tile on the turn, not a rack tile in every new joker set.

`TestOfficialFourJokerReleases` encodes the illustrated positions and all three
replacement choices from the first example, in both editions. The booklet's
printed yellow tiles use the repository's explicit orange label in these manual
fixtures; this is not a photo-color inference. Additional tests cover table-only
joker reuse with a rack tile played elsewhere, successive release/reuse,
initial role ambiguity, duplicate identities, malformed stages, and four jokers.

## Physical certificates

A move supplies its final table and optional `releases` checkpoints. A checkpoint
contains a valid `table`, IDs `freed` into the held-joker set, and optionally IDs
`reused` out of that set. A checkpoint can also prepare a role-preserving table
rearrangement. Each held joker must be an original table occurrence. Every other
original table occurrence remains present, and physical values cannot change.
Previously introduced rack pieces must remain played. Each retrieved joker is
reused in a new physical set; changing order/role alone does not make the same
physical set new. Chained reuse compares with the set at that retrieval, not
only the original photographed set.

Initially unspecified roles may take any interpretation consistent with their
original displayed meld. A concrete supplied/established role stays fixed until
release. Nested extensions/contractions compare represented support values;
substitution of equal numbered physical copies or release of another joker does
not spuriously force this retained joker to be retrieved. Physical occurrence
conservation is checked separately and never collapses equal-looking copies.

Both the public transition checker and witness finder replay the certificate
through the same accounting/rule gates. Search tries a single release first,
then explores a graph of valid table checkpoints, including preparations,
release/reuse, and sequential joker exchanges. The graph has at most 512 retained
states and shares the search work/cancellation budget. All returned moves are
validated independently at the application boundary before exact rendering.

## Honest optimum/absence proofs

The final-partition enumeration is unrestricted over all meld templates and all
physical table occurrences, with optional rack occurrences. It supplies a
relaxed upper bound that ignores joker retrieval requirements. Once exhausted:

- Bound zero proves no rack tile can be added even under the relaxation, hence
  no legal move, including when the original table contains jokers.
- A verified move attaining that bound proves the optimal primary objective.
- A verified move emptying the rack reaches the absolute upper bound immediately.
- A larger bound than the best verified result leaves unresolved transitions;
  return incomplete, even if the enumerated checkpoint graph found no witness.
- A work/deadline/cancellation limit never establishes absence or optimality.

`upper_bound` is returned only when established by exhaustive final partitions
or the absolute rack-size bound. One-table-joker bounds are checked against
31 nonempty rack subsets using an independent wildcard-substitution and
set-partition oracle, alongside the original 576 numbered oracle cases.

The valid-table checkpoint certificate remains sufficient rather than a proven
complete representation of every possible temporary loose-tile manipulation.
The booklet guarantees legitimate sets at the end of manipulation; this code
does not equate failure to find such a certificate with an illegal final table.
The `table_joker_transition_coverage_incomplete` reason remains when the relaxed
bound exceeds the verified result, with `joker_witness_state_limit` when relevant.
This is an explicit search limitation, not a new house rule or proof of no move.
