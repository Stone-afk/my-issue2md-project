# issue2md Minimal Project Map / 最小项目地图

> Purpose / 目的
>
> 这是一份给 AI 与维护者使用的轻量级项目地图，只保留高价值骨架信息：入口、核心包、关键签名、依赖方向、常用检索点。
>
> This is a lightweight project map for AI and maintainers. It keeps only the high-value skeleton: entrypoints, core packages, key signatures, dependency direction, and common lookup points.

## 1. Repo shape / 仓库骨架

```text
cmd/
  issue2md/          CLI 入口 / CLI entrypoint
  issue2mdweb/       Web 入口占位 / placeholder web entrypoint
internal/
  cli/               参数解析 + 组装执行流 / arg parsing + orchestration
  config/            环境变量加载 / env loading
  converter/         DocumentData -> Markdown
  fetchprovider/     provider 抽象 + GitHub/GitLab 实现
  model/             核心数据模型 / core document model
  parser/            URL -> parser.Target
.github/workflows/
  ci.yml             CI + Claude Headless failure diagnosis
Dockerfile           双模式镜像 / dual-mode image
Makefile             标准构建、测试、Docker 命令 / standard build, test, docker commands
```

## 2. Runtime flow / 运行主链路

```text
main()
  -> config.LoadFromEnv(envMap())
  -> assemble Providers map[model.Provider]fetchprovider.Provider
  -> cli.App.Run(args)
      -> parse flags
      -> parser.Parse(rawURL)
      -> provider.Fetch(ctx, target, fetchOpts)
      -> converter.NewRenderer(opts).Render(doc)
      -> stdout or output file
```

## 3. Entrypoints / 入口文件

### `cmd/issue2md/main.go`
- `func main()`
- `func envMap() map[string]string`
- Responsibility / 职责:
  - 读取环境变量 / read environment variables
  - 构造 provider registry / build provider registry
  - 注入 renderer factory / inject renderer factory
  - 启动 `cli.App` / start `cli.App`

### `cmd/issue2mdweb/main.go`
- `func main()`
- Current state / 当前状态:
  - 占位实现，暂无实际 Web 行为 / placeholder only, no real web behavior yet

## 4. Core packages / 核心包

### `internal/cli` — orchestration / 执行编排
**Key file:** `internal/cli/app.go`

**Important types / 关键类型**
- `type Renderer interface { Render(doc model.DocumentData) (string, error) }`
- `type Options struct { EnableUserLinks bool; EnableReactions bool }`
- `type App struct { Stdout, Stderr, Providers, Renderer, NewRenderer }`

**Important methods / 关键方法**
- `func (a App) Run(ctx context.Context, args []string) int`
- `func (a App) fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error)`
- `func (a App) renderer(opts Options) (Renderer, error)`

**What to inspect here / 这里适合查什么**
- CLI flags
- output file behavior
- provider dispatch
- top-level error wrapping

---

### `internal/parser` — URL parsing / URL 解析
**Key file:** `internal/parser/parser.go`

**Important type / 关键类型**
- `type Target struct { Provider, Kind, Owner, Repo, Project, Number, URL }`

**Important functions / 关键函数**
- `func Parse(rawURL string) (Target, error)`
- `func parseGitHubTarget(rawURL string, hostAndPath []string) (Target, error)`
- `func parseGitLabTarget(rawURL string, hostAndPath []string) (Target, error)`

**Current supported URLs / 当前支持的 URL**
- GitHub Issue
- GitHub Pull Request
- GitHub Discussion
- GitLab Issue
- HTTPS only

---

### `internal/model` — shared document model / 共享文档模型
**Key file:** `internal/model/document.go`

**Important enums / 关键枚举**
- `type Provider string`
  - `ProviderGitHub`
  - `ProviderGitLab`
- `type ContentKind string`
  - `KindIssue`
  - `KindPullRequest`
  - `KindDiscussion`

