package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/MGaulia/rummy-analizer/model"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name, input, want string
		count             int
	}{
		{"empty EOF", "", "Available colors:", 0},
		{"blank lines", "\n \t\n", "Available colors:", 0},
		{"piped lines", "grey 1\ngrey 2\ngrey 3\n", "[grey 1, grey 2, grey 3]\nValid group/run: true", 3},
		{"unterminated last line", "grey 1\ngrey 2\ngrey 3", "[grey 1, grey 2, grey 3]\nValid group/run: true", 3},
		{"CRLF", "grey 1\r\ngrey 2\r\ngrey 3\r\n", "[grey 1, grey 2, grey 3]\nValid group/run: true", 3},
		{"recover after invalid input", "grey1\ngrey 1\nred 7\ngrey 2\ngrey 3\n", "[grey 1, grey 2, grey 3]\nValid group/run: true", 3},
		{"long run", "grey 1\ngrey 2\ngrey 3\ngrey 4\ngrey 5\n", "[grey 1, grey 2, grey 3, grey 4, grey 5]\nValid group/run: true", 5},
		{"Rummikub no wrap", "grey 12\ngrey 13\ngrey 1\n", "[grey 12, grey 13, grey 1]\nValid group/run: false", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := run(strings.NewReader(tt.input), &output); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), tt.want) {
				t.Fatalf("output %q does not contain %q", output.String(), tt.want)
			}
			if got := strings.Count(output.String(), "Valid group/run:"); got != tt.count {
				t.Fatalf("processed %d tiles, want %d; output %q", got, tt.count, output.String())
			}
		})
	}
}

var errIO = errors.New("test I/O failure")

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errIO }

type failingWriter struct{ writesLeft int }

func (w *failingWriter) Write(p []byte) (int, error) {
	if w.writesLeft == 0 {
		return 0, errIO
	}
	w.writesLeft--
	return len(p), nil
}

func TestRunIOErrors(t *testing.T) {
	input := io.MultiReader(strings.NewReader("grey 1\n"), failingReader{})
	if err := run(input, io.Discard); !errors.Is(err, errIO) {
		t.Fatalf("read error = %v, want %v", err, errIO)
	}
	for _, input := range []string{"grey 1\n", "invalid\n"} {
		for writesLeft := 0; writesLeft <= 2; writesLeft++ {
			err := run(strings.NewReader(input), &failingWriter{writesLeft: writesLeft})
			if !errors.Is(err, errIO) {
				t.Fatalf("write error after %d writes = %v, want %v", writesLeft, err, errIO)
			}
		}
	}
	if err := run(strings.NewReader(strings.Repeat("x", 70_000)), io.Discard); err == nil {
		t.Fatal("expected scanner error for an oversized line")
	}
}

func TestFormatTiles(t *testing.T) {
	tests := []struct {
		tiles []model.Card
		want  string
	}{
		{nil, "[]"},
		{[]model.Card{{Color: model.COLOR_GREY, Number: 1}}, "[grey 1]"},
		{[]model.Card{{Color: model.COLOR_GREY, Number: 1}, {Color: model.COLOR_PURPLE, Number: 13}}, "[grey 1, purple 13]"},
	}
	for _, tt := range tests {
		if got := formatTiles(tt.tiles); got != tt.want {
			t.Errorf("formatTiles() = %q, want %q", got, tt.want)
		}
	}
}
