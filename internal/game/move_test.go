package game

import (
	"context"
	"encoding/json"
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
	if got.Status != NoMove || got.UpperBound == nil || *got.UpperBound != 0 {
		t.Fatalf("expected proven empty relaxed partition bound: %+v", got)
	}
}

// Independent tiny oracle: enumerate every rack subset and every set partition;
// decide melds by map/sort reference, not production template enumeration.
func tinyBest(table, rack []Tile) int {
	return tinyBestWithMeldCheck(table, rack, func(tiles []Tile) bool { return referenceMeld(tiles) != "" })
}
func tinyBestWithMeldCheck(table, rack []Tile, valid func([]Tile) bool) int {
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
			if valid(meld) && partition(rest) {
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

// Booklet D-2600-1236-0041 (120819), visually checked pp. 3-4.
// Its yellow illustration tiles use this repository's canonical orange label.
func TestOfficialFourJokerReleases(t *testing.T) {
	cases := []struct {
		name        string
		table       [][]Piece
		rack        []Piece
		replacement [][]Piece
		role        Color
		number      int
	}{
		{"1 either missing group color", [][]Piece{{mp("r3", Red, 3), mp("b3", Blue, 3), mj("j", Black, 3)}}, []Piece{mp("o3", Orange, 3)}, [][]Piece{{mp("r3", Red, 3), mp("b3", Blue, 3), mp("o3", Orange, 3)}}, Black, 10},
		{"1 other missing group color", [][]Piece{{mp("r3", Red, 3), mp("b3", Blue, 3), mj("j", Black, 3)}}, []Piece{mp("k3", Black, 3)}, [][]Piece{{mp("r3", Red, 3), mp("b3", Blue, 3), mp("k3", Black, 3)}}, Black, 10},
		{"1 both missing group colors", [][]Piece{{mp("r3", Red, 3), mp("b3", Blue, 3), mj("j", Black, 3)}}, []Piece{mp("o3", Orange, 3), mp("k3", Black, 3)}, [][]Piece{{mp("r3", Red, 3), mp("b3", Blue, 3), mp("o3", Orange, 3), mp("k3", Black, 3)}}, Black, 10},
		{"2 split around joker", [][]Piece{{mp("r2", Red, 2), mp("r3", Red, 3), mj("j", Red, 4), mp("r5", Red, 5), mp("r6", Red, 6)}}, []Piece{mp("r1", Red, 1), mp("r7", Red, 7)}, [][]Piece{{mp("r1", Red, 1), mp("r2", Red, 2), mp("r3", Red, 3)}, {mp("r5", Red, 5), mp("r6", Red, 6), mp("r7", Red, 7)}}, Black, 10},
		{"3 shift endpoint", [][]Piece{{mp("b6", Blue, 6), mp("b7", Blue, 7), mj("j", Blue, 8)}}, []Piece{mp("b5", Blue, 5)}, [][]Piece{{mp("b5", Blue, 5), mp("b6", Blue, 6), mp("b7", Blue, 7)}}, Black, 10},
		{"4 distribute source into table groups", [][]Piece{{mp("k1", Black, 1), mp("k2", Black, 2), mj("j", Black, 3)}, {mp("b1", Blue, 1), mp("o1", Orange, 1), mp("r1", Red, 1)}, {mp("b2", Blue, 2), mp("o2", Orange, 2), mp("r2", Red, 2)}}, nil, [][]Piece{{mp("k1", Black, 1), mp("b1", Blue, 1), mp("o1", Orange, 1), mp("r1", Red, 1)}, {mp("k2", Black, 2), mp("b2", Blue, 2), mp("o2", Orange, 2), mp("r2", Red, 2)}}, Black, 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The booklet shows only the release. Complete each example by reusing that
			// physical joker in a fresh 10-group on the same turn, with two rack tiles.
			rack := append(clonePieces(tc.rack), mp("new-blue", Blue, 10), mp("new-red", Red, 10))
			before := Position{Table: tc.table, Rack: rack}
			final := append(ClonePosition(Position{Table: tc.replacement}).Table, []Piece{mp("new-blue", Blue, 10), mp("new-red", Red, 10), mj("j", tc.role, tc.number)})
			move := Move{Table: final, Releases: []JokerRelease{{Table: tc.replacement, Freed: []string{"j"}}}}
			for _, edition := range []Edition{FourPlayer, SixPlayer} {
				got := ValidateMove(before, Rules{Edition: edition}, Turn{}, move)
				if !got.Valid || len(got.Played) != len(rack) {
					t.Fatalf("edition %d: %+v", edition, got)
				}
				if ValidateMove(before, DefaultRules(edition), Turn{}, move).Valid {
					t.Fatal("unopened player retrieved a table joker")
				}
			}
		})
	}
}

func TestJokerSearchCanProveSimpleOptimumAndAbsence(t *testing.T) {
	table := [][]Piece{{mp("b1", Blue, 1), mp("b2", Blue, 2), mj("j", Blue, 3)}}
	for _, tc := range []struct {
		name   string
		rack   []Piece
		status SearchStatus
		played int
	}{
		{"no final partition can add rack tile", []Piece{mp("r13", Red, 13)}, NoMove, 0},
		{"one playable, one impossible", []Piece{mp("b4", Blue, 4), mp("r13", Red, 13)}, Optimal, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := Position{Table: table, Rack: tc.rack}
			got := Search(context.Background(), before, Rules{Edition: 4}, Turn{}, SearchOptions{MaxNodes: 2000000})
			if got.Status != tc.status {
				t.Fatalf("got %+v want %s", got, tc.status)
			}
			if tc.played > 0 && (got.Verified == nil || len(got.Verified.Played) != tc.played) {
				t.Fatalf("wrong result %+v", got)
			}
		})
	}
}

