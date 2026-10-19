package mapper

import (
	"errors"
	"testing"

	"github.com/MGaulia/rummy-analizer/model"
)

func TestStringToCard(t *testing.T) {
	tests := []struct {
		input   string
		want    model.Card
		wantErr error
	}{
		{"", model.Card{}, ErrInputFormat},
		{"     ", model.Card{}, ErrInputFormat},
		{"grey", model.Card{}, ErrInputFormat},
		{"grey1", model.Card{}, ErrInputFormat},
		{"grey 1 extra", model.Card{}, ErrInputFormat},
		{"grey  1", model.Card{Color: model.COLOR_GREY, Number: 1}, nil},
		{"purple\t13", model.Card{Color: model.COLOR_PURPLE, Number: 13}, nil},
		{"  GrEeN   12\r\n", model.Card{Color: model.COLOR_GREEN, Number: 12}, nil},
		{"grey 0", model.Card{}, ErrInvalidNumberInput},
		{"grey -1", model.Card{}, ErrInvalidNumberInput},
		{"grey +1", model.Card{}, ErrInvalidNumberInput},
		{"grey 01", model.Card{}, ErrInvalidNumberInput},
		{"grey 99999999999999999999999999", model.Card{}, ErrInvalidNumberInput},
		{"abcd", model.Card{}, ErrInputLength},
		{"very long sentence", model.Card{}, ErrInputLength},
		{"haha 1", model.Card{}, ErrInvalidColorInput},
		{"grey 99", model.Card{}, ErrInvalidNumberInput},
		{"grey 1", model.Card{Color: model.COLOR_GREY, Number: 1}, nil},
		{"  grey 1  ", model.Card{Color: model.COLOR_GREY, Number: 1}, nil},
		{"GREY 1", model.Card{Color: model.COLOR_GREY, Number: 1}, nil},
	}
	for _, test := range tests {
		got, err := StringToCard(test.input)
		if got != test.want {
			t.Errorf("StringToCard(%s) = %v, want %v", test.input, got, test.want)
		}
		if !errors.Is(err, test.wantErr) {
			t.Errorf("StringToCard(%q) error = %v, want %v", test.input, err, test.wantErr)
		}
	}
}

func FuzzStringToCard(f *testing.F) {
	for _, input := range []string{"grey 1", "purple\t13", "grey1", "green  2", "", "grey +1"} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		card, err := StringToCard(input)
		if err != nil {
			if card != (model.Card{}) {
				t.Fatalf("error returned a partial tile: %v", card)
			}
			return
		}
		if card.Number < model.MinNumber || card.Number > model.MaxNumber {
			t.Fatalf("accepted invalid number: %v", card)
		}
		switch card.Color {
		case model.COLOR_GREY, model.COLOR_BLACK, model.COLOR_GREEN, model.COLOR_PURPLE:
		default:
			t.Fatalf("accepted invalid color: %v", card)
		}
	})
}
