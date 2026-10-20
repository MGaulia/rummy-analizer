// Package render produces exact diagrams from confirmed physical positions.
package render

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/MGaulia/rummy-analizer/internal/game"
)

func escape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

// SVG renders every occurrence once. It validates inventory/layout and resolves
// all joker roles before drawing. A played ID must refer to a final table piece.
func SVG(position game.Position, edition game.Edition, played []string) (string, error) {
	checked := game.ValidatePosition(position, edition)
	if !checked.Valid {
		return "", fmt.Errorf("cannot render invalid position: %v", checked.Issues)
	}
	position = game.ClonePosition(position)
	position.Table = checked.Resolved
	additions := map[string]bool{}
	tableIDs := map[string]bool{}
	for _, meld := range position.Table {
		for _, p := range meld {
			tableIDs[p.ID] = true
		}
	}
	for _, id := range played {
		if !tableIDs[id] || additions[id] {
			return "", fmt.Errorf("invalid played occurrence %q", id)
		}
		additions[id] = true
	}
	rackRows := (len(position.Rack) + 12) / 13
	if rackRows == 0 {
		rackRows = 1
	}
	height := 136 + len(position.Table)*100 + rackRows*100
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1020 %d" role="img" aria-label="Confirmed table and remaining rack"><rect width="1020" height="%d" fill="#eef3f0"/><style>text{font-family:system-ui,sans-serif} .heading{font-size:22px;fill:#162f24} .number{font-size:28px;font-weight:700} .caption{font-size:11px;fill:#253d31}</style>`, height, height)
	b.WriteString(`<text x="20" y="32" class="heading">Final table · green border = played from rack</text>`)
	draw := func(p game.Piece, x, y int) {
		tile := p.Tile
		joker := tile.Kind == game.Joker
		if joker && p.As != nil {
			tile = *p.As
		}
		color := map[game.Color]string{game.Black: "#17212b", game.Blue: "#064dd5", game.Orange: "#a44800", game.Red: "#c51e2f"}[tile.Color]
		stroke := "#9ba9a2"
		if additions[p.ID] {
			stroke = "#007a41"
		}
		label := fmt.Sprintf("%s %d", tile.Color, tile.Number)
		if joker {
			if p.As == nil {
				label = "Joker (unassigned on rack)"
			} else {
				label = "Joker as " + label
			}
		}
		roleColor, roleNumber := "", 0
		if joker && p.As != nil {
			roleColor, roleNumber = string(p.As.Color), p.As.Number
		}
		fmt.Fprintf(&b, `<g data-piece-id="%s" data-played="%t" data-tile-color="%s" data-tile-number="%d" data-joker="%t" data-role-color="%s" data-role-number="%d"><title>%s</title><rect x="%d" y="%d" width="68" height="76" rx="7" fill="#fffaf0" stroke="%s" stroke-width="3"/>`, escape(p.ID), additions[p.ID], p.Tile.Color, p.Tile.Number, joker, roleColor, roleNumber, escape(p.ID+": "+label), x, y, stroke)
		top := ""
		if joker {
			top = "JOKER"
		}
		numberLabel := fmt.Sprint(tile.Number)
		colorLabel := string(tile.Color)
		if joker && p.As == nil {
			numberLabel = "★"
			colorLabel = "unassigned"
			color = "#17212b"
		}
		fmt.Fprintf(&b, `<text x="%d" y="%d" text-anchor="middle" class="caption">%s</text><text x="%d" y="%d" text-anchor="middle" class="number" fill="%s">%s</text><text x="%d" y="%d" text-anchor="middle" class="caption">%s</text></g>`, x+34, y+14, top, x+34, y+45, color, numberLabel, x+34, y+64, escape(colorLabel))
	}
	y := 50
	for _, meld := range position.Table {
		for i, p := range meld {
			draw(p, 20+i*76, y)
		}
		y += 100
	}
	if len(position.Table) == 0 {
		b.WriteString(`<text x="20" y="64" class="caption">Empty table</text>`)
		y += 30
	}
	fmt.Fprintf(&b, `<text x="20" y="%d" class="heading">Remaining rack · %d tiles</text>`, y+18, len(position.Rack))
	y += 34
	for i, p := range position.Rack {
		draw(p, 20+(i%13)*76, y+(i/13)*100)
	}
	if len(position.Rack) == 0 {
		fmt.Fprintf(&b, `<text x="20" y="%d" class="caption">Rack empty</text>`, y+30)
	}
	b.WriteString(`</svg>`)
	return b.String(), nil
}
