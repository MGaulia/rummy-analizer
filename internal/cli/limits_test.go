package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestRunCollectionLimit(t *testing.T) {
	for _, count := range []int{12, 13, 14, 1000} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			var output bytes.Buffer
			err := Run(nil, strings.NewReader(strings.Repeat("blue 1\n", count)), &output)
			wantResults := count
			if count > 13 {
				wantResults = 13
				if err == nil || err.Error() != "single-collection input is limited to 13 tiles; start a new process for another collection" {
					t.Fatalf("Run with %d tiles: error = %v, want collection-limit error", count, err)
				}
			} else if err != nil {
				t.Fatalf("Run with %d tiles: %v", count, err)
			}
			if got := strings.Count(output.String(), "Can form group/run:"); got != wantResults {
				t.Errorf("processed %d tiles, want %d", got, wantResults)
			}
		})
	}
}

func TestRunFullCollectionAllowsIgnoredLines(t *testing.T) {
	var input strings.Builder
	for number := 1; number <= 13; number++ {
		fmt.Fprintf(&input, "blue %d\n", number)
	}
	// Invalid and blank lines do not consume collection capacity, even at EOF.
	input.WriteString("\n \t\nnot-a-tile")
	var output bytes.Buffer
	if err := Run(nil, strings.NewReader(input.String()), &output); err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(output.String(), "Can form group/run:"); got != 13 {
		t.Fatalf("processed %d tiles, want 13", got)
	}
	if !strings.Contains(output.String(), "blue 12, blue 13]\nCan form group/run: true") {
		t.Fatal("a complete 13-tile run did not validate")
	}
}

func TestRunCollectionLimitCountsAliasesAndJokers(t *testing.T) {
	input := strings.Repeat("myblue 1\n", 12) + "joker\nred 13"
	var output bytes.Buffer
	err := Run([]string{"-color-aliases", "myblue=blue"}, strings.NewReader(input), &output)
	if err == nil {
		t.Fatal("fourteenth parsed tile was accepted")
	}
	if !strings.Contains(output.String(), "Can form group/run: unsupported") {
		t.Fatal("joker status within the limit was lost")
	}
	if strings.Contains(output.String(), "red 13") {
		t.Fatal("fourteenth tile was appended or printed")
	}
}
