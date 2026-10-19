package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/MGaulia/rummy-analizer/internal/game"
)

func TestParseTile(t *testing.T) {
	tests := []struct {
		input   string
		want    game.Tile
		wantErr error
	}{
		{"", game.Tile{}, ErrInputFormat},
		{"     ", game.Tile{}, ErrInputFormat},
		{"blue", game.Tile{}, ErrInputFormat},
		{"blue1", game.Tile{}, ErrInputFormat},
		{"blue 1 extra", game.Tile{}, ErrInputFormat},
		{"blue  1", game.Tile{Color: game.Blue, Number: 1}, nil},
		{"orange\t13", game.Tile{Color: game.Orange, Number: 13}, nil},
		{"  ReD   12\r\n", game.Tile{Color: game.Red, Number: 12}, nil},
		{"blue 0", game.Tile{}, ErrNumber},
		{"blue -1", game.Tile{}, ErrNumber},
		{"blue +1", game.Tile{}, ErrNumber},
		{"blue 01", game.Tile{}, ErrNumber},
		{"blue 99999999999999999999999999", game.Tile{}, ErrNumber},
		{"abcd", game.Tile{}, ErrInputFormat},
		{"very long sentence", game.Tile{}, ErrInputFormat},
		{"haha 1", game.Tile{}, ErrColor},
		{"blue 99", game.Tile{}, ErrNumber},
		{"blue 1", game.Tile{Color: game.Blue, Number: 1}, nil},
		{"  blue 1  ", game.Tile{Color: game.Blue, Number: 1}, nil},
		{"BLUE 1", game.Tile{Color: game.Blue, Number: 1}, nil},
	}
	for _, test := range tests {
		got, err := (Parser{}).ParseTile(test.input)
		if got != test.want {
			t.Errorf("ParseTile(%s) = %v, want %v", test.input, got, test.want)
		}
		if !errors.Is(err, test.wantErr) {
			t.Errorf("ParseTile(%q) error = %v, want %v", test.input, err, test.wantErr)
		}
	}
}

// referenceParseTile uses a literal number vocabulary, not the production
// parser's integer conversion or Tile.Validate. It checks the requested value
// and error category, including valid input that must not be rejected.
func referenceParseTile(input string) (game.Tile, error) {
	fields := strings.Fields(input)
	if len(fields) == 1 && strings.EqualFold(fields[0], "joker") {
		return game.Tile{Kind: game.Joker}, nil
	}
	if len(fields) != 2 {
		return game.Tile{}, ErrInputFormat
	}
	var color game.Color
	switch strings.ToLower(fields[0]) {
	case "black":
		color = game.Black
	case "blue":
		color = game.Blue
	case "orange":
		color = game.Orange
	case "red":
		color = game.Red
	default:
		return game.Tile{}, ErrColor
	}
	for index, token := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13"} {
		if fields[1] == token {
			return game.Tile{Color: color, Number: index + 1}, nil
		}
	}
	return game.Tile{}, ErrNumber
}

func FuzzParseTile(f *testing.F) {
	for _, color := range []string{"black", "blue", "orange", "red"} {
		for _, number := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13"} {
			f.Add(color + " " + number)
			f.Add("\u2003" + strings.ToUpper(color) + "\t" + number + "\r\n")
		}
	}
	for _, input := range []string{"blue1", "", "blue +1", "blue 01", "blue 14", "unknown 1", "JOKER", "joker 1", "blue 1 extra"} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		want, wantErr := referenceParseTile(input)
		got, err := (Parser{}).ParseTile(input)
		if got != want || !errors.Is(err, wantErr) {
			t.Fatalf("ParseTile(%q) = %+v, %v; want %+v, %v", input, got, err, want, wantErr)
		}
	})
}
