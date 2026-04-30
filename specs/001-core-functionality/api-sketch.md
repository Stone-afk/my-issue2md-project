# API Sketch: GitHub Fetching and Markdown Conversion

本文档是 `issue2md` MVP 的包级 API 草图，重点描述 `internal/github` 和 `internal/converter` 两个包的职责、核心数据结构与接口边界。

This document sketches the package-level API for the `issue2md` MVP, focusing on `internal/github` and `internal/converter`.

## 设计原则 / Design Principles

- 包职责保持内聚：`github` 只负责从 GitHub 获取并归一化远端数据；`converter` 只负责将归一化后的内容渲染为 GitHub Flavored Markdown。
- Keep packages cohesive: `github` fetches and normalizes remote GitHub data; `converter` renders normalized content into GitHub Flavored Markdown.
- 优先使用简单结构体和函数，不提前引入复杂接口层级。
- Prefer simple structs and functions; avoid premature interface hierarchies.
- 运行时依赖显式传入，例如 `http.Client`、`context.Context`、选项结构体。
- Pass runtime dependencies explicitly, such as `http.Client`, `context.Context`, and option structs.
- 错误必须显式处理，并在向上传递时使用 `fmt.Errorf("...: %w", err)` 包装。
- Handle errors explicitly and wrap propagated errors with `fmt.Errorf("...: %w", err)`.

## `internal/github`

### 职责 / Responsibilities

`internal/github` 负责和 GitHub API 交互，并把 Issue、Pull Request、Discussion 转换为统一的文档数据模型。

`internal/github` interacts with GitHub APIs and converts Issues, Pull Requests, and Discussions into a shared document data model.

该包应负责：

This package should handle:

- 根据 owner、repo、number、kind 获取远端内容。
- Fetching remote content by owner, repo, number, and kind.
- 在存在 `GITHUB_TOKEN` 时设置 GitHub API 认证头。
- Setting GitHub API authentication headers when `GITHUB_TOKEN` is available.
- 透传 GitHub API 错误，包括 rate limit。
- Surfacing GitHub API errors, including rate limit errors.
- 将 GitHub API 响应归一化为 `Document`、`Comment` 等内部结构。
- Normalizing GitHub API responses into internal structs such as `Document` and `Comment`.

该包不应负责：

This package should not handle:

- CLI flags 解析。
- CLI flag parsing.
- URL 字符串解析与类型识别。
- URL string parsing and type detection.
- Markdown 渲染。
- Markdown rendering.
- 文件输出。
- File output.

### Core Types

```go
package github

import "time"

type Kind string

const (
	KindIssue      Kind = "issue"
	KindPull       Kind = "pull_request"
	KindDiscussion Kind = "discussion"
)

type Target struct {
	Owner  string
	Repo   string
	Kind   Kind
	Number int
	URL    string
}

type User struct {
	Login   string
	HTMLURL string
}

type ReactionSummary struct {
	Total      int
	PlusOne    int
	MinusOne   int
	Laugh      int
	Hooray     int
	Confused   int
	Heart      int
	Rocket     int
	Eyes       int
}

type Document struct {
	Kind           Kind
	Title          string
	URL            string
	Author         User
	CreatedAt      time.Time
	State          string
	Body           string
	Comments       []Comment
	ReviewComments []Comment
	AcceptedAnswer *Comment
	Reactions      ReactionSummary
}

type Comment struct {
	ID        string
	Author    User
	CreatedAt time.Time
	Body      string
	Reactions ReactionSummary
}
```

### Client Sketch

```go
type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

type ClientOptions struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      string
}

func NewClient(opts ClientOptions) *Client

func (c *Client) Fetch(ctx context.Context, target Target, opts FetchOptions) (Document, error)

type FetchOptions struct {
	IncludeReactions bool
}
```

### Fetch Behavior

- `Fetch` 根据 `target.Kind` 分派到 Issue、Pull Request 或 Discussion 的获取逻辑。
- `Fetch` dispatches to Issue, Pull Request, or Discussion fetching based on `target.Kind`.
- Issue 获取主楼和全部 comments。
- Issue fetching includes the main issue body and all comments.
- Pull Request 获取 PR 描述和 review comments，不获取 diff 或 commit history。
- Pull Request fetching includes the PR description and review comments, but not diffs or commit history.
- Discussion 获取主楼、全部 comments 和 accepted answer 信息。
- Discussion fetching includes the discussion body, all comments, and accepted answer information.
- 所有 comments 在返回前按 `CreatedAt` 升序排序。
- All comments are sorted by `CreatedAt` ascending before being returned.

## `internal/converter`

### 职责 / Responsibilities

`internal/converter` 负责把 `github.Document` 渲染为标准 GitHub Flavored Markdown。

`internal/converter` renders `github.Document` into standard GitHub Flavored Markdown.

该包应负责：

This package should handle:

- 输出 YAML Frontmatter。
- Emitting YAML frontmatter.
- 输出 Summary / Structured Notes / Raw Archive 双层结构。
- Emitting the Summary / Structured Notes / Raw Archive dual-layer structure.
- 根据选项决定是否输出用户链接。
- Rendering user links based on options.
- 根据输入数据决定是否输出 reactions。
- Rendering reactions when present in the input document.
- 显著标记 Discussion accepted answer。
- Prominently marking Discussion accepted answers.

该包不应负责：

This package should not handle:

- GitHub API 请求。
- GitHub API requests.
- 环境变量读取。
- Reading environment variables.
- CLI 参数解析。
- CLI argument parsing.
- 文件写入或 stdout 选择。
- Choosing file output or stdout.

### Renderer Sketch

```go
package converter

import "github.com/stoneafk/issue2md/internal/github"

type Options struct {
	EnableUserLinks bool
}

type Renderer struct {
	Options Options
}

func NewRenderer(opts Options) Renderer

func (r Renderer) Render(doc github.Document) (string, error)
```

### Rendering Rules

- `Render` 返回完整 Markdown 字符串。
- `Render` returns a complete Markdown string.
- Frontmatter 至少包含 `title`、`url`、`author`、`created_at`。
- Frontmatter includes at least `title`, `url`, `author`, and `created_at`.
- `EnableUserLinks=false` 时，作者和评论者输出为纯文本 login。
- When `EnableUserLinks=false`, authors and commenters are rendered as plain login names.
- `EnableUserLinks=true` 时，如果 `User.HTMLURL` 非空，则输出 Markdown link。
- When `EnableUserLinks=true`, users are rendered as Markdown links if `User.HTMLURL` is available.
- `doc.Reactions` 或 `Comment.Reactions` 有数据时，渲染在对应内容附近。
- When `doc.Reactions` or `Comment.Reactions` contains data, render it near the related content.
- PR 文档只渲染描述和 review comments，不渲染 diff 或 commit history。
- Pull Request documents render only the description and review comments, not diffs or commit history.
- Discussion 的 `AcceptedAnswer` 必须用显著标记展示。
- Discussion `AcceptedAnswer` must be rendered with a prominent marker.

## Package Boundary Summary

```text
cmd/issue2md
  -> internal/cli
      -> internal/config
      -> internal/parser
      -> internal/github
      -> internal/converter

cmd/issue2mdweb
  -> future Web entrypoint

web/templates, web/static
  -> future Web assets
```

`parser.Target` 后续可以转换为 `github.Target`，或者直接复用 `github.Target`，以实现简单为准。

A future `parser.Target` may be converted into `github.Target`, or the project may directly reuse `github.Target`; choose the simpler option during implementation.