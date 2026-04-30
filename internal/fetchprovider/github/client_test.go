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
			name              string
			token             string
			target            parser.Target
			handler           http.HandlerFunc
			wantErrParts      []string
			wantAuthHeader    string
			wantTitle         string
			wantCommentBodies []string
			wantCreatedAt     time.Time
		}{
			{
				name:   "issue fetch uses mock server",
				target: parser.Target{Provider: model.ProviderGitHub, Kind: model.KindIssue, Owner: "OWNER", Repo: "REPO", Number: 123, URL: "https://github.com/OWNER/REPO/issues/123"},
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
							"user": {"login": "octocat", "html_url": "https://github.com/octocat"}
						}`))
					case "/repos/OWNER/REPO/issues/123/comments":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`[
							{
								"id": 2,
								"body": "Second comment",
								"html_url": "https://github.com/OWNER/REPO/issues/123#issuecomment-2",
								"created_at": "2026-04-28T11:30:00Z",
								"user": {"login": "hubot", "html_url": "https://github.com/hubot"}
							},
							{
								"id": 1,
								"body": "First comment",
								"html_url": "https://github.com/OWNER/REPO/issues/123#issuecomment-1",
								"created_at": "2026-04-28T11:00:00Z",
								"user": {"login": "monalisa", "html_url": "https://github.com/monalisa"}
							}
						]`))
					default:
						http.NotFound(w, r)
					}
				},
				wantTitle:         "Mock GitHub Issue",
				wantCommentBodies: []string{"First comment", "Second comment"},
				wantCreatedAt:     time.Date(2026, 4, 28, 10, 30, 0, 0, time.UTC),
			},
			{
				name:            "authorization header when token configured",
				token:           "secret-token",
				target:          parser.Target{Provider: model.ProviderGitHub, Kind: model.KindIssue, Owner: "OWNER", Repo: "REPO", Number: 123, URL: "https://github.com/OWNER/REPO/issues/123"},
				wantAuthHeader:  "Bearer secret-token",
				wantTitle:       "Issue",
				wantCommentBodies: []string{},
				wantCreatedAt:   time.Date(2026, 4, 28, 10, 30, 0, 0, time.UTC),
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
				doc, err := client.Fetch(context.Background(), tt.target, model.FetchOptions{})
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
				if doc.Provider != model.ProviderGitHub {
					t.Fatalf("Provider = %q, want %q", doc.Provider, model.ProviderGitHub)
				}
				if doc.Kind != model.KindIssue {
					t.Fatalf("Kind = %q, want %q", doc.Kind, model.KindIssue)
				}
				if doc.Issue == nil {
					t.Fatal("Issue should not be nil")
				}
				if doc.Issue.Title != tt.wantTitle {
					t.Fatalf("Title = %q, want %q", doc.Issue.Title, tt.wantTitle)
				}
				if len(doc.Issue.Comments) != len(tt.wantCommentBodies) {
					t.Fatalf("len(Comments) = %d, want %d", len(doc.Issue.Comments), len(tt.wantCommentBodies))
				}
				for i, wantCommentBody := range tt.wantCommentBodies {
					if doc.Issue.Comments[i].Body != wantCommentBody {
						t.Fatalf("Comments[%d].Body = %q, want %q", i, doc.Issue.Comments[i].Body, wantCommentBody)
					}
				}
				if !doc.Issue.CreatedAt.Equal(tt.wantCreatedAt) {
					t.Fatalf("CreatedAt = %v, want %v", doc.Issue.CreatedAt, tt.wantCreatedAt)
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
				name:   "rejects discussion kind",
				target: parser.Target{Provider: model.ProviderGitHub, Kind: model.KindDiscussion, Owner: "OWNER", Repo: "REPO", Number: 123, URL: "https://github.com/OWNER/REPO/discussions/123"},
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