**Important structs / 关键结构**
- `type UserData`
- `type ReactionSummary`
- `type CommentData`
- `type IssueData`
- `type PullRequestData`
- `type DiscussionData`
- `type FetchOptions`
- `type DocumentData`

**Why it matters / 为什么重要**
- 这是 parser/provider/converter 之间的公共契约 / This is the shared contract between parser, providers, and converter.

---

### `internal/fetchprovider` — provider abstraction / provider 抽象层
**Key file:** `internal/fetchprovider/types.go`

**Important interface / 关键接口**
- `type Provider interface { Fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error) }`

**Important errors / 关键错误**
- `ErrProviderNotRegistered`
- `UnsupportedCapabilityError`
- `UnsupportedCapability(...)`
- `IsUnsupportedCapability(err error)`

**Design note / 设计说明**
- `App` 只依赖 `Provider.Fetch(...)`，不做 provider-specific 分支。
- `App` depends only on `Provider.Fetch(...)` and does not branch on provider-specific behavior.

---

### `internal/fetchprovider/github` — GitHub implementation
**Key file:** `internal/fetchprovider/github/client.go`

**Important types / 关键类型**
- `type Client`
- `type Options struct { HTTPClient *http.Client; BaseURL string; Token string }`

**Important functions / 关键函数**
- `func NewClient(opts Options) *Client`
- `func (c *Client) Fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error)`
- `func (c *Client) GetIssue(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error)`
- `func (c *Client) GetPullRequest(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error)`
- `func (c *Client) GetDiscussion(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error)`

**Current capabilities / 当前能力**
- Issue: REST
- Pull Request: REST issue + review comments
- Discussion: GraphQL
- Reactions: supported when `IncludeReactions` is true

**What to inspect here / 这里适合查什么**
- API integration bugs
- auth behavior via `GITHUB_TOKEN`
- reaction mapping
- PR/discussion support details

---

### `internal/fetchprovider/gitlab` — GitLab implementation
**Key file:** `internal/fetchprovider/gitlab/client.go`

**Important types / 关键类型**
- `type Client`
- `type Options struct { HTTPClient *http.Client; BaseURL string }`

**Important functions / 关键函数**
- `func NewClient(opts Options) *Client`
- `func (c *Client) Fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error)`
- `func (c *Client) GetIssue(ctx context.Context, target parser.Target, _ model.FetchOptions) (model.DocumentData, error)`

**Current limitation / 当前限制**
- 仅支持 GitLab Issue / supports GitLab issues only

---

### `internal/converter` — Markdown renderer / Markdown 渲染器
**Key files / 关键文件**
- `internal/converter/renderer.go`
- `internal/converter/frontmatter.go`
- `internal/converter/reactions.go`
- `internal/converter/users.go`

**Important types / 关键类型**
- `type Options struct { EnableUserLinks bool; EnableReactions bool }`
- `type Renderer struct { Options Options }`

**Important functions / 关键函数**
- `func NewRenderer(opts Options) Renderer`
- `func (r Renderer) Render(doc model.DocumentData) (string, error)`
- `func renderFrontmatter(doc model.DocumentData) (string, error)`
- `func renderReactions(summary model.ReactionSummary, enabled bool) string`
- `func (r Renderer) renderUser(user model.UserData) string`

**Rendered sections / 当前输出结构**
- YAML frontmatter
- `Summary / 摘要`
- `Structured Notes / 结构化笔记`
- `Raw Archive / 原始归档`
- comment sections
- accepted answer section for discussions

---

### `internal/config` — config loading / 配置加载
**Key file:** `internal/config/config.go`

**Important type / 关键类型**
- `type Config struct { GitHubToken string }`

**Important functions / 关键函数**
- `func LoadFromEnv(env map[string]string) Config`
- `func LoadFromLookup(lookup func(string) (string, bool)) Config`

**Current limitation / 当前限制**
- 当前只读取 `GITHUB_TOKEN` / currently reads only `GITHUB_TOKEN`

## 5. Dependency map / 依赖关系图

