# issue2md 核心功能规格 / Core Functionality Specification

## 变更记录 / Change Log

### 2026-04-28: 新增 GitLab Issue 支持 / Add GitLab Issue Support

- 变更：MVP 额外支持 GitLab 公有项目的 Issue URL。
- Change: The MVP additionally supports Issue URLs from public GitLab projects.
- 范围：仅支持 GitLab Issue；不支持 GitLab Merge Request、Epic、Snippet 或 Discussion。
- Scope: Only GitLab Issues are supported; GitLab Merge Requests, Epics, Snippets, and Discussions are not supported.
- 认证：本次变更不新增 GitLab token 认证；GitLab Issue MVP 仅支持公开可访问内容。
- Authentication: This change does not add GitLab token authentication; GitLab Issue support is limited to publicly accessible content.

## 概述 / Overview

`issue2md` 是一个命令行工具，用于将 GitHub Issue、Pull Request、Discussion URL 或 GitLab Issue URL 转换为结构清晰、格式良好的 GitHub Flavored Markdown 文档。

`issue2md` is a command-line tool that converts a GitHub Issue, Pull Request, Discussion URL, or GitLab Issue URL into a well-structured GitHub Flavored Markdown document.

MVP 的核心目标是知识沉淀：用户可以将 GitHub 上的协作讨论转换成本地 Markdown 文件。输出文档同时包含“结构化知识层”和“原始归档层”，既方便快速阅读，也方便长期保存和后续二次处理。

The MVP focuses on knowledge preservation. Users can convert GitHub conversations into local Markdown files that contain both a structured knowledge layer and a raw archive layer. The tool must be simple, scriptable, and suitable for shell workflows.

## 用户故事 / User Stories

### CLI MVP 用户故事 / CLI MVP User Stories

- 作为开发者，我希望运行 `issue2md <url>`，从而将 GitHub 或 GitLab 对话以 Markdown 形式输出到 stdout。
- As a developer, I want to run `issue2md <url>` so that I can print a GitHub or GitLab conversation as Markdown to stdout.

- 作为开发者，我希望运行 `issue2md <url> output.md`，从而将转换后的 Markdown 直接保存到文件。
- As a developer, I want to run `issue2md <url> output.md` so that I can save the converted Markdown directly to a file.

- 作为维护者，我希望工具能自动识别 Issue URL，从而不需要手动指定内容类型。
- As a maintainer, I want Issue URLs to be detected automatically so that I do not need to specify the content type manually.

- 【变更 / Changed】作为用户，我希望工具能自动识别 GitLab Issue URL，从而可以用同一个 CLI 归档 GitLab Issue。
- 【Changed】As a user, I want GitLab Issue URLs to be detected automatically so that I can archive GitLab Issues with the same CLI.

- 作为维护者，我希望工具能自动识别 Pull Request URL，从而可以归档 PR 描述和 Review 评论。
- As a maintainer, I want Pull Request URLs to be detected automatically so that review discussions can be archived without copying content by hand.

- 作为社区管理者，我希望工具能自动识别 Discussion URL，从而可以保存讨论内容、评论和已采纳答案。
- As a community manager, I want Discussion URLs to be detected automatically so that accepted answers and comments can be preserved as Markdown.

- 作为知识工作者，我希望评论按照时间正序排列，从而让 Markdown 保留原始讨论脉络。
- As a knowledge worker, I want comments sorted from oldest to newest so that the Markdown reflects the conversation timeline.

- 作为用户，我希望 reactions 是可选输出，从而在默认情况下保持文档简洁，在需要详细归档时再启用。
- As a user, I want reactions to be optional so that the default output stays focused, while detailed archives can include reaction context.

- 作为用户，我希望作者链接是可选输出，从而可以在纯净文本和富链接输出之间选择。
- As a user, I want author links to be optional so that I can choose between clean plain-text output and richer link-preserving output.

- 作为重视安全的用户，我希望认证只通过 `GITHUB_TOKEN` 环境变量完成，从而避免 token 出现在 shell history 中。
- As a security-conscious user, I want authentication to use only the `GITHUB_TOKEN` environment variable so that tokens do not appear in shell history.

