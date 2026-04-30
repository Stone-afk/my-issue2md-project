package model

import "time"

type Provider string

const (
	ProviderGitHub Provider = "github"
	ProviderGitLab Provider = "gitlab"
)

type ContentKind string

const (
	KindIssue       ContentKind = "issue"
	KindPullRequest ContentKind = "pull_request"
	KindDiscussion  ContentKind = "discussion"
)

type UserData struct {
	Login string
	URL   string
}

type ReactionSummary struct {
	Total    int
	PlusOne  int
	MinusOne int
	Laugh    int
	Hooray   int
	Confused int
	Heart    int
	Rocket   int
	Eyes     int
}

type CommentData struct {
	ID        string
	Author    UserData
	CreatedAt time.Time
	Body      string
	URL       string
	Reactions ReactionSummary
}

type IssueData struct {
	Provider  Provider
	Title     string
	URL       string
	Author    UserData
	CreatedAt time.Time
	State     string
	Body      string
	Comments  []CommentData
	Reactions ReactionSummary
}

type PullRequestData struct {
	Title          string
	URL            string
	Author         UserData
	CreatedAt      time.Time
	State          string
	Body           string
	ReviewComments []CommentData
	Reactions      ReactionSummary
}

type DiscussionData struct {
	Title          string
	URL            string
	Author         UserData
	CreatedAt      time.Time
	State          string
	Body           string
	Comments       []CommentData
	AcceptedAnswer *CommentData
	Reactions      ReactionSummary
}

type FetchOptions struct {
	IncludeReactions bool
}

type DocumentData struct {
	Provider    Provider
	Kind        ContentKind
	Issue       *IssueData
	PullRequest *PullRequestData
	Discussion  *DiscussionData
}
