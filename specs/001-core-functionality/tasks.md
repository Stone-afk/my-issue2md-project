# issue2md 核心功能任务清单 / Core Functionality Tasks

**规格文档 / Spec**: `specs/001-core-functionality/spec.md`  
**技术方案 / Plan**: `specs/001-core-functionality/plan.md`  
**任务清单 / Tasks**: `specs/001-core-functionality/tasks.md`  
**日期 / Date**: 2026-04-28

## 任务格式 / Task Format

- `[ ] T### [P] 任务描述 / Task description`
- `[P]` 表示该任务在依赖满足后，可以与同阶段其他任务并行执行。
- `[P]` means the task can run in parallel with other tasks in the same phase after its dependencies are satisfied.
- 每个实现任务都必须依赖一个更早的失败测试任务，以满足 TDD 要求。
- Every implementation task must depend on an earlier failing test task when it changes runtime behavior.
- 每个任务应只创建或修改一个主要文件。
- Each task should touch one primary file only.

## 依赖总览 / Dependency Overview

- Phase 1 Foundation 阻塞所有后续阶段。
- Phase 1 Foundation blocks all later phases.
- Phase 2 GitHub Fetcher 依赖 Phase 1 的 model/parser 契约。
- Phase 2 GitHub Fetcher depends on Phase 1 model/parser contracts.
- Phase 3 Markdown Converter 依赖 Phase 1 的 model 契约。
- Phase 3 Markdown Converter depends on Phase 1 model contracts.
- Phase 4 CLI Assembly 依赖 Phase 1 parser/model、Phase 2 fetchers 和 Phase 3 converter。
- Phase 4 CLI Assembly depends on Phase 1 parser/model, Phase 2 fetchers, and Phase 3 converter.

---

## Phase 1: Foundation / 数据结构定义

### Tests First / 测试先行

- [ ] T001 [P] 创建 `internal/model/document_test.go`，使用表格驱动测试验证 `DocumentData`、`IssueData`、`PullRequestData`、`DiscussionData`、`CommentData` 和 `ReactionSummary` 可被安全构造。
  - Create `internal/model/document_test.go` with table tests validating zero-value-safe construction for the core data structs.
  - Dependencies: none.
  - Acceptance: 测试引用 `plan.md` 要求的所有核心结构体，并在 `internal/model/document.go` 创建前失败。
  - Acceptance: tests reference all core structs required by `plan.md` and fail before `internal/model/document.go` exists.

- [ ] T002 [P] 创建 `internal/parser/parser_test.go`，使用表格驱动测试覆盖 GitHub Issue、GitHub PR、GitHub Discussion、GitLab Issue、不支持的 GitHub URL、GitLab MR 拒绝、非法 URL 和 query string 处理。
  - Create `internal/parser/parser_test.go` with table tests for all supported and unsupported URL cases.
  - Dependencies: none.
  - Acceptance: 断言 provider、kind、owner/repo 或 project path、number 和原始 source URL。
  - Acceptance: tests assert provider, kind, owner/repo or project path, number, and source URL.

- [ ] T003 [P] 创建 `internal/config/config_test.go`，验证只从环境输入读取 `GITHUB_TOKEN`，且不读取也不要求 GitLab token。
  - Create `internal/config/config_test.go` proving `GITHUB_TOKEN` is read and GitLab token is not required.
  - Dependencies: none.
  - Acceptance: 测试在 `internal/config/config.go` 创建前失败。
  - Acceptance: tests fail before `internal/config/config.go` exists.

### Implementation / 实现

- [ ] T004 创建 `internal/model/document.go`，定义 `Provider`、`ContentKind`、`UserData`、`ReactionSummary`、`CommentData`、`IssueData`、`PullRequestData`、`DiscussionData`、`FetchOptions` 和 `DocumentData`。
  - Create `internal/model/document.go` with all core model structs.
  - Dependencies: T001.
  - Acceptance: T001 通过；结构体包含 spec 要求的 reactions 字段。
  - Acceptance: T001 passes; structs include reaction fields required by spec.

- [ ] T005 创建 `internal/parser/parser.go`，实现 `Parse(rawURL string) (Target, error)` 和 `Target`，支持 GitHub 与 GitLab URL 形式。
  - Create `internal/parser/parser.go` implementing URL parsing and target detection.
  - Dependencies: T002, T004.
  - Acceptance: T002 通过；错误信息显式包含 invalid/unsupported URL 上下文。
  - Acceptance: T002 passes; errors include invalid/unsupported URL context.

