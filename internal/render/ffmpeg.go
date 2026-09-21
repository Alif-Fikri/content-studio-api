package render

import (
	"fmt"
	"strings"

	"github.com/alchemist/content-studio-api/internal/content"
)

const (
	outputWidth  = 1080
	outputHeight = 1920
)

func BuildFFmpegArgs(inputPath, backgroundAudioPath, outputPath string, script []content.ScriptBeat) []string {
	args := []string{
		"-y",
		"-i", inputPath,
		"-i", backgroundAudioPath,
	}

	var filter strings.Builder
	fmt.Fprintf(&filter,
		"[0:v]scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d[base]",
		outputWidth, outputHeight, outputWidth, outputHeight,
	)

	stage := "base"
	for i, beat := range script {
		next := fmt.Sprintf("v%d", i)
		fmt.Fprintf(&filter,
			";[%s]drawtext=text='%s':enable='between(t,%f,%f)':fontcolor=white:fontsize=64:x=(w-text_w)/2:y=h*0.75[%s]",
			stage, escapeDrawtext(beat.Text), beat.StartSeconds, beat.StartSeconds+beat.DurationSeconds, next,
		)
		stage = next
	}

	filter.WriteString(";[0:a][1:a]amix=inputs=2:duration=first:weights=1 0.2[aout]")

	args = append(args,
		"-filter_complex", filter.String(),
		"-map", "["+stage+"]",
		"-map", "[aout]",
		outputPath,
	)

	return args
}

func escapeDrawtext(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ":", "\\:")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}
