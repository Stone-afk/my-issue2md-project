package converter

import (
	"strings"
	"testing"
	"time"

	"github.com/stoneafk/issue2md/internal/model"
)

func TestRenderFrontmatter(t *testing.T) {
	tests := []struct {
		name                string
		doc                 model.DocumentData
		wantTitleFragment   string
		wantURLFragment     string
		wantAuthorFragment  string
		wantCreatedFragment string
	}{
		{
			name: "github issue frontmatter",
			doc: model.DocumentData{
				Provider: model.ProviderGitHub,
				Kind:     model.KindIssue,
				Issue: &model.IssueData{
					Provider:  model.ProviderGitHub,
					Title:     "GitHub Issue",
					URL:       "https://github.com/OWNER/REPO/issues/1",
					Author:    model.UserData{Login: "octocat"},
					CreatedAt: time.Date(2026, 4, 28, 10, 30, 0, 0, time.UTC),
				},
			},
			wantTitleFragment:   `title: "GitHub Issue"`,
			wantURLFragment:     `url: "https://github.com/OWNER/REPO/issues/1"`,
			wantAuthorFragment:  `author: "octocat"`,
			wantCreatedFragment: `created_at: "2026-04-28T10:30:00Z"`,
		},
		{
			name: "gitlab issue frontmatter",
			doc: model.DocumentData{
				Provider: model.ProviderGitLab,
				Kind:     model.KindIssue,
				Issue: &model.IssueData{
					Provider:  model.ProviderGitLab,
					Title:     "GitLab Issue",
					URL:       "https://gitlab.com/GROUP/PROJECT/-/issues/2",
					Author:    model.UserData{Login: "gitlab-user"},
					CreatedAt: time.Date(2026, 4, 28, 11, 30, 0, 0, time.UTC),
				},
			},
			wantTitleFragment:   `title: "GitLab Issue"`,
			wantURLFragment:     `url: "https://gitlab.com/GROUP/PROJECT/-/issues/2"`,
			wantAuthorFragment:  `author: "gitlab-user"`,
			wantCreatedFragment: `created_at: "2026-04-28T11:30:00Z"`,
		},
	}

	renderer := NewRenderer(Options{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderer.Render(tt.doc)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			for _, fragment := range []string{
				"---",
				tt.wantTitleFragment,
				tt.wantURLFragment,
				tt.wantAuthorFragment,
				tt.wantCreatedFragment,
			} {
				if !strings.Contains(got, fragment) {
					t.Fatalf("expected frontmatter fragment %q in output %q", fragment, got)
				}
			}
		})
	}
}
