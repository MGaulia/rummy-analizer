package game

import "fmt"

// Move specifies physical occurrences, not just tile values. Release stages are
// executable evidence: each leaves a valid table while the listed table jokers
// are held for immediate reuse in the final table. Held tiles cannot be rack tiles.
type Move struct {
	Table    [][]Piece      `json:"table"`
	Releases []JokerRelease `json:"releases,omitempty"`
}

type JokerRelease struct {
	Table  [][]Piece `json:"table"`
	Freed  []string  `json:"freed,omitempty"`
	Reused []string  `json:"reused,omitempty"`
}

type MoveReport struct {
	Valid        bool     `json:"valid"`
	Code         string   `json:"code,omitempty"`
	Message      string   `json:"message,omitempty"`
	Played       []string `json:"played"`
	OpeningScore int      `json:"opening_score"`
	After        Position `json:"after"`
}

// ValidateMove checks the original position, turn, conservation, opening, and
// joker release evidence independently of search. It never mutates/aliases input.
// The currently encoded release certificate is sufficient, not a claim to cover
// every booklet manipulation; search explicitly exposes that coverage limit.
func ValidateMove(before Position, rules Rules, turn Turn, move Move) MoveReport {
	fail := func(code, message string) MoveReport { return MoveReport{Code: code, Message: message} }
	original := ValidatePosition(before, rules.Edition)
	if !original.Valid {
		return fail("invalid_position", fmt.Sprint(original.Issues))
	}
	if turn.Drawn {
		return fail("turn_ended", "drawing ends the turn")
	}
	state := physicalTableChecker(before, rules.Edition)
	tableIDs := map[string]bool{}
	for _, meld := range before.Table {
		for _, p := range meld {
			tableIDs[p.ID] = true
		}
	}

	after, problem := state(move.Table, nil)
	if problem != "" {
		return fail("tile_accounting", problem)
	}
	final := ValidatePosition(after, rules.Edition)
	after = ClonePosition(after)
	after.Table = final.Resolved
	played := []string{}
	remaining := map[string]bool{}
	for _, p := range after.Rack {
		remaining[p.ID] = true
	}
	for _, p := range before.Rack {
		if !remaining[p.ID] {
			played = append(played, p.ID)
		}
	}
	if len(played) == 0 {
		return fail("no_rack_tile", "a move must play at least one rack tile")
	}
	openingScore := 0
	if rules.NeedsOpening(turn) {
		if len(move.Releases) != 0 {
			return fail("opening_manipulation", "opening cannot release table jokers")
		}
		for _, old := range before.Table {
			found := false
			for _, meld := range after.Table {
				if sameInitialMeld(old, meld) {
					found = true
					break
				}
			}
			if !found {
				return fail("opening_manipulation", "opening must leave existing table melds unchanged")
			}
		}
		for m, meld := range after.Table {
			newMeld := true
			for _, p := range meld {
				if tableIDs[p.ID] {
					newMeld = false
				}
			}
			if newMeld {
				openingScore += final.Melds[m].Score
			} else {
				for _, p := range meld {
					if !tableIDs[p.ID] {
						return fail("opening_manipulation", "opening uses rack-only new melds")
					}
				}
			}
		}
		if openingScore < 30 {
			return fail("opening_score", "opening must total at least 30 represented points")
		}

	} else {
		replay, code, message := replayJokerStages(before, rules.Edition, move.Releases)
		if code != "" {
			return fail(code, message)
		}
		if !retainedJokers(replay.Table, after.Table, replay.Held) {
			return fail("joker_witness_required", "changed table joker needs a valid release stage before reuse")
		}
		for id := range replay.UsedRack {
			if remaining[id] {
				return fail("joker_release", "release-stage rack tile was not retained")
			}
		}
		for id := range replay.Held {
			if samePhysicalMeld(replay.Sources[id], findMeld(after.Table, id)) {
				return fail("joker_reuse", "released joker must be used in a new set on this turn")
			}
		}
	}

	return MoveReport{Valid: true, Played: played, OpeningScore: openingScore, After: after}
}

func findMeld(table [][]Piece, id string) []Piece {
	for _, meld := range table {
		for _, p := range meld {
			if p.ID == id {
				return meld
			}
		}
	}
	return nil
}

func samePiece(a, b Piece) bool {
	return a.ID == b.ID && a.Tile == b.Tile && ((a.As == nil && b.As == nil) || (a.As != nil && b.As != nil && *a.As == *b.As))
}

func sameMeld(a, b []Piece) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !samePiece(a[i], b[i]) {
			return false
		}
	}
	return true
}

