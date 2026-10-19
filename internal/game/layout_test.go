package game

import (
	"reflect"
	"testing"
)

func TestJokerMeldsAndLayout(t *testing.T) {
	j := Tile{Kind: Joker}
	for _, tc := range []struct {
		name               string
		tiles              []Tile
		group, run, layout bool
	}{
		{"gap joker", []Tile{numbered(Blue, 1), j, numbered(Blue, 3)}, false, true, true},
		{"endpoint joker", []Tile{numbered(Blue, 6), numbered(Blue, 7), j}, false, true, true},
		{"ambiguous kind", []Tile{j, numbered(Red, 7), j}, true, true, true},
		{"duplicate group color", []Tile{numbered(Red, 7), numbered(Red, 7), j}, false, false, false},
		{"group", []Tile{numbered(Red, 7), numbered(Blue, 7), j}, true, false, true},
		{"four group", []Tile{numbered(Red, 7), numbered(Blue, 7), j, j}, true, false, true},
		{"five group", []Tile{numbered(Red, 7), numbered(Blue, 7), numbered(Black, 7), j, j}, false, false, false},
		{"all jokers", []Tile{j, j, j}, true, true, true},
		{"all four jokers", []Tile{j, j, j, j}, true, true, true},
		{"descending", []Tile{numbered(Blue, 3), j, numbered(Blue, 1)}, false, true, true},
		{"wrong order", []Tile{numbered(Blue, 3), numbered(Blue, 1), j}, false, true, false},
		{"no wrapping", []Tile{numbered(Blue, 13), j, numbered(Blue, 1)}, false, false, false},
		{"two jokers bridge", []Tile{numbered(Blue, 1), j, j, numbered(Blue, 4)}, false, true, true},
		{"too many missing", []Tile{numbered(Blue, 1), j, j, numbered(Blue, 5)}, false, false, false},
		{"short", []Tile{j, j}, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := append([]Tile(nil), tc.tiles...)
			if CanFormGroup(tc.tiles) != tc.group || CanFormRun(tc.tiles) != tc.run || AnalyzeLayout(tc.tiles, nil).Valid() != tc.layout {
				t.Fatalf("%+v: group=%v run=%v layout=%+v", tc.tiles, CanFormGroup(tc.tiles), CanFormRun(tc.tiles), AnalyzeLayout(tc.tiles, nil))
			}
			for _, got := range []MeldResult{AnalyzeMeld(tc.tiles), AnalyzeLayout(tc.tiles, nil)} {
				if got.Valid() {
					if referenceMeld(got.Roles[:len(tc.tiles)]) == "" {
						t.Fatalf("invalid concrete witness %+v", got)
					}
					for i, tile := range tc.tiles {
						if tile.Kind == Numbered && tile != got.Roles[i] {
							t.Fatal("changed real tile to fit")
						}
					}
				}
			}
			if !reflect.DeepEqual(tc.tiles, before) {
				t.Fatal("mutated input")
			}
		})
	}
}

func TestJokerBindings(t *testing.T) {
	tiles := []Tile{numbered(Blue, 6), numbered(Blue, 7), {Kind: Joker}}
	five, eight := numbered(Blue, 5), numbered(Blue, 8)
	if AnalyzeLayout(tiles, []*Tile{nil, nil, &five}).Valid() {
		t.Fatal("ignored explicit role in wrong position")
	}
	got := AnalyzeLayout(tiles, []*Tile{nil, nil, &eight})
	if !got.Valid() || got.Score != 21 || got.Roles[2] != eight {
		t.Fatalf("wrong bound result: %+v", got)
	}
	if AnalyzeLayout(tiles, []*Tile{nil}).Issue.Code != InvalidRole {
		t.Fatal("role length ignored")
	}
	invalid := Tile{Kind: Joker}
	if AnalyzeLayout(tiles, []*Tile{nil, nil, &invalid}).Issue.Code != InvalidRole {
		t.Fatal("joker-as-joker accepted")
	}
	if AnalyzeLayout(tiles, []*Tile{&five, nil, nil}).Issue.Code != InvalidRole {
		t.Fatal("role on numbered tile accepted")
	}
	if !AnalyzeMeld([]Tile{{Kind: Joker}, numbered(Red, 7), {Kind: Joker}}).Ambiguous {
		t.Fatal("ambiguity not reported")
	}
	if AnalyzeLayout([]Tile{numbered(Blue, 1), {Kind: Joker}, numbered(Blue, 3)}, nil).Ambiguous {
		t.Fatal("unique role reported ambiguous")
	}
}

// For one wildcard, compare against every possible literal substitution using
// the older independent map/sort reference, not the new resolution algorithm.
func TestOneJokerExhaustiveReference(t *testing.T) {
	colors := []Color{Black, Blue, Orange, Red}
	for a := 0; a < 52; a++ {
		for b := 0; b < 52; b++ {
			tiles := []Tile{numbered(colors[a/13], a%13+1), {Kind: Joker}, numbered(colors[b/13], b%13+1)}
			wantGroup, wantRun := false, false
			for substitute := 0; substitute < 52; substitute++ {
				resolved := append([]Tile(nil), tiles...)
				resolved[1] = numbered(colors[substitute/13], substitute%13+1)
				switch referenceMeld(resolved) {
				case Group:
					wantGroup = true
				case Run:
					wantRun = true
				}
			}
			if CanFormGroup(tiles) != wantGroup || CanFormRun(tiles) != wantRun {
				t.Fatalf("joker oracle %+v: group=%v run=%v", tiles, wantGroup, wantRun)
			}
		}
	}
}