### 未来 Web 版用户故事 / Future Web User Stories

以下用户故事用于指导未来产品设计，但不属于 MVP 实现范围。

The following stories are intentionally out of scope for the MVP implementation but should guide future product design.

- 作为非 CLI 用户，我希望在 Web 界面中粘贴 GitHub URL，从而无需使用终端也能下载生成的 Markdown。
- As a non-CLI user, I want to paste a GitHub URL into a Web interface so that I can download the generated Markdown without using a terminal.

- 作为团队成员，我希望 Web 界面可以预览生成的 Markdown，从而在下载前确认内容结构。
- As a team member, I want a Web interface to preview generated Markdown before downloading it.

- 作为组织管理员，我希望 Web 版本支持受控认证流程，从而安全处理私有或内部仓库内容。
- As an organization admin, I want the Web version to support a controlled authentication flow so that private or internal repositories can be handled safely.

## 功能性需求 / Functional Requirements

### URL 识别 / URL Recognition

- 工具必须自动识别受支持的 GitHub 和 GitLab URL 类型。
- The tool must automatically recognize supported GitHub and GitLab URL types.

- MVP 支持以下 URL 类型：
- Supported MVP URL types:
  - GitHub Issue URL，例如 / for example: `https://github.com/OWNER/REPO/issues/123`
  - GitHub Pull Request URL，例如 / for example: `https://github.com/OWNER/REPO/pull/123`
  - GitHub Discussion URL，例如 / for example: `https://github.com/OWNER/REPO/discussions/123`
  - 【变更 / Changed】GitLab Issue URL，例如 / for example: `https://gitlab.com/GROUP/PROJECT/-/issues/123`

- 用户不需要传入类型标记或 type flag。
- The user must not need to pass a type flag.

- 不支持或格式错误的 URL 必须输出错误到 stderr，并以非零状态码退出。
- Unsupported or malformed URLs must be reported to stderr and must cause a non-zero exit code.

### CLI 接口 / CLI Interface

MVP 命令格式如下：

The MVP command syntax is:

```sh
issue2md [flags] <url> [output_file]
```

支持的 flags：

Supported flags:

- `-enable-reactions`
  - 启用时，在 Markdown 中包含可用的 reaction 信息。
  - When present, include available reaction information in the Markdown output.
  - 未启用时，省略 reaction 信息。
  - When absent, omit reaction information.

- `-enable-user-links`
  - 启用时，如果 GitHub 用户主页 URL 可用，则将用户名渲染为 Markdown 链接。
  - When present, render user names as links to their GitHub profiles when profile URLs are available.
  - 未启用时，将用户名渲染为纯文本。
  - When absent, render user names as plain text.

参数说明：

Arguments:

- `<url>` 是必填参数。
- `<url>` is required.
- `[output_file]` 是可选参数。
- `[output_file]` is optional.
- 如果省略 `[output_file]`，工具必须将 Markdown 写入 stdout。
- If `[output_file]` is omitted, the tool must write Markdown to stdout.
- 如果提供 `[output_file]`，工具必须将 Markdown 写入该文件。
- If `[output_file]` is provided, the tool must write Markdown to that file.

### 认证 / Authentication

- MVP 支持 GitHub 和 GitLab 公有仓库或项目。
- The MVP supports public GitHub repositories and public GitLab projects.

- 当设置了 `GITHUB_TOKEN` 环境变量时，工具可以使用其中的 GitHub Personal Access Token。
- The tool may use a GitHub Personal Access Token when the `GITHUB_TOKEN` environment variable is set.

- 【变更 / Changed】GitLab Issue 支持不新增认证方式；MVP 仅访问公开可读的 GitLab Issue。
- 【Changed】GitLab Issue support does not add a new authentication method; the MVP only accesses publicly readable GitLab Issues.

- token 不得通过命令行 flag 或位置参数传入。
- The token must not be accepted through command-line flags or positional arguments.

- 如果未设置 `GITHUB_TOKEN`，工具必须尝试以未认证方式访问公有 GitHub 内容。
- If `GITHUB_TOKEN` is not set, the tool must attempt unauthenticated access to public GitHub content.

### GitHub 内容获取 / GitHub Content Fetching

#### Issues

