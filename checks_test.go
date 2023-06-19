package main

import "testing"

func TestIsValidSequence(t *testing.T) {
	tests := []struct {
		name  string
		input []Card
		want  bool
	}{
		{
			name: "colour mismatch",
			input: []Card{
				{Colour: GREY, Number: 1},
				{Colour: BLACK, Number: 2},
			},
		},
		{
			name: "number same",
			input: []Card{
				{Colour: GREY, Number: 1},
				{Colour: GREY, Number: 1},
			},
		},
		{
			name: "number too big",
			input: []Card{
				{Colour: GREY, Number: 1},
				{Colour: GREY, Number: 3},
			},
		},
		{
			name: "true 3 cards",
			want: true,
			input: []Card{
				{Colour: GREY, Number: 1},
				{Colour: GREY, Number: 2},
				{Colour: GREY, Number: 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidSequence(tt.input)
			if result != tt.want {
				t.Errorf("isValidSequence() = %v, want %v", result, tt.want)
			}
		})
	}
}
