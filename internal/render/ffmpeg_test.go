package render

import (
	"strings"
	"testing"

	"github.com/alchemist/content-studio-api/internal/content"
)

func TestBuildFFmpegArgs_NoBeats(t *testing.T) {
	args := BuildFFmpegArgs("in.mp4", "bg.mp3", "out.mp4", nil)

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-i in.mp4") {
		t.Errorf("expected input path in args, got: %s", joined)
	}
	if !strings.Contains(joined, "-i bg.mp3") {
		t.Errorf("expected background audio path in args, got: %s", joined)
	}
	if !strings.Contains(joined, "-map [base]") {
		t.Errorf("expected final video stage to be [base] with no beats, got: %s", joined)
	}
	if args[len(args)-1] != "out.mp4" {
		t.Errorf("expected output path as last arg, got: %s", args[len(args)-1])
	}
}

func TestBuildFFmpegArgs_WithBeats(t *testing.T) {
	script := []content.ScriptBeat{
		{Text: "hello", StartSeconds: 0, DurationSeconds: 2},
		{Text: "world", StartSeconds: 2, DurationSeconds: 3},
	}

	args := BuildFFmpegArgs("in.mp4", "bg.mp3", "out.mp4", script)
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "text='hello'") {
		t.Errorf("expected first beat text in filter, got: %s", joined)
	}
	if !strings.Contains(joined, "text='world'") {
		t.Errorf("expected second beat text in filter, got: %s", joined)
	}
	if !strings.Contains(joined, "between(t,0.000000,2.000000)") {
		t.Errorf("expected first beat timing window, got: %s", joined)
	}
	if !strings.Contains(joined, "-map [v1]") {
		t.Errorf("expected final video stage to be last beat's output [v1], got: %s", joined)
	}
}

func TestEscapeDrawtext(t *testing.T) {
	cases := map[string]string{
		"plain text": "plain text",
		"a:b":        `a\:b`,
		"it's":       `it\'s`,
		`back\slash`: `back\\slash`,
	}

	for input, want := range cases {
		if got := escapeDrawtext(input); got != want {
			t.Errorf("escapeDrawtext(%q) = %q, want %q", input, got, want)
		}
	}
}
