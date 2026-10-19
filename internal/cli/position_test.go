package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/MGaulia/rummy-analizer/internal/app"
	"github.com/MGaulia/rummy-analizer/internal/game"
)

const openingPosition = `{"table":[],"rack":[{"id":"b","tile":{"color":"blue","number":10}},{"id":"r","tile":{"color":"red","number":10}},{"id":"j","tile":{"joker":true}}]}`

func TestPositionCommands(t *testing.T) {
	for _, edition := range []string{"4", "6"} {
		var output bytes.Buffer
		if err := Run([]string{"solve", "--edition=" + edition}, strings.NewReader(openingPosition), &output); err != nil {
			t.Fatal(err)
		}
		var response app.Response
		if err := json.Unmarshal(output.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.Search.Status != game.Optimal || len(response.Search.Verified.Played) != 3 || response.SVG == "" {
			t.Fatalf("wrong result %+v", response)
		}
	}
	for _, args := range [][]string{{"solve"}, {"solve", "--edition=5"}, {"solve", "--edition=4", "--max-nodes=0"}, {"solve", "--edition=4", "--time-limit=0s"}, {"validate", "--edition=4", "extra"}} {
		var out bytes.Buffer
		if Run(args, strings.NewReader(openingPosition), &out) == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	var out bytes.Buffer
	if err := Run([]string{"validate", "--edition=4", "--format=svg"}, strings.NewReader(openingPosition), &out); err != nil || !strings.HasPrefix(out.String(), "<svg") {
		t.Fatalf("validate svg: %v", err)
	}
}

func TestOpeningFlagsThroughCLI(t *testing.T) {
	input := `{"table":[[{"id":"a","tile":{"color":"blue","number":1}},{"id":"b","tile":{"color":"blue","number":2}},{"id":"c","tile":{"color":"blue","number":3}}]],"rack":[{"id":"d","tile":{"color":"blue","number":4}}]}`
	for _, tc := range []struct {
		extra []string
		want  game.SearchStatus
	}{{nil, game.NoMove}, {[]string{"--opened"}, game.Optimal}, {[]string{"--opening-required=false"}, game.Optimal}, {[]string{"--opened=false", "--opening-required=false"}, game.Optimal}, {[]string{"--drawn"}, game.NoMove}} {
		var out bytes.Buffer
		args := append([]string{"solve", "--edition=4"}, tc.extra...)
		if err := Run(args, strings.NewReader(input), &out); err != nil {
			t.Fatal(err)
		}
		var result app.Response
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Search.Status != tc.want {
			t.Fatalf("%v: got %+v want %s", args, result.Search, tc.want)
		}
	}
}
