package game

import (
	"encoding/json"
	"sort"
)

// The graph is a fallback after the inexpensive single-release certificate.
// Its memory is bounded independently of the shared work/cancellation budget.
const maxJokerWitnessStates = 512

type jokerWitnessNode struct {
	state  jokerReplay
	parent int
	step   JokerRelease
}

// findJokerStages explores valid-table checkpoints, with release/reuse and
// role-preserving preparations. Every edge is replayed by the same independent
// physical transition checker used by callers. It does not assume a legitimate
// final partition alone licenses changing a table joker's role.
func findJokerStages(before Position, rules Rules, turn Turn, final [][]Piece, budget *searchBudget) ([]JokerRelease, bool) {
	initial, code, _ := replayJokerStages(before, rules.Edition, nil)
	if code != "" {
		return nil, false
	}
	jokers := []string{}
	for _, meld := range before.Table {
		for _, p := range meld {
			if p.Tile.Kind == Joker {
				jokers = append(jokers, p.ID)
			}
		}
	}
	played := map[string]bool{}
	for _, meld := range final {
		for _, p := range meld {
			played[p.ID] = true
		}
	}
	nodes := []jokerWitnessNode{{state: initial, parent: -1}}
	seen := map[string]bool{jokerStateKey(initial): true}
	path := func(index int) []JokerRelease {
		reverse := []JokerRelease{}
		for nodes[index].parent >= 0 {
			reverse = append(reverse, nodes[index].step)
			index = nodes[index].parent
		}
		result := make([]JokerRelease, len(reverse))
		for i := range reverse {
			result[i] = reverse[len(reverse)-1-i]
		}
		return result
	}
	for index := 0; index < len(nodes); index++ {
		if !budget.tick() {
			return nil, false
		}
		current := nodes[index].state
		prefix := path(index)
		// Prefer freeing/replacing jokers before unconstrained preparation states.
		for order := 1; order <= (1 << len(jokers)); order++ {
			mask := order % (1 << len(jokers))
			held := map[string]bool{}
			freed, reused := []string{}, []string{}
			for j, id := range jokers {
				if mask&(1<<j) != 0 {
					held[id] = true
					if !current.Held[id] {
						freed = append(freed, id)
					}
				} else if current.Held[id] {
					reused = append(reused, id)
				}
			}
			pool := []Piece{}
			for _, meld := range before.Table {
				for _, p := range meld {
					if !held[p.ID] {
						pool = append(pool, p)
					}
				}
			}
			for _, p := range before.Rack {
				if current.UsedRack[p.ID] {
					pool = append(pool, p)
				}
			}
			required := len(pool)
			for _, p := range before.Rack {
				if played[p.ID] && !current.UsedRack[p.ID] {
					pool = append(pool, p)
				}
			}
			var solution []JokerRelease
			limited := false
			walkPartitions(pool, required, budget, func(table [][]Piece) bool {
				stages := append(append([]JokerRelease{}, prefix...), JokerRelease{Table: table, Freed: freed, Reused: reused})
				replay, code, _ := replayJokerStages(before, rules.Edition, stages)
				if code != "" {
					return false
				}
				key := jokerStateKey(replay)
				if seen[key] {
					return false
				}
				seen[key] = true
				// Freeze callback-owned slices before retaining or descending.
				step := JokerRelease{Table: replay.Table, Freed: append([]string{}, freed...), Reused: append([]string{}, reused...)}
				stages[len(stages)-1] = step
				if ValidateMove(before, rules, turn, Move{Table: final, Releases: stages}).Valid {
					solution = stages
					return true
				}
				if len(nodes) >= maxJokerWitnessStates {
					limited = true
					return true
				}
				nodes = append(nodes, jokerWitnessNode{state: replay, parent: index, step: step})
				return false
			})
			if solution != nil {
				return solution, false
			}
			if limited {
				return nil, true
			}
			if budget.stopped != "" {
				return nil, false
			}
		}
	}
	return nil, false
}

func jokerStateKey(state jokerReplay) string {
	held, used := []string{}, []string{}
	sources := map[string][]string{}
	for id := range state.Held {
		held = append(held, id)
		for _, p := range state.Sources[id] {
			sources[id] = append(sources[id], p.ID)
		}
		sort.Strings(sources[id])
	}
	for id := range state.UsedRack {
		used = append(used, id)
	}
	sort.Strings(held)
	sort.Strings(used)
	// All tiles were validated and IDs are strings, so encoding cannot fail.
	data, err := json.Marshal(struct {
		Table      [][]Piece
		Held, Used []string
		Sources    map[string][]string
	}{state.Table, held, used, sources})
	if err != nil {
		panic("validated joker state could not be encoded: " + err.Error())
	}
	return string(data)
}
