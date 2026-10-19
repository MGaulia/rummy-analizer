# Local development and verification

From a Git checkout with Go, Bash, and a working race-detector/C toolchain:

```sh
bash scripts/check.sh
bash scripts/check.sh --fuzz
```

The script resolves the project root, prints the current commit and worktree
status, and checks formatting, working/staged/HEAD whitespace, uncached tests,
vet, races, and an executable built in a temporary directory. Smoke checks cover
help, invalid arguments and exit status, stdout/stderr separation, piped input
without a final newline, input-error recovery, aliases, and unsupported jokers.
It does not stage, commit, push, format files in place, install dependencies, or
configure CI. Build and smoke artifacts are cleaned up; Go uses its normal caches.
Worktree edits are tested as they stand, so include status/diff information when
reporting results from a dirty checkout rather than attributing them to HEAD alone.

The optional fuzz mode requests 10 seconds per target with two workers. Fuzzing
may spend additional time minimizing an interesting input or shutting down. A
new failure can create a regression corpus file under testdata/fuzz; inspect and
retain that case when fixing the defect. A fuzz duration or coverage percentage
is not a proof of correctness. Windows users need a Bash-compatible environment
such as WSL for the script; the Go commands can also be run individually.

Use a supported Go release for development. The go.mod language/minimum directive
is not a recommendation to use an old compiler. Report the version actually run;
do not claim latest-toolchain verification when only an older runtime is available.
The script deliberately fails rather than silently skipping unavailable race checks.

## Review evidence

For fixes, show the regression failing before the fix and passing afterward.
For test improvements, a small temporary mutation can verify that the test catches
the intended wrong behavior; restore it and rerun checks before committing.
Compare published trees with tested trees, distinguish local results from hosted
CI, and disclose checks that could not execute. Run the combined changes locally
when several independent PRs are intended to be used together.

Keep benchmarks separate from correctness gates:

```sh
go test ./internal/game -run '^$' -bench . -benchmem
```

## Guidance consulted

These sources informed the short project-specific AGENTS.md and verification
workflow. They are guidance, not extra dependencies or permission to expand scope.
Reviewed October 19, 2026; prefer the live documentation over copied model advice.

- [Go code review comments](https://go.dev/wiki/CodeReviewComments): formatting,
  clear naming, error handling, and small, concrete APIs.
- [Go test comments](https://go.dev/wiki/TestComments): informative failures and
  helpers that point back to the failing case.
- [Codex best practices](https://developers.openai.com/codex/learn/best-practices):
  concise durable repository instructions and explicit verification criteria.
- [Claude Code best practices](https://code.claude.com/docs/en/best-practices):
  executable checks, scoped changes, and evidence rather than unsupported success claims.
