package github

import (
	"bytes"
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

	switch target.Kind {
	case model.KindIssue:
		return c.GetIssue(ctx, target, opts)
	case model.KindPullRequest:
		return c.GetPullRequest(ctx, target, opts)
	case model.KindDiscussion:
		return c.GetDiscussion(ctx, target, opts)
	default:
		return model.DocumentData{}, fetchprovider.UnsupportedCapability(target.Provider, target.Kind)
	}
}

func (c *Client) GetIssue(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error) {
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

	comments, err := mapGitHubComments(commentsResp, opts.IncludeReactions)
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
			Reactions: mapRESTReactions(issueResp.Reactions, opts.IncludeReactions),
		},
	}, nil
}

func (c *Client) GetPullRequest(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error) {
	issueURL := fmt.Sprintf("%s/repos/%s/%s/issues/%d", c.baseURL, target.Owner, target.Repo, target.Number)
	reviewCommentsURL := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/comments", c.baseURL, target.Owner, target.Repo, target.Number)

	var issueResp githubIssueResponse
	if err := c.getJSON(ctx, issueURL, &issueResp); err != nil {
		return model.DocumentData{}, fmt.Errorf("get pull request: %w", err)
	}

	var reviewCommentsResp []githubCommentResponse
	if err := c.getJSON(ctx, reviewCommentsURL, &reviewCommentsResp); err != nil {
		return model.DocumentData{}, fmt.Errorf("get pull request review comments: %w", err)
	}

	reviewComments, err := mapGitHubComments(reviewCommentsResp, opts.IncludeReactions)
	if err != nil {
		return model.DocumentData{}, fmt.Errorf("map pull request review comments: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, issueResp.CreatedAt)
	if err != nil {
		return model.DocumentData{}, fmt.Errorf("parse pull request created_at: %w", err)
	}

	return model.DocumentData{
		Provider: model.ProviderGitHub,
		Kind:     model.KindPullRequest,
		PullRequest: &model.PullRequestData{
			Title:          issueResp.Title,
			URL:            issueResp.HTMLURL,
			Author:         model.UserData{Login: issueResp.User.Login, URL: issueResp.User.HTMLURL},
			CreatedAt:      createdAt,
			State:          issueResp.State,
			Body:           issueResp.Body,
			ReviewComments: reviewComments,
			Reactions:      mapRESTReactions(issueResp.Reactions, opts.IncludeReactions),
		},
	}, nil
}

func (c *Client) GetDiscussion(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error) {
	request := githubGraphQLRequest{
		Query: githubDiscussionQuery,
		Variables: map[string]any{
			"owner":  target.Owner,
			"repo":   target.Repo,
			"number": target.Number,
		},
	}

	var response githubDiscussionQueryResponse
	if err := c.postJSON(ctx, c.baseURL+"/graphql", request, &response); err != nil {
		return model.DocumentData{}, fmt.Errorf("get discussion: %w", err)
	}
	if len(response.Errors) > 0 {
		return model.DocumentData{}, fmt.Errorf("get discussion: graphql response: %s", response.Errors[0].Message)
	}
	if response.Data.Repository.Discussion == nil {
		return model.DocumentData{}, fmt.Errorf("get discussion: discussion not found")
	}

	discussion, err := mapGitHubDiscussion(*response.Data.Repository.Discussion, opts.IncludeReactions)
	if err != nil {
		return model.DocumentData{}, fmt.Errorf("map discussion: %w", err)
	}

	return model.DocumentData{
		Provider:   model.ProviderGitHub,
		Kind:       model.KindDiscussion,
		Discussion: discussion,
	}, nil
}

type githubIssueResponse struct {
	Title     string                `json:"title"`
	HTMLURL   string                `json:"html_url"`
	State     string                `json:"state"`
	Body      string                `json:"body"`
	CreatedAt string                `json:"created_at"`
	User      githubRESTUser        `json:"user"`
	Reactions githubReactionsObject `json:"reactions"`
}

type githubRESTUser struct {
	Login   string `json:"login"`
	HTMLURL string `json:"html_url"`
}

type githubCommentResponse struct {
	ID        int64                 `json:"id"`
	Body      string                `json:"body"`
	HTMLURL   string                `json:"html_url"`
	CreatedAt string                `json:"created_at"`
	User      githubRESTUser        `json:"user"`
	Reactions githubReactionsObject `json:"reactions"`
}

type githubReactionsObject struct {
	TotalCount int `json:"total_count"`
	PlusOne    int `json:"+1"`
	MinusOne   int `json:"-1"`
	Laugh      int `json:"laugh"`
	Hooray     int `json:"hooray"`
	Confused   int `json:"confused"`
	Heart      int `json:"heart"`
	Rocket     int `json:"rocket"`
	Eyes       int `json:"eyes"`
}

type githubGraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type githubGraphQLError struct {
	Message string `json:"message"`
}

type githubDiscussionQueryResponse struct {
	Data struct {
		Repository struct {
			Discussion *githubDiscussionNode `json:"discussion"`
		} `json:"repository"`
	} `json:"data"`
	Errors []githubGraphQLError `json:"errors"`
}

type githubDiscussionNode struct {
	Title          string                       `json:"title"`
	URL            string                       `json:"url"`
	Body           string                       `json:"body"`
	CreatedAt      string                       `json:"createdAt"`
	IsAnswered     bool                         `json:"isAnswered"`
	Author         *githubGraphQLUser           `json:"author"`
	ReactionGroups []githubReactionGroup        `json:"reactionGroups"`
	Answer         *githubDiscussionCommentNode `json:"answer"`
	Comments       struct {
		Nodes []githubDiscussionCommentNode `json:"nodes"`
	} `json:"comments"`
}

type githubDiscussionCommentNode struct {
	ID             string                    `json:"id"`
	Body           string                    `json:"body"`
	URL            string                    `json:"url"`
	CreatedAt      string                    `json:"createdAt"`
	Author         *githubGraphQLUser        `json:"author"`
	ReactionGroups []githubReactionGroup     `json:"reactionGroups"`
	Replies        githubDiscussionReplyList `json:"replies"`
}

type githubDiscussionReplyList struct {
	Nodes []githubDiscussionCommentNode `json:"nodes"`
}

type githubGraphQLUser struct {
	Login string `json:"login"`
	URL   string `json:"url"`
}

type githubReactionGroup struct {
	Content string `json:"content"`
	Users   struct {
		TotalCount int `json:"totalCount"`
	} `json:"users"`
}

func (c *Client) getJSON(ctx context.Context, requestURL string, out any) error {
	return c.doJSON(ctx, http.MethodGet, requestURL, nil, out)
}

func (c *Client) postJSON(ctx context.Context, requestURL string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}
	return c.doJSON(ctx, http.MethodPost, requestURL, bytes.NewReader(body), out)
}

func (c *Client) doJSON(ctx context.Context, method, requestURL string, body io.Reader, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
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
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func mapGitHubComments(input []githubCommentResponse, includeReactions bool) ([]model.CommentData, error) {
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
			Reactions: mapRESTReactions(item.Reactions, includeReactions),
		})
	}

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	return comments, nil
}

