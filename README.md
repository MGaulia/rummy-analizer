# Rummy analyzer

A small, dependency-free Go CLI and rules foundation for **Rummikub**.
It checks whether a single collection of numbered tiles **can form** a group or
run after rearrangement. It does not validate displayed order, a whole table,
or a legal move. Jokers are modeled but their meld rules remain unsupported.

## Run

Requires Go 1.18 or newer.

```sh
go run .
printf 'blue 3\nblue 1\nblue 2\n' | go run .
```

Enter one tile per line as `<color> <number>`, or `joker`. Canonical colors are
`black`, `blue`, `orange`, and `red`; numbers are `1` through `13`. Color input is
case-insensitive. Extra whitespace is allowed; signs and leading zeros are not.

The CLI accumulates **one collection** and prints its feasibility after each
tile. Start a fresh process to check another collection. Blank lines are ignored.
Malformed input reports an error without adding a tile. End input with Ctrl-D
(Unix) or Ctrl-Z followed by Enter (Windows). Piped input and a final line without
a newline are supported.

After the example above, the output includes:

```text
[blue 3, blue 1, blue 2]
Can form group/run: true
```

This is not an assertion that the displayed order is correct. Invalid collections
include an explanatory reason; duplicate reports use one-based input positions.
A joker-containing collection prints `Can form group/run: unsupported`, not a
claim that it is impossible to play.

### Explicit color aliases

Legacy `grey`, `green`, and `purple` labels have no assumed physical-color
mapping. Prefer canonical colors. For a custom label, explicitly choose its
meaning, for example:

```sh
printf 'myblue 1\nmyblue 2\nmyblue 3\n' | go run . -color-aliases myblue=blue
```

Multiple mappings are comma-separated. Aliases are case-insensitive; conflicting
aliases and attempts to reassign canonical colors or `joker` are rejected.
Aliases normalize to canonical tile values, so two labels for the same color do
not create distinct colors. See `go run . -h` for usage.

## Structure

- `internal/game`: tile values, duplicate-safe inventory, and meld analysis.
- `internal/cli`: text parsing, explicit color aliases, formatting, and I/O.
- `main.go`: process entry point only; `go run .` remains supported.

`Inventory` has value semantics. `Consume` removes one copy; `ConsumeAll` consumes
exactly the requested list (including its duplicate counts), or returns the
original state unchanged on failure. Neither removes every matching value.
Construction validates representation and at most two copies of each value,
including two jokers. Check combined known table+rack multiplicity, not only each
collection independently. Inventory counts alone do not establish move legality.

`AnalyzeMeld` returns a meld kind or a structured issue with an issue code and
zero-based tile references where applicable. `CanFormMeld`, `CanFormGroup`, and
`CanFormRun` are boolean conveniences backed by that same rule implementation.
Use `AnalyzeMeld` when unsupported input must be distinguished from invalid input.

## Development

```sh
go test ./...
go vet ./...
go test -race ./...
go test ./internal/game -run '^$' -bench . -benchmem
go test ./internal/game -run '^$' -fuzz FuzzInventoryConsumption -fuzztime 10s
go test ./internal/cli -run '^$' -fuzz FuzzParseTile -fuzztime 10s
```

No CI workflow or third-party dependency is required. See
[rules and scope](docs/rules.md) and [architecture and migration](docs/architecture.md).

## Compatibility decision

This is a single-user application, not a supported public Go library. The old
`model`, `mapper`, `check`, `combinations`, and `repository` packages have been
removed; their public import paths and card-style constants are not retained.
The obsolete `Delete`/`RemoveAll` behavior is not part of the game API. Existing
text input must use canonical colors or an explicit alias; output now says
`Can form group/run` rather than `Valid group/run`. The module path stays
`github.com/MGaulia/rummy-analizer`.
