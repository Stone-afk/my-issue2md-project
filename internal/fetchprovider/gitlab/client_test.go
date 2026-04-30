package gitlab

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
			target            parser.Target
			handler           http.HandlerFunc
			wantErrParts      []string
			wantTitle         string
			wantCommentBodies []string
			wantCreatedAt     time.Time
		}{
			{
				name:   "issue fetch uses mock server",
				target: parser.Target{Provider: model.ProviderGitLab, Kind: model.KindIssue, Project: "GROUP/PROJECT", Number: 321, URL: "https://gitlab.com/GROUP/PROJECT/-/issues/321"},
				handler: func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/v4/projects/GROUP%2FPROJECT/issues/321":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{
							"iid": 321,
							"title": "Mock GitLab Issue",
							"web_url": "https://gitlab.com/GROUP/PROJECT/-/issues/321",
							"state": "opened",
							"description": "GitLab issue body from mock server",
							"created_at": "2026-04-28T09:30:00Z",
							"author": {"username": "gitlab-user", "web_url": "https://gitlab.com/gitlab-user"}
						}`))
					case "/api/v4/projects/GROUP%2FPROJECT/issues/321/notes":
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`[
							{
								"id": 2,
								"body": "Later note",
								"created_at": "2026-04-28T10:30:00Z",
								"author": {"username": "note-user-2", "web_url": "https://gitlab.com/note-user-2"}
							},
							{
								"id": 1,
								"body": "Earlier note",
								"created_at": "2026-04-28T10:00:00Z",
								"author": {"username": "note-user-1", "web_url": "https://gitlab.com/note-user-1"}
							}
						]`))
					default:
						http.NotFound(w, r)
					}
				},
				wantTitle:         "Mock GitLab Issue",
				wantCommentBodies: []string{"Earlier note", "Later note"},
				wantCreatedAt:     time.Date(2026, 4, 28, 9, 30, 0, 0, time.UTC),
			},
			{
				name:         "surfaces unexpected status",
				target:       parser.Target{Provider: model.ProviderGitLab, Kind: model.KindIssue, Project: "GROUP/PROJECT", Number: 321, URL: "https://gitlab.com/GROUP/PROJECT/-/issues/321"},
				wantErrParts: []string{"get issue", "unexpected status 418"},
				handler: func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "boom", http.StatusTeapot)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(tt.handler))
				defer server.Close()

				client := NewClient(Options{HTTPClient: server.Client(), BaseURL: server.URL})
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
				if doc.Provider != model.ProviderGitLab {
					t.Fatalf("Provider = %q, want %q", doc.Provider, model.ProviderGitLab)
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
				target: parser.Target{Provider: model.ProviderGitLab, Kind: model.KindDiscussion, Project: "GROUP/PROJECT", Number: 321, URL: "https://gitlab.com/GROUP/PROJECT/-/discussions/321"},
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
