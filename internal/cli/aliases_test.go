package cli

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/MGaulia/rummy-analizer/internal/game"
)

func TestParserCanonicalValues(t *testing.T) {
	for _, color := range []game.Color{game.Black, game.Blue, game.Orange, game.Red} {
		for number := game.MinNumber; number <= game.MaxNumber; number++ {
			tiles := []game.Tile{{Color: color, Number: number}}
			text := strings.Trim(formatTiles(tiles), "[]")
			got, err := (Parser{}).ParseTile(text)
			if err != nil || got != tiles[0] {
				t.Fatalf("%s: %+v, %v", text, got, err)
			}
		}
	}
	for _, text := range []string{"joker", " JoKeR \r\n"} {
		if got, err := (Parser{}).ParseTile(text); err != nil || got != (game.Tile{Kind: game.Joker}) {
			t.Fatalf("joker: %+v %v", got, err)
		}
	}
	for _, text := range []string{"red 14", "red 1.0", "red one", "joker 1", "grey 1", "green 1", "purple 1"} {
		if got, err := (Parser{}).ParseTile(text); err == nil || got != (game.Tile{}) {
			t.Fatalf("accepted %q: %+v", text, got)
		}
	}
}

func TestExplicitAliases(t *testing.T) {
	aliases := map[string]game.Color{" gReY ": game.Blue}
	parser, err := NewParser(aliases)
	if err != nil {
		t.Fatal(err)
	}
	aliases[" gReY "] = game.Red
	got, err := parser.ParseTile("GREY 7")
	if err != nil || got != (game.Tile{Color: game.Blue, Number: 7}) {
		t.Fatalf("alias/copy: %+v %v", got, err)
	}
	for _, aliases := range []map[string]game.Color{
		{"": game.Blue}, {"two words": game.Blue}, {"joker": game.Blue},
		{"blue": game.Red}, {"grey": "unknown"}, {"grey": game.Blue, "GREY": game.Red},
		{"a,b": game.Blue}, {"a=b": game.Blue},
	} {
		if _, err := NewParser(aliases); !errors.Is(err, ErrColorAlias) {
			t.Fatalf("accepted aliases %+v", aliases)
		}
	}
	for _, spec := range []string{"", "grey=blue", " GREY = BLUE ", "blue=blue", "grey=blue,GREY=blue"} {
		if _, err := parserFromAliases(spec); err != nil {
			t.Errorf("%q: %v", spec, err)
		}
	}
	for _, spec := range []string{"grey", "=blue", "grey=", "grey=blue,", "grey=blue=red", "grey=blue,GREY=red"} {
		if _, err := parserFromAliases(spec); !errors.Is(err, ErrColorAlias) {
			t.Errorf("accepted %q", spec)
		}
	}
}

func TestRunArgumentsAndUnsupportedInput(t *testing.T) {
	for _, tc := range []struct {
		args        []string
		input, want string
	}{
		{[]string{"-h"}, "", "Usage:"},
		{[]string{"-color-aliases", "grey=blue"}, "grey 1\ngrey 2\ngrey 3", "[blue 1, blue 2, blue 3]\nCan form group/run: true"},
		{nil, "joker", "Can form group/run: unsupported"},
		{nil, "blue 7\nred 7\nblue 7", "tile 3 repeats tile 1"},
		{nil, "blue 1\nred 2\nblue 3", "neither one shared number nor one shared color"},
		{nil, "blue 1\nblue 2\nblue 4", "consecutive numbers"},
		{[]string{"-color-aliases", "a=blue,b=blue"}, "a 7\nred 7\nb 7", "tile 3 repeats tile 1"},
	} {
		var output bytes.Buffer
		if err := Run(tc.args, strings.NewReader(tc.input), &output); err != nil || !strings.Contains(output.String(), tc.want) {
			t.Fatalf("%+v: %q, %v", tc, output.String(), err)
		}
	}
	for _, args := range [][]string{{"-unknown"}, {"-color-aliases"}, {"extra"}, {"-color-aliases", "grey"}, {"--", "extra"}} {
		if err := Run(args, strings.NewReader(""), io.Discard); err == nil {
			t.Fatalf("accepted arguments %v", args)
		}
	}
	if err := Run([]string{"-h"}, strings.NewReader(""), &failingWriter{}); !errors.Is(err, errIO) {
		t.Fatalf("help write: %v", err)
	}
	if err := Run(nil, strings.NewReader("blue 1\n"), &failingWriter{writesLeft: 3}); !errors.Is(err, errIO) {
		t.Fatalf("reason write: %v", err)
	}
	if got := formatTiles([]game.Tile{{Kind: game.Joker}}); got != "[joker]" {
		t.Fatal(got)
	}
	if got := issueMessage(game.ValidationIssue{Code: game.InvalidTile, TileIndex: 0}); !strings.Contains(got, "tile 1") {
		t.Fatal(got)
	}
	if got := issueMessage(game.ValidationIssue{}); got != "unrecognized validation result" {
		t.Fatal(got)
	}
}
