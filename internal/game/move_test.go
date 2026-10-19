package game

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func mp(id string, c Color, n int) Piece { return Piece{ID: id, Tile: Tile{Color: c, Number: n}} }
func mj(id string, c Color, n int) Piece {
	r := Tile{Color: c, Number: n}
	return Piece{ID: id, Tile: Tile{Kind: Joker}, As: &r}
}
func blueRun(prefix string, start, end int) []Piece {
	p := []Piece{}
	for n := start; n <= end; n++ {
		p = append(p, mp(fmt.Sprintf("%s%d", prefix, n), Blue, n))
	}
	return p
}

func TestMoveOpeningAndConservation(t *testing.T) {
	table := blueRun("t", 1, 3)
	rack := []Piece{mp("r", Red, 10), mp("b", Black, 10), mp("o", Orange, 10), mp("x", Blue, 4)}
	before := Position{Table: [][]Piece{table}, Rack: rack}
	legal := Move{Table: [][]Piece{table, rack[:3]}}
	snapshot := ClonePosition(before)
	for _, tc := range []struct {
		name  string
		rules Rules
		turn  Turn
		move  Move
		code  string
	}{
		{"opening", DefaultRules(4), Turn{}, legal, ""},
		{"drawn", DefaultRules(4), Turn{Drawn: true}, legal, "turn_ended"},
		{"first turn cannot extend", DefaultRules(4), Turn{}, Move{Table: [][]Piece{append(clonePieces(table), rack[3])}}, "opening_manipulation"},
		{"opening disabled", Rules{Edition: 4}, Turn{}, Move{Table: [][]Piece{append(clonePieces(table), rack[3])}}, ""},
		{"opened", DefaultRules(4), Turn{Opened: true}, Move{Table: [][]Piece{append(clonePieces(table), rack[3])}}, ""},
		{"missing table", Rules{Edition: 4}, Turn{}, Move{Table: [][]Piece{rack[:3]}}, "tile_accounting"},
		{"no rack", Rules{Edition: 4}, Turn{}, Move{Table: [][]Piece{table}}, "no_rack_tile"},
		{"duplicate ID", Rules{Edition: 4}, Turn{}, Move{Table: [][]Piece{table, table}}, "tile_accounting"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateMove(before, tc.rules, tc.turn, tc.move)
			if got.Code != tc.code || got.Valid != (tc.code == "") {
				t.Fatalf("got %+v, want code %q", got, tc.code)
			}
			if got.Valid && len(got.Played) == 0 {
				t.Fatal("no played tiles")
			}
		})
	}
	if !reflect.DeepEqual(before, snapshot) {
		t.Fatal("mutated input")
	}
	got := ValidateMove(before, DefaultRules(4), Turn{}, legal)
	if got.OpeningScore != 30 || len(got.After.Rack) != 1 || got.After.Rack[0].ID != "x" {
		t.Fatalf("wrong result %+v", got)
	}
	got.After.Table[0][0].ID = "changed"
	if before.Table[0][0].ID == "changed" {
		t.Fatal("aliased input")
	}
	low := Position{Rack: blueRun("r", 1, 3)}
	if got := ValidateMove(low, DefaultRules(4), Turn{}, Move{Table: [][]Piece{low.Rack}}); got.Code != "opening_score" {
		t.Fatalf("accepted low opening %+v", got)
	}
}

func TestMoveJokerReleaseAndIdentity(t *testing.T) {
	source := []Piece{mp("t1", Blue, 1), mp("t2", Blue, 2), mj("j", Blue, 3)}
	rack := []Piece{mp("r3", Blue, 3), mp("r7", Red, 7), mp("o7", Orange, 7)}
	before := Position{Table: [][]Piece{source}, Rack: rack}
	final := [][]Piece{{source[0], source[1], rack[0]}, {rack[1], rack[2], mj("j", Black, 7)}}
	rules := Rules{Edition: 4}
	if got := ValidateMove(before, rules, Turn{}, Move{Table: final}); got.Code != "joker_witness_required" {
		t.Fatalf("accepted missing witness %+v", got)
	}
	move := Move{Table: final, Releases: []JokerRelease{{Table: [][]Piece{{source[0], source[1], rack[0]}}, Freed: []string{"j"}}}}
	got := ValidateMove(before, rules, Turn{}, move)
	if !got.Valid || len(got.Played) != 3 {
		t.Fatalf("valid replacement/reuse rejected %+v", got)
	}
	bad := ClonePosition(Position{Table: final}).Table
	bad[1][2].Tile = Tile{Color: Black, Number: 7}
	if ValidateMove(before, rules, Turn{}, Move{Table: bad}).Valid {
		t.Fatal("changed physical joker")
	}
	// A split release uses only original table numbered tiles. The rack tile is
	// required once for the turn, not once per new joker-containing meld.
	run := blueRun("t", 1, 7)
	run[3] = mj("j", Blue, 4)
	before = Position{Table: [][]Piece{run}, Rack: []Piece{mp("r8", Blue, 8)}}
	final = [][]Piece{run[:3], append(append(clonePieces(run[4:]), mj("j", Blue, 8)), mp("r8", Blue, 8))}
	// The extra real blue 8 cannot share a run with joker-as-8. Instead reuse the
	// joker as 9 after extending 5,6,7 with the rack 8.
	final[1] = append(append(clonePieces(run[4:]), before.Rack[0]), mj("j", Blue, 9))
	stage := [][]Piece{run[:3], run[4:]}
	if got := ValidateMove(before, rules, Turn{}, Move{Table: final, Releases: []JokerRelease{{Table: stage, Freed: []string{"j"}}}}); !got.Valid {
		t.Fatalf("split release rejected %+v", got)
	}
}

