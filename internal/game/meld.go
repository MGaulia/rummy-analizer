package game

// MeldKind identifies an unordered collection that can form a group or run.
type MeldKind string

const (
	Group MeldKind = "group"
	Run   MeldKind = "run"
)

type IssueCode string

const (
	InvalidSize      IssueCode = "invalid_size"
	InvalidTile      IssueCode = "invalid_tile"
	UnsupportedJoker IssueCode = "unsupported_joker"
	DuplicateTile    IssueCode = "duplicate_tile"
	RunGap           IssueCode = "run_gap"
	MixedMeld        IssueCode = "mixed_meld"
)

// ValidationIssue uses zero-based input indices; -1 means no specific tile.
// RelatedIndex identifies the earlier occurrence for a duplicate. A future
// table/photo caller can attach its own group ID without changing tile values.
type ValidationIssue struct {
	Code         IssueCode
	TileIndex    int
	RelatedIndex int
}

type MeldResult struct {
	Kind  MeldKind
	Issue ValidationIssue
}

// Valid deliberately returns false for the zero result. Inspect Issue.Code to
// distinguish invalid numbered melds from unsupported joker-containing input.
func (r MeldResult) Valid() bool {
	return (r.Kind == Group || r.Kind == Run) && r.Issue.Code == ""
}

func invalidMeld(code IssueCode, index, related int) MeldResult {
	return MeldResult{Issue: ValidationIssue{Code: code, TileIndex: index, RelatedIndex: related}}
}

// AnalyzeMeld checks whether tiles CAN FORM one meld after rearrangement. It
// never validates displayed order, multiple groups, inventory, or a legal turn.
// It does not mutate tiles and uses O(n) time and constant scratch space.
// Representation errors and unsupported jokers are checked in input order
// before size and arrangement errors. Only the first issue is returned; no
// joker is silently treated as a numbered tile.
func AnalyzeMeld(tiles []Tile) MeldResult {
	for i, tile := range tiles {
		if err := tile.Validate(); err != nil {
			return invalidMeld(InvalidTile, i, -1)
		}
		if tile.Kind == Joker {
			return invalidMeld(UnsupportedJoker, i, -1)
		}
	}
	if len(tiles) < 3 || len(tiles) > MaxNumber {
		return invalidMeld(InvalidSize, -1, -1)
	}

	var first [4 * MaxNumber]int // input index + 1; zero means unseen
	minimum, maximum := MaxNumber, MinNumber
	sameColor, sameNumber := true, true
	for i, tile := range tiles {
		index := tileIndex(tile)
		if first[index] != 0 {
			return invalidMeld(DuplicateTile, i, first[index]-1)
		}
		first[index] = i + 1
		sameColor = sameColor && tile.Color == tiles[0].Color
		sameNumber = sameNumber && tile.Number == tiles[0].Number
		if tile.Number < minimum {
			minimum = tile.Number
		}
		if tile.Number > maximum {
			maximum = tile.Number
		}
	}
	// At this point all tiles are valid and distinct. Equal numbers therefore
	// imply distinct colors (and at most four tiles).
	if sameNumber {
		return MeldResult{Kind: Group}
	}
	if sameColor {
		if maximum-minimum+1 != len(tiles) {
			return invalidMeld(RunGap, -1, -1)
		}
		return MeldResult{Kind: Run}
	}
	return invalidMeld(MixedMeld, -1, -1)
}

// Boolean conveniences share the same implementation. False is not a proof
// that a joker-containing collection cannot form a meld; use AnalyzeMeld when
// the distinction between unsupported and invalid matters.
func CanFormMeld(tiles []Tile) bool { return AnalyzeMeld(tiles).Valid() }

func CanFormGroup(tiles []Tile) bool {
	result := AnalyzeMeld(tiles)
	return result.Valid() && result.Kind == Group
}

func CanFormRun(tiles []Tile) bool {
	result := AnalyzeMeld(tiles)
	return result.Valid() && result.Kind == Run
}
