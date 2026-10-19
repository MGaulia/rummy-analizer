# Rummy analyzer

A small Go foundation for a **Rummikub** analyzer. It currently parses numbered
tiles and checks whether a single collection can form a valid group or run.
Photo recognition, whole-table validation, jokers, and move recommendations are
planned, but are not implemented yet.

## Run

Requires Go 1.18 or newer. There are no third-party dependencies.

```sh
go run .
```

Enter one tile per line as `<color> <number>`. The current project color labels
are `grey`, `black`, `green`, and `purple`; numbers are `1` through `13`.
Color input is case-insensitive, and extra spaces or tabs are accepted.
These legacy labels will need mapping to the physical set's colors for photos.

The CLI accumulates tiles into **one group/run** and prints its validity after
each tile. Start a fresh process to check another group. Empty lines are ignored;
invalid lines print an error without adding a tile. End input with Ctrl-D (Unix)
or Ctrl-Z followed by Enter (Windows).

Input can also be piped; the final line does not need a newline:

```sh
printf 'grey 1\ngrey 2\ngrey 3\n' | go run .
```

After the third tile, the output includes:

```text
[grey 1, grey 2, grey 3]
Valid group/run: true
```

## Packages

- `model`: numbered tiles, colors, and number bounds. `Card` is the existing type
  name retained for compatibility.
- `mapper`: text parsing with explicit format, color, and number errors.
- `check`: allocation-free group/run validation for any supported length.
- `combinations`: compatibility helpers for exactly three or four tiles.
- `repository`: copy-based deletion of **all** matching tiles. This is not yet a
  physical-tile inventory API for a move solver.

See [Rummikub rules and scope](docs/rules.md) for supported rules and the next
steps toward photo validation and move search.

## Development

```sh
go test ./...
go vet ./...
go test -race ./...
go test ./combinations -run '^$' -bench . -benchmem
go test ./mapper -run '^$' -fuzz FuzzStringToCard -fuzztime 10s
```

Malformed input always returns an error and a zero-value tile. `ErrInputFormat`
replaces the former character-count error; `ErrInputLength` remains an alias for
source compatibility. The module path now matches this repository:
`github.com/MGaulia/rummy-analizer`.
