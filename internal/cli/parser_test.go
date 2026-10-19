package cli

import (
	"errors"
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

func FuzzParseTile(f *testing.F) {
	for _, input := range []string{"blue 1", "orange\t13", "blue1", "red  2", "", "blue +1", "JOKER"} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		card, err := (Parser{}).ParseTile(input)
		if err != nil {
			if card != (game.Tile{}) {
				t.Fatalf("error returned a partial tile: %v", card)
			}
			return
		}
		if err := card.Validate(); err != nil {
			t.Fatalf("parser produced invalid tile: %+v", card)
		}
		if card.Kind == game.Joker {
			return
		}
		if card.Number < game.MinNumber || card.Number > game.MaxNumber {
			t.Fatalf("accepted invalid number: %v", card)
		}
		switch card.Color {
		case game.Blue, game.Black, game.Red, game.Orange:
		default:
			t.Fatalf("accepted invalid color: %v", card)
		}
	})
}