- [ ] T006 创建 `internal/config/config.go`，实现从 environment map 或 lookup function 显式加载配置，只包含 `GITHUB_TOKEN`。
  - Create `internal/config/config.go` implementing explicit config loading for `GITHUB_TOKEN` only.
  - Dependencies: T003.
  - Acceptance: T003 通过；不引入 GitLab token 配置。
  - Acceptance: T003 passes; no GitLab token config is introduced.

- [ ] T007 [P] 创建 `cmd/issue2mdweb/main.go`，作为未来 Web 入口的最小占位文件，只使用标准库或不引入 import。
  - Create `cmd/issue2mdweb/main.go` as a minimal future Web entrypoint placeholder.
  - Dependencies: none.
  - Acceptance: 文件可构建；不实现 Web UI 行为。
  - Acceptance: file builds; does not implement Web UI behavior.

---

## Phase 2: GitHub Fetcher / API 交互逻辑，TDD

### Tests First / 测试先行

- [ ] T008 [P] 创建 `internal/github/client_test.go`，测试 `NewClient` 默认值、注入 HTTP client/GraphQL URL 行为，以及拒绝非 GitHub target。
  - Create `internal/github/client_test.go` for GitHub client construction and target validation.
  - Dependencies: T004, T005.
  - Acceptance: 测试在 `internal/github/client.go` 创建前失败。
  - Acceptance: tests fail before `internal/github/client.go` exists.

- [ ] T009 [P] 创建 `internal/github/issues_test.go`，用 `httptest.Server` 或注入的 `go-github` client 测试 GitHub Issue 和 comments 映射到 `model.DocumentData`。
  - Create `internal/github/issues_test.go` for GitHub Issue mapping tests.
  - Dependencies: T004, T005.
  - Acceptance: 覆盖 title、URL、author、created time、state、body、comments、comments 时间正序和可选 reactions。
  - Acceptance: tests cover required fields, chronological comments, and optional reactions.

- [ ] T010 [P] 创建 `internal/github/pulls_test.go`，测试 GitHub Pull Request 和 review comments 映射到 `model.DocumentData`。
  - Create `internal/github/pulls_test.go` for Pull Request mapping tests.
  - Dependencies: T004, T005.
  - Acceptance: 断言包含 PR 描述和 review comments，不建模 diff/commit history，review comments 时间正序。
  - Acceptance: tests assert PR description/review comments, no diff/commit history, and ascending ordering.

- [ ] T011 [P] 创建 `internal/github/discussions_test.go`，用 `httptest.Server` 测试 GitHub GraphQL v4 Discussion 请求和响应映射。
  - Create `internal/github/discussions_test.go` for Discussion GraphQL request/response mapping.
  - Dependencies: T004, T005.
  - Acceptance: 覆盖 title、URL、author、created time、state、body、comments、accepted answer 数据和 API 错误透传。
  - Acceptance: tests cover required fields, accepted answer data, and API error propagation.

- [ ] T012 [P] 创建 `internal/github/reactions_test.go`，用表格驱动测试 GitHub reaction payload/counts 到 `model.ReactionSummary` 的映射。
  - Create `internal/github/reactions_test.go` for reaction mapping helpers.
  - Dependencies: T004.
  - Acceptance: 覆盖 total、+1、-1、laugh、hooray、confused、heart、rocket、eyes。
  - Acceptance: tests cover total, +1, -1, laugh, hooray, confused, heart, rocket, and eyes.

- [ ] T013 [P] 创建 `internal/gitlab/client_test.go`，用 `httptest.Server` 测试公开 GitLab Issue metadata 和 notes 映射。
  - Create `internal/gitlab/client_test.go` for public GitLab Issue and notes mapping.
  - Dependencies: T004, T005.
  - Acceptance: 覆盖 title、URL、author、created time、state、body、notes as comments、note ordering 和 API 错误透传。
  - Acceptance: tests cover required fields, note ordering, and API error propagation.

### Implementation / 实现

- [ ] T014 创建 `internal/github/client.go`，实现 `Client`、`Options`、`NewClient` 和 GitHub target 的 `Fetch` 分派。
  - Create `internal/github/client.go` implementing GitHub client construction and dispatch.
  - Dependencies: T008, T004, T005.
  - Acceptance: T008 通过；`Fetch` 能分派 Issue、PR、Discussion，且不渲染 Markdown。
  - Acceptance: T008 passes; `Fetch` routes Issue, PR, and Discussion without rendering Markdown.

