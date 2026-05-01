package converter

import (
	"strings"
	"testing"
	"time"

	"github.com/stoneafk/issue2md/internal/model"
)

func TestRenderIssuePullRequestAndDiscussion(t *testing.T) {
	renderer := NewRenderer(Options{})

	t.Run("issue rendering", func(t *testing.T) {
		tests := []struct {
			name      string
			doc       model.DocumentData
			want      []string
			wantOrder []string
			imageURL  string
		}{
			{
				name: "github issue",
				doc: model.DocumentData{
					Provider: model.ProviderGitHub,
					Kind:     model.KindIssue,
					Issue: &model.IssueData{
						Provider:  model.ProviderGitHub,
						Title:     "GitHub Issue",
						URL:       "https://github.com/OWNER/REPO/issues/1",
						Author:    model.UserData{Login: "octocat"},
						CreatedAt: time.Date(2026, 4, 28, 10, 30, 0, 0, time.UTC),
						State:     "open",
						Body:      "Issue body with ![image](https://example.com/image.png)",
						Comments:  []model.CommentData{{Author: model.UserData{Login: "hubot"}, Body: "First comment"}},
					},
				},
				want: []string{
					"# GitHub Issue",
					"## Summary / 摘要",
					"## Structured Notes / 结构化笔记",
					"## Raw Archive / 原始归档",
					"### Comment by hubot",
					"Issue body with ![image](https://example.com/image.png)",
					"First comment",
					"- State: open",
				},
				wantOrder: []string{
					"## Summary / 摘要",
					"## Structured Notes / 结构化笔记",
					"## Raw Archive / 原始归档",
					"First comment",
				},
				imageURL: "https://example.com/image.png",
			},
			{
				name: "gitlab issue",
				doc: model.DocumentData{
					Provider: model.ProviderGitLab,
					Kind:     model.KindIssue,
					Issue: &model.IssueData{
						Provider:  model.ProviderGitLab,
						Title:     "GitLab Issue",
						URL:       "https://gitlab.com/GROUP/PROJECT/-/issues/2",
						Author:    model.UserData{Login: "gitlab-user"},
						CreatedAt: time.Date(2026, 4, 28, 11, 30, 0, 0, time.UTC),
						State:     "opened",
						Body:      "GitLab issue body",
						Comments:  []model.CommentData{{Author: model.UserData{Login: "note-user"}, Body: "Note body"}},
					},
				},
				want: []string{
					"# GitLab Issue",
					"## Summary / 摘要",
					"## Structured Notes / 结构化笔记",
					"## Raw Archive / 原始归档",
					"### Comment by note-user",
					"GitLab issue body",
					"Note body",
					"- State: opened",
				},
				wantOrder: []string{
					"## Summary / 摘要",
					"## Structured Notes / 结构化笔记",
					"## Raw Archive / 原始归档",
					"Note body",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := renderer.Render(tt.doc)
				if err != nil {
					t.Fatalf("Render() error = %v", err)
				}

				for _, fragment := range tt.want {
					if !strings.Contains(got, fragment) {
						t.Fatalf("expected fragment %q in output %q", fragment, got)
					}
				}
				if tt.imageURL != "" && !strings.Contains(got, tt.imageURL) {
					t.Fatalf("expected image link to remain unchanged in output %q", got)
				}

				lastIndex := -1
				for _, fragment := range tt.wantOrder {
					index := strings.Index(got, fragment)
					if index < 0 {
						t.Fatalf("expected ordered fragment %q in output %q", fragment, got)
					}
					if index <= lastIndex {
						t.Fatalf("expected ordered fragments %v in output %q", tt.wantOrder, got)
					}
					lastIndex = index
				}
			})
		}
	})

	t.Run("non-issue rendering", func(t *testing.T) {
		tests := []struct {
			name      string
			doc       model.DocumentData
			want      []string
			not       []string
			wantOrder []string
		}{
			{
				name: "pull request rendering",
				doc: model.DocumentData{
					Provider: model.ProviderGitHub,
					Kind:     model.KindPullRequest,
					PullRequest: &model.PullRequestData{
						Title:          "PR Title",
						URL:            "https://github.com/OWNER/REPO/pull/2",
						Author:         model.UserData{Login: "octocat"},
						CreatedAt:      time.Date(2026, 4, 28, 12, 30, 0, 0, time.UTC),
						State:          "open",
						Body:           "PR description",
						ReviewComments: []model.CommentData{{Author: model.UserData{Login: "reviewer"}, Body: "Review comment"}},
					},
				},
				want: []string{
					"# PR Title",
					"## Summary / 摘要",
					"## Structured Notes / 结构化笔记",
					"## Raw Archive / 原始归档",
					"PR description",
					"### Review Comment by reviewer",
					"Review comment",
				},
				not: []string{"diff", "commit history"},
				wantOrder: []string{
					"## Summary / 摘要",
					"## Structured Notes / 结构化笔记",
					"## Raw Archive / 原始归档",
					"Review comment",
				},
			},
			{
				name: "discussion rendering with accepted answer marker",
				doc: model.DocumentData{
					Provider: model.ProviderGitHub,
					Kind:     model.KindDiscussion,
					Discussion: &model.DiscussionData{
						Title:          "Discussion Title",
						URL:            "https://github.com/OWNER/REPO/discussions/3",
						Author:         model.UserData{Login: "octocat"},
						CreatedAt:      time.Date(2026, 4, 28, 13, 30, 0, 0, time.UTC),
						State:          "answered",
						Body:           "Discussion body",
						Comments:       []model.CommentData{{Author: model.UserData{Login: "commenter"}, Body: "Discussion comment"}},
						AcceptedAnswer: &model.CommentData{Author: model.UserData{Login: "helper"}, Body: "Accepted answer body"},
					},
				},
				want: []string{
					"# Discussion Title",
					"## Summary / 摘要",
					"## Structured Notes / 结构化笔记",
					"## Raw Archive / 原始归档",
					"Discussion body",
					"### Accepted answer",
					"Accepted answer body",
					"### Comment by commenter",
					"Discussion comment",
				},
				wantOrder: []string{
					"## Summary / 摘要",
					"## Structured Notes / 结构化笔记",
					"### Accepted answer",
					"## Raw Archive / 原始归档",
					"Discussion comment",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
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
					if strings.Contains(strings.ToLower(got), fragment) {
						t.Fatalf("did not expect fragment %q in output %q", fragment, got)
					}
				}

				lastIndex := -1
				for _, fragment := range tt.wantOrder {
					index := strings.Index(got, fragment)
					if index < 0 {
						t.Fatalf("expected ordered fragment %q in output %q", fragment, got)
					}
					if index <= lastIndex {
						t.Fatalf("expected ordered fragments %v in output %q", tt.wantOrder, got)
					}
					lastIndex = index
				}
			})
		}
	})
}