func TestSearchRearrangesAndVerifies(t *testing.T) {
	// Two runs can be split into three groups, consuming two rack tiles.
	before := Position{Table: [][]Piece{blueRun("b", 1, 3), {mp("r1", Red, 1), mp("r2", Red, 2), mp("r3", Red, 3)}}, Rack: []Piece{mp("o1", Orange, 1), mp("o2", Orange, 2), mp("o3", Orange, 3)}}
	snapshot := ClonePosition(before)
	got := Search(context.Background(), before, Rules{Edition: 4}, Turn{}, SearchOptions{MaxNodes: 1000000})
	if got.Status != Optimal || got.Verified == nil || len(got.Verified.Played) != 3 {
		t.Fatalf("got %+v", got)
	}
	if !ValidateMove(before, Rules{Edition: 4}, Turn{}, *got.Move).Valid {
		t.Fatal("unverified result")
	}
	if !reflect.DeepEqual(before, snapshot) {
		t.Fatal("search mutated input")
	}
	// Joker rack substitution and high-scoring opening.
	rack := []Piece{mp("r", Red, 10), mp("b", Blue, 10), {ID: "j", Tile: Tile{Kind: Joker}}}
	got = Search(context.Background(), Position{Rack: rack}, DefaultRules(4), Turn{}, SearchOptions{})
	if got.Status != Optimal || got.Verified.OpeningScore != 30 {
		t.Fatalf("joker opening %+v", got)
	}
}