// Wildcard substitution is independent of production template enumeration.
func referenceWildcardMeld(tiles []Tile) bool {
	if len(tiles) < 3 || len(tiles) > 13 {
		return false
	}
	for i, tile := range tiles {
		if tile.Kind != Joker {
			continue
		}
		copyTiles := append([]Tile(nil), tiles...)
		for _, color := range []Color{Black, Blue, Orange, Red} {
			for n := 1; n <= 13; n++ {
				copyTiles[i] = numbered(color, n)
				if referenceWildcardMeld(copyTiles) {
					return true
				}
			}
		}
		return false
	}
	return referenceMeld(tiles) != ""
}

func TestJokerPartitionBoundAgainstIndependentOracle(t *testing.T) {
	table := []Piece{mp("t1", Blue, 1), mp("t2", Blue, 2), mj("j", Blue, 3)}
	originalValues := []Tile{numbered(Blue, 1), numbered(Blue, 2), {Kind: Joker}}
	universe := []Tile{numbered(Blue, 4), numbered(Blue, 5), numbered(Red, 1), numbered(Red, 2), numbered(Red, 3)}
	for mask := 0; mask < 32; mask++ {
		rack, values := []Piece{}, []Tile{}
		for i, tile := range universe {
			if mask&(1<<i) != 0 {
				rack = append(rack, Piece{ID: fmt.Sprint(i), Tile: tile})
				values = append(values, tile)
			}
		}
		if len(rack) == 0 {
			continue
		}
		want := tinyBestWithMeldCheck(originalValues, values, referenceWildcardMeld)
		before := Position{Table: [][]Piece{table}, Rack: rack}
		got := Search(context.Background(), before, Rules{Edition: 4}, Turn{}, SearchOptions{MaxNodes: 2000000})
		if got.UpperBound == nil || *got.UpperBound != want {
			t.Fatalf("mask %d: got %+v, oracle bound %d", mask, got, want)
		}
		if got.Status == Optimal && (got.Verified == nil || len(got.Verified.Played) != want) {
			t.Fatalf("false optimum: %+v", got)
		}
		if got.Status == NoMove && want != 0 {
			t.Fatalf("false absence: %+v", got)
		}
		if got.Move != nil && !ValidateMove(before, Rules{Edition: 4}, Turn{}, *got.Move).Valid {
			t.Fatal("unverified move")
		}
	}
}

func TestUnboundOriginalJokerRetainsAnyConsistentRole(t *testing.T) {
	before := Position{Table: [][]Piece{{mp("r3", Red, 3), mp("b3", Blue, 3), {ID: "j", Tile: Tile{Kind: Joker}}}}, Rack: []Piece{mp("k3", Black, 3)}}
	final := [][]Piece{{before.Table[0][0], before.Table[0][1], mj("j", Orange, 3), before.Rack[0]}}
	got := ValidateMove(before, Rules{Edition: 4}, Turn{}, Move{Table: final})
	if !got.Valid {
		t.Fatalf("initially unbound joker may already represent orange 3: %+v", got)
	}
	fixed := numbered(Black, 3)
	before.Table[0][2].As = &fixed
	if ValidateMove(before, Rules{Edition: 4}, Turn{}, Move{Table: final}).Valid {
		t.Fatal("changed an explicit joker role without release")
	}
}

