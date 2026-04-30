package converter

import (
	"strings"

	"github.com/stoneafk/issue2md/internal/model"
)

func renderReactions(summary model.ReactionSummary, enabled bool) string {
	if !enabled || summary.Total == 0 {
		return ""
	}

	parts := make([]string, 0, 8)
	if summary.PlusOne > 0 {
		parts = append(parts, "+1")
	}
	if summary.MinusOne > 0 {
		parts = append(parts, "-1")
	}
	if summary.Laugh > 0 {
		parts = append(parts, "laugh")
	}
	if summary.Hooray > 0 {
		parts = append(parts, "hooray")
	}
	if summary.Confused > 0 {
		parts = append(parts, "confused")
	}
	if summary.Heart > 0 {
		parts = append(parts, "heart")
	}
	if summary.Rocket > 0 {
		parts = append(parts, "rocket")
	}
	if summary.Eyes > 0 {
		parts = append(parts, "eyes")
	}
	if len(parts) == 0 {
		return ""
	}
	return "Reactions: " + strings.Join(parts, ", ") + "\n"
}
