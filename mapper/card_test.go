package mapper

import (
	"main/model"
	"testing"
)

func TestStringToCard(t *testing.T) {
	tests := []struct {
		input   string
		want    model.Card
		wantErr error
	}{
		{"abcd", model.Card{}, ErrInputLength},
		{"very long sentence", model.Card{}, ErrInputLength},
		{"haha 1", model.Card{}, ErrInvalidColorInput},
		{"grey 99", model.Card{}, ErrInvalidNumberInput},
		{"grey 1", model.Card{Colour: model.COLOR_GREY, Number: 1}, nil},
		{"  grey 1  ", model.Card{Colour: model.COLOR_GREY, Number: 1}, nil},
		{"GREY 1", model.Card{Colour: model.COLOR_GREY, Number: 1}, nil},
	}
	for _, test := range tests {
		got, err := StringToCard(test.input)
		if got != test.want {
			t.Errorf("StringToCard(%s) = %v, want %v", test.input, got, test.want)
		}
		if err != nil && err.Error() != test.wantErr.Error() {
			t.Errorf("StringToCard(%s) = %s, want %s", test.input, err, test.wantErr)
		}
	}
}
