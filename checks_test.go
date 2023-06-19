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
				{Colour: GREY, Number: 4},
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

func TestIsValidSet(t *testing.T) {
	tests := []struct {
		name  string
		input []Card
		want  bool
	}{
		{
			name: "colour same",
			input: []Card{
				{Colour: GREY, Number: 1},
				{Colour: GREY, Number: 1},
			},
		},
		{
			name: "not equal numbers",
			input: []Card{
				{Colour: GREY, Number: 1},
				{Colour: BLACK, Number: 2},
			},
		},
		{
			name: "valid",
			want: true,
			input: []Card{
				{Colour: GREY, Number: 1},
				{Colour: BLACK, Number: 1},
				{Colour: GREEN, Number: 1},
				{Colour: PURPLE, Number: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidSet(tt.input)
			if result != tt.want {
				t.Errorf("isValidSet() = %v, want %v", result, tt.want)
			}
		})
	}
}