- [ ] T015 创建 `internal/github/reactions.go`，实现 GitHub fetchers 使用的 reaction 映射辅助函数。
  - Create `internal/github/reactions.go` implementing reaction mapping helpers.
  - Dependencies: T012, T004.
  - Acceptance: T012 通过。
  - Acceptance: T012 passes.

- [ ] T016 创建 `internal/github/issues.go`，通过 `google/go-github` 实现 GitHub Issue 和 issue comments 获取。
  - Create `internal/github/issues.go` implementing GitHub Issue and comments fetching via `google/go-github`.
  - Dependencies: T009, T014, T015.
  - Acceptance: T009 通过；comments 按 `CreatedAt` 排序；错误带上下文包装。
  - Acceptance: T009 passes; comments sorted by `CreatedAt`; errors wrapped with context.

- [ ] T017 创建 `internal/github/pulls.go`，通过 `google/go-github` 实现 GitHub Pull Request 描述和 review comments 获取。
  - Create `internal/github/pulls.go` implementing PR description and review comments fetching.
  - Dependencies: T010, T014, T015.
  - Acceptance: T010 通过；不引入 diff 或 commit history 获取/渲染数据。
  - Acceptance: T010 passes; no diff or commit history data is introduced.

- [ ] T018 创建 `internal/github/discussions.go`，通过 `net/http` 和 `encoding/json` 实现 GitHub Discussion GraphQL v4 获取。
  - Create `internal/github/discussions.go` implementing GitHub Discussion GraphQL v4 fetching.
  - Dependencies: T011, T014.
  - Acceptance: T011 通过；accepted answer 映射到 `DiscussionData.AcceptedAnswer`；API 错误不重试并透传。
  - Acceptance: T011 passes; accepted answer is mapped and API errors are surfaced without retry.

- [ ] T019 创建 `internal/gitlab/client.go`，实现公开 GitLab Issue target 的 `Client`、`Options`、`NewClient` 和 `FetchIssue`。
  - Create `internal/gitlab/client.go` implementing GitLab client construction and `FetchIssue`.
  - Dependencies: T013, T004, T005.
  - Acceptance: T013 的 client 构造、公开 API 错误处理和 target 校验通过。
  - Acceptance: T013 passes for client construction, API errors, and target validation.

- [ ] T020 创建 `internal/gitlab/issues.go`，实现 GitLab Issue 和 notes HTTP 请求/响应映射。
  - Create `internal/gitlab/issues.go` implementing GitLab Issue and notes HTTP mapping.
  - Dependencies: T013, T019.
  - Acceptance: T013 通过；不引入 GitLab token 支持。
  - Acceptance: T013 passes; no GitLab token support is introduced.

---

## Phase 3: Markdown Converter / 转换逻辑，TDD

### Tests First / 测试先行

- [ ] T021 [P] 创建 `internal/converter/frontmatter_test.go`，测试 YAML frontmatter 包含 `title`、`url`、`author`、`created_at`。
  - Create `internal/converter/frontmatter_test.go` for YAML frontmatter tests.
  - Dependencies: T004.
  - Acceptance: 覆盖 GitHub Issue 和 GitLab Issue 文档。
  - Acceptance: tests cover GitHub Issue and GitLab Issue documents.

- [ ] T022 [P] 创建 `internal/converter/users_test.go`，用表格驱动测试纯文本用户渲染和 `EnableUserLinks=true` 时的 Markdown 链接渲染。
  - Create `internal/converter/users_test.go` for user rendering tests.
  - Dependencies: T004.
  - Acceptance: 覆盖 URL 为空时回退到纯文本。
  - Acceptance: tests cover empty URL fallback to plain text.

- [ ] T023 [P] 创建 `internal/converter/reactions_test.go`，测试 reactions 默认省略，且只在 `EnableReactions=true` 时渲染。
  - Create `internal/converter/reactions_test.go` for reaction rendering tests.
  - Dependencies: T004.
  - Acceptance: 覆盖主内容 reactions 和评论 reactions。
  - Acceptance: tests cover main content reactions and comment reactions.

- [ ] T024 [P] 创建 `internal/converter/renderer_test.go`，测试 GitHub Issue 和 GitLab Issue 通过共享 Issue Markdown 路径渲染。
  - Create `internal/converter/renderer_test.go` for shared Issue rendering tests.
  - Dependencies: T004.
  - Acceptance: 断言 Summary、Structured Notes、Raw Archive、原始正文保留、comments 和图片链接不变。
  - Acceptance: tests assert required sections, original body preservation, comments, and unchanged image links.