func TestChainedJokerReleaseReuse(t *testing.T) {
	before := Position{Table: [][]Piece{{mp("r1", Red, 1), mp("b1", Blue, 1), mj("a", Black, 1)}, {mp("b2", Blue, 2), mp("b3", Blue, 3), mj("b", Blue, 4)}}, Rack: []Piece{mp("o1", Orange, 1), mp("r7", Red, 7), mp("o7", Orange, 7)}}
	// Clear A with orange 1, then use A to clear B, then reuse B as black 7.
	first := [][]Piece{{mp("r1", Red, 1), mp("b1", Blue, 1), mp("o1", Orange, 1)}, before.Table[1]}
	second := [][]Piece{first[0], {mp("b2", Blue, 2), mp("b3", Blue, 3), mj("a", Blue, 4)}}
	final := append(ClonePosition(Position{Table: second}).Table, []Piece{mp("r7", Red, 7), mp("o7", Orange, 7), mj("b", Black, 7)})
	// Decode new witness metadata through JSON so this regression also compiles
	// against the preceding implementation, which did not support reuse stages.
	raw := fmt.Sprintf(`{"table":%s,"releases":[{"table":%s,"freed":["a"]},{"table":%s,"freed":["b"],"reused":["a"]}]}`, mustJSON(t, final), mustJSON(t, first), mustJSON(t, second))
	var move Move
	if err := json.Unmarshal([]byte(raw), &move); err != nil {
		t.Fatal(err)
	}
	got := ValidateMove(before, Rules{Edition: 4}, Turn{}, move)
	if !got.Valid || len(got.Played) != 3 {
		t.Fatalf("valid chained release/reuse rejected: %+v", got)
	}

	witness, limited := findJokerStages(before, Rules{Edition: 4}, Turn{}, final, &searchBudget{ctx: context.Background(), max: 3000000})
	if limited || len(witness) < 2 || !ValidateMove(before, Rules{Edition: 4}, Turn{}, Move{Table: final, Releases: witness}).Valid {
		t.Fatalf("chained witness finder: limited=%v witness=%+v", limited, witness)
	}
	gotSearch := Search(context.Background(), before, Rules{Edition: 4}, Turn{}, SearchOptions{MaxNodes: 3000000})
	if gotSearch.Status != Optimal || gotSearch.Verified == nil || len(gotSearch.Verified.Played) != 3 {
		t.Fatalf("chained release search: %+v", gotSearch)
	}
}
func mustJSON(t *testing.T, value interface{}) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestReleaseOneOfTwoJokersKeepsOtherBinding(t *testing.T) {
	source := []Piece{mp("r3", Red, 3), mj("a", Black, 3), mj("b", Blue, 3)}
	before := Position{Table: [][]Piece{source}, Rack: []Piece{mp("k3", Black, 3), mp("r7", Red, 7), mp("o7", Orange, 7)}}
	stage := [][]Piece{{source[0], before.Rack[0], source[2]}}
	final := append(ClonePosition(Position{Table: stage}).Table, []Piece{before.Rack[1], before.Rack[2], mj("a", Blue, 7)})
	got := ValidateMove(before, Rules{Edition: 6}, Turn{}, Move{Table: final, Releases: []JokerRelease{{Table: stage, Freed: []string{"a"}}}})
	if !got.Valid {
		t.Fatalf("retained joker should not depend on released joker's new binding: %+v", got)
	}
}

