package game

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestNumberedSearchDuplicateOptima(t *testing.T) {
	for _, tc := range []struct {
		name   string
		before Position
		rules  Rules
		want   int
	}{
		{
			name: "duplicate rack copies with a partial optimum",
			before: Position{Table: [][]Piece{blueRun("table-", 1, 3)}, Rack: []Piece{
				mp("rack-b1", Blue, 1), mp("rack-b2", Blue, 2),
				mp("rack-r1", Red, 1), mp("rack-o1", Orange, 1), mp("rack-k1", Black, 1),
			}},
			rules: Rules{Edition: FourPlayer},
			want:  4,
		},
		{
			name:   "an extra physical copy cannot join its run",
			before: Position{Table: [][]Piece{blueRun("table-", 1, 3)}, Rack: []Piece{mp("rack-b2", Blue, 2)}},
			rules:  Rules{Edition: FourPlayer},
			want:   0,
		},
		{
			name: "six-player third copies with an unplayable tile",
			before: Position{
				Table: [][]Piece{blueRun("first-", 1, 3), blueRun("second-", 1, 3)},
				Rack:  append(blueRun("rack-", 1, 3), mp("rack-r13", Red, 13)),
			},
			rules: Rules{Edition: SixPlayer},
			want:  3,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := ClonePosition(tc.before)
			table, rack := numberedSearchValues(tc.before)
			if want := tinyBest(table, rack); want != tc.want {
				t.Fatalf("independent oracle got %d, want fixture optimum %d", want, tc.want)
			}
			got := Search(context.Background(), tc.before, tc.rules, Turn{}, SearchOptions{MaxNodes: 200000})
			checkNumberedSearchOptimum(t, tc.before, tc.rules, got, tc.want)
			if !reflect.DeepEqual(tc.before, snapshot) {
				t.Fatal("search mutated the original table or rack")
			}
		})
	}
}

func TestNumberedSearchDuplicateRackSubsetsAgainstOracle(t *testing.T) {
	// Each numbered value has a separate physical table and rack occurrence.
	// Exercise both ends of the number range, including runs ending at 13 and
	// groups of four, using the independent map/sort partition oracle.
	for _, start := range []int{1, 11} {
		table := blueRun("table-", start, start+2)
		universe := []Tile{
			numbered(Blue, start), numbered(Blue, start+1), numbered(Blue, start+2),
			numbered(Red, start+2), numbered(Orange, start+2), numbered(Black, start+2),
		}
		for mask := 1; mask < 1<<len(universe); mask++ {
			before := Position{Table: [][]Piece{table}}
			for i, tile := range universe {
				if mask&(1<<i) != 0 {
					before.Rack = append(before.Rack, Piece{ID: fmt.Sprintf("rack-%d", i), Tile: tile})
				}
			}
			t.Run(fmt.Sprintf("start=%d/mask=%02x", start, mask), func(t *testing.T) {
				snapshot := ClonePosition(before)
				tableValues, rackValues := numberedSearchValues(before)
				want := tinyBest(tableValues, rackValues)
				rules := Rules{Edition: FourPlayer}
				got := Search(context.Background(), before, rules, Turn{}, SearchOptions{MaxNodes: 200000})
				checkNumberedSearchOptimum(t, before, rules, got, want)
				if !reflect.DeepEqual(before, snapshot) {
					t.Fatal("search mutated the original table or rack")
				}
			})
		}
	}
}

func TestNumberedSearchCancellationAndLimit(t *testing.T) {
	before := Position{Table: [][]Piece{blueRun("table-", 1, 3)}, Rack: []Piece{mp("rack-b4", Blue, 4)}}
	snapshot := ClonePosition(before)
	rules := Rules{Edition: FourPlayer}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, tc := range []struct {
		name    string
		ctx     context.Context
		options SearchOptions
		reason  string
	}{
		{"pre-canceled", ctx, SearchOptions{}, context.Canceled.Error()},
		{"one work unit", context.Background(), SearchOptions{MaxNodes: 1}, "node_limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := Search(tc.ctx, before, rules, Turn{}, tc.options)
			if got.Status != Incomplete || got.Reason != tc.reason || got.Move != nil || got.Verified != nil || got.UpperBound != nil {
				t.Fatalf("got %+v; want incomplete without a move or proved bound, reason %q", got, tc.reason)
			}
			if got.Nodes > tc.options.MaxNodes {
				t.Fatalf("got %d work units, limit %d", got.Nodes, tc.options.MaxNodes)
			}
			if !reflect.DeepEqual(before, snapshot) {
				t.Fatal("interrupted search mutated the original table or rack")
			}
		})
	}
}

