package render

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/MGaulia/rummy-analizer/internal/game"
)

func TestExactOccurrenceDiagram(t *testing.T) {
	role := game.Tile{Color: game.Blue, Number: 3}
	position := game.Position{Table: [][]game.Piece{{{ID: `a"<&`, Tile: game.Tile{Color: game.Blue, Number: 1}}, {ID: "b", Tile: game.Tile{Color: game.Blue, Number: 2}}, {ID: "j", Tile: game.Tile{Kind: game.Joker}, As: &role}}}, Rack: []game.Piece{{ID: "rack-joker", Tile: game.Tile{Kind: game.Joker}}}}
	svg, err := SVG(position, 4, []string{"b"})
	if err != nil {
		t.Fatal(err)
	}
	decoder := xml.NewDecoder(strings.NewReader(svg))
	seen := map[string]int{}
	played := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		element, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		for _, attr := range element.Attr {
			if attr.Name.Local == "data-piece-id" {
				seen[attr.Value]++
			}
			if attr.Name.Local == "data-played" && attr.Value == "true" {
				played++
			}
		}
	}
	for _, id := range []string{`a"<&`, "b", "j", "rack-joker"} {
		if seen[id] != 1 {
			t.Fatalf("occurrence %q count %d", id, seen[id])
		}
	}
	if len(seen) != 4 || played != 1 || !strings.Contains(svg, "Joker as blue 3") || !strings.Contains(svg, "unassigned") {
		t.Fatal("missing exact labels or additions")
	}
	again, _ := SVG(position, 4, []string{"b"})
	if again != svg {
		t.Fatal("nondeterministic rendering")
	}
	if _, err := SVG(position, 4, []string{"rack-joker"}); err == nil {
		t.Fatal("rack marked played")
	}
	position.Table[0][0].Tile.Number = 7
	if _, err := SVG(position, 4, nil); err == nil {
		t.Fatal("rendered invalid layout")
	}
}
