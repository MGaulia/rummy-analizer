package game

import (
	"testing"
)

func TestCanFormRun(t *testing.T) {
	tests := []struct {
		name  string
		input []Tile
		want  bool
	}{
		{
			name: "colour mismatch",
			input: []Tile{
				{Color: Blue, Number: 1},
				{Color: Black, Number: 2},
			},
		},
		{
			name: "number same",
			input: []Tile{
				{Color: Blue, Number: 1},
				{Color: Blue, Number: 1},
			},
		},
		{
			name: "number too big",
			input: []Tile{
				{Color: Blue, Number: 1},
				{Color: Blue, Number: 3},
			},
		},
		{
			name: "true",
			want: true,
			input: []Tile{
				{Color: Blue, Number: 1},
				{Color: Blue, Number: 2},
				{Color: Blue, Number: 3},
				{Color: Blue, Number: 4},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanFormRun(tt.input)
			if result != tt.want {
				t.Errorf("isValidSequence() = %v, want %v", result, tt.want)
			}
		})
	}
}
