package game

import "sort"

// walkNumberedPartitions keeps table counts mandatory and rack counts optional.
// Without jokers, equal numbered copies are interchangeable for move
// legality, so each value consumes stable table IDs before rack IDs. Joker
// positions keep the physical search and its release-certificate handling.
func walkNumberedPartitions(pool []Piece, required int, budget *searchBudget, visit func([][]Piece) bool) bool {
	const values = 4 * MaxNumber
	var pieces [values][]Piece
	var available, needed, cursor [values]uint8
	for i, piece := range pool {
		value := tileIndex(piece.Tile)
		pieces[value] = append(pieces[value], piece)
		available[value]++
		// A one-tile rack has no legal partial move: that tile is mandatory too.
		if i < required || len(pool) == required+1 {
			needed[value]++
		}
	}
	var templates [][]int
	var containing [values][]int
	add := func(roles []int) bool {
		if !budget.tick() {
			return false
		}
		for _, value := range roles {
			if available[value] == 0 {
				return true
			}
		}
		index := len(templates)
		templates = append(templates, append([]int(nil), roles...))
		for _, value := range roles {
			containing[value] = append(containing[value], index)
		}
		return true
	}
	// These are all 65 distinct-color groups and 264 non-wrapping runs. A
	// template absent from the initial inventory cannot become available later.
	for number := 0; number < MaxNumber; number++ {
		for mask := 1; mask < 16; mask++ {
			var roles [4]int
			length := 0
			for color := 0; color < 4; color++ {
				if mask&(1<<color) != 0 {
					roles[length] = color*MaxNumber + number
					length++
				}
			}
			if length >= 3 && !add(roles[:length]) {
				return false
			}
		}
	}
	for color := 0; color < 4; color++ {
		for start := 0; start < MaxNumber-2; start++ {
			for end := start + 2; end < MaxNumber; end++ {
				var roles [MaxNumber]int
				for number := start; number <= end; number++ {
					roles[number-start] = color*MaxNumber + number
				}
				if !add(roles[:end-start+1]) {
					return false
				}
			}
		}
	}
	possible := func(template []int) bool {
		for _, value := range template {
			if available[value] == 0 {
				return false
			}
		}
		return true
	}
	// Equal remaining counts have the same future choices. A path with fewer
	// selected tiles is dominated, even if it omitted different rack copies.
	type state struct{ available, needed [values]uint8 }
	seen := make(map[state]int)
	// A legal move must improve on keeping the validated original table alone.
	best := required
	var path [][]Piece
	var walk func(int) bool
	walk = func(selected int) bool {
		if !budget.tick() {
			return false
		}
		upper := selected
		mandatoryLeft := false
		optionalLeft := false
		for value, count := range available {
			upper += int(count)
			mandatoryLeft = mandatoryLeft || needed[value] > 0
			optionalLeft = optionalLeft || count > needed[value]
		}
		if upper <= best {
			return false
		}
		key := state{available, needed}
		if previous, ok := seen[key]; ok && previous >= selected {
			return false
		}
		seen[key] = selected
		if !mandatoryLeft {
			if selected > best {
				best = selected
			}
			if visit(path) {
				return true
			}
		}
		anchor, fewest := -1, len(templates)+1
		// Prefer scarce choices per copy while deciding rack participation;
		// once every remaining copy is mandatory, use ordinary exact-cover MRV.
		// Cross multiplication compares the ratios without rounding.
		weight := func(count uint8) int {
			if optionalLeft {
				return int(count)
			}
			return 1
		}
		for value, count := range available {
			if count == 0 {
				continue
			}
			if !budget.tick() {
				return false
			}
			choices := 0
			for _, index := range containing[value] {
				if possible(templates[index]) {
					choices++
					if anchor >= 0 && choices*weight(available[anchor]) > fewest*weight(count) {
						break
					}
				}
			}
			if choices == 0 && needed[value] > 0 {
				return false
			}
			if anchor < 0 || choices*weight(available[anchor]) < fewest*weight(count) ||
				(choices*weight(available[anchor]) == fewest*weight(count) && available[value] > needed[value] && available[anchor] == needed[anchor]) {
				anchor, fewest = value, choices
			}
		}
		if anchor < 0 {
			return false
		}
		choices := make([]int, 0, len(containing[anchor]))
		for _, index := range containing[anchor] {
			if !budget.tick() {
				return false
			}
			if possible(templates[index]) {
				choices = append(choices, index)
			}
		}
		rackUse := func(template []int) (played, potential int) {
			for _, value := range template {
				if needed[value] == 0 {
					played++
				}
				if available[value] > needed[value] {
					potential++
				}
			}
			return played, potential
		}
		// This changes branch order only; every feasible template is still tried.
		sort.SliceStable(choices, func(i, j int) bool {
			a, b := templates[choices[i]], templates[choices[j]]
			playedA, potentialA := rackUse(a)
			playedB, potentialB := rackUse(b)
			if playedA != playedB {
				return playedA > playedB
			}
			if potentialA != potentialB {
				return potentialA > potentialB
			}
			return len(a) > len(b)
		})
		for _, index := range choices {
			template := templates[index]
			meld := make([]Piece, len(template))
			var consumed [MaxNumber]bool
			for i, value := range template {
				meld[i] = pieces[value][cursor[value]]
				cursor[value]++
				available[value]--
				if needed[value] > 0 {
					needed[value]--
					consumed[i] = true
				}
			}
			path = append(path, meld)
			stop := walk(selected + len(template))
			path = path[:len(path)-1]
			for i, value := range template {
				cursor[value]--
				available[value]++
				if consumed[i] {
					needed[value]++
				}
			}
			if stop {
				return true
			}
		}
		if available[anchor] > needed[anchor] && budget.stopped == "" {
			// Omit the last remaining copy: table IDs precede rack IDs.
			available[anchor]--
			stop := walk(selected)
			available[anchor]++
			return stop
		}
		return false
	}
	return walk(0)
}