对于 Issue URL，生成的文档必须包含：

For an Issue URL, the generated document must include:

- 标题 / Title
- URL
- 作者 / Author
- 创建时间 / Created time
- 状态 / State
- Issue 主体内容 / Main issue body
- 所有 Issue 评论 / All issue comments

#### GitLab Issues / GitLab Issues

【变更 / Changed】对于 GitLab Issue URL，生成的文档必须包含：

【Changed】For a GitLab Issue URL, the generated document must include:

- 标题 / Title
- URL
- 作者 / Author
- 创建时间 / Created time
- 状态 / State
- GitLab Issue 主体内容 / Main GitLab issue body
- 所有 GitLab Issue 评论或 notes / All GitLab issue comments or notes

GitLab Issue MVP 不支持 GitLab Merge Request、Epic、Snippet 或其他 GitLab 对象。

The GitLab Issue MVP does not support GitLab Merge Requests, Epics, Snippets, or other GitLab objects.

#### Pull Requests

对于 Pull Request URL，生成的文档必须包含：

For a Pull Request URL, the generated document must include:

- 标题 / Title
- URL
- 作者 / Author
- 创建时间 / Created time
- 状态 / State
- PR 描述 / Main PR description
- Review 评论 / Review comments

MVP 中 Pull Request 输出不得包含：

For the MVP, Pull Request output must not include:

- Diff 内容 / Diff content
- Commit 历史 / Commit history

PR Review Comments 不需要分组，必须按照时间线渲染为一个评论列表。

PR review comments must not be grouped. They must be rendered as a chronological list.

#### Discussions

对于 Discussion URL，生成的文档必须包含：

For a Discussion URL, the generated document must include:

- 标题 / Title
- URL
- 作者 / Author
- 创建时间 / Created time
- 状态或等价讨论状态（如果可用）/ State or equivalent discussion status when available
- Discussion 主体内容 / Main discussion body
- 所有 Discussion 评论 / All discussion comments
- 如果存在已采纳答案，必须使用显著标记展示 / A prominent marker for the accepted answer when an accepted answer exists

### 排序 / Ordering

- 评论必须按创建时间升序排列。
- Comments must be sorted by creation time in ascending order.

- PR Review Comments 也必须按创建时间升序排列。
- PR review comments must also be sorted by creation time in ascending order.

- 输出应尽可能保留原始讨论时间线。
- The output should preserve the conversation timeline as closely as possible.

### Markdown 渲染 / Markdown Rendering

- 输出格式必须是标准 GitHub Flavored Markdown。
- The output format must be standard GitHub Flavored Markdown.

- Markdown 文档必须包含 YAML Frontmatter，至少包含：
- The Markdown document must contain YAML frontmatter with at least:
  - `title`
  - `url`
  - `author`
  - `created_at`

- 文档必须使用双层结构：
- The document must use a dual-layer structure:
  - 结构化知识层，用于快速阅读和后续总结。
  - A structured knowledge layer for quick reading and future summarization.
  - 原始归档层，用于保存原始对话内容。
  - A raw archive layer that preserves the original conversation content.

- 图片和附件必须保留为原始链接。
- Images and attachments must remain as original links.

- MVP 不下载图片或附件到本地。
- The MVP must not download images or attachments to local storage.

### Reactions

- Reactions 是可选内容，由 `-enable-reactions` 控制。
- Reactions are optional and controlled by `-enable-reactions`.

- 未启用该 flag 时，输出中不得出现 reactions。
- When the flag is absent, reactions must not appear in the output.

- 启用该 flag 时，如果 GitHub 提供 reaction 信息，应将其展示在对应内容附近。
- When the flag is present, reactions should be included near the item they belong to when available from GitHub.

### 用户链接 / User Links

- 用户链接是可选内容，由 `-enable-user-links` 控制。
- User links are optional and controlled by `-enable-user-links`.

- 未启用该 flag 时，用户必须渲染为纯文本，例如 `octocat`。
- When the flag is absent, users must be rendered as plain names, for example `octocat`.

- 启用该 flag 时，如果用户主页 URL 可用，应渲染为 Markdown 链接，例如 `[octocat](https://github.com/octocat)`。
- When the flag is present, users should be rendered as Markdown links when profile URLs are available, for example `[octocat](https://github.com/octocat)`.