- [ ] T025 [P] 扩展 `internal/converter/renderer_test.go`，添加 PR 渲染测试。
  - Extend `internal/converter/renderer_test.go` with PR rendering tests.
  - Dependencies: T024.
  - Acceptance: 断言 PR 描述和 review comments 被渲染，review comments 不分组，diff/commit history 文本不存在。
  - Acceptance: tests assert PR description/review comments, no grouping, and no diff/commit history text.

- [ ] T026 [P] 扩展 `internal/converter/renderer_test.go`，添加 Discussion 渲染测试。
  - Extend `internal/converter/renderer_test.go` with Discussion rendering tests.
  - Dependencies: T024.
  - Acceptance: 断言 discussion body/comments 被渲染，accepted answer 有显著标记。
  - Acceptance: tests assert discussion body/comments render and accepted answer has a prominent marker.

### Implementation / 实现

- [ ] T027 创建 `internal/converter/frontmatter.go`，实现 YAML frontmatter 渲染辅助函数。
  - Create `internal/converter/frontmatter.go` implementing frontmatter helpers.
  - Dependencies: T021.
  - Acceptance: T021 通过；只使用标准库。
  - Acceptance: T021 passes; implementation uses standard library only.

- [ ] T028 创建 `internal/converter/users.go`，实现纯文本和链接形式的用户渲染。
  - Create `internal/converter/users.go` implementing plain and linked user rendering.
  - Dependencies: T022.
  - Acceptance: T022 通过。
  - Acceptance: T022 passes.

- [ ] T029 创建 `internal/converter/reactions.go`，实现 reaction summary 渲染。
  - Create `internal/converter/reactions.go` implementing reaction summary rendering.
  - Dependencies: T023.
  - Acceptance: T023 通过。
  - Acceptance: T023 passes.

- [ ] T030 创建 `internal/converter/renderer.go`，实现 `Renderer`、`Options`、`NewRenderer` 和 `Render` 编排。
  - Create `internal/converter/renderer.go` implementing renderer orchestration.
  - Dependencies: T024, T027, T028, T029.
  - Acceptance: T024 的 GitHub/GitLab Issue 渲染测试通过。
  - Acceptance: T024 passes for GitHub and GitLab Issue rendering.

- [ ] T031 更新 `internal/converter/renderer.go`，添加 Pull Request 渲染。
  - Update `internal/converter/renderer.go` to add Pull Request rendering.
  - Dependencies: T025, T030.
  - Acceptance: T025 通过；不输出 diff 或 commit history。
  - Acceptance: T025 passes; no diff or commit history output.

- [ ] T032 更新 `internal/converter/renderer.go`，添加 Discussion 渲染和 accepted answer 标记。
  - Update `internal/converter/renderer.go` to add Discussion rendering and accepted answer marker.
  - Dependencies: T026, T030.
  - Acceptance: T026 通过。
  - Acceptance: T026 passes.

---

## Phase 4: CLI Assembly / 命令行入口集成

### Tests First / 测试先行

- [ ] T033 [P] 创建 `internal/cli/app_test.go`，测试缺少 URL、非法 URL、不支持 URL，以及 stderr/非零退出行为。
  - Create `internal/cli/app_test.go` for CLI error behavior tests.
  - Dependencies: T005, T006.
  - Acceptance: 测试在 `internal/cli/app.go` 创建前失败。
  - Acceptance: tests fail before `internal/cli/app.go` exists.

- [ ] T034 [P] 扩展 `internal/cli/app_test.go`，使用 fake fetchers 和 renderer 测试 stdout 输出和 output file 行为。
  - Extend `internal/cli/app_test.go` with stdout and output-file behavior tests.
  - Dependencies: T033, T030.
  - Acceptance: 覆盖默认 stdout 和在 `t.TempDir()` 中创建 `[output_file]`。
  - Acceptance: tests cover default stdout and `[output_file]` creation in `t.TempDir()`.

- [ ] T035 [P] 扩展 `internal/cli/app_test.go`，测试 `-enable-reactions` 和 `-enable-user-links` flag 传递。
  - Extend `internal/cli/app_test.go` with flag propagation tests.
  - Dependencies: T033, T030.
  - Acceptance: 证明 fetch options 收到 `IncludeReactions`，renderer options 收到 `EnableUserLinks`/`EnableReactions`。
  - Acceptance: tests prove fetch and renderer options receive the expected flag values.