```text
cmd/issue2md
  -> internal/config
  -> internal/cli
  -> internal/converter
  -> internal/fetchprovider
  -> internal/fetchprovider/github
  -> internal/fetchprovider/gitlab
  -> internal/model

internal/cli
  -> internal/parser
  -> internal/model
  -> internal/fetchprovider

internal/fetchprovider/github
  -> internal/parser
  -> internal/model
  -> internal/fetchprovider

internal/fetchprovider/gitlab
  -> internal/parser
  -> internal/model
  -> internal/fetchprovider

internal/converter
  -> internal/model

internal/parser
  -> internal/model
```

## 6. Build, test, runtime surfaces / 构建、测试、运行入口

### `Makefile`
**Core targets / 核心目标**
- `make build`
- `make test`
- `make lint`
- `make format`
- `make verify`
- `make clean`
- `make web`
- `make docker-build`
- `make docker-run-cli`
- `make docker-run-web`
- `make dev-setup`

### `Dockerfile`
- Multi-stage build
- Builds both `issue2md` and `issue2mdweb`
- Final image: Alpine
- Non-root runtime
- Uses `/app/entrypoint.sh`

### `docker-entrypoint.sh`
- first arg = `web` -> exec `issue2mdweb`
- otherwise -> exec `issue2md`

### `.github/workflows/ci.yml`
- Trigger: push to `main`, pull_request
- Steps:
  - `make test`
  - `make lint`
  - `make build`
- On failure:
  - runs `anthropics/claude-code-action@v1` for CI log diagnosis

## 7. Fast lookup cheatsheet / 快速检索清单

### If you need... / 如果你要查...
- CLI 参数与输出行为 / CLI flags and output behavior  
  -> `internal/cli/app.go`
- 支持哪些 URL / supported URL patterns  
  -> `internal/parser/parser.go`
- 核心文档数据结构 / core document structs  
  -> `internal/model/document.go`
- provider 抽象或错误边界 / provider abstraction or capability errors  
  -> `internal/fetchprovider/types.go`
- GitHub API 行为 / GitHub API behavior  
  -> `internal/fetchprovider/github/client.go`
- GitLab API 行为 / GitLab API behavior  
  -> `internal/fetchprovider/gitlab/client.go`
- Markdown 渲染逻辑 / Markdown rendering logic  
  -> `internal/converter/*.go`
- 环境变量来源 / environment variables  
  -> `internal/config/config.go`, `cmd/issue2md/main.go`
- 构建与 Docker 运行 / build and docker runtime  
  -> `Makefile`, `Dockerfile`, `docker-entrypoint.sh`
- CI 行为 / CI behavior  
  -> `.github/workflows/ci.yml`

### Suggested Grep targets / 建议 Grep 关键词
- `EnableReactions`
- `EnableUserLinks`
- `ProviderGitHub`
- `ProviderGitLab`
- `KindDiscussion`
- `UnsupportedCapability`
- `Render(`
- `Parse(`
- `GITHUB_TOKEN`
- `docker-run-cli`

## 8. Current facts to avoid forgetting / 当前容易忘的事实

- `issue2mdweb` 目前是占位入口，不是完整 Web 服务。
- GitHub provider 支持 Issue / PR / Discussion。
- GitLab provider 当前仅支持 Issue。
- CLI 当前只读 `GITHUB_TOKEN`。
- 输出目标是 stdout 或 `[output_file]`，不是两者同时。
- reactions 只有在 `-enable-reactions` 打开时才渲染。
- user links 只有在 `-enable-user-links` 打开且用户 URL 存在时才渲染。

## 9. Suggested import snippet / 建议导入方式

如果要把这份地图接入 `CLAUDE.md`，建议只加一行引用：

If you want to wire this into `CLAUDE.md`, add a single import line:

```md
@./.claude/project-map.md
```

这样既保持主 `CLAUDE.md` 简洁，也能让 AI 在需要时加载项目骨架上下文。

This keeps the main `CLAUDE.md` compact while still making the project skeleton available when needed.
