package game

import (
	"testing"
)

var benchmarkResult bool

func BenchmarkCombinations(b *testing.B) {
	tests := []struct {
		name  string
		check func([]Tile) bool
		cards []Tile
	}{
		{"ThreeSet", CanFormGroup, []Tile{{Color: Orange, Number: 13}, {Color: Blue, Number: 13}, {Color: Red, Number: 13}}},
		{"FourSet", CanFormGroup, []Tile{{Color: Black, Number: 13}, {Color: Red, Number: 13}, {Color: Blue, Number: 13}, {Color: Orange, Number: 13}}},
		{"ThreeSequence", CanFormRun, []Tile{{Color: Orange, Number: 11}, {Color: Orange, Number: 12}, {Color: Orange, Number: 13}}},
		{"FourSequence", CanFormRun, []Tile{{Color: Orange, Number: 10}, {Color: Orange, Number: 11}, {Color: Orange, Number: 12}, {Color: Orange, Number: 13}}},
		{"InvalidThreeSet", CanFormGroup, []Tile{{Color: Blue, Number: 1}, {Color: Blue, Number: 2}, {Color: Blue, Number: 3}}},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmarkResult = tt.check(tt.cards)
			}
		})
	}
}
