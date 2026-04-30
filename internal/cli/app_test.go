package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stoneafk/issue2md/internal/fetchprovider"
	"github.com/stoneafk/issue2md/internal/model"
	"github.com/stoneafk/issue2md/internal/parser"
)

type fakeProvider struct {
	doc      model.DocumentData
	err      error
	called   bool
	target   parser.Target
	fetchOps model.FetchOptions
}

func (f *fakeProvider) Fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error) {
	f.called = true
	f.target = target
	f.fetchOps = opts
	return f.doc, f.err
}

type fakeRenderer struct {
	output string
	err    error
	called bool
	doc    model.DocumentData
	opts   Options
}

func (f *fakeRenderer) Render(doc model.DocumentData) (string, error) {
	f.called = true
	f.doc = doc
	return f.output, f.err
}

func TestAppRun(t *testing.T) {
	baseDoc := model.DocumentData{
		Provider: model.ProviderGitHub,
		Kind:     model.KindIssue,
		Issue: &model.IssueData{
			Provider: model.ProviderGitHub,
			Title:    "Issue",
			URL:      "https://github.com/OWNER/REPO/issues/123",
			Author:   model.UserData{Login: "octocat"},
			Body:     "Issue body",
		},
	}

	tests := []struct {
		name           string
		args           []string
		setup          func(t *testing.T) (App, *fakeProvider, *fakeProvider, *fakeRenderer, string)
		wantCode       int
		wantStdout     string
		wantStderrPart []string
		check          func(t *testing.T, stdout string, stderr string, githubProvider *fakeProvider, gitlabProvider *fakeProvider, renderer *fakeRenderer, outputPath string)
	}{
		{
			name: "missing url exits non-zero and writes stderr",
			setup: func(t *testing.T) (App, *fakeProvider, *fakeProvider, *fakeRenderer, string) {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				app := App{Stdout: &stdout, Stderr: &stderr}
				return app, nil, nil, nil, ""
			},
			wantCode:       1,
			wantStderrPart: []string{"url"},
		},
		{
			name: "invalid url exits non-zero and writes stderr",
			args: []string{"://not-a-url"},
			setup: func(t *testing.T) (App, *fakeProvider, *fakeProvider, *fakeRenderer, string) {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				app := App{Stdout: &stdout, Stderr: &stderr}
				return app, nil, nil, nil, ""
			},
			wantCode:       1,
			wantStderrPart: []string{"invalid"},
		},
		{
			name: "unsupported url exits non-zero and writes stderr",
			args: []string{"https://github.com/OWNER/REPO"},
			setup: func(t *testing.T) (App, *fakeProvider, *fakeProvider, *fakeRenderer, string) {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				app := App{Stdout: &stdout, Stderr: &stderr}
				return app, nil, nil, nil, ""
			},
			wantCode:       1,
			wantStderrPart: []string{"unsupported"},
		},
		{
			name: "writes rendered markdown to stdout by default",
			args: []string{"https://github.com/OWNER/REPO/issues/123"},
			setup: func(t *testing.T) (App, *fakeProvider, *fakeProvider, *fakeRenderer, string) {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				provider := &fakeProvider{doc: baseDoc}
				renderer := &fakeRenderer{output: "rendered markdown"}
				app := App{Stdout: &stdout, Stderr: &stderr, Providers: map[model.Provider]fetchprovider.Provider{model.ProviderGitHub: provider}, NewRenderer: func(opts Options) Renderer {
					renderer.opts = opts
					return renderer
				}}
				return app, provider, nil, renderer, ""
			},
			wantCode:   0,
			wantStdout: "rendered markdown",
		},
		{
			name: "writes rendered markdown to output file",
			args: []string{"https://github.com/OWNER/REPO/issues/123"},
			setup: func(t *testing.T) (App, *fakeProvider, *fakeProvider, *fakeRenderer, string) {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				provider := &fakeProvider{doc: baseDoc}
				renderer := &fakeRenderer{output: "file markdown"}
				outputPath := filepath.Join(t.TempDir(), "out.md")
				app := App{Stdout: &stdout, Stderr: &stderr, Providers: map[model.Provider]fetchprovider.Provider{model.ProviderGitHub: provider}, NewRenderer: func(opts Options) Renderer {
					renderer.opts = opts
					return renderer
				}}
				return app, provider, nil, renderer, outputPath
			},
			wantCode: 0,
			check: func(t *testing.T, stdout string, stderr string, githubProvider *fakeProvider, gitlabProvider *fakeProvider, renderer *fakeRenderer, outputPath string) {
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatalf("ReadFile() error = %v", err)
				}
				if string(content) != "file markdown" {
					t.Fatalf("file contents = %q, want %q", string(content), "file markdown")
				}
			},
		},
		{
			name: "propagates flags to fetcher and renderer",
			args: []string{"-enable-reactions", "-enable-user-links", "https://github.com/OWNER/REPO/issues/123"},
			setup: func(t *testing.T) (App, *fakeProvider, *fakeProvider, *fakeRenderer, string) {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				provider := &fakeProvider{doc: baseDoc}
				renderer := &fakeRenderer{output: "flag markdown"}
				app := App{Stdout: &stdout, Stderr: &stderr, Providers: map[model.Provider]fetchprovider.Provider{model.ProviderGitHub: provider}, NewRenderer: func(opts Options) Renderer {
					renderer.opts = opts
					return renderer
				}}
				return app, provider, nil, renderer, ""
			},
			wantCode:   0,
			wantStdout: "flag markdown",
			check: func(t *testing.T, stdout string, stderr string, githubProvider *fakeProvider, gitlabProvider *fakeProvider, renderer *fakeRenderer, outputPath string) {
				if !githubProvider.fetchOps.IncludeReactions {
					t.Fatal("expected fetch options to include reactions")
				}
				if renderer.opts != (Options{EnableUserLinks: true, EnableReactions: true}) {
					t.Fatalf("renderer options = %#v", renderer.opts)
				}
			},
		},
		{
			name: "dispatches by provider registry",
			args: []string{"https://gitlab.com/GROUP/PROJECT/-/issues/321"},
			setup: func(t *testing.T) (App, *fakeProvider, *fakeProvider, *fakeRenderer, string) {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				githubProvider := &fakeProvider{doc: baseDoc}
				gitlabProvider := &fakeProvider{doc: model.DocumentData{Provider: model.ProviderGitLab, Kind: model.KindIssue, Issue: &model.IssueData{Provider: model.ProviderGitLab, Title: "GitLab"}}}
				renderer := &fakeRenderer{output: "dispatch markdown"}
				app := App{Stdout: &stdout, Stderr: &stderr, Providers: map[model.Provider]fetchprovider.Provider{
					model.ProviderGitHub: githubProvider,
					model.ProviderGitLab: gitlabProvider,
				}, NewRenderer: func(opts Options) Renderer {
					renderer.opts = opts
					return renderer
				}}
				return app, githubProvider, gitlabProvider, renderer, ""
			},
			wantCode:   0,
			wantStdout: "dispatch markdown",
			check: func(t *testing.T, stdout string, stderr string, githubProvider *fakeProvider, gitlabProvider *fakeProvider, renderer *fakeRenderer, outputPath string) {
				if githubProvider.called {
					t.Fatal("did not expect github provider to be called")
				}
				if !gitlabProvider.called {
					t.Fatal("expected gitlab provider to be called")
				}
			},
		},
		{
			name: "missing provider reports provider not registered",
			args: []string{"https://github.com/OWNER/REPO/issues/123"},
			setup: func(t *testing.T) (App, *fakeProvider, *fakeProvider, *fakeRenderer, string) {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				app := App{Stdout: &stdout, Stderr: &stderr, Providers: map[model.Provider]fetchprovider.Provider{}, NewRenderer: func(opts Options) Renderer {
					return &fakeRenderer{output: "unused"}
				}}
				return app, nil, nil, nil, ""
			},
			wantCode:       1,
			wantStderrPart: []string{"fetch document via github", fetchprovider.ErrProviderNotRegistered.Error()},
		},
		{
			name: "unsupported capability is wrapped with provider context",
			args: []string{"https://github.com/OWNER/REPO/discussions/123"},
			setup: func(t *testing.T) (App, *fakeProvider, *fakeProvider, *fakeRenderer, string) {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				provider := &fakeProvider{err: fetchprovider.UnsupportedCapability(model.ProviderGitHub, model.KindDiscussion)}
				app := App{Stdout: &stdout, Stderr: &stderr, Providers: map[model.Provider]fetchprovider.Provider{model.ProviderGitHub: provider}, NewRenderer: func(opts Options) Renderer {
					return &fakeRenderer{output: "unused"}
				}}
				return app, provider, nil, nil, ""
			},
			wantCode:       1,
			wantStderrPart: []string{"fetch document via github", "does not support kind"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, githubProvider, gitlabProvider, renderer, outputPath := tt.setup(t)

			stdoutBuffer := app.Stdout.(*bytes.Buffer)
			stderrBuffer := app.Stderr.(*bytes.Buffer)
			args := append([]string{}, tt.args...)
			if outputPath != "" {
				args = append(args, outputPath)
			}

			code := app.Run(context.Background(), args)
			if code != tt.wantCode {
				t.Fatalf("exit code = %d, want %d", code, tt.wantCode)
			}
			if stdoutBuffer.String() != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", stdoutBuffer.String(), tt.wantStdout)
			}
			for _, wantStderrPart := range tt.wantStderrPart {
				if !strings.Contains(strings.ToLower(stderrBuffer.String()), strings.ToLower(wantStderrPart)) {
					t.Fatalf("expected stderr to contain %q, got %q", wantStderrPart, stderrBuffer.String())
				}
			}
			if tt.check != nil {
				tt.check(t, stdoutBuffer.String(), stderrBuffer.String(), githubProvider, gitlabProvider, renderer, outputPath)
			}
		})
	}
}
