package checks

import (
	"main/models"
	"testing"
)

func TestIsValidSequence(t *testing.T) {
	tests := []struct {
		name  string
		input []models.Card
		want  bool
	}{
		{
			name: "colour mismatch",
			input: []models.Card{
				{Colour: models.COLOR_GREY, Number: 1},
				{Colour: models.COLOR_BLACK, Number: 2},
			},
		},
		{
			name: "number same",
			input: []models.Card{
				{Colour: models.COLOR_GREY, Number: 1},
				{Colour: models.COLOR_GREY, Number: 1},
			},
		},
		{
			name: "number too big",
			input: []models.Card{
				{Colour: models.COLOR_GREY, Number: 1},
				{Colour: models.COLOR_GREY, Number: 3},
			},
		},
		{
			name: "true 3 cards",
			want: true,
			input: []models.Card{
				{Colour: models.COLOR_GREY, Number: 1},
				{Colour: models.COLOR_GREY, Number: 2},
				{Colour: models.COLOR_GREY, Number: 3},
				{Colour: models.COLOR_GREY, Number: 4},
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
