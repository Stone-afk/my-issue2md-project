package converter

import (
	"strings"
	"testing"

	"github.com/stoneafk/issue2md/internal/model"
)

func TestRenderReactions(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		doc  model.DocumentData
		want []string
		not  []string
	}{
		{
			name: "reactions omitted by default",
			opts: Options{},
			doc: model.DocumentData{
				Provider: model.ProviderGitHub,
				Kind:     model.KindIssue,
				Issue: &model.IssueData{
					Provider:  model.ProviderGitHub,
					Title:     "Issue",
					URL:       "https://github.com/OWNER/REPO/issues/1",
					Author:    model.UserData{Login: "octocat"},
					Body:      "Body",
					Reactions: model.ReactionSummary{Total: 2, PlusOne: 1, Heart: 1},
					Comments:  []model.CommentData{{Body: "Comment", Reactions: model.ReactionSummary{Total: 1, Rocket: 1}}},
				},
			},
			not: []string{"+1", "heart", "rocket"},
		},
		{
			name: "reactions rendered when enabled",
			opts: Options{EnableReactions: true},
			doc: model.DocumentData{
				Provider: model.ProviderGitHub,
				Kind:     model.KindIssue,
				Issue: &model.IssueData{
					Provider:  model.ProviderGitHub,
					Title:     "Issue",
					URL:       "https://github.com/OWNER/REPO/issues/1",
					Author:    model.UserData{Login: "octocat"},
					Body:      "Body",
					Reactions: model.ReactionSummary{Total: 2, PlusOne: 1, Heart: 1},
					Comments:  []model.CommentData{{Body: "Comment", Reactions: model.ReactionSummary{Total: 1, Rocket: 1}}},
				},
			},
			want: []string{"+1", "heart", "rocket"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewRenderer(tt.opts)
			got, err := renderer.Render(tt.doc)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			for _, fragment := range tt.want {
				if !strings.Contains(got, fragment) {
					t.Fatalf("expected fragment %q in output %q", fragment, got)
				}
			}
			for _, fragment := range tt.not {
				if strings.Contains(got, fragment) {
					t.Fatalf("did not expect fragment %q in output %q", fragment, got)
				}
			}
		})
	}
}
