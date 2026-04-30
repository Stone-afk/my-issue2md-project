package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stoneafk/issue2md/internal/fetchprovider"
	"github.com/stoneafk/issue2md/internal/model"
	"github.com/stoneafk/issue2md/internal/parser"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

type Options struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      string
}

func NewClient(opts Options) *Client {
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      opts.Token,
	}
}

func (c *Client) Fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error) {
	if target.Provider != model.ProviderGitHub {
		return model.DocumentData{}, fmt.Errorf("unsupported provider %q", target.Provider)
	}
	if target.Kind != model.KindIssue {
		return model.DocumentData{}, fetchprovider.UnsupportedCapability(target.Provider, target.Kind)
	}
	return c.GetIssue(ctx, target, opts)
}

func (c *Client) GetIssue(ctx context.Context, target parser.Target, _ model.FetchOptions) (model.DocumentData, error) {
	issueURL := fmt.Sprintf("%s/repos/%s/%s/issues/%d", c.baseURL, target.Owner, target.Repo, target.Number)
	commentsURL := issueURL + "/comments"

	var issueResp githubIssueResponse
	if err := c.getJSON(ctx, issueURL, &issueResp); err != nil {
		return model.DocumentData{}, fmt.Errorf("get issue: %w", err)
	}

	var commentsResp []githubCommentResponse
	if err := c.getJSON(ctx, commentsURL, &commentsResp); err != nil {
		return model.DocumentData{}, fmt.Errorf("get issue comments: %w", err)
	}

	comments, err := mapGitHubComments(commentsResp)
	if err != nil {
		return model.DocumentData{}, fmt.Errorf("map issue comments: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, issueResp.CreatedAt)
	if err != nil {
		return model.DocumentData{}, fmt.Errorf("parse issue created_at: %w", err)
	}

	return model.DocumentData{
		Provider: model.ProviderGitHub,
		Kind:     model.KindIssue,
		Issue: &model.IssueData{
			Provider:  model.ProviderGitHub,
			Title:     issueResp.Title,
			URL:       issueResp.HTMLURL,
			Author:    model.UserData{Login: issueResp.User.Login, URL: issueResp.User.HTMLURL},
			CreatedAt: createdAt,
			State:     issueResp.State,
			Body:      issueResp.Body,
			Comments:  comments,
		},
	}, nil
}

type githubIssueResponse struct {
	Title     string `json:"title"`
	HTMLURL   string `json:"html_url"`
	State     string `json:"state"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Login   string `json:"login"`
		HTMLURL string `json:"html_url"`
	} `json:"user"`
}

type githubCommentResponse struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Login   string `json:"login"`
		HTMLURL string `json:"html_url"`
	} `json:"user"`
}

func (c *Client) getJSON(ctx context.Context, requestURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func mapGitHubComments(input []githubCommentResponse) ([]model.CommentData, error) {
	comments := make([]model.CommentData, 0, len(input))
	for _, item := range input {
		createdAt, err := time.Parse(time.RFC3339, item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse comment created_at: %w", err)
		}

		comments = append(comments, model.CommentData{
			ID:        strconv.FormatInt(item.ID, 10),
			Author:    model.UserData{Login: item.User.Login, URL: item.User.HTMLURL},
			CreatedAt: createdAt,
			Body:      item.Body,
			URL:       item.HTMLURL,
		})
	}

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	return comments, nil
}