### 错误处理 / Error Handling

- 无效或不支持的 URL 必须输出到 stderr，并以非零状态码退出。
- Invalid or unsupported URLs must be printed to stderr and must exit with a non-zero status.

- GitHub API 错误和 GitLab API 错误必须作为 API 错误透传给用户。
- GitHub API errors and GitLab API errors must be surfaced to the user as API errors.

- Rate limit 响应必须作为 GitHub 或 GitLab API 错误透传。
- Rate limit responses must be passed through as GitHub or GitLab API errors.

- MVP 不实现重试机制。
- The MVP must not implement retry behavior.

## 非功能性需求 / Non-Functional Requirements

- 实现应保持 URL 解析、GitHub 数据获取、内容归一化和 Markdown 渲染之间的解耦。
- The implementation should keep URL parsing, GitHub fetching, content normalization, and Markdown rendering decoupled.

- 实现必须优先使用 Go 标准库，避免不必要的依赖。
- The implementation must prefer the Go standard library and avoid unnecessary dependencies.

- HTTP client、writer、环境变量等运行时依赖应显式传递，不应隐藏在全局状态中。
- Runtime dependencies such as HTTP clients, writers, and environment values should be passed explicitly rather than hidden behind global state.

- 后续 Go 实现必须显式处理所有错误。
- Future Go implementation must explicitly handle errors.

- Go 代码中向上传递错误时，必须使用 `fmt.Errorf("...: %w", err)` 添加上下文。
- Propagated errors in Go code must be wrapped with contextual messages using `fmt.Errorf("...: %w", err)`.

- 新功能实现必须从失败测试开始，并遵循 Red-Green-Refactor 流程。
- New implementation work must start with failing tests and follow a Red-Green-Refactor workflow.

- 单元测试应优先采用表格驱动测试。
- Unit tests should prefer table-driven tests.

- CLI 必须适合脚本化使用：stdout 输出确定性内容，stderr 输出错误，并提供有意义的退出码。
- The CLI must be script-friendly: deterministic stdout output, stderr for errors, and meaningful exit codes.

## 验收标准 / Acceptance Criteria

### CLI 行为 / CLI Behavior

- 给定一个有效 Issue URL 且未提供输出文件，当用户运行 `issue2md <issue-url>` 时，Markdown 被写入 stdout。
- Given a valid Issue URL and no output file, when the user runs `issue2md <issue-url>`, then Markdown is written to stdout.

- 给定一个有效 Issue URL 且提供输出文件，当用户运行 `issue2md <issue-url> issue.md` 时，Markdown 被写入 `issue.md`。
- Given a valid Issue URL and an output file, when the user runs `issue2md <issue-url> issue.md`, then Markdown is written to `issue.md`.

- 给定未提供 URL，当用户运行 `issue2md` 时，usage 或错误信息被写入 stderr，并以非零状态码退出。
- Given no URL, when the user runs `issue2md`, then usage or an error is written to stderr and the command exits non-zero.

- 给定无效 URL，当用户运行 `issue2md not-a-url` 时，无效 URL 错误被写入 stderr，并以非零状态码退出。
- Given an invalid URL, when the user runs `issue2md not-a-url`, then an invalid URL error is written to stderr and the command exits non-zero.

- 给定不支持的 GitHub URL，当用户运行 `issue2md https://github.com/OWNER/REPO/actions` 时，不支持 URL 错误被写入 stderr，并以非零状态码退出。
- Given an unsupported GitHub URL, when the user runs `issue2md https://github.com/OWNER/REPO/actions`, then an unsupported URL error is written to stderr and the command exits non-zero.

### URL 检测 / URL Detection

- 给定 `https://github.com/OWNER/REPO/issues/123`，工具将内容类型识别为 GitHub Issue。
- Given `https://github.com/OWNER/REPO/issues/123`, the tool detects the content type as GitHub Issue.

- 给定 `https://github.com/OWNER/REPO/pull/123`，工具将内容类型识别为 Pull Request。
- Given `https://github.com/OWNER/REPO/pull/123`, the tool detects the content type as Pull Request.

