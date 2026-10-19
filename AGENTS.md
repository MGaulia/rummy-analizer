# Working on rummy-analizer

## Scope and structure

This is a personal offline Rummikub assistant with a Go CLI development interface.
Preserve `go run .`. See docs/product-spec.md and docs/implementation-plan.md
for the accepted product scope and current milestone status.
Read README.md and docs/architecture.md for the current scope; consult
[development checks](docs/development.md) before publishing changes.
Keep dependencies directed `main -> internal/cli -> internal/game`.
Do not add CI workflows, dependencies, persistence, network services, compatibility
wrappers, or speculative abstractions unless the requested task needs them.

## Invariants

- Tile values and physical tile identity are different. Preserve duplicate counts.
- `Consume` removes one copy. `ConsumeAll` consumes exact requested multiplicities
  atomically and returns the original inventory on failure. No `RemoveAll` API.
- The zero Tile is invalid; a joker is explicit. Keep invalid and unsupported
  results distinct. Do not imply that unsupported joker rules prove no move exists.
- `AnalyzeMeld` means unordered single-meld feasibility. It must not mutate input
  or be used as a substitute for layout, inventory, or move validation.
- Color aliases belong at the text boundary; never guess physical photo colors.
- Keep game logic independent of I/O. Add table, solver, and photo APIs only when
  implementing those features, with their contracts and tests.

## Change and verification workflow

Keep changes focused. Prefer clear Go and concrete types over new layers.
Run gofmt, check errors, and write useful input/actual/expected test failures.
For a bug fix, reproduce it in a regression test and run that test before and
after the fix. Do not weaken a test merely to make changed code pass.
For validation or parsing, check expected values and rejected valid inputs too;
a no-panic check or production Validate call alone is not a correctness oracle.

Run `bash scripts/check.sh` before publishing. Run its `--fuzz` mode for parser or
inventory changes, and use benchmarks only to support a specific performance claim.
Record the exact commit/tree, Go version, commands, results, and any checks that
could not run. Never relabel source inspection or old PR results as fresh tests.
Keep changes in a separate branch/worktree and preserve unrelated owner changes.

## Pull requests

Use focused PRs against `main` unless a dependency truly requires stacking.
Verify the base and diff before opening a PR. Keep PRs unmerged unless the owner
explicitly requests a merge. After an owner merge, verify the commit reached
`main`; a merged PR targeting another branch does not establish that.
Do not claim background progress without an actual scheduled execution.