- [ ] T036 [P] 扩展 `internal/cli/app_test.go`，测试 GitHub Issue、GitHub PR、GitHub Discussion 和 GitLab Issue 的 provider 分派。
  - Extend `internal/cli/app_test.go` with provider dispatch tests.
  - Dependencies: T033, T014, T019.
  - Acceptance: 证明 GitHub targets 调用 GitHub fetcher，GitLab Issue targets 调用 GitLab fetcher。
  - Acceptance: tests prove GitHub targets call GitHub fetcher and GitLab Issue targets call GitLab fetcher.

- [ ] T037 [P] 创建 `cmd/issue2md/main_test.go`，通过测试或构建检查证明 `main` 只委托给 `internal/cli`，不包含业务逻辑。
  - Create `cmd/issue2md/main_test.go` proving `main` delegates to `internal/cli`.
  - Dependencies: T033.
  - Acceptance: 测试/构建在 `cmd/issue2md/main.go` 创建前失败。
  - Acceptance: test/build fails before `cmd/issue2md/main.go` exists.

### Implementation / 实现

- [ ] T038 创建 `internal/cli/app.go`，实现 `App.Run`、flag parsing、URL parsing、config loading、provider dispatch、rendering、stdout/file output 和 stderr error handling。
  - Create `internal/cli/app.go` implementing CLI orchestration.
  - Dependencies: T033, T034, T035, T036, T005, T006, T014, T019, T030.
  - Acceptance: T033-T036 通过；错误显式处理，向上传递时包装上下文。
  - Acceptance: T033-T036 pass; errors are explicit and wrapped where propagated.

- [ ] T039 创建 `cmd/issue2md/main.go`，连接真实 config、GitHub client、GitLab client、converter renderer、stdout、stderr、args 和 process exit code。
  - Create `cmd/issue2md/main.go` wiring real runtime dependencies.
  - Dependencies: T037, T038, T014, T019, T030.
  - Acceptance: T037 通过；`go build ./cmd/issue2md` 成功。
  - Acceptance: T037 passes; `go build ./cmd/issue2md` succeeds.

- [ ] T040 [P] 创建 `Makefile`，包含 `test: go test ./...`；只有在不引入 Web UI 行为的前提下，添加保守的 `web:` target。
  - Create `Makefile` with `test: go test ./...` and a conservative `web:` target if appropriate.
  - Dependencies: T039, T007.
  - Acceptance: `make test` 运行 Go 测试；`make web` 不引入不支持的 Web framework。
  - Acceptance: `make test` runs the Go test suite; `make web` does not introduce unsupported Web framework dependencies.

- [ ] T041 运行 `go test ./...` 或 `make test`，只修复每个失败对应的单一责任文件。
  - Run `go test ./...` or `make test` and fix failures in the responsible single file.
  - Dependencies: T004-T040.
  - Acceptance: 所有测试通过。
  - Acceptance: all tests pass.

- [ ] T042 运行 `go build ./cmd/issue2md`，只修复每个构建失败对应的单一责任文件。
  - Run `go build ./cmd/issue2md` and fix build failures in the responsible single file.
  - Dependencies: T041.
  - Acceptance: CLI binary 构建成功。
  - Acceptance: CLI binary builds successfully.

## Final Validation Checklist / 最终验证清单

- [ ] 所有测试任务都早于对应实现任务。
- [ ] All test tasks precede related implementation tasks.
- [ ] 每个任务只创建或修改一个主要文件。
- [ ] Each task modifies or creates one primary file.
- [ ] 可并行任务已标记 `[P]`。
- [ ] Parallelizable tasks are marked `[P]`.
- [ ] GitLab 范围仍限定为公开 Issue URL。
- [ ] GitLab scope remains limited to public Issue URLs.
- [ ] 不引入 GitLab token 支持。
- [ ] No GitLab token support is introduced.
- [ ] 不引入 Web framework。
- [ ] No Web framework is introduced.
- [ ] Markdown 渲染使用标准库字符串构建。
- [ ] Markdown rendering uses standard library string construction.
- [ ] 错误被显式处理，向上传递时包装上下文。
- [ ] Errors are explicit and wrapped when propagated.
- [ ] 不通过全局变量传递运行时状态。
- [ ] No runtime state is passed through globals.