- 给定 `https://github.com/OWNER/REPO/discussions/123`，工具将内容类型识别为 Discussion。
- Given `https://github.com/OWNER/REPO/discussions/123`, the tool detects the content type as Discussion.

- 【变更 / Changed】给定 `https://gitlab.com/GROUP/PROJECT/-/issues/123`，工具将内容类型识别为 GitLab Issue。
- 【Changed】Given `https://gitlab.com/GROUP/PROJECT/-/issues/123`, the tool detects the content type as GitLab Issue.

- 给定一个带额外 query 参数的受支持 URL，工具仍能正确识别基础内容类型。
- Given a supported URL with additional query parameters, the tool still identifies the base content type correctly.

### 认证 / Authentication

- 给定已设置 `GITHUB_TOKEN`，工具使用该 token 发起 GitHub API 请求。
- Given `GITHUB_TOKEN` is set, the tool uses it for GitHub API requests.

- 给定未设置 `GITHUB_TOKEN`，工具尝试未认证访问公有内容。
- Given `GITHUB_TOKEN` is not set, the tool attempts unauthenticated public access.

- 给定 token-like 值通过 CLI 参数或 flag 传入，工具不得将其视为认证 token。
- Given a token-like value is provided as a CLI argument or flag, the tool does not treat it as an authentication token.

### Issue 转换 / Issue Conversion

- 给定一个包含标题、作者、创建时间、状态、正文和评论的 Issue，Markdown 包含所有必需字段。
- Given an Issue with a title, author, created time, state, body, and comments, the Markdown includes all required fields.

- 给定多个创建时间不同的 Issue 评论，Markdown 按创建时间升序渲染评论。
- Given Issue comments with different creation times, the Markdown renders comments in ascending creation-time order.

- 给定未启用 `-enable-reactions`，Issue reactions 被省略。
- Given `-enable-reactions` is absent, Issue reactions are omitted.

- 给定启用 `-enable-reactions`，可用的 Issue reactions 被包含在输出中。
- Given `-enable-reactions` is present, available Issue reactions are included.

- 给定未启用 `-enable-user-links`，Issue 作者和评论者被渲染为纯文本。
- Given `-enable-user-links` is absent, Issue authors and commenters are rendered as plain names.

- 给定启用 `-enable-user-links`，可用的 Issue 作者和评论者被渲染为 Markdown 链接。
- Given `-enable-user-links` is present, available Issue authors and commenters are rendered as Markdown links.

### GitLab Issue 转换 / GitLab Issue Conversion

- 【变更 / Changed】给定一个包含标题、作者、创建时间、状态、正文和 notes 的 GitLab Issue，Markdown 包含所有必需字段。
- 【Changed】Given a GitLab Issue with a title, author, created time, state, body, and notes, the Markdown includes all required fields.

- 【变更 / Changed】给定多个创建时间不同的 GitLab Issue notes，Markdown 按创建时间升序渲染。
- 【Changed】Given GitLab Issue notes with different creation times, the Markdown renders notes in ascending creation-time order.

- 【变更 / Changed】给定 GitLab Issue URL，输出仍使用与 GitHub Issue 一致的双层 Markdown 结构。
- 【Changed】Given a GitLab Issue URL, the output still uses the same dual-layer Markdown structure as GitHub Issues.

### Pull Request 转换 / Pull Request Conversion

- 给定一个包含标题、作者、创建时间、状态、描述和 Review 评论的 PR，Markdown 包含这些字段。
- Given a PR with a title, author, created time, state, description, and review comments, the Markdown includes those fields.

- 给定一个包含 diff 和 commits 的 PR，Markdown 不包含 diff 内容或 commit 历史。
- Given a PR with diff content and commits, the Markdown does not include diff content or commit history.

- 给定多个创建时间不同的 PR Review Comments，Markdown 按创建时间升序渲染。
- Given PR review comments with different creation times, the Markdown renders review comments in ascending creation-time order.

- 给定来自同一个 review 的多个 PR Review Comments，Markdown 不按 review 分组。
- Given multiple PR review comments from the same review, the Markdown does not group them by review.

### Discussion 转换 / Discussion Conversion

