package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/MGaulia/rummy-analizer/internal/game"
)

var (
	ErrInputFormat = errors.New("expected <color> <number> or joker")
	ErrColor       = errors.New("unknown color; use black, blue, orange, red, or an explicit -color-aliases mapping")
	ErrNumber      = errors.New("expected a number from 1 to 13, without signs or leading zeros")
	ErrColorAlias  = errors.New("invalid color alias")
)

// Parser is a text adapter. The zero value accepts canonical colors. Aliases
// are copied at construction, so caller mutations cannot change later parsing.
type Parser struct {
	aliases map[string]game.Color
}

func canonicalColor(label string) (game.Color, bool) {
	color := game.Color(label)
	return color, (game.Tile{Color: color, Number: game.MinNumber}).Validate() == nil
}

// NewParser takes an explicit user-chosen mapping, never a guessed mapping of
// the old grey/green/purple labels. Canonical labels cannot be reassigned and
// joker is reserved. Conflicting aliases after case normalization are rejected.
func NewParser(aliases map[string]game.Color) (Parser, error) {
	result := Parser{aliases: make(map[string]game.Color, len(aliases))}
	for label, color := range aliases {
		label = strings.ToLower(strings.TrimSpace(label))
		_, validTarget := canonicalColor(string(color))
		canonical, isCanonical := canonicalColor(label)
		prior, exists := result.aliases[label]
		if len(strings.Fields(label)) != 1 || strings.ContainsAny(label, ",=") || label == "joker" || !validTarget ||
			(isCanonical && canonical != color) || (exists && prior != color) {
			return Parser{}, fmt.Errorf("%w: %q", ErrColorAlias, label)
		}
		result.aliases[label] = color
	}
	return result, nil
}

// ParseTile returns a zero Tile on every error. Joker syntax is accepted as a
// representation; the game validator explicitly reports unsupported rules.
func (p Parser) ParseTile(input string) (game.Tile, error) {
	fields := strings.Fields(input)
	if len(fields) == 1 && strings.EqualFold(fields[0], "joker") {
		return game.Tile{Kind: game.Joker}, nil
	}
	if len(fields) != 2 {
		return game.Tile{}, ErrInputFormat
	}
	label := strings.ToLower(fields[0])
	color, ok := canonicalColor(label)
	if !ok {
		color, ok = p.aliases[label]
	}
	if !ok {
		return game.Tile{}, ErrColor
	}
	number, err := strconv.Atoi(fields[1])
	if err != nil || number < game.MinNumber || number > game.MaxNumber || strconv.Itoa(number) != fields[1] {
		return game.Tile{}, ErrNumber
	}
	return game.Tile{Color: color, Number: number}, nil
}

func parserFromAliases(spec string) (Parser, error) {
	aliases := make(map[string]game.Color)
	if strings.TrimSpace(spec) != "" {
		for _, pair := range strings.Split(spec, ",") {
			parts := strings.Split(pair, "=")
			if len(parts) != 2 {
				return Parser{}, fmt.Errorf("%w: expected label=color", ErrColorAlias)
			}
			label := strings.ToLower(strings.TrimSpace(parts[0]))
			color := game.Color(strings.ToLower(strings.TrimSpace(parts[1])))
			if prior, exists := aliases[label]; exists && prior != color {
				return Parser{}, fmt.Errorf("%w: conflicting label %q", ErrColorAlias, label)
			}
			aliases[label] = color
		}
	}
	return NewParser(aliases)
}
