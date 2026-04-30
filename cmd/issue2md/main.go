package main

import (
	"context"
	"os"

	"github.com/stoneafk/issue2md/internal/cli"
	"github.com/stoneafk/issue2md/internal/config"
	"github.com/stoneafk/issue2md/internal/converter"
	"github.com/stoneafk/issue2md/internal/fetchprovider"
	gh "github.com/stoneafk/issue2md/internal/fetchprovider/github"
	gl "github.com/stoneafk/issue2md/internal/fetchprovider/gitlab"
	"github.com/stoneafk/issue2md/internal/model"
)

func main() {
	env := envMap()
	cfg := config.LoadFromEnv(env)
	app := cli.App{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Providers: map[model.Provider]fetchprovider.Provider{
			model.ProviderGitHub: gh.NewClient(gh.Options{Token: cfg.GitHubToken}),
			model.ProviderGitLab: gl.NewClient(gl.Options{}),
		},
		NewRenderer: func(opts cli.Options) cli.Renderer {
			return converter.NewRenderer(converter.Options{
				EnableUserLinks: opts.EnableUserLinks,
				EnableReactions: opts.EnableReactions,
			})
		},
	}
	os.Exit(app.Run(context.Background(), os.Args[1:]))
}

func envMap() map[string]string {
	values := make(map[string]string)
	for _, entry := range os.Environ() {
		for i := 0; i < len(entry); i++ {
			if entry[i] == '=' {
				values[entry[:i]] = entry[i+1:]
				break
			}
		}
	}
	return values
}