- 给定一个包含标题、作者、创建时间、正文和评论的 Discussion，Markdown 包含所有必需字段。
- Given a Discussion with a title, author, created time, body, and comments, the Markdown includes all required fields.

- 给定一个包含已采纳答案的 Discussion，Markdown 使用显著标记渲染已采纳答案。
- Given a Discussion with an accepted answer, the Markdown renders the accepted answer with a prominent marker.

- 给定多个创建时间不同的 Discussion 评论，Markdown 按创建时间升序渲染。
- Given Discussion comments with different creation times, the Markdown renders comments in ascending creation-time order.

### 错误处理 / Error Handling

- 给定 GitHub 返回 not found 响应，工具报告 GitHub API 错误并以非零状态码退出。
- Given GitHub returns a not found response, the tool reports the GitHub API error and exits non-zero.

- 给定 GitHub 返回 rate limit 响应，工具报告 GitHub API 错误并以非零状态码退出。
- Given GitHub returns a rate limit response, the tool reports the GitHub API error and exits non-zero.

- 给定文件写入失败，工具将文件输出错误写入 stderr 并以非零状态码退出。
- Given a file write fails, the tool reports the file output error to stderr and exits non-zero.

- 给定网络请求失败，工具将请求错误写入 stderr 并以非零状态码退出。
- Given a network request fails, the tool reports the request error to stderr and exits non-zero.

## 具体测试用例 / Concrete Test Cases

- GitLab Issue 输出到 stdout / GitLab Issue to stdout
  - Command: `issue2md https://gitlab.com/GROUP/PROJECT/-/issues/1`
  - Expected: 将 GitLab Issue 转换为 GFM Markdown 并写入 stdout / Converts the GitLab Issue to GFM Markdown and writes it to stdout
- GitLab Issue comments 排序 / GitLab Issue comments ordering
  - Command: `issue2md https://gitlab.com/GROUP/PROJECT/-/issues/1`
  - Expected: 按创建时间升序渲染 GitLab Issue notes / Renders GitLab Issue notes in ascending creation-time order
- GitLab 非 Issue URL / GitLab non-Issue URL
  - Command: `issue2md https://gitlab.com/GROUP/PROJECT/-/merge_requests/1`
  - Expected: 写入 unsupported URL 错误并非零退出 / Writes unsupported URL error to stderr and exits non-zero
- GitLab Issue 输出到文件 / GitLab Issue to file
  - Command: `issue2md https://gitlab.com/GROUP/PROJECT/-/issues/1 gitlab-issue.md`
  - Expected: 创建或覆盖 `gitlab-issue.md` / Creates or overwrites `gitlab-issue.md` with Markdown
- GitLab Issue 公开访问失败 / GitLab Issue public access failure
  - Command: `issue2md https://gitlab.com/GROUP/PROJECT/-/issues/1`
  - Expected: 透传 GitLab API 错误并非零退出 / Surfaces GitLab API error and exits non-zero
- Issue 输出到 stdout / Issue to stdout
  - Command: `issue2md https://github.com/OWNER/REPO/issues/1`
  - Expected: 将 GFM Markdown 写入 stdout / Writes GFM Markdown to stdout
- Issue 输出到文件 / Issue to file
  - Command: `issue2md https://github.com/OWNER/REPO/issues/1 issue.md`
  - Expected: 创建或覆盖 `issue.md` / Creates or overwrites `issue.md` with Markdown
- PR 输出到 stdout / PR to stdout
  - Command: `issue2md https://github.com/OWNER/REPO/pull/2`
  - Expected: 包含 PR 描述和 Review 评论 / Includes PR description and review comments
- PR 排除 diff / PR excludes diff
  - Command: `issue2md https://github.com/OWNER/REPO/pull/2`
  - Expected: 不包含 diff hunks 或 commit history / Does not include diff hunks or commit history
- Discussion 输出到 stdout / Discussion to stdout
  - Command: `issue2md https://github.com/OWNER/REPO/discussions/3`
  - Expected: 包含 discussion body 和 comments / Includes discussion body and comments
- Discussion 已采纳答案 / Discussion accepted answer
  - Command: `issue2md https://github.com/OWNER/REPO/discussions/3`
  - Expected: 已采纳答案被显著标记 / Accepted answer is visibly marked
