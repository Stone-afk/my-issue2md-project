package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stoneafk/issue2md/internal/fetchprovider"
	"github.com/stoneafk/issue2md/internal/model"
	"github.com/stoneafk/issue2md/internal/parser"
)

func TestFetch(t *testing.T) {
	t.Run("server-backed cases", func(t *testing.T) {
		tests := []struct {
			name           string
			token          string
			target         parser.Target
			opts           model.FetchOptions
			handler        http.HandlerFunc
			wantErrParts   []string
			wantAuthHeader string
			check          func(t *testing.T, doc model.DocumentData)
		}{
			{
				name:   "issue fetch uses mock server",
				target: parser.Target{Provider: model.ProviderGitHub, Kind: model.KindIssue, Owner: "OWNER", Repo: "REPO", Number: 123, URL: "https://github.com/OWNER/REPO/issues/123"},
				opts:   model.FetchOptions{IncludeReactions: true},
				handler: func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/repos/OWNER/REPO/issues/123":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{
							"number": 123,
							"title": "Mock GitHub Issue",
							"html_url": "https://github.com/OWNER/REPO/issues/123",
							"state": "open",
							"body": "Issue body from mock server",
							"created_at": "2026-04-28T10:30:00Z",
							"user": {"login": "octocat", "html_url": "https://github.com/octocat"},
							"reactions": {"total_count": 2, "+1": 1, "heart": 1}
						}`))
					case "/repos/OWNER/REPO/issues/123/comments":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`[
							{
								"id": 2,
								"body": "Second comment",
								"html_url": "https://github.com/OWNER/REPO/issues/123#issuecomment-2",
								"created_at": "2026-04-28T11:30:00Z",
								"user": {"login": "hubot", "html_url": "https://github.com/hubot"},
								"reactions": {"total_count": 1, "rocket": 1}
							},
							{
								"id": 1,
								"body": "First comment",
								"html_url": "https://github.com/OWNER/REPO/issues/123#issuecomment-1",
								"created_at": "2026-04-28T11:00:00Z",
								"user": {"login": "monalisa", "html_url": "https://github.com/monalisa"},
								"reactions": {"total_count": 1, "+1": 1}
							}
						]`))
					default:
						http.NotFound(w, r)
					}
				},
				check: func(t *testing.T, doc model.DocumentData) {
					if doc.Provider != model.ProviderGitHub {
						t.Fatalf("Provider = %q, want %q", doc.Provider, model.ProviderGitHub)
					}
					if doc.Kind != model.KindIssue {
						t.Fatalf("Kind = %q, want %q", doc.Kind, model.KindIssue)
					}
					if doc.Issue == nil {
						t.Fatal("Issue should not be nil")
					}
					if doc.Issue.Title != "Mock GitHub Issue" {
						t.Fatalf("Title = %q", doc.Issue.Title)
					}
					if !doc.Issue.CreatedAt.Equal(time.Date(2026, 4, 28, 10, 30, 0, 0, time.UTC)) {
						t.Fatalf("CreatedAt = %v", doc.Issue.CreatedAt)
					}
					if doc.Issue.Reactions != (model.ReactionSummary{Total: 2, PlusOne: 1, Heart: 1}) {
						t.Fatalf("Issue reactions = %#v", doc.Issue.Reactions)
					}
					if len(doc.Issue.Comments) != 2 {
						t.Fatalf("len(Comments) = %d, want 2", len(doc.Issue.Comments))
					}
					if doc.Issue.Comments[0].Body != "First comment" || doc.Issue.Comments[1].Body != "Second comment" {
						t.Fatalf("comment order = %q, %q", doc.Issue.Comments[0].Body, doc.Issue.Comments[1].Body)
					}
					if doc.Issue.Comments[0].Reactions != (model.ReactionSummary{Total: 1, PlusOne: 1}) {
						t.Fatalf("first comment reactions = %#v", doc.Issue.Comments[0].Reactions)
					}
					if doc.Issue.Comments[1].Reactions != (model.ReactionSummary{Total: 1, Rocket: 1}) {
						t.Fatalf("second comment reactions = %#v", doc.Issue.Comments[1].Reactions)
					}
				},
			},
			{
				name:   "pull request fetch uses mock server",
				target: parser.Target{Provider: model.ProviderGitHub, Kind: model.KindPullRequest, Owner: "OWNER", Repo: "REPO", Number: 456, URL: "https://github.com/OWNER/REPO/pull/456"},
				opts:   model.FetchOptions{IncludeReactions: true},
				handler: func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/repos/OWNER/REPO/issues/456":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{
							"number": 456,
							"title": "Mock Pull Request",
							"html_url": "https://github.com/OWNER/REPO/pull/456",
							"state": "open",
							"body": "Pull request body from mock server",
							"created_at": "2026-04-28T12:30:00Z",
							"user": {"login": "octocat", "html_url": "https://github.com/octocat"},
							"reactions": {"total_count": 1, "eyes": 1}
						}`))
					case "/repos/OWNER/REPO/pulls/456/comments":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`[
							{
								"id": 2,
								"body": "Later review comment",
								"html_url": "https://github.com/OWNER/REPO/pull/456#discussion_r2",
								"created_at": "2026-04-28T13:30:00Z",
								"user": {"login": "reviewer-2", "html_url": "https://github.com/reviewer-2"},
								"reactions": {"total_count": 1, "rocket": 1}
							},
							{
								"id": 1,
								"body": "Earlier review comment",
								"html_url": "https://github.com/OWNER/REPO/pull/456#discussion_r1",
								"created_at": "2026-04-28T13:00:00Z",
								"user": {"login": "reviewer-1", "html_url": "https://github.com/reviewer-1"},
								"reactions": {"total_count": 1, "heart": 1}
							}
						]`))
					default:
						http.NotFound(w, r)
					}
				},
				check: func(t *testing.T, doc model.DocumentData) {
					if doc.Provider != model.ProviderGitHub {
						t.Fatalf("Provider = %q, want %q", doc.Provider, model.ProviderGitHub)
					}
					if doc.Kind != model.KindPullRequest {
						t.Fatalf("Kind = %q, want %q", doc.Kind, model.KindPullRequest)
					}
					if doc.PullRequest == nil {
						t.Fatal("PullRequest should not be nil")
					}
					if doc.PullRequest.Title != "Mock Pull Request" {
						t.Fatalf("Title = %q", doc.PullRequest.Title)
					}
					if !doc.PullRequest.CreatedAt.Equal(time.Date(2026, 4, 28, 12, 30, 0, 0, time.UTC)) {
						t.Fatalf("CreatedAt = %v", doc.PullRequest.CreatedAt)
					}
					if doc.PullRequest.Reactions != (model.ReactionSummary{Total: 1, Eyes: 1}) {
						t.Fatalf("PR reactions = %#v", doc.PullRequest.Reactions)
					}
					if len(doc.PullRequest.ReviewComments) != 2 {
						t.Fatalf("len(ReviewComments) = %d, want 2", len(doc.PullRequest.ReviewComments))
					}
					if doc.PullRequest.ReviewComments[0].Body != "Earlier review comment" || doc.PullRequest.ReviewComments[1].Body != "Later review comment" {
						t.Fatalf("review comment order = %q, %q", doc.PullRequest.ReviewComments[0].Body, doc.PullRequest.ReviewComments[1].Body)
					}
					if doc.PullRequest.ReviewComments[0].Reactions != (model.ReactionSummary{Total: 1, Heart: 1}) {
						t.Fatalf("first review comment reactions = %#v", doc.PullRequest.ReviewComments[0].Reactions)
					}
					if doc.PullRequest.ReviewComments[1].Reactions != (model.ReactionSummary{Total: 1, Rocket: 1}) {
						t.Fatalf("second review comment reactions = %#v", doc.PullRequest.ReviewComments[1].Reactions)
					}
				},
			},
			{
				name:   "discussion fetch uses mock server",
				target: parser.Target{Provider: model.ProviderGitHub, Kind: model.KindDiscussion, Owner: "OWNER", Repo: "REPO", Number: 789, URL: "https://github.com/OWNER/REPO/discussions/789"},
				opts:   model.FetchOptions{IncludeReactions: true},
				handler: func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/graphql":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{
							"data": {
								"repository": {
									"discussion": {
										"title": "Mock Discussion",
										"url": "https://github.com/OWNER/REPO/discussions/789",
										"body": "Discussion body from mock server",
										"createdAt": "2026-04-28T13:30:00Z",
										"isAnswered": true,
										"author": {"login": "octocat", "url": "https://github.com/octocat"},
										"reactionGroups": [
											{"content": "HEART", "users": {"totalCount": 1}},
											{"content": "THUMBS_UP", "users": {"totalCount": 1}}
										],
										"answer": {
											"id": "ANSWER_1",
											"body": "Accepted answer body",
											"url": "https://github.com/OWNER/REPO/discussions/789#discussioncomment-3",
											"createdAt": "2026-04-28T14:30:00Z",
											"author": {"login": "helper", "url": "https://github.com/helper"},
											"reactionGroups": [
												{"content": "HOORAY", "users": {"totalCount": 1}}
											]
										},
										"comments": {
											"nodes": [
												{
													"id": "COMMENT_2",
													"body": "Later top-level comment",
													"url": "https://github.com/OWNER/REPO/discussions/789#discussioncomment-2",
													"createdAt": "2026-04-28T15:00:00Z",
													"author": {"login": "later", "url": "https://github.com/later"},
													"reactionGroups": [
														{"content": "EYES", "users": {"totalCount": 1}}
													],
													"replies": {
														"nodes": [
															{
																"id": "REPLY_1",
																"body": "Earlier reply",
																"url": "https://github.com/OWNER/REPO/discussions/789#discussioncomment-4",
																"createdAt": "2026-04-28T14:45:00Z",
																"author": {"login": "replier", "url": "https://github.com/replier"},
																"reactionGroups": [
																	{"content": "ROCKET", "users": {"totalCount": 1}}
																],
																"replies": {"nodes": []}
															}
														]
													}
												},
												{
													"id": "COMMENT_1",
													"body": "Earlier top-level comment",
													"url": "https://github.com/OWNER/REPO/discussions/789#discussioncomment-1",
													"createdAt": "2026-04-28T14:00:00Z",
													"author": {"login": "earlier", "url": "https://github.com/earlier"},
													"reactionGroups": [
														{"content": "HEART", "users": {"totalCount": 1}}
													],
													"replies": {"nodes": []}
												}
											]
										}
									}
								}
							}
						}`))
					default:
						http.NotFound(w, r)
					}
				},
				check: func(t *testing.T, doc model.DocumentData) {
					if doc.Provider != model.ProviderGitHub {
						t.Fatalf("Provider = %q, want %q", doc.Provider, model.ProviderGitHub)
					}
					if doc.Kind != model.KindDiscussion {
						t.Fatalf("Kind = %q, want %q", doc.Kind, model.KindDiscussion)
					}
					if doc.Discussion == nil {
						t.Fatal("Discussion should not be nil")
					}
					if doc.Discussion.Title != "Mock Discussion" {
						t.Fatalf("Title = %q", doc.Discussion.Title)
					}
					if doc.Discussion.State != "answered" {
						t.Fatalf("State = %q", doc.Discussion.State)
					}
					if doc.Discussion.Reactions != (model.ReactionSummary{Total: 2, PlusOne: 1, Heart: 1}) {
						t.Fatalf("discussion reactions = %#v", doc.Discussion.Reactions)
					}
					if doc.Discussion.AcceptedAnswer == nil {
						t.Fatal("AcceptedAnswer should not be nil")
					}
					if doc.Discussion.AcceptedAnswer.Body != "Accepted answer body" {
						t.Fatalf("accepted answer body = %q", doc.Discussion.AcceptedAnswer.Body)
					}
					if doc.Discussion.AcceptedAnswer.Reactions != (model.ReactionSummary{Total: 1, Hooray: 1}) {
						t.Fatalf("accepted answer reactions = %#v", doc.Discussion.AcceptedAnswer.Reactions)
					}
					if len(doc.Discussion.Comments) != 3 {
						t.Fatalf("len(Comments) = %d, want 3", len(doc.Discussion.Comments))
					}
					wantBodies := []string{"Earlier top-level comment", "Earlier reply", "Later top-level comment"}
					for i, wantBody := range wantBodies {
						if doc.Discussion.Comments[i].Body != wantBody {
							t.Fatalf("Comments[%d].Body = %q, want %q", i, doc.Discussion.Comments[i].Body, wantBody)
						}
					}
					if doc.Discussion.Comments[0].Reactions != (model.ReactionSummary{Total: 1, Heart: 1}) {
						t.Fatalf("first discussion comment reactions = %#v", doc.Discussion.Comments[0].Reactions)
					}
					if doc.Discussion.Comments[1].Reactions != (model.ReactionSummary{Total: 1, Rocket: 1}) {
						t.Fatalf("reply reactions = %#v", doc.Discussion.Comments[1].Reactions)
					}
					if doc.Discussion.Comments[2].Reactions != (model.ReactionSummary{Total: 1, Eyes: 1}) {
						t.Fatalf("last discussion comment reactions = %#v", doc.Discussion.Comments[2].Reactions)
					}
				},
			},
			{
				name:           "authorization header when token configured",
				token:          "secret-token",
				target:         parser.Target{Provider: model.ProviderGitHub, Kind: model.KindIssue, Owner: "OWNER", Repo: "REPO", Number: 123, URL: "https://github.com/OWNER/REPO/issues/123"},
				wantAuthHeader: "Bearer secret-token",
				handler: func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/repos/OWNER/REPO/issues/123":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{
							"title": "Issue",
							"html_url": "https://github.com/OWNER/REPO/issues/123",
							"state": "open",
							"body": "Body",
							"created_at": "2026-04-28T10:30:00Z",
							"user": {"login": "octocat", "html_url": "https://github.com/octocat"}
						}`))
					case "/repos/OWNER/REPO/issues/123/comments":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`[]`))
					default:
						http.NotFound(w, r)
					}
				},
				check: func(t *testing.T, doc model.DocumentData) {
					if doc.Issue == nil || doc.Issue.Title != "Issue" {
						t.Fatalf("Issue = %#v", doc.Issue)
					}
				},
			},
			{
				name:   "reactions omitted when disabled",
				target: parser.Target{Provider: model.ProviderGitHub, Kind: model.KindIssue, Owner: "OWNER", Repo: "REPO", Number: 321, URL: "https://github.com/OWNER/REPO/issues/321"},
				handler: func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/repos/OWNER/REPO/issues/321":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{
							"title": "Issue without fetched reactions",
							"html_url": "https://github.com/OWNER/REPO/issues/321",
							"state": "open",
							"body": "Body",
							"created_at": "2026-04-28T10:30:00Z",
							"user": {"login": "octocat", "html_url": "https://github.com/octocat"},
							"reactions": {"total_count": 2, "+1": 1, "heart": 1}
						}`))
					case "/repos/OWNER/REPO/issues/321/comments":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`[
							{
								"id": 1,
								"body": "Comment",
								"html_url": "https://github.com/OWNER/REPO/issues/321#issuecomment-1",
								"created_at": "2026-04-28T11:00:00Z",
								"user": {"login": "hubot", "html_url": "https://github.com/hubot"},
								"reactions": {"total_count": 1, "rocket": 1}
							}
						]`))
					default:
						http.NotFound(w, r)
					}
				},
				check: func(t *testing.T, doc model.DocumentData) {
					if doc.Issue == nil {
						t.Fatal("Issue should not be nil")
					}
					if doc.Issue.Reactions != (model.ReactionSummary{}) {
						t.Fatalf("Issue reactions = %#v, want zero", doc.Issue.Reactions)
					}
					if len(doc.Issue.Comments) != 1 {
						t.Fatalf("len(Comments) = %d, want 1", len(doc.Issue.Comments))
					}
					if doc.Issue.Comments[0].Reactions != (model.ReactionSummary{}) {
						t.Fatalf("Comment reactions = %#v, want zero", doc.Issue.Comments[0].Reactions)
					}
				},
			},
			{
				name:         "surfaces unexpected status",
				target:       parser.Target{Provider: model.ProviderGitHub, Kind: model.KindIssue, Owner: "OWNER", Repo: "REPO", Number: 123, URL: "https://github.com/OWNER/REPO/issues/123"},
				wantErrParts: []string{"get issue", "unexpected status 418"},
				handler: func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "boom", http.StatusTeapot)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var authHeader string
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					authHeader = r.Header.Get("Authorization")
					tt.handler(w, r)
				}))
				defer server.Close()

				client := NewClient(Options{HTTPClient: server.Client(), BaseURL: server.URL, Token: tt.token})
				doc, err := client.Fetch(context.Background(), tt.target, tt.opts)
				if len(tt.wantErrParts) > 0 {
					if err == nil {
						t.Fatal("expected error")
					}
					for _, wantErrPart := range tt.wantErrParts {
						if !strings.Contains(err.Error(), wantErrPart) {
							t.Fatalf("expected error to contain %q, got %q", wantErrPart, err.Error())
						}
					}
					return
				}

				if err != nil {
					t.Fatalf("Fetch() error = %v", err)
				}
				if tt.wantAuthHeader != "" && authHeader != tt.wantAuthHeader {
					t.Fatalf("Authorization header = %q, want %q", authHeader, tt.wantAuthHeader)
				}
				if tt.check != nil {
					tt.check(t, doc)
				}
			})
		}
	})

	t.Run("unsupported kind", func(t *testing.T) {
		tests := []struct {
			name   string
			target parser.Target
		}{
			{
				name:   "rejects unknown kind",
				target: parser.Target{Provider: model.ProviderGitHub, Kind: model.ContentKind("unknown"), Owner: "OWNER", Repo: "REPO", Number: 123, URL: "https://github.com/OWNER/REPO/unknown/123"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				client := NewClient(Options{})
				_, err := client.Fetch(context.Background(), tt.target, model.FetchOptions{})
				if err == nil {
					t.Fatal("expected error")
				}
				if !fetchprovider.IsUnsupportedCapability(err) {
					t.Fatalf("expected unsupported capability error, got %v", err)
				}
			})
		}
	})
}
