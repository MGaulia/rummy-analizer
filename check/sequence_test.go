package check

import (
	"main/model"
	"testing"
)

func TestIsValidSequence(t *testing.T) {
	tests := []struct {
		name  string
		input []model.Card
		want  bool
	}{
		{
			name: "colour mismatch",
			input: []model.Card{
				{Colour: model.COLOR_GREY, Number: 1},
				{Colour: model.COLOR_BLACK, Number: 2},
			},
		},
		{
			name: "number same",
			input: []model.Card{
				{Colour: model.COLOR_GREY, Number: 1},
				{Colour: model.COLOR_GREY, Number: 1},
			},
		},
		{
			name: "number too big",
			input: []model.Card{
				{Colour: model.COLOR_GREY, Number: 1},
				{Colour: model.COLOR_GREY, Number: 3},
			},
		},
		{
			name: "true",
			want: true,
			input: []model.Card{
				{Colour: model.COLOR_GREY, Number: 1},
				{Colour: model.COLOR_GREY, Number: 2},
				{Colour: model.COLOR_GREY, Number: 3},
				{Colour: model.COLOR_GREY, Number: 4},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSequence(tt.input)
			if result != tt.want {
				t.Errorf("isValidSequence() = %v, want %v", result, tt.want)
			}
		})
	}
}
