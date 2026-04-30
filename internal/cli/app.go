package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/stoneafk/issue2md/internal/fetchprovider"
	"github.com/stoneafk/issue2md/internal/model"
	"github.com/stoneafk/issue2md/internal/parser"
)

type Renderer interface {
	Render(doc model.DocumentData) (string, error)
}

type Options struct {
	EnableUserLinks bool
	EnableReactions bool
}

type App struct {
	Stdout      io.Writer
	Stderr      io.Writer
	Providers   map[model.Provider]fetchprovider.Provider
	Renderer    Renderer
	NewRenderer func(Options) Renderer
}

func (a App) Run(ctx context.Context, args []string) int {
	stdout := a.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := a.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	fs := flag.NewFlagSet("issue2md", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	enableReactions := fs.Bool("enable-reactions", false, "include reactions")
	enableUserLinks := fs.Bool("enable-user-links", false, "render user links")
	if err := fs.Parse(args); err != nil {
		_, _ = fmt.Fprintln(stderr, fmt.Errorf("parse flags: %w", err))
		return 1
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		_, _ = fmt.Fprintln(stderr, fmt.Errorf("missing url argument"))
		return 1
	}

	rawURL := remaining[0]
	target, err := parser.Parse(rawURL)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, fmt.Errorf("parse input url: %w", err))
		return 1
	}

	fetchOpts := model.FetchOptions{IncludeReactions: *enableReactions}
	doc, err := a.fetch(ctx, target, fetchOpts)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}

	rendererOpts := Options{EnableUserLinks: *enableUserLinks, EnableReactions: *enableReactions}
	renderer, err := a.renderer(rendererOpts)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, fmt.Errorf("create renderer: %w", err))
		return 1
	}

	markdown, err := renderer.Render(doc)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, fmt.Errorf("render document: %w", err))
		return 1
	}

	if len(remaining) > 1 {
		if err := os.WriteFile(remaining[1], []byte(markdown), 0o644); err != nil {
			_, _ = fmt.Fprintln(stderr, fmt.Errorf("write output file %q: %w", remaining[1], err))
			return 1
		}
		return 0
	}

	_, err = io.WriteString(stdout, markdown)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, fmt.Errorf("write stdout: %w", err))
		return 1
	}

	return 0
}

func (a App) fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error) {
	provider, ok := a.Providers[target.Provider]
	if !ok {
		return model.DocumentData{}, fmt.Errorf("fetch document via %s: %w", target.Provider, fetchprovider.ErrProviderNotRegistered)
	}

	doc, err := provider.Fetch(ctx, target, opts)
	if err != nil {
		return model.DocumentData{}, fmt.Errorf("fetch document via %s: %w", target.Provider, err)
	}
	return doc, nil
}

func (a App) renderer(opts Options) (Renderer, error) {
	if a.NewRenderer != nil {
		return a.NewRenderer(opts), nil
	}
	if a.Renderer != nil {
		return a.Renderer, nil
	}
	return nil, fmt.Errorf("renderer is nil")
}