- 包含 reactions / Include reactions
  - Command: `issue2md -enable-reactions https://github.com/OWNER/REPO/issues/1`
  - Expected: 包含可用 reaction 信息 / Includes available reaction counts/details
- 默认省略 reactions / Omit reactions by default
  - Command: `issue2md https://github.com/OWNER/REPO/issues/1`
  - Expected: 不包含 reactions / Does not include reactions
- 包含用户链接 / Include user links
  - Command: `issue2md -enable-user-links https://github.com/OWNER/REPO/issues/1`
  - Expected: 将可用用户渲染为 Markdown 链接 / Renders available users as Markdown links
- 默认省略用户链接 / Omit user links by default
  - Command: `issue2md https://github.com/OWNER/REPO/issues/1`
  - Expected: 将用户渲染为纯文本 / Renders users as plain text
- 无效 URL / Invalid URL
  - Command: `issue2md not-a-url`
  - Expected: 写入 stderr 并非零退出 / Writes error to stderr and exits non-zero
- 不支持 URL / Unsupported URL
  - Command: `issue2md https://github.com/OWNER/REPO/actions`
  - Expected: 写入 unsupported URL 错误并非零退出 / Writes unsupported URL error to stderr and exits non-zero
- 缺少 URL / Missing URL
  - Command: `issue2md`
  - Expected: 写入 usage/error 并非零退出 / Writes usage/error to stderr and exits non-zero
- API rate limit
  - Command: `issue2md https://github.com/OWNER/REPO/issues/1`
  - Expected: 透传 GitHub API 错误并非零退出 / Surfaces GitHub API error and exits non-zero
- 文件写入失败 / File write failure
  - Command: `issue2md https://github.com/OWNER/REPO/issues/1 /no/such/dir/out.md`
  - Expected: 写入文件错误并非零退出 / Writes file error to stderr and exits non-zero

## 输出格式示例 / Output Format Examples

### 通用 Markdown 结构 / General Markdown Structure

```markdown
---
title: "Example Issue Title"
url: "https://github.com/OWNER/REPO/issues/123"
author: "octocat"
created_at: "2026-04-28T10:30:00Z"
---

# Example Issue Title

## Summary / 摘要

- Type: Issue
- State: open
- Author: octocat
- Created: 2026-04-28T10:30:00Z
- Source: https://github.com/OWNER/REPO/issues/123

## Structured Notes / 结构化笔记

### Background / 背景

_To be filled by the reader or a future summarization workflow._

_由读者或未来的总结流程补充。_

### Key Discussion Points / 关键讨论点

_To be filled by the reader or a future summarization workflow._

_由读者或未来的总结流程补充。_

### Decisions / 决策

_To be filled by the reader or a future summarization workflow._

_由读者或未来的总结流程补充。_

### Action Items / 行动项

_To be filled by the reader or a future summarization workflow._

_由读者或未来的总结流程补充。_

## Raw Archive / 原始归档

### Original Post / 原始内容

**Author / 作者:** octocat  
**Created / 创建时间:** 2026-04-28T10:30:00Z

Original issue body in GitHub Flavored Markdown.

Images remain as original links:

![screenshot](https://github.com/OWNER/REPO/assets/example.png)

### Comments / 评论

#### Comment 1 / 评论 1

**Author / 作者:** contributor  
**Created / 创建时间:** 2026-04-28T11:00:00Z

First comment body.

#### Comment 2 / 评论 2

**Author / 作者:** maintainer  
**Created / 创建时间:** 2026-04-28T12:00:00Z

Second comment body.
```

### Pull Request 示例说明 / Pull Request Example Notes

```markdown
## Summary / 摘要

- Type: Pull Request
- State: open
- Author: octocat
- Created: 2026-04-28T10:30:00Z
- Source: https://github.com/OWNER/REPO/pull/456

## Raw Archive / 原始归档

### Pull Request Description / PR 描述

**Author / 作者:** octocat  
**Created / 创建时间:** 2026-04-28T10:30:00Z

PR description body.

### Review Comments / Review 评论

#### Review Comment 1 / Review 评论 1

**Author / 作者:** reviewer  
**Created / 创建时间:** 2026-04-28T11:15:00Z

Review comment body.
```

