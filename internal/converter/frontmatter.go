package converter

import (
	"fmt"
	"strings"
	"time"

	"github.com/stoneafk/issue2md/internal/model"
)

func renderFrontmatter(doc model.DocumentData) (string, error) {
	title, url, author, createdAt, err := frontmatterFields(doc)
	if err != nil {
		return "", fmt.Errorf("frontmatter fields: %w", err)
	}

	var builder strings.Builder
	builder.WriteString("---\n")
	builder.WriteString("title: \"")
	builder.WriteString(title)
	builder.WriteString("\"\n")
	builder.WriteString("url: \"")
	builder.WriteString(url)
	builder.WriteString("\"\n")
	builder.WriteString("author: \"")
	builder.WriteString(author)
	builder.WriteString("\"\n")
	builder.WriteString("created_at: \"")
	builder.WriteString(createdAt.Format(time.RFC3339))
	builder.WriteString("\"\n")
	builder.WriteString("---\n\n")
	return builder.String(), nil
}

func frontmatterFields(doc model.DocumentData) (string, string, string, time.Time, error) {
	switch doc.Kind {
	case model.KindIssue:
		if doc.Issue == nil {
			return "", "", "", time.Time{}, fmt.Errorf("missing issue data")
		}
		return doc.Issue.Title, doc.Issue.URL, doc.Issue.Author.Login, doc.Issue.CreatedAt, nil
	case model.KindPullRequest:
		if doc.PullRequest == nil {
			return "", "", "", time.Time{}, fmt.Errorf("missing pull request data")
		}
		return doc.PullRequest.Title, doc.PullRequest.URL, doc.PullRequest.Author.Login, doc.PullRequest.CreatedAt, nil
	case model.KindDiscussion:
		if doc.Discussion == nil {
			return "", "", "", time.Time{}, fmt.Errorf("missing discussion data")
		}
		return doc.Discussion.Title, doc.Discussion.URL, doc.Discussion.Author.Login, doc.Discussion.CreatedAt, nil
	default:
		return "", "", "", time.Time{}, fmt.Errorf("unsupported document kind %q", doc.Kind)
	}
}