func TestNumberedSearchNodeLimitKeepsVerifiedBest(t *testing.T) {
	before := Position{Table: [][]Piece{blueRun("table-", 1, 3)}, Rack: []Piece{
		mp("rack-b1", Blue, 1), mp("rack-b2", Blue, 2),
		mp("rack-r1", Red, 1), mp("rack-o1", Orange, 1), mp("rack-k1", Black, 1),
	}}
	snapshot := ClonePosition(before)
	rules := Rules{Edition: FourPlayer}
	complete := Search(context.Background(), before, rules, Turn{}, SearchOptions{MaxNodes: 200000})
	table, rack := numberedSearchValues(before)
	want := tinyBest(table, rack)
	checkNumberedSearchOptimum(t, before, rules, complete, want)

	// Locate the first budget that retains a witness. This checks an interrupted
	// prefix without depending on the traversal's exact number of work units.
	low, high := 1, complete.Nodes
	for low < high {
		limit := low + (high-low)/2
		got := Search(context.Background(), before, rules, Turn{}, SearchOptions{MaxNodes: limit})
		if got.Move != nil {
			high = limit
		} else {
			low = limit + 1
		}
	}
	got := Search(context.Background(), before, rules, Turn{}, SearchOptions{MaxNodes: low})
	if got.Status != FeasibleIncomplete || got.Reason != "node_limit" || got.UpperBound != nil || got.Nodes > low {
		t.Fatalf("budget %d of complete %d: got %+v; want a retained feasible move with an unproved optimum", low, complete.Nodes, got)
	}
	played := checkNumberedSearchMove(t, before, rules, got)
	if played == 0 || played > want {
		t.Fatalf("retained move plays %d rack tiles; independent optimum is %d", played, want)
	}
	if !reflect.DeepEqual(before, snapshot) {
		t.Fatal("bounded searches mutated the original table or rack")
	}
}

func numberedSearchValues(before Position) (table, rack []Tile) {
	for _, meld := range before.Table {
		for _, piece := range meld {
			table = append(table, piece.Tile)
		}
	}
	for _, piece := range before.Rack {
		rack = append(rack, piece.Tile)
	}
	return table, rack
}

func checkNumberedSearchOptimum(t *testing.T, before Position, rules Rules, got SearchResult, want int) {
	t.Helper()
	wantStatus := Optimal
	if want == 0 {
		wantStatus = NoMove
	}
	if got.Status != wantStatus || got.UpperBound == nil || *got.UpperBound != want {
		t.Fatalf("got %+v; want status %s and independent upper bound %d", got, wantStatus, want)
	}
	if want == 0 {
		if got.Move != nil || got.Verified != nil {
			t.Fatalf("no-move result included a witness: %+v", got)
		}
		return
	}
	if played := checkNumberedSearchMove(t, before, rules, got); played != want {
		t.Fatalf("returned move plays %d rack tiles; independent optimum is %d", played, want)
	}
}

func checkNumberedSearchMove(t *testing.T, before Position, rules Rules, got SearchResult) int {
	t.Helper()
	if got.Move == nil || got.Verified == nil {
		t.Fatalf("missing move or verification: %+v", got)
	}
	checked := ValidateMove(before, rules, Turn{}, *got.Move)
	if !checked.Valid || !reflect.DeepEqual(checked.Played, got.Verified.Played) {
		t.Fatalf("independent move validation got %+v; search verification %+v", checked, got.Verified)
	}
	original, mandatory := map[string]Tile{}, map[string]bool{}
	for _, meld := range before.Table {
		for _, piece := range meld {
			original[piece.ID], mandatory[piece.ID] = piece.Tile, true
		}
	}
	for _, piece := range before.Rack {
		original[piece.ID] = piece.Tile
	}
	seen := map[string]bool{}
	for _, meld := range got.Move.Table {
		for _, piece := range meld {
			tile, ok := original[piece.ID]
			if !ok || seen[piece.ID] || tile != piece.Tile {
				t.Fatalf("unexpected, repeated, or changed physical piece: %+v", piece)
			}
			seen[piece.ID] = true
		}
	}
	for id := range mandatory {
		if !seen[id] {
			t.Fatalf("original table piece %q was omitted", id)
		}
	}
	return len(checked.Played)
}
