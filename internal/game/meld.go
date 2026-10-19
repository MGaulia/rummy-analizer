package game

// MeldKind identifies a collection that can form a group or run.
type MeldKind string

const (
	Group MeldKind = "group"
	Run   MeldKind = "run"
)

type IssueCode string

const (
	InvalidSize   IssueCode = "invalid_size"
	InvalidTile   IssueCode = "invalid_tile"
	InvalidRole   IssueCode = "invalid_joker_role"
	DuplicateTile IssueCode = "duplicate_tile"
	RunGap        IssueCode = "run_gap"
	MixedMeld     IssueCode = "mixed_meld"
	InvalidOrder  IssueCode = "invalid_order"
)

// ValidationIssue uses zero-based input indices; -1 means no specific tile.
type ValidationIssue struct {
	Code         IssueCode `json:"code"`
	TileIndex    int       `json:"tile_index"`
	RelatedIndex int       `json:"related_index"`
}

// MeldResult contains one concrete interpretation. Roles use input positions;
// only the first len(input) entries are meaningful. Ambiguous means another
// interpretation also exists, not that recognition confidence was measured.
// Score counts a joker's represented number, not its end-of-game penalty.
type MeldResult struct {
	Kind      MeldKind
	Issue     ValidationIssue
	Roles     [MaxNumber]Tile
	Score     int
	Ambiguous bool
}

func (r MeldResult) Valid() bool {
	return (r.Kind == Group || r.Kind == Run) && r.Issue.Code == ""
}

func invalidMeld(code IssueCode, index, related int) MeldResult {
	return MeldResult{Issue: ValidationIssue{Code: code, TileIndex: index, RelatedIndex: related}}
}

// AnalyzeMeld checks unordered feasibility, including jokers, without mutation.
// It does not validate whole-table inventory or a legal turn. On ambiguity it
// chooses the highest-scoring interpretation; callers must not assume it is the
// only interpretation. Small fixed bounds keep scratch storage on the stack.
func AnalyzeMeld(tiles []Tile) MeldResult { return analyzeMeld(tiles, nil, false, "") }

// AnalyzeLayout preserves positions. Runs must be consecutive in one direction
// (ascending or descending); groups allow any color ordering. roles may be nil,
// or contain one optional numbered role per input tile. Only jokers may bind a
// role. A nil joker role is resolved rather than guessed from a zero-value tile.
func AnalyzeLayout(tiles []Tile, roles []*Tile) MeldResult {
	return analyzeMeld(tiles, roles, true, "")
}

func CanFormMeld(tiles []Tile) bool  { return AnalyzeMeld(tiles).Valid() }
func CanFormGroup(tiles []Tile) bool { return analyzeMeld(tiles, nil, false, Group).Valid() }
func CanFormRun(tiles []Tile) bool   { return analyzeMeld(tiles, nil, false, Run).Valid() }

func analyzeMeld(tiles []Tile, roles []*Tile, ordered bool, only MeldKind) MeldResult {
	if roles != nil && len(roles) != len(tiles) {
		return invalidMeld(InvalidRole, -1, -1)
	}
	// Validate representations before size, without indexing bounded storage.
	for i, tile := range tiles {
		if tile.Validate() != nil {
			return invalidMeld(InvalidTile, i, -1)
		}
		if roles != nil && roles[i] != nil && (tile.Kind != Joker || roles[i].Kind != Numbered || roles[i].Validate() != nil) {
			return invalidMeld(InvalidRole, i, -1)
		}
	}
	if len(tiles) < 3 || len(tiles) > MaxNumber {
		return invalidMeld(InvalidSize, -1, -1)
	}
	var fixed [MaxNumber]Tile
	var first [4 * MaxNumber]int
	known, unknown := 0, 0
	sameColor, sameNumber := true, true
	var reference Tile
	for i, tile := range tiles {
		if tile.Kind == Joker {
			if roles == nil || roles[i] == nil {
				unknown++
				continue
			}
			tile = *roles[i]
		}
		index := tileIndex(tile)
		if first[index] != 0 {
			return invalidMeld(DuplicateTile, i, first[index]-1)
		}
		first[index] = i + 1
		fixed[i] = tile
		if known == 0 {
			reference = tile
		}
		sameColor = sameColor && tile.Color == reference.Color
		sameNumber = sameNumber && tile.Number == reference.Number
		known++
	}
	var best MeldResult
	consider := func(candidate MeldResult) {
		if best.Valid() {
			if candidate.Score > best.Score {
				candidate.Ambiguous = true
				best = candidate
			} else {
				best.Ambiguous = true
			}
		} else {
			best = candidate
		}
	}
	colors := [4]Color{Black, Blue, Orange, Red}
	if only != Run && len(tiles) <= 4 && sameNumber {
		number := reference.Number
		if known == 0 {
			number = MaxNumber
		}
		candidate := MeldResult{Kind: Group, Score: number * len(tiles), Roles: fixed}
		var used [4]bool
		for _, tile := range fixed[:len(tiles)] {
			if tile.Color != "" {
				used[colorIndex(tile.Color)] = true
			}
		}
		for i := range tiles {
			if fixed[i].Color != "" {
				continue
			}
			for c, color := range colors {
				if !used[c] {
					candidate.Roles[i] = Tile{Color: color, Number: number}
					used[c] = true
					break
				}
			}
		}
		candidate.Ambiguous = known == 0 || (unknown > 0 && (unknown > 1 || len(tiles) < 4))
		consider(candidate)
	}
	if only != Group && sameColor {
		for _, color := range colors {
			if known > 0 && color != reference.Color {
				continue
			}
			for start := MinNumber; start <= MaxNumber-len(tiles)+1; start++ {
				for _, direction := range []int{1, -1} {
					if !ordered && direction == -1 {
						continue
					}
					candidate := MeldResult{Kind: Run, Roles: fixed}
					var used [MaxNumber + 1]bool
					fits := true
					for i, tile := range fixed[:len(tiles)] {
						if tile.Color == "" {
							continue
						}
						want := start + i
						if direction == -1 {
							want = start + len(tiles) - 1 - i
						}
						if tile.Number < start || tile.Number >= start+len(tiles) || (ordered && tile.Number != want) {
							fits = false
							break
						}
						used[tile.Number] = true
					}
					if !fits {
						continue
					}
					for i := range tiles {
						if fixed[i].Color != "" {
							continue
						}
						number := start + i
						if direction == -1 {
							number = start + len(tiles) - 1 - i
						}
						if !ordered {
							for number = start; used[number]; number++ {
							}
							used[number] = true
						}
						candidate.Roles[i] = Tile{Color: color, Number: number}
					}
					candidate.Score = len(tiles) * (2*start + len(tiles) - 1) / 2
					candidate.Ambiguous = !ordered && unknown > 1
					consider(candidate)
				}
			}
		}
	}
	if best.Valid() {
		return best
	}
	if ordered {
		if unordered := analyzeMeld(tiles, roles, false, only); unordered.Valid() {
			return invalidMeld(InvalidOrder, -1, -1)
		}
	}
	if sameColor {
		return invalidMeld(RunGap, -1, -1)
	}
	return invalidMeld(MixedMeld, -1, -1)
}
