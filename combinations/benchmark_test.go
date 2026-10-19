package combinations

import (
	"testing"

	"github.com/MGaulia/rummy-analizer/model"
)

var benchmarkResult bool

func BenchmarkCombinations(b *testing.B) {
	tests := []struct {
		name  string
		check func([]model.Card) bool
		cards []model.Card
	}{
		{"ThreeSet", IsThreeSetCombination, []model.Card{{Color: model.COLOR_PURPLE, Number: 13}, {Color: model.COLOR_GREY, Number: 13}, {Color: model.COLOR_GREEN, Number: 13}}},
		{"FourSet", IsFourSetCombination, []model.Card{{Color: model.COLOR_BLACK, Number: 13}, {Color: model.COLOR_GREEN, Number: 13}, {Color: model.COLOR_GREY, Number: 13}, {Color: model.COLOR_PURPLE, Number: 13}}},
		{"ThreeSequence", IsThreeSequenceCombination, []model.Card{{Color: model.COLOR_PURPLE, Number: 11}, {Color: model.COLOR_PURPLE, Number: 12}, {Color: model.COLOR_PURPLE, Number: 13}}},
		{"FourSequence", IsFourSequenceCombination, []model.Card{{Color: model.COLOR_PURPLE, Number: 10}, {Color: model.COLOR_PURPLE, Number: 11}, {Color: model.COLOR_PURPLE, Number: 12}, {Color: model.COLOR_PURPLE, Number: 13}}},
		{"InvalidThreeSet", IsThreeSetCombination, []model.Card{{Color: model.COLOR_GREY, Number: 1}, {Color: model.COLOR_GREY, Number: 2}, {Color: model.COLOR_GREY, Number: 3}}},
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