func mapGitHubDiscussion(input githubDiscussionNode, includeReactions bool) (*model.DiscussionData, error) {
	createdAt, err := time.Parse(time.RFC3339, input.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse discussion created_at: %w", err)
	}

	state := "open"
	if input.IsAnswered {
		state = "answered"
	}

	var acceptedAnswer *model.CommentData
	acceptedAnswerID := ""
	if input.Answer != nil {
		mappedAnswer, err := mapGitHubDiscussionComment(*input.Answer, includeReactions)
		if err != nil {
			return nil, fmt.Errorf("map accepted answer: %w", err)
		}
		acceptedAnswer = &mappedAnswer
		acceptedAnswerID = mappedAnswer.ID
	}

	comments, err := flattenGitHubDiscussionComments(input.Comments.Nodes, includeReactions, acceptedAnswerID)
	if err != nil {
		return nil, fmt.Errorf("map discussion comments: %w", err)
	}

	return &model.DiscussionData{
		Title:          input.Title,
		URL:            input.URL,
		Author:         mapGraphQLUser(input.Author),
		CreatedAt:      createdAt,
		State:          state,
		Body:           input.Body,
		Comments:       comments,
		AcceptedAnswer: acceptedAnswer,
		Reactions:      mapGraphQLReactions(input.ReactionGroups, includeReactions),
	}, nil
}

