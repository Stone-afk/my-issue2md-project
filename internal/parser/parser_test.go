package parser_test

import (
	"strings"
	"testing"

	"github.com/stoneafk/issue2md/internal/model"
	"github.com/stoneafk/issue2md/internal/parser"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		rawURL      string
		want        parser.Target
		wantErrPart string
	}{
		{
			name:   "valid github issue url",
			rawURL: "https://github.com/OWNER/REPO/issues/123",
			want: parser.Target{
				Provider: model.ProviderGitHub,
				Kind:     model.KindIssue,
				Owner:    "OWNER",
				Repo:     "REPO",
				Number:   123,
				URL:      "https://github.com/OWNER/REPO/issues/123",
			},
		},
		{
			name:   "valid github pull request url",
			rawURL: "https://github.com/OWNER/REPO/pull/456",
			want: parser.Target{
				Provider: model.ProviderGitHub,
				Kind:     model.KindPullRequest,
				Owner:    "OWNER",
				Repo:     "REPO",
				Number:   456,
				URL:      "https://github.com/OWNER/REPO/pull/456",
			},
		},
		{
			name:   "valid github discussion url",
			rawURL: "https://github.com/OWNER/REPO/discussions/789",
			want: parser.Target{
				Provider: model.ProviderGitHub,
				Kind:     model.KindDiscussion,
				Owner:    "OWNER",
				Repo:     "REPO",
				Number:   789,
				URL:      "https://github.com/OWNER/REPO/discussions/789",
			},
		},
		{
			name:   "valid gitlab issue url",
			rawURL: "https://gitlab.com/GROUP/PROJECT/-/issues/321",
			want: parser.Target{
				Provider: model.ProviderGitLab,
				Kind:     model.KindIssue,
				Project:  "GROUP/PROJECT",
				Number:   321,
				URL:      "https://gitlab.com/GROUP/PROJECT/-/issues/321",
			},
		},
		{
			name:   "preserves source url with query string",
			rawURL: "https://github.com/OWNER/REPO/issues/123?foo=bar",
			want: parser.Target{
				Provider: model.ProviderGitHub,
				Kind:     model.KindIssue,
				Owner:    "OWNER",
				Repo:     "REPO",
				Number:   123,
				URL:      "https://github.com/OWNER/REPO/issues/123?foo=bar",
			},
		},
		{
			name:        "invalid scheme",
			rawURL:      "http://github.com/OWNER/REPO/issues/123",
			wantErrPart: "invalid",
		},
		{
			name:        "missing path segment",
			rawURL:      "https://github.com/OWNER//issues/123",
			wantErrPart: "unsupported",
		},
		{
			name:        "non positive number",
			rawURL:      "https://github.com/OWNER/REPO/issues/0",
			wantErrPart: "invalid",
		},
		{
			name:        "unsupported github url",
			rawURL:      "https://github.com/OWNER/REPO/actions",
			wantErrPart: "unsupported",
		},
		{
			name:        "unsupported gitlab merge request url",
			rawURL:      "https://gitlab.com/GROUP/PROJECT/-/merge_requests/123",
			wantErrPart: "unsupported",
		},
		{
			name:        "invalid url",
			rawURL:      "://not-a-url",
			wantErrPart: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.rawURL)
			if tt.wantErrPart != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(strings.ToLower(err.Error()), tt.wantErrPart) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErrPart, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
