package game

import (
	"fmt"
	"strings"
)

// Piece identifies a physical occurrence within one photographed snapshot.
// Equal tile values must still have distinct IDs. As belongs to a particular
// table meld; changing a joker's role does not change its inventory value.
type Piece struct {
	ID   string `json:"id"`
	Tile Tile   `json:"tile"`
	As   *Tile  `json:"as,omitempty"`
}

type Position struct {
	Table [][]Piece `json:"table"`
	Rack  []Piece   `json:"rack"`
}

type PositionIssue struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Zone      string `json:"zone"`
	MeldIndex int    `json:"meld_index"`
	TileIndex int    `json:"tile_index"`
	TileID    string `json:"tile_id,omitempty"`
	RelatedID string `json:"related_id,omitempty"`
}

type MeldSummary struct {
	Kind      MeldKind `json:"kind"`
	Score     int      `json:"score"`
	Ambiguous bool     `json:"ambiguous"`
}

type PositionReport struct {
	Valid    bool            `json:"valid"`
	Issues   []PositionIssue `json:"issues"`
	Melds    []MeldSummary   `json:"melds"`
	Resolved [][]Piece       `json:"resolved_table,omitempty"`
}

// ValidatePosition checks the displayed table plus combined known table/rack
// inventory. An empty table is allowed; an empty meld is not. Opening state is
// deliberately irrelevant. Explicit joker bindings are respected. Unbound
// jokers can have several interpretations; the report marks that ambiguity.
func ValidatePosition(position Position, edition Edition) PositionReport {
	report := PositionReport{Issues: []PositionIssue{}, Melds: []MeldSummary{}}
	numbered, jokers, total, err := edition.Limits()
	if err != nil {
		report.Issues = append(report.Issues, PositionIssue{Code: "invalid_edition", Message: err.Error(), MeldIndex: -1, TileIndex: -1})
		return report
	}
	count := len(position.Rack)
	for _, meld := range position.Table {
		count += len(meld)
	}
	if count > total || len(position.Table) > total/3 {
		report.Issues = append(report.Issues, PositionIssue{Code: "inventory_size", Message: fmt.Sprintf("known position exceeds the %d-tile edition", total), MeldIndex: -1, TileIndex: -1})
		return report
	}
	seen := make(map[string]bool, count)
	var counts [4*MaxNumber + 1]int
	inspect := func(piece Piece, zone string, meld, index int) {
		add := func(code, message string) {
			report.Issues = append(report.Issues, PositionIssue{Code: code, Message: message, Zone: zone, MeldIndex: meld, TileIndex: index, TileID: piece.ID})
		}
		if piece.ID == "" || len(piece.ID) > 128 || strings.TrimSpace(piece.ID) != piece.ID {
			add("invalid_id", "piece IDs must be nonempty, trimmed, and at most 128 bytes")
		}
		if seen[piece.ID] {
			add("duplicate_id", "physical piece ID appears more than once")
		}
		seen[piece.ID] = true
		if err := piece.Tile.Validate(); err != nil {
			add("invalid_tile", err.Error())
			return
		}
		if piece.As != nil && (zone == "rack" || piece.Tile.Kind != Joker || piece.As.Kind != Numbered || piece.As.Validate() != nil) {
			add("invalid_joker_role", "only table jokers can carry valid numbered roles")
		}
		value := tileIndex(piece.Tile)
		counts[value]++
		limit := numbered
		if piece.Tile.Kind == Joker {
			limit = jokers
		}
		if counts[value] > limit {
			add("excess_copies", fmt.Sprintf("known inventory exceeds %d copies of this tile value", limit))
		}
	}
	for m, meld := range position.Table {
		for i, piece := range meld {
			inspect(piece, "table", m, i)
		}
		analysis := analyzePieces(meld)
		report.Melds = append(report.Melds, MeldSummary{analysis.Kind, analysis.Score, analysis.Ambiguous})
		if !analysis.Valid() {
			issue := PositionIssue{Code: string(analysis.Issue.Code), Message: "table meld is not valid in displayed order", Zone: "table", MeldIndex: m, TileIndex: analysis.Issue.TileIndex}
			if issue.TileIndex >= 0 && issue.TileIndex < len(meld) {
				issue.TileID = meld[issue.TileIndex].ID
			}
			if analysis.Issue.RelatedIndex >= 0 {
				issue.RelatedID = meld[analysis.Issue.RelatedIndex].ID
			}
			report.Issues = append(report.Issues, issue)
			continue
		}
		resolved := clonePieces(meld)
		for i := range resolved {
			if resolved[i].Tile.Kind == Joker {
				role := analysis.Roles[i]
				resolved[i].As = &role
			}
		}
		report.Resolved = append(report.Resolved, resolved)
	}
	for i, piece := range position.Rack {
		inspect(piece, "rack", -1, i)
	}
	report.Valid = len(report.Issues) == 0
	if !report.Valid {
		report.Resolved = nil
	}
	return report
}

func analyzePieces(meld []Piece) MeldResult {
	if len(meld) > MaxNumber {
		return invalidMeld(InvalidSize, -1, -1)
	}
	var tiles [MaxNumber]Tile
	var roles [MaxNumber]*Tile
	for i, piece := range meld {
		tiles[i] = piece.Tile
		roles[i] = piece.As
	}
	return AnalyzeLayout(tiles[:len(meld)], roles[:len(meld)])
}

func clonePieces(pieces []Piece) []Piece {
	result := make([]Piece, len(pieces))
	copy(result, pieces)
	for i := range result {
		if result[i].As != nil {
			role := *result[i].As
			result[i].As = &role
		}
	}
	return result
}

// ClonePosition deep-copies mutable slices and joker-role pointers.
func ClonePosition(position Position) Position {
	result := Position{Table: make([][]Piece, len(position.Table)), Rack: clonePieces(position.Rack)}
	for i, meld := range position.Table {
		result.Table[i] = clonePieces(meld)
	}
	return result
}
