package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/stoneafk/issue2md/internal/model"
)

type Target struct {
	Provider model.Provider
	Kind     model.ContentKind
	Owner    string
	Repo     string
	Project  string
	Number   int
	URL      string
}

func Parse(rawURL string) (Target, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return Target{}, fmt.Errorf("parse url %q: invalid url", rawURL)
	}
	if parsedURL.Scheme != "https" {
		return Target{}, fmt.Errorf("parse url %q: invalid url", rawURL)
	}

	hostAndPath, err := splitAndValidatePath(parsedURL)
	if err != nil {
		return Target{}, fmt.Errorf("parse url %q: %s", rawURL, err)
	}

	switch hostAndPath[0] {
	case "github.com":
		return parseGitHubTarget(rawURL, hostAndPath)
	case "gitlab.com":
		return parseGitLabTarget(rawURL, hostAndPath)
	default:
		return Target{}, fmt.Errorf("parse url %q: unsupported url", rawURL)
	}
}

func splitAndValidatePath(parsedURL *url.URL) ([]string, error) {
	hostAndPath := append([]string{parsedURL.Host}, strings.Split(strings.TrimPrefix(parsedURL.EscapedPath(), "/"), "/")...)
	if len(hostAndPath) < 3 || hostAndPath[0] == "" {
		return nil, fmt.Errorf("unsupported url")
	}

	return hostAndPath, nil
}

func parseGitHubTarget(rawURL string, hostAndPath []string) (Target, error) {
	if len(hostAndPath) != 5 || hostAndPath[1] == "" || hostAndPath[2] == "" {
		return Target{}, fmt.Errorf("parse url %q: unsupported url", rawURL)
	}

	number, err := parsePositiveNumber(hostAndPath[4])
	if err != nil {
		return Target{}, fmt.Errorf("parse url %q: invalid url", rawURL)
	}

	var kind model.ContentKind
	switch hostAndPath[3] {
	case "issues":
		kind = model.KindIssue
	case "pull":
		kind = model.KindPullRequest
	case "discussions":
		kind = model.KindDiscussion
	default:
		return Target{}, fmt.Errorf("parse url %q: unsupported url", rawURL)
	}

	return Target{
		Provider: model.ProviderGitHub,
		Kind:     kind,
		Owner:    hostAndPath[1],
		Repo:     hostAndPath[2],
		Number:   number,
		URL:      rawURL,
	}, nil
}

func parseGitLabTarget(rawURL string, hostAndPath []string) (Target, error) {
	if len(hostAndPath) < 6 {
		return Target{}, fmt.Errorf("parse url %q: unsupported url", rawURL)
	}

	marker := -1
	for i := 1; i < len(hostAndPath); i++ {
		if hostAndPath[i] == "-" {
			marker = i
			break
		}
	}
	if marker < 2 || marker+2 >= len(hostAndPath) {
		return Target{}, fmt.Errorf("parse url %q: unsupported url", rawURL)
	}
	if hostAndPath[marker+1] != "issues" || marker+3 != len(hostAndPath) {
		return Target{}, fmt.Errorf("parse url %q: unsupported url", rawURL)
	}

	number, err := parsePositiveNumber(hostAndPath[marker+2])
	if err != nil {
		return Target{}, fmt.Errorf("parse url %q: invalid url", rawURL)
	}

	project := strings.Join(hostAndPath[1:marker], "/")
	if project == "" {
		return Target{}, fmt.Errorf("parse url %q: unsupported url", rawURL)
	}

	return Target{
		Provider: model.ProviderGitLab,
		Kind:     model.KindIssue,
		Project:  project,
		Number:   number,
		URL:      rawURL,
	}, nil
}

func parsePositiveNumber(value string) (int, error) {
	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("invalid url")
	}
	return number, nil
}
