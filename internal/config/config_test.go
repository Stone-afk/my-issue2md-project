package config_test

import (
	"reflect"
	"testing"

	"github.com/stoneafk/issue2md/internal/config"
)

func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		wantToken string
	}{
		{
			name:      "reads github token",
			env:       map[string]string{"GITHUB_TOKEN": "github-token"},
			wantToken: "github-token",
		},
		{
			name:      "github token missing is allowed",
			env:       map[string]string{},
			wantToken: "",
		},
		{
			name:      "gitlab token is ignored",
			env:       map[string]string{"GITLAB_TOKEN": "gitlab-token"},
			wantToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.LoadFromEnv(tt.env)
			if got.GitHubToken != tt.wantToken {
				t.Fatalf("GitHubToken = %q, want %q", got.GitHubToken, tt.wantToken)
			}
			if _, ok := reflect.TypeOf(got).FieldByName("GitLabToken"); ok {
				t.Fatal("Config must not include GitLabToken")
			}
		})
	}
}

func TestLoadFromLookup(t *testing.T) {
	tests := []struct {
		name      string
		lookup    func(string) (string, bool)
		wantToken string
	}{
		{
			name: "returns github token when present",
			lookup: func(key string) (string, bool) {
				values := map[string]string{
					"GITHUB_TOKEN": "github-token",
					"GITLAB_TOKEN": "gitlab-token",
				}
				value, ok := values[key]
				return value, ok
			},
			wantToken: "github-token",
		},
		{
			name: "missing github token returns empty string",
			lookup: func(key string) (string, bool) {
				values := map[string]string{
					"GITLAB_TOKEN": "gitlab-token",
				}
				value, ok := values[key]
				return value, ok
			},
			wantToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.LoadFromLookup(tt.lookup)
			if got.GitHubToken != tt.wantToken {
				t.Fatalf("GitHubToken = %q, want %q", got.GitHubToken, tt.wantToken)
			}
		})
	}
}
