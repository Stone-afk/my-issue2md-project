package converter

import (
	"fmt"
	"strings"

	"github.com/stoneafk/issue2md/internal/model"
)

type Options struct {
	EnableUserLinks bool
	EnableReactions bool
}

type Renderer struct {
	Options Options
}

func NewRenderer(opts Options) Renderer {
	return Renderer{Options: opts}
}

func (r Renderer) Render(doc model.DocumentData) (string, error) {
	frontmatter, err := renderFrontmatter(doc)
	if err != nil {
		return "", fmt.Errorf("render frontmatter: %w", err)
	}

	var builder strings.Builder
	builder.WriteString(frontmatter)

	switch doc.Kind {
	case model.KindIssue:
		if doc.Issue == nil {
			return "", fmt.Errorf("render issue: missing issue data")
		}
		r.renderIssue(&builder, *doc.Issue)
	case model.KindPullRequest:
		if doc.PullRequest == nil {
			return "", fmt.Errorf("render pull request: missing pull request data")
		}
		r.renderPullRequest(&builder, *doc.PullRequest)
	case model.KindDiscussion:
		if doc.Discussion == nil {
			return "", fmt.Errorf("render discussion: missing discussion data")
		}
		r.renderDiscussion(&builder, *doc.Discussion)
	default:
		return "", fmt.Errorf("render document: unsupported kind %q", doc.Kind)
	}

	return builder.String(), nil
}

func (r Renderer) renderIssue(builder *strings.Builder, issue model.IssueData) {
	r.writeTitle(builder, issue.Title)
	r.writeSummary(builder, issue.Author, issue.State)
	r.writeStructuredNotes(builder, issue.Body, issue.Reactions)
	r.writeRawArchive(builder, issue.Body)
	for _, comment := range issue.Comments {
		r.writeComment(builder, "Comment by ", comment)
	}
}

func (r Renderer) renderPullRequest(builder *strings.Builder, pr model.PullRequestData) {
	r.writeTitle(builder, pr.Title)
	r.writeSummary(builder, pr.Author, pr.State)
	r.writeStructuredNotes(builder, pr.Body, pr.Reactions)
	r.writeRawArchive(builder, pr.Body)
	for _, comment := range pr.ReviewComments {
		r.writeComment(builder, "Review Comment by ", comment)
	}
}

func (r Renderer) renderDiscussion(builder *strings.Builder, discussion model.DiscussionData) {
	r.writeTitle(builder, discussion.Title)
	r.writeSummary(builder, discussion.Author, discussion.State)
	r.writeStructuredNotes(builder, discussion.Body, discussion.Reactions)
	if discussion.AcceptedAnswer != nil {
		builder.WriteString("### Accepted answer\n\n")
		builder.WriteString(discussion.AcceptedAnswer.Body)
		builder.WriteString("\n\n")
	}
	r.writeRawArchive(builder, discussion.Body)
	for _, comment := range discussion.Comments {
		r.writeComment(builder, "Comment by ", comment)
	}
}

func (r Renderer) writeTitle(builder *strings.Builder, title string) {
	builder.WriteString("# ")
	builder.WriteString(title)
	builder.WriteString("\n\n")
}

func (r Renderer) writeSummary(builder *strings.Builder, author model.UserData, state string) {
	builder.WriteString("## Summary / 摘要\n\n")
	builder.WriteString("- Author: ")
	builder.WriteString(r.renderUser(author))
	builder.WriteString("\n")
	builder.WriteString("- State: ")
	builder.WriteString(state)
	builder.WriteString("\n\n")
}

func (r Renderer) writeStructuredNotes(builder *strings.Builder, body string, reactions model.ReactionSummary) {
	builder.WriteString("## Structured Notes / 结构化笔记\n\n")
	builder.WriteString(body)
	builder.WriteString("\n")
	r.writeReactions(builder, reactions)
	builder.WriteString("\n")
}

func (r Renderer) writeRawArchive(builder *strings.Builder, body string) {
	builder.WriteString("## Raw Archive / 原始归档\n\n")
	builder.WriteString(body)
	builder.WriteString("\n\n")
}

func (r Renderer) writeComment(builder *strings.Builder, heading string, comment model.CommentData) {
	builder.WriteString("### ")
	builder.WriteString(heading)
	builder.WriteString(r.renderUser(comment.Author))
	builder.WriteString("\n\n")
	builder.WriteString(comment.Body)
	builder.WriteString("\n")
	r.writeReactions(builder, comment.Reactions)
	builder.WriteString("\n")
}

func (r Renderer) writeReactions(builder *strings.Builder, reactions model.ReactionSummary) {
	if rendered := renderReactions(reactions, r.Options.EnableReactions); rendered != "" {
		builder.WriteString(rendered)
	}
}