Pull Request 输出不得包含 diff 内容或 commit 历史。

The Pull Request output must not include diff content or commit history.

### Discussion 示例说明 / Discussion Example Notes

```markdown
## Summary / 摘要

- Type: Discussion
- State: answered
- Author: octocat
- Created: 2026-04-28T10:30:00Z
- Source: https://github.com/OWNER/REPO/discussions/789

## Raw Archive / 原始归档

### Discussion Body / Discussion 正文

**Author / 作者:** octocat  
**Created / 创建时间:** 2026-04-28T10:30:00Z

Discussion body.

### Comments / 评论

#### Comment 1 / 评论 1

**Author / 作者:** community-member  
**Created / 创建时间:** 2026-04-28T11:00:00Z

Comment body.

#### Accepted Answer / 已采纳答案

> **Accepted Answer / 已采纳答案**

**Author / 作者:** maintainer  
**Created / 创建时间:** 2026-04-28T12:00:00Z

Accepted answer body.
```

## 合宪性审查 / Constitution Compliance Review

### 需求变更 / Requirement Change

- 本次变更新增 GitLab Issue URL 支持。
- This change adds GitLab Issue URL support.
- 变更范围限定为公开 GitLab Issue，不新增 GitLab Merge Request、Epic、Snippet、Discussion 或 GitLab token 认证。
- The scope is limited to public GitLab Issues and does not add GitLab Merge Requests, Epics, Snippets, Discussions, or GitLab token authentication.

### 审查结论 / Review Result

- 结论：通过。
- Result: Pass.

### 原则核对 / Principle Checklist

- 简单性原则：通过。变更只增加 spec 明确要求的 GitLab Issue URL 支持，不引入额外平台抽象或复杂认证体系。
- Simplicity First: Pass. The change only adds explicitly requested GitLab Issue URL support and does not introduce broad provider abstractions or complex authentication.

- 标准库优先：通过。后续实现可继续使用 Go 标准库的 `net/http`、`net/url`、`encoding/json` 等能力。
- Standard Library First: Pass. Future implementation can continue using Go standard library packages such as `net/http`, `net/url`, and `encoding/json`.

- 测试先行：通过。后续实现必须先补充 GitLab Issue URL 解析、获取、排序和 Markdown 输出的失败测试。
- Test First: Pass. Future implementation must first add failing tests for GitLab Issue URL parsing, fetching, ordering, and Markdown output.

- 明确错误处理：通过。GitLab API 错误和公开访问失败必须与 GitHub API 错误一样透传并显式处理。
- Explicit Error Handling: Pass. GitLab API errors and public access failures must be surfaced and handled explicitly like GitHub API errors.

- 无全局状态：通过。GitLab 相关 HTTP client、base URL 和运行时选项应通过结构体或函数参数显式传入。
- No Global State: Pass. GitLab-related HTTP clients, base URLs, and runtime options should be passed explicitly through structs or function parameters.

## 非 MVP 范围 / Out of Scope

- MVP 不实现 Web UI。
- Web UI implementation for the MVP.

- MVP 不保证支持私有仓库。
- Private repository support as a guaranteed MVP capability.

- 不允许通过 CLI flag 或位置参数传入 token。
- Passing tokens through CLI flags or positional arguments.

- 不下载图片或附件到本地。
- Downloading images or attachments to local storage.

- 不导出 Pull Request diff。
- Pull Request diff export.

- 不导出 Pull Request commit history。
- Pull Request commit history export.

- 不实现失败请求重试逻辑。
- Retry logic for failed GitHub API requests.

- 不内置 AI 摘要生成。
- Built-in AI summary generation.

- GitLab Merge Request 支持。
- GitLab Merge Request support.

- GitLab Epic、Snippet 或其他 GitLab 对象支持。
- GitLab Epic, Snippet, or other GitLab object support.

- GitLab token 认证。
- GitLab token authentication.

- 不支持除指定 GitLab Issue 之外的其他非 GitHub 平台能力。
- Non-GitHub provider capabilities other than the specified GitLab Issue support.