package claude

import "fmt"

var productDescriptions = map[string]string{
	"beres":         "Beres, a personal finance app that helps users track debts owed to and by them, with automatic reminders.",
	"kanji_widget":  "Kanji Widget, a home-screen widget app for learning Japanese kanji through daily exposure.",
}

const targetScriptSecondsMin = 10
const targetScriptSecondsMax = 12

func BuildPrompt(product, brief string) string {
	description, ok := productDescriptions[product]
	if !ok {
		description = product
	}

	return fmt.Sprintf(`You are writing a short vertical video ad script and caption for %s.

Brief: %s

Write a hook -> beats -> CTA script totaling %d-%d seconds, split into short on-screen text beats with timing, plus a caption for the post.`,
		description, brief, targetScriptSecondsMin, targetScriptSecondsMax)
}