func flattenGitHubDiscussionComments(nodes []githubDiscussionCommentNode, includeReactions bool, acceptedAnswerID string) ([]model.CommentData, error) {
	comments := make([]model.CommentData, 0)
	for _, node := range nodes {
		mapped, err := mapGitHubDiscussionComment(node, includeReactions)
		if err != nil {
			return nil, err
		}
		if mapped.ID != acceptedAnswerID {
			comments = append(comments, mapped)
		}

		replies, err := flattenGitHubDiscussionComments(node.Replies.Nodes, includeReactions, acceptedAnswerID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, replies...)
	}

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	return comments, nil
}

func mapGitHubDiscussionComment(input githubDiscussionCommentNode, includeReactions bool) (model.CommentData, error) {
	createdAt, err := time.Parse(time.RFC3339, input.CreatedAt)
	if err != nil {
		return model.CommentData{}, fmt.Errorf("parse discussion comment created_at: %w", err)
	}

	return model.CommentData{
		ID:        input.ID,
		Author:    mapGraphQLUser(input.Author),
		CreatedAt: createdAt,
		Body:      input.Body,
		URL:       input.URL,
		Reactions: mapGraphQLReactions(input.ReactionGroups, includeReactions),
	}, nil
}

func mapGraphQLUser(input *githubGraphQLUser) model.UserData {
	if input == nil {
		return model.UserData{}
	}
	return model.UserData{Login: input.Login, URL: input.URL}
}

func mapRESTReactions(input githubReactionsObject, include bool) model.ReactionSummary {
	if !include {
		return model.ReactionSummary{}
	}
	return model.ReactionSummary{
		Total:    input.TotalCount,
		PlusOne:  input.PlusOne,
		MinusOne: input.MinusOne,
		Laugh:    input.Laugh,
		Hooray:   input.Hooray,
		Confused: input.Confused,
		Heart:    input.Heart,
		Rocket:   input.Rocket,
		Eyes:     input.Eyes,
	}
}

func mapGraphQLReactions(groups []githubReactionGroup, include bool) model.ReactionSummary {
	if !include {
		return model.ReactionSummary{}
	}

	var summary model.ReactionSummary
	for _, group := range groups {
		count := group.Users.TotalCount
		summary.Total += count
		switch group.Content {
		case "THUMBS_UP":
			summary.PlusOne += count
		case "THUMBS_DOWN":
			summary.MinusOne += count
		case "LAUGH":
			summary.Laugh += count
		case "HOORAY":
			summary.Hooray += count
		case "CONFUSED":
			summary.Confused += count
		case "HEART":
			summary.Heart += count
		case "ROCKET":
			summary.Rocket += count
		case "EYES":
			summary.Eyes += count
		}
	}
	return summary
}

const githubDiscussionQuery = `query Discussion($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    discussion(number: $number) {
      title
      url
      body
      createdAt
      isAnswered
      author {
        login
        url
      }
      reactionGroups {
        content
        users {
          totalCount
        }
      }
      answer {
        id
        body
        url
        createdAt
        author {
          login
          url
        }
        reactionGroups {
          content
          users {
            totalCount
          }
        }
      }
      comments(first: 100) {
        nodes {
          id
          body
          url
          createdAt
          author {
            login
            url
          }
          reactionGroups {
            content
            users {
              totalCount
            }
          }
          replies(first: 100) {
            nodes {
              id
              body
              url
              createdAt
              author {
                login
                url
              }
              reactionGroups {
                content
                users {
                  totalCount
                }
              }
              replies(first: 100) {
                nodes {
                  id
                  body
                  url
                  createdAt
                  author {
                    login
                    url
                  }
                  reactionGroups {
                    content
                    users {
                      totalCount
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`
