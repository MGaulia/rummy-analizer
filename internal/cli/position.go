package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/MGaulia/rummy-analizer/internal/app"
	"github.com/MGaulia/rummy-analizer/internal/game"
)

func runPosition(operation string, args []string, input io.Reader, output io.Writer) error {
	flags := flag.NewFlagSet(operation, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	edition := flags.Int("edition", 0, "required tile pool: 4 or 6")
	opening := flags.Bool("opening-required", true, "require rack-only 30-point opening")
	opened := flags.Bool("opened", false, "player opened on an earlier turn")
	drawn := flags.Bool("drawn", false, "player has already drawn on this turn")
	nodes := flags.Int("max-nodes", 200000, "search work budget including candidate generation")
	timeout := flags.Duration("time-limit", 5*time.Second, "search deadline")
	format := flags.String("format", "json", "json or svg")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err = fmt.Fprintln(output, "Usage: rummy-analizer "+operation+" --edition=4|6 [--opening-required=false] [--opened] [--max-nodes=200000] [--time-limit=5s] [--format=json|svg]\nRead a JSON position from stdin; check-move reads {position,move}.")
			return err
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	if err := game.Edition(*edition).Validate(); err != nil {
		return err
	}
	if *nodes <= 0 || *timeout <= 0 {
		return errors.New("max-nodes and time-limit must be positive")
	}
	if *format != "json" && *format != "svg" {
		return errors.New("format must be json or svg")
	}
	request := app.Request{Operation: operation, Settings: app.Settings{Edition: game.Edition(*edition), OpeningRequired: opening}, Turn: game.Turn{Opened: *opened, Drawn: *drawn}, MaxNodes: *nodes}
	if operation == "check-move" {
		var wire struct {
			Position game.Position `json:"position"`
			Move     game.Move     `json:"move"`
		}
		if err := app.Decode(input, &wire); err != nil {
			return err
		}
		request.Position, request.Move = wire.Position, &wire.Move
	} else {
		if err := app.Decode(input, &request.Position); err != nil {
			return err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	response := app.Execute(ctx, request)
	if *format == "svg" {
		if response.SVG == "" {
			return fmt.Errorf("no renderable result: %s", response.Error)
		}
		_, err := io.WriteString(output, response.SVG)
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		return err
	}
	if response.Error != "" {
		return errors.New(response.Error)
	}
	if response.Move != nil && !response.Move.Valid {
		return fmt.Errorf("move rejected: %s", response.Move.Code)
	}
	return nil
}