func TestSearchLimitsAndNoMove(t *testing.T) {
	before := Position{Rack: []Piece{mp("r", Red, 1)}}
	got := Search(context.Background(), before, Rules{Edition: 4}, Turn{}, SearchOptions{MaxNodes: 1})
	if got.Status != Incomplete || got.Reason != "node_limit" {
		t.Fatalf("limit mislabeled %+v", got)
	}
	got = Search(context.Background(), before, Rules{Edition: 4}, Turn{}, SearchOptions{})
	if got.Status != NoMove {
		t.Fatalf("complete negative %+v", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got = Search(ctx, before, Rules{Edition: 4}, Turn{}, SearchOptions{})
	if got.Status != Incomplete {
		t.Fatalf("cancellation %+v", got)
	}
	before = Position{Table: [][]Piece{{mp("r1", Red, 1), mp("r2", Red, 2), mj("j", Red, 3)}}, Rack: []Piece{mp("b", Blue, 13)}}
	got = Search(context.Background(), before, Rules{Edition: 4}, Turn{}, SearchOptions{})
	if got.Status != Incomplete {
		t.Fatalf("unsupported joker coverage claimed proof %+v", got)
	}
}

// Independent tiny oracle: enumerate every rack subset and every set partition;
// decide melds by map/sort reference, not production template enumeration.
func tinyBest(table, rack []Tile) int {
	all := append(append([]Tile{}, table...), rack...)
	var partition func([]Tile) bool
	partition = func(left []Tile) bool {
		if len(left) == 0 {
			return true
		}
		for mask := 1; mask < (1 << len(left)); mask += 2 {
			meld, rest := []Tile{}, []Tile{}
			for i, p := range left {
				if mask&(1<<i) != 0 {
					meld = append(meld, p)
				} else {
					rest = append(rest, p)
				}
			}
			if referenceMeld(meld) != "" && partition(rest) {
				return true
			}
		}
		return false
	}
	best := 0
	for mask := 0; mask < (1 << len(rack)); mask++ {
		selected := append([]Tile{}, all[:len(table)]...)
		n := 0
		for i, p := range rack {
			if mask&(1<<i) != 0 {
				selected = append(selected, p)
				n++
			}
		}
		if n > best && partition(selected) {
			best = n
		}
	}
	return best
}

func TestSearchAgainstIndependentPartitions(t *testing.T) {
	universe := []Tile{numbered(Blue, 1), numbered(Blue, 2), numbered(Blue, 3), numbered(Red, 1), numbered(Red, 2), numbered(Red, 3), numbered(Orange, 1), numbered(Orange, 2), numbered(Orange, 3)}
	for mask := 0; mask < (1 << len(universe)); mask++ {
		rack, tiles := []Piece{}, []Tile{}
		for i, tile := range universe {
			if mask&(1<<i) != 0 {
				rack = append(rack, Piece{ID: fmt.Sprint(i), Tile: tile})
				tiles = append(tiles, tile)
			}
		}
		got := Search(context.Background(), Position{Rack: rack}, Rules{Edition: 4}, Turn{}, SearchOptions{MaxNodes: 1000000})
		want := tinyBest(nil, tiles)
		actual := 0
		if got.Verified != nil {
			actual = len(got.Verified.Played)
		}
		if actual != want || (got.Status != Optimal && got.Status != NoMove) {
			t.Fatalf("mask %d: got %+v, want %d", mask, got, want)
		}
	}
}

func BenchmarkSearchRearrangement(b *testing.B) {
	before := Position{Table: [][]Piece{blueRun("b", 1, 3)}, Rack: []Piece{mp("x", Blue, 4), mp("r", Red, 7), mp("o", Orange, 7), mp("k", Black, 7)}}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Search(context.Background(), before, Rules{Edition: 4}, Turn{}, SearchOptions{})
	}
}

func TestSearchWithMandatoryTableReference(t *testing.T) {
	universe := []Tile{numbered(Red, 1), numbered(Red, 2), numbered(Red, 3), numbered(Orange, 1), numbered(Orange, 2), numbered(Orange, 3)}
	table := blueRun("b", 1, 3)
	tableValues := []Tile{numbered(Blue, 1), numbered(Blue, 2), numbered(Blue, 3)}
	for mask := 0; mask < 64; mask++ {
		rack, values := []Piece{}, []Tile{}
		for i, tile := range universe {
			if mask&(1<<i) != 0 {
				rack = append(rack, Piece{ID: fmt.Sprint(i), Tile: tile})
				values = append(values, tile)
			}
		}
		before := Position{Table: [][]Piece{table}, Rack: rack}
		got := Search(context.Background(), before, Rules{Edition: 4}, Turn{}, SearchOptions{MaxNodes: 1000000})
		want := tinyBest(tableValues, values)
		actual := 0
		if got.Verified != nil {
			actual = len(got.Verified.Played)
		}
		if actual != want || (got.Status != Optimal && got.Status != NoMove) {
			t.Fatalf("mask %d: got %+v, want %d", mask, got, want)
		}
	}
}

func TestSearchFindsJokerRelease(t *testing.T) {
	before := Position{Table: [][]Piece{{mp("t1", Blue, 1), mp("t2", Blue, 2), mj("j", Blue, 3)}}, Rack: []Piece{mp("r3", Blue, 3), mp("r7", Red, 7), mp("o7", Orange, 7)}}
	got := Search(context.Background(), before, Rules{Edition: 4}, Turn{}, SearchOptions{MaxNodes: 2000000})
	if got.Status != Optimal || len(got.Verified.Played) != 3 || !ValidateMove(before, Rules{Edition: 4}, Turn{}, *got.Move).Valid {
		t.Fatalf("release search %+v", got)
	}
}

func TestReleaseCannotReturnToReorderedOriginalSet(t *testing.T) {
	original := []Piece{mp("red", Red, 7), mp("blue", Blue, 7), mj("j", Black, 7)}
	rack := []Piece{mp("black7", Black, 7), mp("black8", Black, 8), mp("black9", Black, 9)}
	before := Position{Table: [][]Piece{original}, Rack: rack}
	final := [][]Piece{{original[1], original[0], original[2]}, rack}
	move := Move{Table: final, Releases: []JokerRelease{{Table: [][]Piece{{original[0], original[1], rack[0]}}, Freed: []string{"j"}}}}
	if got := ValidateMove(before, Rules{Edition: 4}, Turn{}, move); got.Valid || got.Code != "joker_reuse" {
		t.Fatalf("reordering original set must not prove new-set reuse: %+v", got)
	}
}