// An unspecified original role may take any interpretation consistent with
// its displayed set. Once specified at a checkpoint, it remains fixed until
// release. Equal numbered physical copies can substitute without resetting it.
func bindRetainedRoles(old, next []Piece) ([]Piece, bool) {
	bound := clonePieces(old)
	for i, p := range bound {
		if p.Tile.Kind != Joker {
			continue
		}
		for _, q := range next {
			if p.ID == q.ID {
				if q.As == nil {
					return nil, false
				}
				if p.As != nil && *p.As != *q.As {
					return nil, false
				}
				role := *q.As
				bound[i].As = &role
				break
			}
		}
	}
	return bound, analyzePieces(bound).Valid()
}
func sameInitialMeld(old, next []Piece) bool {
	bound, ok := bindRetainedRoles(old, next)
	return ok && sameMeld(bound, next)
}
func retainedJokers(before, after [][]Piece, freed map[string]bool) bool {
	for _, old := range before {
		for _, j := range old {
			if j.Tile.Kind != Joker || freed[j.ID] {
				continue
			}
			next := findMeld(after, j.ID)
			bound, ok := bindRetainedRoles(old, next)
			if !ok || len(next) == 0 {
				return false
			}
			nextAnalysis := analyzePieces(next)
			var tiles [MaxNumber]Tile
			var roles [MaxNumber]*Tile
			for i, p := range bound {
				tiles[i] = p.Tile
				roles[i] = p.As
			}
			oldAnalysis := analyzeMeld(tiles[:len(bound)], roles[:len(bound)], true, nextAnalysis.Kind)
			if !oldAnalysis.Valid() {
				return false
			}
			oldSet, newSet := map[Tile]int{}, map[Tile]int{}
			for i, p := range old {
				if p.ID != j.ID {
					oldSet[oldAnalysis.Roles[i]]++
				}
			}
			for i, p := range next {
				if p.ID != j.ID {
					newSet[nextAnalysis.Roles[i]]++
				}
			}
			sub, super := true, true
			for key, count := range oldSet {
				sub = sub && newSet[key] >= count
			}
			for key, count := range newSet {
				super = super && oldSet[key] >= count
			}
			if !sub && !super {
				return false
			}
		}
	}
	return true
}

// Reordering a group or changing wildcard roles alone does not make a new
// physical set for purposes of the release/reuse certificate.
func samePhysicalMeld(a, b []Piece) bool {
	if len(a) != len(b) {
		return false
	}
	ids := map[string]bool{}
	for _, p := range a {
		ids[p.ID] = true
	}
	for _, p := range b {
		if !ids[p.ID] {
			return false
		}
	}
	return true
}

// Every checkpoint conserves physical occurrences. Only held original table
// jokers may be absent; unused rack pieces remain on the rack.
func physicalTableChecker(before Position, edition Edition) func([][]Piece, map[string]bool) (Position, string) {
	known, tableIDs := map[string]Piece{}, map[string]bool{}
	for _, meld := range before.Table {
		for _, p := range meld {
			known[p.ID] = p
			tableIDs[p.ID] = true
		}
	}
	for _, p := range before.Rack {
		known[p.ID] = p
	}
	return func(table [][]Piece, held map[string]bool) (Position, string) {
		seen := map[string]bool{}
		for _, meld := range table {
			for _, p := range meld {
				old, ok := known[p.ID]
				if !ok || p.Tile != old.Tile {
					return Position{}, "invented or changed physical tile: " + p.ID
				}
				if seen[p.ID] || held[p.ID] {
					return Position{}, "duplicate physical tile: " + p.ID
				}
				seen[p.ID] = true
			}
		}
		for id := range tableIDs {
			if !seen[id] && !held[id] {
				return Position{}, "missing table tile: " + id
			}
		}
		position := Position{Table: table}
		for _, p := range before.Rack {
			if !seen[p.ID] {
				position.Rack = append(position.Rack, p)
			}
		}
		report := ValidatePosition(position, edition)
		if !report.Valid {
			return Position{}, fmt.Sprint(report.Issues)
		}
		position.Table = report.Resolved
		return position, ""
	}
}

type jokerReplay struct {
	Table    [][]Piece
	Held     map[string]bool
	UsedRack map[string]bool
	Sources  map[string][]Piece
}

// Replay is shared by the independent validator and witness search. A stage
// can release, reuse, or prepare a rearrangement while retaining concrete roles.
func replayJokerStages(before Position, edition Edition, stages []JokerRelease) (jokerReplay, string, string) {
	replay := jokerReplay{Table: ClonePosition(before).Table, Held: map[string]bool{}, UsedRack: map[string]bool{}, Sources: map[string][]Piece{}}
	tableIDs := map[string]bool{}
	for _, meld := range before.Table {
		for _, p := range meld {
			tableIDs[p.ID] = true
		}
	}
	state := physicalTableChecker(before, edition)
	for _, stage := range stages {
		mentioned := map[string]bool{}
		for _, id := range stage.Reused {
			if mentioned[id] || !replay.Held[id] {
				return replay, "joker_release", "reuse must name a held joker exactly once"
			}
			mentioned[id] = true
			delete(replay.Held, id)
		}
		for _, id := range stage.Freed {
			source := findMeld(replay.Table, id)
			if mentioned[id] || replay.Held[id] || len(source) == 0 {
				return replay, "joker_release", "release must name an available original table joker exactly once"
			}
			valid := false
			for _, p := range source {
				valid = valid || (p.ID == id && p.Tile.Kind == Joker && tableIDs[id])
			}
			if !valid {
				return replay, "joker_release", "only original table jokers may be held"
			}
			mentioned[id] = true
			replay.Held[id] = true
			replay.Sources[id] = clonePieces(source)
		}
		position, problem := state(stage.Table, replay.Held)
		if problem != "" {
			return replay, "joker_release", problem
		}
		if !retainedJokers(replay.Table, position.Table, replay.Held) {
			return replay, "joker_release", "a retained joker changed without release evidence"
		}
		for _, id := range stage.Reused {
			if samePhysicalMeld(replay.Sources[id], findMeld(position.Table, id)) {
				return replay, "joker_reuse", "released joker must be reused in a new physical set"
			}
		}
		present := map[string]bool{}
		for _, meld := range position.Table {
			for _, p := range meld {
				present[p.ID] = true
				if !tableIDs[p.ID] {
					replay.UsedRack[p.ID] = true
				}
			}
		}
		for id := range replay.UsedRack {
			if !present[id] {
				return replay, "joker_release", "a played rack tile cannot be returned to the rack"
			}
		}
		replay.Table = position.Table
	}
	return replay, "", ""
}
