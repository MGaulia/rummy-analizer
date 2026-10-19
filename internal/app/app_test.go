package app

import (
	"context"
	"strings"
	"testing"

	"github.com/MGaulia/rummy-analizer/internal/game"
)

func TestSharedOfflineRequest(t *testing.T) {
	var request Request
	raw := `{"operation":"solve","settings":{"edition":6},"position":{"table":[],"rack":[{"id":"r","tile":{"color":"red","number":10}},{"id":"b","tile":{"color":"blue","number":10}},{"id":"j","tile":{"joker":true}}]}}`
	if err := Decode(strings.NewReader(raw), &request); err != nil {
		t.Fatal(err)
	}
	response := Execute(context.Background(), request)
	if response.Error != "" || response.Search.Status != game.Optimal || response.Search.Verified.OpeningScore != 30 || !strings.Contains(response.SVG, "Joker as") {
		t.Fatalf("bad response %+v", response)
	}
	request.Position.Rack[0].Tile.Number = 1
	response = Execute(context.Background(), request)
	if response.Search.Status != game.NoMove {
		t.Fatalf("default opening skipped %+v", response)
	}
	request.Settings.Edition = 0
	if Execute(context.Background(), request).Error == "" {
		t.Fatal("guessed edition")
	}
}

func TestStrictRequests(t *testing.T) {
	for _, raw := range []string{`{"unexpected":true}`, `{} {}`, `{"position":{"rack":[{"id":"x","tile":{"colour":"blue"}}]}}`, strings.Repeat(" ", MaxRequestBytes+1)} {
		var request Request
		if Decode(strings.NewReader(raw), &request) == nil {
			t.Fatalf("accepted malformed input %.100s", raw)
		}
	}
}
