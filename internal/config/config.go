package config

type Config struct {
	GitHubToken string
}

func LoadFromEnv(env map[string]string) Config {
	return Config{GitHubToken: env["GITHUB_TOKEN"]}
}

func LoadFromLookup(lookup func(string) (string, bool)) Config {
	token, _ := lookup("GITHUB_TOKEN")
	return Config{GitHubToken: token}
}
