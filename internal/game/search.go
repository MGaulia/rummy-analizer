package game

import (
	"context"
	"fmt"
)

type SearchStatus string

const (
	Optimal            SearchStatus = "optimal"
	FeasibleIncomplete SearchStatus = "feasible_incomplete"
	NoMove             SearchStatus = "no_move"
	Incomplete         SearchStatus = "incomplete"
	InvalidInput       SearchStatus = "invalid_input"
)

type SearchOptions struct {
	// Includes candidate construction and partition exploration, not just leaves.
	MaxNodes int `json:"max_nodes"`
}

type SearchResult struct {
	Status   SearchStatus `json:"status"`
	Reason   string       `json:"reason,omitempty"`
	Nodes    int          `json:"nodes"`
	Move     *Move        `json:"move,omitempty"`
	Verified *MoveReport  `json:"verified,omitempty"`
}

type searchBudget struct {
	ctx        context.Context
	max, nodes int
	stopped    string
}

func (b *searchBudget) tick() bool {
	if b.stopped != "" {
		return false
	}
	if err := b.ctx.Err(); err != nil {
		b.stopped = err.Error()
		return false
	}
	if b.nodes >= b.max {
		b.stopped = "node_limit"
		return false
	}
	b.nodes++
	return true
}

// Search enumerates meld templates and exact physical partitions. Numbered
// table tiles are mandatory; rack pieces may be omitted. There is no limit on
// table rearrangement. Jokers fill any numbered template slot, retaining ID.
// Every candidate passes the separate transition validator before acceptance.
// Table-joker certificates are sufficient but not complete for the booklet;
// unless the rack is emptied, this is explicitly an incomplete search scope.
func Search(ctx context.Context, before Position, rules Rules, turn Turn, options SearchOptions) SearchResult {
	report := ValidatePosition(before, rules.Edition)
	if !report.Valid {
		return SearchResult{Status: InvalidInput, Reason: fmt.Sprint(report.Issues)}
	}
	if options.MaxNodes < 0 {
		return SearchResult{Status: InvalidInput, Reason: "max_nodes must not be negative"}
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 200000
	}
	budget := &searchBudget{ctx: ctx, max: options.MaxNodes}
	result := SearchResult{Status: NoMove}
	if turn.Drawn {
		result.Reason = "drawing ended this turn"
		return result
	}
	if len(before.Rack) == 0 {
		result.Reason = "rack is empty"
		return result
	}
	before = ClonePosition(before)
	before.Table = report.Resolved
	opening := rules.NeedsOpening(turn)
	pool := []Piece{}
	required := 0
	tableJokers := []string{}
	fixed := [][]Piece{}
	if opening {
		fixed = before.Table
	} else {
		for _, meld := range before.Table {
			for _, p := range meld {
				pool = append(pool, p)
				if p.Tile.Kind == Joker {
					tableJokers = append(tableJokers, p.ID)
				}
			}
		}
		required = len(pool)
	}
	pool = append(pool, before.Rack...)
	best := 0
	// Cache release certificates by played rack subset. They do not depend on
	// final arrangement; ValidateMove still checks each candidate's actual reuse.
	certificates := map[string][]JokerRelease{}
	certTried := map[string]bool{}
	walkPartitions(pool, required, budget, func(table [][]Piece) bool {
		move := Move{Table: append(append([][]Piece{}, fixed...), table...)}
		checked := ValidateMove(before, rules, turn, move)
		if !checked.Valid && checked.Code == "joker_witness_required" {
			present := map[string]bool{}
			for _, meld := range table {
				for _, p := range meld {
					present[p.ID] = true
				}
			}
			freed := []string{}
			ignored := map[string]bool{}
			for _, id := range tableJokers {
				ignored[id] = true
			}
			for _, id := range tableJokers {
				ignored[id] = false
				if !retainedJokers(before.Table, table, ignored) {
					freed = append(freed, id)
				}
				ignored[id] = true
			}
			key := ""
			for _, p := range before.Rack {
				if present[p.ID] {
					key += "1"
				} else {
					key += "0"
				}
			}
			for _, id := range freed {
				key += "/" + id
			}
			if !certTried[key] {
				certTried[key] = true
				// Free the changed table jokers by a valid table restructuring,
				// then reuse them in the proposed final sets.
				stagePool := []Piece{}
				for _, meld := range before.Table {
					for _, p := range meld {
						remove := false
						for _, id := range freed {
							remove = remove || p.ID == id
						}
						if !remove {
							stagePool = append(stagePool, p)
						}
					}
				}
				n := len(stagePool)
				for _, p := range before.Rack {
					if present[p.ID] {
						stagePool = append(stagePool, p)
					}
				}
				walkPartitions(stagePool, n, budget, func(stage [][]Piece) bool {
					trial := Move{Table: move.Table, Releases: []JokerRelease{{Table: stage, Freed: freed}}}
					if !ValidateMove(before, rules, turn, trial).Valid {
						return false
					}
					certificates[key] = []JokerRelease{{Table: ClonePosition(Position{Table: stage}).Table, Freed: append([]string(nil), freed...)}}
					return true
				})
			}
			move.Releases = certificates[key]
			checked = ValidateMove(before, rules, turn, move)
		}
		if checked.Valid && len(checked.Played) > best {
			best = len(checked.Played)
			copyMove := move
			copyMove.Table = ClonePosition(Position{Table: move.Table}).Table
			result.Move = &copyMove
			result.Verified = &checked
		}
		return best == len(before.Rack)
	})
	result.Nodes = budget.nodes
	if best == len(before.Rack) {
		result.Status = Optimal
		result.Reason = "all rack tiles played; primary upper bound reached"
		return result
	}
	reason := budget.stopped
	if len(tableJokers) > 0 {
		if reason != "" {
			reason += "; "
		}
		reason += "table_joker_transition_coverage_incomplete"
	}
	if reason != "" {
		result.Reason = reason
		result.Status = Incomplete
		if result.Move != nil {
			result.Status = FeasibleIncomplete
		}
	} else if result.Move != nil {
		result.Status = Optimal
		result.Reason = "complete partition search"
	}
	return result
}

