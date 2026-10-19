// Package cli adapts text and process arguments to the game domain.
package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/MGaulia/rummy-analizer/internal/game"
)

const usage = "Usage: rummy-analizer [-color-aliases label=color,...]\nEnter one tile per line: <color> <number>, or joker.\nColors: black, blue, orange, red. Checks whether one collection can form a meld, not its displayed order."

// Run processes one growing collection, not a table. Input errors are printed
// without adding a tile. A fourteenth parsed tile ends the session with an
// error before being added. I/O and argument errors are also returned.
func Run(args []string, input io.Reader, output io.Writer) error {
	flags := flag.NewFlagSet("rummy-analizer", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	aliases := flags.String("color-aliases", "", "explicit comma-separated label=color mappings")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err := fmt.Fprintln(output, usage)
			return err
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments; tiles are read from stdin")
	}
	parser, err := parserFromAliases(*aliases)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(output, "Available colors: black, blue, orange, red (or configured aliases); joker rules unsupported"); err != nil {
		return err
	}
	scanner := bufio.NewScanner(input)
	var tiles []game.Tile
	for {
		if _, err := fmt.Fprint(output, "Enter tile: "); err != nil {
			return err
		}
		if !scanner.Scan() {
			return scanner.Err()
		}
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		tile, err := parser.ParseTile(line)
		if err != nil {
			if _, writeErr := fmt.Fprintln(output, err); writeErr != nil {
				return writeErr
			}
			continue
		}
		// This CLI checks one collection, so retaining more tiles cannot help.
		if len(tiles) == game.MaxNumber {
			return fmt.Errorf("single-collection input is limited to %d tiles; start a new process for another collection", game.MaxNumber)
		}
		tiles = append(tiles, tile)
		result := game.AnalyzeMeld(tiles)
		status := strconv.FormatBool(result.Valid())
		if result.Issue.Code == game.UnsupportedJoker {
			status = "unsupported"
		}
		if _, err := fmt.Fprintf(output, "%s\nCan form group/run: %s\n", formatTiles(tiles), status); err != nil {
			return err
		}
		if !result.Valid() {
			if _, err := fmt.Fprintln(output, "Reason:", issueMessage(result.Issue)); err != nil {
				return err
			}
		}
	}
}

func formatTiles(tiles []game.Tile) string {
	var result strings.Builder
	result.WriteByte('[')
	for i, tile := range tiles {
		if i > 0 {
			result.WriteString(", ")
		}
		if tile.Kind == game.Joker {
			result.WriteString("joker")
			continue
		}
		result.WriteString(string(tile.Color))
		result.WriteByte(' ')
		result.WriteString(strconv.Itoa(tile.Number))
	}
	result.WriteByte(']')
	return result.String()
}

func issueMessage(issue game.ValidationIssue) string {
	switch issue.Code {
	case game.InvalidSize:
		return "a meld needs 3-13 tiles (an equal-number group has at most 4)"
	case game.InvalidTile:
		return fmt.Sprintf("tile %d has an invalid representation", issue.TileIndex+1)
	case game.UnsupportedJoker:
		return fmt.Sprintf("tile %d is a joker; joker meld rules are not implemented", issue.TileIndex+1)
	case game.DuplicateTile:
		return fmt.Sprintf("tile %d repeats tile %d; duplicates cannot share one meld", issue.TileIndex+1, issue.RelatedIndex+1)
	case game.RunGap:
		return "a run must contain consecutive numbers and cannot wrap from 13 to 1"
	case game.MixedMeld:
		return "tiles have neither one shared number nor one shared color"
	default:
		return "unrecognized validation result"
	}
}
