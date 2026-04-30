package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
}

type Options struct {
	HTTPClient *http.Client
	BaseURL    string
}

func NewClient(opts Options) *Client {
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

func (c *Client) Fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error) {
	if target.Provider != model.ProviderGitLab {
		return model.DocumentData{}, fmt.Errorf("unsupported provider %q", target.Provider)
	}
	if target.Kind != model.KindIssue {
		return model.DocumentData{}, fetchprovider.UnsupportedCapability(target.Provider, target.Kind)
	}
	return c.GetIssue(ctx, target, opts)
}

func (c *Client) GetIssue(ctx context.Context, target parser.Target, _ model.FetchOptions) (model.DocumentData, error) {
	project := url.PathEscape(url.PathEscape(target.Project))
	issueURL := fmt.Sprintf("%s/api/v4/projects/%s/issues/%d", c.baseURL, project, target.Number)
	notesURL := issueURL + "/notes"

	var issueResp gitlabIssueResponse
	if err := c.getJSON(ctx, issueURL, &issueResp); err != nil {
		return model.DocumentData{}, fmt.Errorf("get issue: %w", err)
	}

	var notesResp []gitlabNoteResponse
	if err := c.getJSON(ctx, notesURL, &notesResp); err != nil {
		return model.DocumentData{}, fmt.Errorf("get issue notes: %w", err)
	}

	comments, err := mapGitLabNotes(notesResp)
	if err != nil {
		return model.DocumentData{}, fmt.Errorf("map issue notes: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, issueResp.CreatedAt)
	if err != nil {
		return model.DocumentData{}, fmt.Errorf("parse issue created_at: %w", err)
	}

	return model.DocumentData{
		Provider: model.ProviderGitLab,
		Kind:     model.KindIssue,
		Issue: &model.IssueData{
			Provider:  model.ProviderGitLab,
			Title:     issueResp.Title,
			URL:       issueResp.WebURL,
			Author:    model.UserData{Login: issueResp.Author.Username, URL: issueResp.Author.WebURL},
			CreatedAt: createdAt,
			State:     issueResp.State,
			Body:      issueResp.Description,
			Comments:  comments,
		},
	}, nil
}

type gitlabIssueResponse struct {
	Title       string `json:"title"`
	WebURL      string `json:"web_url"`
	State       string `json:"state"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	Author      struct {
		Username string `json:"username"`
		WebURL   string `json:"web_url"`
	} `json:"author"`
}

type gitlabNoteResponse struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	Author    struct {
		Username string `json:"username"`
		WebURL   string `json:"web_url"`
	} `json:"author"`
}

func (c *Client) getJSON(ctx context.Context, requestURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
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

func mapGitLabNotes(input []gitlabNoteResponse) ([]model.CommentData, error) {
	comments := make([]model.CommentData, 0, len(input))
	for _, item := range input {
		createdAt, err := time.Parse(time.RFC3339, item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse note created_at: %w", err)
		}

		comments = append(comments, model.CommentData{
			ID:        strconv.FormatInt(item.ID, 10),
			Author:    model.UserData{Login: item.Author.Username, URL: item.Author.WebURL},
			CreatedAt: createdAt,
			Body:      item.Body,
		})
	}

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	return comments, nil
}