// Enumerate all partitions covering the required prefix and any subset of the
// remaining pieces. Anchor-first removes meld-order permutations. Every member
// of a partition is generated from the finite complete group/run templates.
func walkPartitions(pool []Piece, required int, budget *searchBudget, visit func([][]Piece) bool) bool {
	used := make([]bool, len(pool))
	var path [][]Piece
	var walk func(int) bool
	walk = func(start int) bool {
		if !budget.tick() {
			return false
		}
		anchor := start
		for anchor < len(pool) && used[anchor] {
			anchor++
		}
		if anchor >= required && visit(path) {
			return true
		}
		if anchor == len(pool) {
			return false
		}
		if enumerateMelds(pool, used, anchor, budget, func(meld []Piece, indices []int) bool {
			path = append(path, meld)
			stop := walk(anchor + 1)
			path = path[:len(path)-1]
			return stop
		}) {
			return true
		}
		if anchor >= required && budget.stopped == "" {
			used[anchor] = true
			stop := walk(anchor + 1)
			used[anchor] = false
			return stop
		}
		return false
	}
	return walk(0)
}

func enumerateMelds(pool []Piece, used []bool, anchor int, budget *searchBudget, visit func([]Piece, []int) bool) bool {
	colors := []Color{Black, Blue, Orange, Red}
	apply := func(roles []Tile) bool {
		if !budget.tick() {
			return false
		}
		for anchorSlot, role := range roles {
			p := pool[anchor]
			if p.Tile.Kind != Joker && p.Tile != role {
				continue
			}
			indices := make([]int, len(roles))
			meld := make([]Piece, len(roles))
			p.As = nil
			if p.Tile.Kind == Joker {
				r := role
				p.As = &r
			}
			indices[anchorSlot], meld[anchorSlot] = anchor, p
			used[anchor] = true
			var fill func(int) bool
			fill = func(slot int) bool {
				if !budget.tick() {
					return false
				}
				if slot == len(roles) {
					return visit(meld, indices)
				}
				if slot == anchorSlot {
					return fill(slot + 1)
				}
				for i, q := range pool {
					if used[i] || (q.Tile.Kind != Joker && q.Tile != roles[slot]) {
						continue
					}
					q.As = nil
					if q.Tile.Kind == Joker {
						r := roles[slot]
						q.As = &r
					}
					used[i] = true
					indices[slot], meld[slot] = i, q
					stop := fill(slot + 1)
					used[i] = false
					if stop {
						return true
					}
				}
				return false
			}
			stop := fill(0)
			used[anchor] = false
			if stop {
				return true
			}
			if budget.stopped != "" {
				return false
			}
		}
		return false
	}
	for number := 1; number <= 13; number++ {
		for mask := 1; mask < 16; mask++ {
			roles := []Tile{}
			for i, c := range colors {
				if mask&(1<<i) != 0 {
					roles = append(roles, Tile{Color: c, Number: number})
				}
			}
			if len(roles) >= 3 && apply(roles) {
				return true
			}
			if budget.stopped != "" {
				return false
			}
		}
	}
	for _, c := range colors {
		for start := 1; start <= 11; start++ {
			for end := start + 2; end <= 13; end++ {
				roles := []Tile{}
				for n := start; n <= end; n++ {
					roles = append(roles, Tile{Color: c, Number: n})
				}
				if apply(roles) {
					return true
				}
				if budget.stopped != "" {
					return false
				}
			}
		}
	}
	return false
}
