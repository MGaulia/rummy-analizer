package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/MGaulia/rummy-analizer/check"
	"github.com/MGaulia/rummy-analizer/mapper"
	"github.com/MGaulia/rummy-analizer/model"
)

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run reads one numbered tile per line and checks the accumulated tiles as a
// single group/run. It is not a validator for a table containing several groups.
func run(input io.Reader, output io.Writer) error {
	if _, err := fmt.Fprintln(output, "Available colors: grey, black, green, purple"); err != nil {
		return err
	}
	scanner := bufio.NewScanner(input)
	var tiles []model.Card
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
		card, err := mapper.StringToCard(line)
		if err != nil {
			if _, writeErr := fmt.Fprintln(output, err); writeErr != nil {
				return writeErr
			}
			continue
		}
		tiles = append(tiles, card)
		if _, err := fmt.Fprintf(output, "%s\nValid group/run: %t\n", formatTiles(tiles), check.IsCombination(tiles)); err != nil {
			return err
		}
	}
}

func formatTiles(tiles []model.Card) string {
	var result strings.Builder
	result.WriteByte('[')
	for i, tile := range tiles {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(string(tile.Color))
		result.WriteByte(' ')
		result.WriteString(strconv.Itoa(int(tile.Number)))
	}
	result.WriteByte(']')
	return result.String()
}
