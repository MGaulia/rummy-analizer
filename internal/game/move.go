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
	Table [][]Piece `json:"table"`
	Freed []string  `json:"freed"`
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
	known := map[string]Piece{}
	tableIDs := map[string]bool{}
	for _, meld := range before.Table {
		for _, p := range meld {
			known[p.ID] = p
			tableIDs[p.ID] = true
		}
	}
	for _, p := range before.Rack {
		known[p.ID] = p
	}
	// Validate each complete physical state; absent rack pieces remain on rack,
	// absent original table pieces must be specifically held table jokers.
	state := func(table [][]Piece, held map[string]bool) (Position, string) {
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
		result := Position{Table: table}
		for _, p := range before.Rack {
			if !seen[p.ID] {
				result.Rack = append(result.Rack, p)
			}
		}
		if report := ValidatePosition(result, rules.Edition); !report.Valid {
			return Position{}, fmt.Sprint(report.Issues)
		}
		return result, ""
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
		for _, old := range original.Resolved {
			found := false
			for _, meld := range after.Table {
				if sameMeld(old, meld) {
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
		held := map[string]bool{}
		previous := original.Resolved
		used := map[string]bool{}
		for _, release := range move.Releases {
			if len(release.Freed) == 0 {
				return fail("joker_release", "a release must free at least one table joker")
			}
			for _, id := range release.Freed {
				p, ok := known[id]
				if !ok || !tableIDs[id] || p.Tile.Kind != Joker || held[id] {
					return fail("joker_release", "release must name an unreleased original table joker")
				}
				held[id] = true
			}
			stage, err := state(release.Table, held)
			if err != "" {
				return fail("joker_release", err)
			}
			resolved := ValidatePosition(stage, rules.Edition).Resolved
			if !retainedJokers(previous, resolved, held) {
				return fail("joker_release", "a retained joker changed without release evidence")
			}
			present := map[string]bool{}
			for _, meld := range resolved {
				for _, p := range meld {
					present[p.ID] = true
					if !tableIDs[p.ID] {
						used[p.ID] = true
					}
				}
			}
			for id := range used {
				if !present[id] {
					return fail("joker_release", "a played rack tile cannot be returned to the rack")
				}
			}
			previous = resolved
		}
		if !retainedJokers(previous, after.Table, held) {
			return fail("joker_witness_required", "changed table joker needs a valid release stage before reuse")
		}
		for id := range used {
			if remaining[id] {
				return fail("joker_release", "release-stage rack tile was not retained")
			}
		}
		// A released joker must be used in a new set, not simply put back into
		// exactly its original physical meld with an unchanged role.
		for id := range held {
			if samePhysicalMeld(findMeld(original.Resolved, id), findMeld(after.Table, id)) {
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

// Retention allows adding/removing endpoints or group members while retaining
// the joker's concrete role and a nested physical meld. More complex changes
// require an explicit release certificate instead of assuming final validity.
func retainedJokers(before, after [][]Piece, freed map[string]bool) bool {
	for _, old := range before {
		for _, j := range old {
			if j.Tile.Kind != Joker || freed[j.ID] {
				continue
			}
			next := findMeld(after, j.ID)
			if len(next) == 0 {
				return false
			}
			if analyzePieces(old).Kind != analyzePieces(next).Kind {
				return false
			}
			oldSet, newSet := map[string]bool{}, map[string]bool{}
			for _, p := range old {
				oldSet[p.ID] = true
			}
			for _, p := range next {
				newSet[p.ID] = true
				if p.ID == j.ID && !samePiece(j, p) {
					return false
				}
			}
			sub, super := true, true
			for id := range oldSet {
				sub = sub && newSet[id]
			}
			for id := range newSet {
				super = super && oldSet[id]
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