func TestJokerStageRejectsInvalidReuseAndAccountsOccurrences(t *testing.T) {
	before := Position{Table: [][]Piece{{mp("a1", Blue, 1), mp("a2", Blue, 2), mj("j", Blue, 3)}}, Rack: []Piece{mp("r3", Blue, 3), mp("r7", Red, 7), mp("o7", Orange, 7)}}
	stage := [][]Piece{{before.Table[0][0], before.Table[0][1], before.Rack[0]}}
	final := append(ClonePosition(Position{Table: stage}).Table, []Piece{before.Rack[1], before.Rack[2], mj("j", Black, 7)})
	for _, stages := range [][]JokerRelease{
		{{Table: stage, Freed: []string{"j"}, Reused: []string{"j"}}},
		{{Table: stage, Freed: []string{"j", "j"}}},
		{{Table: stage, Freed: []string{"r3"}}},
		{{Table: stage, Freed: []string{"j"}}, {Table: final, Reused: []string{"j", "j"}}},
		{{Table: stage, Freed: []string{"j"}}, {Table: final, Reused: []string{"unknown"}}},
	} {
		if got := ValidateMove(before, Rules{Edition: 4}, Turn{}, Move{Table: final, Releases: stages}); got.Valid {
			t.Fatalf("accepted invalid stage sequence %+v", stages)
		}
	}
	// A returned original table tile cannot hide in a release's unused rack.
	missing := ClonePosition(Position{Table: stage}).Table
	missing[0] = missing[0][1:]
	if ValidateMove(before, Rules{Edition: 4}, Turn{}, Move{Table: final, Releases: []JokerRelease{{Table: missing, Freed: []string{"j"}}}}).Valid {
		t.Fatal("lost table tile during release")
	}
}

func TestSixPlayerFourJokersWithUnknownInitialRoles(t *testing.T) {
	before := Position{Table: [][]Piece{{{ID: "a", Tile: Tile{Kind: Joker}}, {ID: "b", Tile: Tile{Kind: Joker}}, {ID: "c", Tile: Tile{Kind: Joker}}}}, Rack: []Piece{{ID: "d", Tile: Tile{Kind: Joker}}, mp("r13", Red, 13)}}
	got := Search(context.Background(), before, Rules{Edition: 6}, Turn{}, SearchOptions{MaxNodes: 2000000})
	if got.Status != Optimal || got.Verified == nil || len(got.Verified.Played) != 2 {
		t.Fatalf("four-joker expanded-edition move %+v", got)
	}
	if !ValidateMove(before, Rules{Edition: 6}, Turn{}, *got.Move).Valid {
		t.Fatal("unverified expanded-edition move")
	}
}

func TestRetrievedJokerNewSetMayUseOnlyTableTiles(t *testing.T) {
	source := []Piece{mp("b1", Blue, 1), mp("b2", Blue, 2), mj("j", Blue, 3)}
	first := []Piece{mp("r7a", Red, 7), mp("k7a", Black, 7), mp("o7a", Orange, 7), mp("b7a", Blue, 7)}
	second := []Piece{mp("r7b", Red, 7), mp("k7b", Black, 7), mp("o7b", Orange, 7), mp("b7b", Blue, 7)}
	before := Position{Table: [][]Piece{source, first, second}, Rack: []Piece{mp("r3", Blue, 3)}}
	stage := [][]Piece{{source[0], source[1], before.Rack[0]}, first, second}
	final := [][]Piece{stage[0], first[:3], second[1:], {first[3], second[0], mj("j", Orange, 7)}}
	move := Move{Table: final, Releases: []JokerRelease{{Table: stage, Freed: []string{"j"}}}}
	beforeJSON, moveJSON := mustJSON(t, before), mustJSON(t, move)
	checked := ValidateMove(before, Rules{Edition: 4}, Turn{}, move)
	if !checked.Valid || len(checked.Played) != 1 {
		t.Fatalf("rack requirement is per turn, not per joker's new set: %+v", checked)
	}
	if mustJSON(t, before) != beforeJSON || mustJSON(t, move) != moveJSON {
		t.Fatal("mutated original state or supplied certificate")
	}
}

func TestNodeLimitKeepsVerifiedBest(t *testing.T) {
	before := Position{Rack: []Piece{mp("r10", Red, 10), mp("b10", Blue, 10), mp("k10", Black, 10), mp("o2", Orange, 2), mp("r5", Red, 5), mp("k8", Black, 8)}}
	found := false
	for limit := 100; limit <= 5000; limit += 100 {
		got := Search(context.Background(), before, Rules{Edition: 4}, Turn{}, SearchOptions{MaxNodes: limit})
		if got.Status == FeasibleIncomplete {
			found = true
			if got.Nodes > limit || len(got.Verified.Played) != 3 || !ValidateMove(before, Rules{Edition: 4}, Turn{}, *got.Move).Valid {
				t.Fatalf("bad retained best %+v", got)
			}
			break
		}
	}
	if !found {
		t.Fatal("fixture did not exercise a limit after finding a feasible move")
	}
}
