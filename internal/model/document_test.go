package model_test

import (
	"testing"
	"time"

	"github.com/stoneafk/issue2md/internal/model"
)

func TestDocumentDataCanBeConstructedSafely(t *testing.T) {
	now := time.Date(2026, 4, 28, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		doc  model.DocumentData
	}{
		{
			name: "github issue",
			doc: model.DocumentData{
				Provider: model.ProviderGitHub,
				Kind:     model.KindIssue,
				Issue: &model.IssueData{
					Provider:  model.ProviderGitHub,
					Title:     "Issue title",
					URL:       "https://github.com/owner/repo/issues/1",
					Author:    model.UserData{Login: "octocat", URL: "https://github.com/octocat"},
					CreatedAt: now,
					State:     "open",
					Body:      "Issue body",
					Comments: []model.CommentData{
						{
							ID:        "comment-1",
							Author:    model.UserData{Login: "hubot"},
							CreatedAt: now.Add(time.Hour),
							Body:      "Comment body",
							URL:       "https://github.com/owner/repo/issues/1#issuecomment-1",
							Reactions: model.ReactionSummary{Total: 1, Heart: 1},
						},
					},
					Reactions: model.ReactionSummary{Total: 2, PlusOne: 1, Rocket: 1},
				},
			},
		},
		{
			name: "pull request",
			doc: model.DocumentData{
				Provider: model.ProviderGitHub,
				Kind:     model.KindPullRequest,
				PullRequest: &model.PullRequestData{
					Title:          "PR title",
					URL:            "https://github.com/owner/repo/pull/2",
					Author:         model.UserData{Login: "octocat"},
					CreatedAt:      now,
					State:          "open",
					Body:           "PR body",
					ReviewComments: []model.CommentData{{ID: "review-1", Body: "Review comment"}},
					Reactions:      model.ReactionSummary{Total: 1, Eyes: 1},
				},
			},
		},
		{
			name: "discussion",
			doc: model.DocumentData{
				Provider: model.ProviderGitHub,
				Kind:     model.KindDiscussion,
				Discussion: &model.DiscussionData{
					Title:          "Discussion title",
					URL:            "https://github.com/owner/repo/discussions/3",
					Author:         model.UserData{Login: "octocat"},
					CreatedAt:      now,
					State:          "answered",
					Body:           "Discussion body",
					Comments:       []model.CommentData{{ID: "discussion-comment-1", Body: "Discussion comment"}},
					AcceptedAnswer: &model.CommentData{ID: "answer-1", Body: "Accepted answer"},
					Reactions:      model.ReactionSummary{Total: 1, Hooray: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doc.Provider == "" {
				t.Fatal("provider should be constructible")
			}
			if tt.doc.Kind == "" {
				t.Fatal("kind should be constructible")
			}
		})
	}
}

func TestZeroValuesAreSafe(t *testing.T) {
	var doc model.DocumentData
	var issue model.IssueData
	var pr model.PullRequestData
	var discussion model.DiscussionData
	var comment model.CommentData
	var reactions model.ReactionSummary

	if doc.Issue != nil || doc.PullRequest != nil || doc.Discussion != nil {
		t.Fatal("zero-value document should not contain content pointers")
	}
	if len(issue.Comments) != 0 || len(pr.ReviewComments) != 0 || len(discussion.Comments) != 0 {
		t.Fatal("zero-value comment slices should be empty")
	}
	if discussion.AcceptedAnswer != nil {
		t.Fatal("zero-value discussion should not have accepted answer")
	}
	if comment.Reactions != reactions {
		t.Fatal("zero-value comment reactions should be safe")
	}
}
