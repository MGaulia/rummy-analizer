package game

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestMeldIssues(t *testing.T) {
	for _, tc := range []struct {
		name           string
		tiles          []Tile
		code           IssueCode
		index, related int
	}{
		{"empty", nil, InvalidSize, -1, -1},
		{"short", []Tile{numbered(Blue, 1)}, InvalidSize, -1, -1},
		{"invalid before short", []Tile{{}}, InvalidTile, 0, -1},
		{"joker alone is too short", []Tile{{Kind: Joker}}, InvalidSize, -1, -1},
		{"malformed joker", []Tile{{Kind: Joker, Number: 1}}, InvalidTile, 0, -1},
		{"duplicate", []Tile{numbered(Blue, 7), numbered(Red, 7), numbered(Blue, 7)}, DuplicateTile, 2, 0},
		{"gap", []Tile{numbered(Blue, 1), numbered(Blue, 2), numbered(Blue, 4)}, RunGap, -1, -1},
		{"wrap", []Tile{numbered(Blue, 12), numbered(Blue, 13), numbered(Blue, 1)}, RunGap, -1, -1},
		{"mixed", []Tile{numbered(Blue, 1), numbered(Red, 2), numbered(Blue, 3)}, MixedMeld, -1, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := append([]Tile(nil), tc.tiles...)
			result := AnalyzeMeld(tc.tiles)
			want := ValidationIssue{Code: tc.code, TileIndex: tc.index, RelatedIndex: tc.related}
			if result.Valid() || result.Kind != "" || result.Issue != want {
				t.Fatalf("got %+v, want issue %+v", result, want)
			}
			if !reflect.DeepEqual(tc.tiles, before) {
				t.Fatal("input mutated")
			}
		})
	}
	if (MeldResult{}).Valid() || (MeldResult{Kind: "unknown"}).Valid() || (MeldResult{Kind: Run, Issue: ValidationIssue{Code: RunGap}}).Valid() {
		t.Fatal("invalid result reports success")
	}
}

func TestMeldDoesNotValidateDisplayedOrder(t *testing.T) {
	tiles := []Tile{numbered(Red, 3), numbered(Red, 1), numbered(Red, 2)}
	before := append([]Tile(nil), tiles...)
	result := AnalyzeMeld(tiles)
	if !result.Valid() || result.Kind != Run || !reflect.DeepEqual(tiles, before) {
		t.Fatal("unordered feasibility contract changed")
	}
}

// An independent map/sort reference avoids sharing the implementation's
// min/max test and fixed-size duplicate table.
func referenceMeld(tiles []Tile) MeldKind {
	if len(tiles) < 3 || len(tiles) > 13 {
		return ""
	}
	validColors := map[Color]bool{Black: true, Blue: true, Orange: true, Red: true}
	seen := make(map[Tile]bool)
	for _, tile := range tiles {
		if tile.Kind != Numbered || !validColors[tile.Color] || tile.Number < 1 || tile.Number > 13 || seen[tile] {
			return ""
		}
		seen[tile] = true
	}
	if len(tiles) <= 4 {
		group := true
		for _, tile := range tiles {
			group = group && tile.Number == tiles[0].Number
		}
		if group {
			return Group
		}
	}
	ordered := append([]Tile(nil), tiles...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Number < ordered[j].Number })
	for i := 1; i < len(ordered); i++ {
		if ordered[i].Color != ordered[0].Color || ordered[i].Number != ordered[i-1].Number+1 {
			return ""
		}
	}
	return Run
}

func TestMeldRandomizedReference(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	colors := []Color{Black, Blue, Orange, Red, "unknown", ""}
	maxInt := int(^uint(0) >> 1)
	for trial := 0; trial < 10000; trial++ {
		var tiles []Tile
		switch trial % 3 {
		case 0:
			color := colors[rng.Intn(4)]
			start := 1 + rng.Intn(11)
			for n := start; n <= 13; n++ {
				tiles = append(tiles, numbered(color, n))
			}
		case 1:
			number := 1 + rng.Intn(13)
			for _, i := range rng.Perm(4)[:3+rng.Intn(2)] {
				tiles = append(tiles, numbered(colors[i], number))
			}
		case 2:
			for n := rng.Intn(17); n > 0; n-- {
				number := rng.Intn(17) - 2
				if trial%7 == 0 {
					number = maxInt
				}
				if trial%11 == 0 {
					number = -maxInt - 1
				}
				tiles = append(tiles, Tile{Color: colors[rng.Intn(len(colors))], Number: number, Kind: Numbered})
			}
		}
		rng.Shuffle(len(tiles), func(i, j int) { tiles[i], tiles[j] = tiles[j], tiles[i] })
		before := append([]Tile(nil), tiles...)
		want := referenceMeld(tiles)
		got := AnalyzeMeld(tiles)
		if got.Kind != want || got.Valid() != (want != "") || !reflect.DeepEqual(tiles, before) {
			t.Fatalf("%+v: got %+v want %s", tiles, got, want)
		}
	}
}

func TestMeldAnalysisAllocations(t *testing.T) {
	for _, tiles := range [][]Tile{
		{numbered(Blue, 1), numbered(Blue, 2), numbered(Blue, 3)},
		{numbered(Blue, 7), numbered(Red, 7), numbered(Black, 7)},
		{numbered(Blue, 7), numbered(Red, 7), numbered(Blue, 7)},
		{{Kind: Joker}},
	} {
		if allocations := testing.AllocsPerRun(100, func() { benchmarkResult = CanFormMeld(tiles) }); allocations != 0 {
			t.Fatalf("analysis allocates: %v", allocations)
		}
	}
}
