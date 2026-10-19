package game

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func piece(id string, color Color, number int) Piece {
	return Piece{ID: id, Tile: numbered(color, number)}
}
func jokerPiece(id string) Piece { return Piece{ID: id, Tile: Tile{Kind: Joker}} }
func TestPositionValidation(t *testing.T) {
	base := Position{Table: [][]Piece{{piece("t1", Blue, 1), jokerPiece("j"), piece("t3", Blue, 3)}}, Rack: []Piece{piece("r", Blue, 1)}}
	before := ClonePosition(base)
	got := ValidatePosition(base, FourPlayer)
	if !got.Valid || got.Melds[0].Score != 6 || got.Resolved[0][1].As == nil || *got.Resolved[0][1].As != numbered(Blue, 2) {
		t.Fatalf("valid photo: %+v", got)
	}
	got.Resolved[0][0].ID = "mutated"
	got.Resolved[0][1].As.Number = 13
	if !reflect.DeepEqual(base, before) {
		t.Fatal("report aliases original position")
	}
	for _, tc := range []struct {
		name   string
		change func(*Position)
		code   string
	}{
		{"duplicate ID", func(p *Position) { p.Rack[0].ID = "t1" }, "duplicate_id"},
		{"missing ID", func(p *Position) { p.Rack[0].ID = "" }, "invalid_id"},
		{"invalid tile", func(p *Position) { p.Rack[0].Tile.Number = 0 }, "invalid_tile"},
		{"empty meld", func(p *Position) { p.Table = append(p.Table, []Piece{}) }, "invalid_size"},
		{"wrong order", func(p *Position) { p.Table[0][0], p.Table[0][1] = p.Table[0][1], p.Table[0][0] }, "invalid_order"},
		{"combined copies", func(p *Position) { p.Rack = append(p.Rack, piece("r2", Blue, 1)) }, "excess_copies"},
		{"bound rack", func(p *Position) { r := numbered(Red, 2); p.Rack[0].As = &r }, "invalid_joker_role"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := ClonePosition(base)
			tc.change(&p)
			report := ValidatePosition(p, FourPlayer)
			found := false
			for _, i := range report.Issues {
				found = found || i.Code == tc.code
			}
			if report.Valid || !found || report.Resolved != nil {
				t.Fatalf("wrong report %+v", report)
			}
		})
	}
	p := ClonePosition(base)
	p.Rack = append(p.Rack, piece("r2", Blue, 1), jokerPiece("j2"), jokerPiece("j3"), jokerPiece("j4"))
	if !ValidatePosition(p, SixPlayer).Valid || ValidatePosition(p, FourPlayer).Valid {
		t.Fatal("edition ignored")
	}
	if !ValidatePosition(Position{}, FourPlayer).Valid || ValidatePosition(Position{}, 5).Valid {
		t.Fatal("empty or invalid edition mishandled")
	}
	p = Position{}
	for i := 0; i < 161; i++ {
		p.Rack = append(p.Rack, piece(fmt.Sprint(i), Blue, 1))
	}
	if ValidatePosition(p, SixPlayer).Valid {
		t.Fatal("oversized position")
	}
}

func TestTileJSON(t *testing.T) {
	for _, tile := range []Tile{numbered(Blue, 13), {Kind: Joker}} {
		data, err := json.Marshal(tile)
		if err != nil {
			t.Fatal(err)
		}
		var roundTrip Tile
		if err := json.Unmarshal(data, &roundTrip); err != nil || roundTrip != tile {
			t.Fatalf("%s: %+v %v", data, roundTrip, err)
		}
	}
	for _, text := range []string{`null`, `[]`, `{"number":"1"}`, `{"kind":1}`} {
		var tile Tile
		if json.Unmarshal([]byte(text), &tile) == nil {
			t.Fatalf("accepted %s", text)
		}
	}
	if _, err := json.Marshal(Tile{}); err == nil {
		t.Fatal("serialized invalid tile")
	}
	var tile Tile
	if err := json.Unmarshal([]byte(`{"joker":true,"number":4}`), &tile); err != nil || tile.Validate() == nil {
		t.Fatal("malformed joker was silently repaired")
	}
}
