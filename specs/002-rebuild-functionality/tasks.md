# issue2md 重构任务清单 / Rebuild Functionality Tasks

**规格文档 / Spec**: `specs/002-rebuild-functionality/spec.md`  
**技术方案 / Plan**: `specs/002-rebuild-functionality/plan.md`  
**任务清单 / Tasks**: `specs/002-rebuild-functionality/tasks.md`  
**日期 / Date**: 2026-04-30

## 任务格式 / Task Format

- `[ ] T### [P] 任务描述 / Task description`
- `[P]` 表示该任务在依赖满足后，可以与同阶段其他任务并行执行。
- `[P]` means the task can run in parallel with other tasks in the same phase after dependencies are satisfied.
- 每个任务只应修改一个主要文件；验证任务可运行命令但不得顺手做跨文件重构。
- Each task should touch only one primary file; validation tasks may run commands but must not turn into broad cross-file refactors.
- 所有会改变运行时行为的实现任务，都必须排在更早的失败测试任务之后，以满足 TDD。
- Every runtime-behavior change must come after an earlier failing test task to satisfy TDD.

## 依赖总览 / Dependency Overview

- Phase 1 Foundation 阻塞所有后续阶段。
- Phase 1 Foundation blocks all later phases.
- Phase 2 GitHub Fetcher 依赖 Phase 1 的抽象与错误模型。
- Phase 2 GitHub Fetcher depends on the abstraction and error model from Phase 1.
- Phase 3 Markdown Converter 主要用于锁定输出兼容性，可在 Phase 1 完成后推进，但最终验证依赖 Phase 2 和 Phase 4。
- Phase 3 Markdown Converter mainly locks output compatibility and may proceed after Phase 1, but final validation depends on Phase 2 and Phase 4.
- Phase 4 CLI Assembly 依赖新 provider 包与共享错误模型完成。
- Phase 4 CLI Assembly depends on the new provider packages and shared error model.

---

## Phase 1: Foundation / 数据结构定义

### Tests First / 测试先行

- [ ] T001 [P] 创建 `internal/fetchprovider/types_test.go`，使用表格驱动测试定义共享抓取抽象的最小契约：`Provider` 只暴露 `Fetch(...)`，且调用方可区分“provider 未注册”和“provider 不支持该 kind”两类错误。
  - Create `internal/fetchprovider/types_test.go` with table-driven tests defining the minimal shared fetch contract: `Provider` exposes only `Fetch(...)`, and callers can distinguish “provider not registered” from “provider does not support this kind”.
  - Dependencies: none.
  - Acceptance: 测试在 `internal/fetchprovider/types.go` 创建前失败，并明确锁定错误分类行为。
  - Acceptance: tests fail before `internal/fetchprovider/types.go` exists and lock explicit error classification behavior.

### Implementation / 实现

- [ ] T002 创建 `internal/fetchprovider/types.go`，定义 `Provider` 接口、`ErrProviderNotRegistered`、共享的 unsupported-capability 错误类型，以及必要的错误判断辅助函数。
  - Create `internal/fetchprovider/types.go` defining the `Provider` interface, `ErrProviderNotRegistered`, a shared unsupported-capability error type, and any necessary error inspection helpers.
  - Dependencies: T001.
  - Acceptance: T001 通过；`Provider` 接口不暴露 `GetIssue(...)` 等 provider-specific 方法。
  - Acceptance: T001 passes, and the `Provider` interface does not expose provider-specific methods such as `GetIssue(...)`.

---

## Phase 2: GitHub Fetcher / API 交互逻辑，TDD

### Tests First / 测试先行

- [ ] T003 [P] 创建 `internal/fetchprovider/github/client_test.go`，迁移并补强 GitHub provider 的表格驱动测试：覆盖 `Fetch(...)` 对 Issue、Pull Request、Discussion 的分派，`GITHUB_TOKEN` 注入后的认证头行为，以及 provider-specific 错误上下文。
  - Create `internal/fetchprovider/github/client_test.go` and migrate/strengthen table-driven tests for the GitHub provider: cover `Fetch(...)` dispatch for Issue, Pull Request, and Discussion, authenticated requests when `GITHUB_TOKEN` is injected, and provider-specific error context.
  - Dependencies: T002.
  - Acceptance: 测试在新 `internal/fetchprovider/github/client.go` 实现前失败。
  - Acceptance: tests fail before the new `internal/fetchprovider/github/client.go` implementation exists.

- [ ] T004 [P] 创建 `internal/fetchprovider/gitlab/client_test.go`，迁移并补强 GitLab provider 的表格驱动测试：覆盖公开 Issue 抓取、notes 时间升序、显式 unsupported-capability 错误，以及 provider-specific 错误上下文。
  - Create `internal/fetchprovider/gitlab/client_test.go` and migrate/strengthen table-driven tests for the GitLab provider: cover public Issue fetching, chronological notes ordering, explicit unsupported-capability errors, and provider-specific error context.
  - Dependencies: T002.
  - Acceptance: 测试在新 `internal/fetchprovider/gitlab/client.go` 实现前失败。
  - Acceptance: tests fail before the new `internal/fetchprovider/gitlab/client.go` implementation exists.

### Implementation / 实现

- [ ] T005 创建 `internal/fetchprovider/github/client.go`，定义 GitHub provider 的 `Client`、`Options`、`NewClient`、共享 HTTP 辅助逻辑和 `Fetch(...)` 分派骨架，并将“不支持的 kind”映射到共享 unsupported-capability 错误。
  - Create `internal/fetchprovider/github/client.go` defining the GitHub provider `Client`, `Options`, `NewClient`, shared HTTP helper logic, and the `Fetch(...)` dispatch skeleton, mapping unsupported kinds to the shared unsupported-capability error.
  - Dependencies: T003.
  - Acceptance: T003 中关于构造、分派和 unsupported-kind 的断言通过。
  - Acceptance: the constructor, dispatch, and unsupported-kind assertions from T003 pass.

- [ ] T006 更新 `internal/fetchprovider/github/client.go`，补齐 GitHub Issue 和 Pull Request 抓取迁移，保持评论时间升序和现有字段映射不变。
  - Update `internal/fetchprovider/github/client.go` to complete GitHub Issue and Pull Request fetching migration while preserving chronological comment ordering and existing field mappings.
  - Dependencies: T005.
  - Acceptance: T003 中关于 Issue、Pull Request、评论排序和错误上下文的断言通过。
  - Acceptance: T003 assertions for Issue, Pull Request, comment ordering, and error context pass.

- [ ] T007 更新 `internal/fetchprovider/github/client.go`，补齐 GitHub Discussion 抓取迁移，保留 accepted answer 映射和现有 GraphQL 错误传播语义。
  - Update `internal/fetchprovider/github/client.go` to complete GitHub Discussion fetching migration while preserving accepted-answer mapping and existing GraphQL error propagation semantics.
  - Dependencies: T006.
  - Acceptance: T003 中关于 Discussion、accepted answer 和 GraphQL 错误的断言通过。
  - Acceptance: T003 assertions for Discussion, accepted answer, and GraphQL errors pass.

- [ ] T008 创建 `internal/fetchprovider/gitlab/client.go`，定义 GitLab provider 的 `Client`、`Options`、`NewClient`、`Fetch(...)` 和 Issue/notes 抓取逻辑，并将不支持的 kind 映射到共享 unsupported-capability 错误。
  - Create `internal/fetchprovider/gitlab/client.go` defining the GitLab provider `Client`, `Options`, `NewClient`, `Fetch(...)`, and Issue/notes fetching logic, mapping unsupported kinds to the shared unsupported-capability error.
  - Dependencies: T004.
  - Acceptance: T004 全部通过；不引入 GitLab token 或统一跨-provider `Options`。
  - Acceptance: T004 passes completely; no GitLab token or shared cross-provider `Options` is introduced.

- [ ] T009 [P] 删除 `internal/github/client_test.go`，在新 `internal/fetchprovider/github/client_test.go` 成为唯一权威测试后清理旧路径测试文件。
  - Delete `internal/github/client_test.go` after `internal/fetchprovider/github/client_test.go` becomes the single source of truth for GitHub provider tests.
  - Dependencies: T007.
  - Acceptance: 仓库中不再保留旧路径 GitHub provider 测试文件。
  - Acceptance: the repository no longer keeps the legacy-path GitHub provider test file.

- [ ] T010 [P] 删除 `internal/gitlab/client_test.go`，在新 `internal/fetchprovider/gitlab/client_test.go` 成为唯一权威测试后清理旧路径测试文件。
  - Delete `internal/gitlab/client_test.go` after `internal/fetchprovider/gitlab/client_test.go` becomes the single source of truth for GitLab provider tests.
  - Dependencies: T008.
  - Acceptance: 仓库中不再保留旧路径 GitLab provider 测试文件。
  - Acceptance: the repository no longer keeps the legacy-path GitLab provider test file.

- [ ] T011 [P] 删除 `internal/github/client.go`，在 `internal/fetchprovider/github/client.go` 接管实现后清理旧路径实现文件。
  - Delete `internal/github/client.go` after `internal/fetchprovider/github/client.go` takes over the implementation.
  - Dependencies: T007.
  - Acceptance: GitHub provider 的运行时实现仅位于 `internal/fetchprovider/github`。
  - Acceptance: the GitHub provider runtime implementation exists only under `internal/fetchprovider/github`.

- [ ] T012 [P] 删除 `internal/gitlab/client.go`，在 `internal/fetchprovider/gitlab/client.go` 接管实现后清理旧路径实现文件。
  - Delete `internal/gitlab/client.go` after `internal/fetchprovider/gitlab/client.go` takes over the implementation.
  - Dependencies: T008.
  - Acceptance: GitLab provider 的运行时实现仅位于 `internal/fetchprovider/gitlab`。
  - Acceptance: the GitLab provider runtime implementation exists only under `internal/fetchprovider/gitlab`.

---

## Phase 3: Markdown Converter / 转换逻辑，TDD

### Tests First / 测试先行

- [ ] T013 [P] 更新 `internal/converter/frontmatter_test.go`，锁定 provider 抽象迁移前后 frontmatter 的 `title`、`url`、`author`、`created_at` 输出完全一致。
  - Update `internal/converter/frontmatter_test.go` to lock `title`, `url`, `author`, and `created_at` frontmatter output so it remains identical before and after the provider abstraction migration.
  - Dependencies: T002.
  - Acceptance: 测试先提供回归保护，不允许因 provider 重构改变 frontmatter 输出。
  - Acceptance: the tests provide regression protection and do not allow provider refactoring to change frontmatter output.

- [ ] T014 [P] 更新 `internal/converter/renderer_test.go`，锁定 Issue、Pull Request、Discussion 的标题、Summary、Structured Notes、Raw Archive、comments/review comments/accepted answer 段落顺序与文本位置。
  - Update `internal/converter/renderer_test.go` to lock the section order and text placement for Issue, Pull Request, and Discussion output: title, Summary, Structured Notes, Raw Archive, comments/review comments/accepted answer.
  - Dependencies: T002.
  - Acceptance: 测试覆盖 GitHub 和 GitLab Issue 共享渲染路径，且迁移期间若输出变化会先失败。
  - Acceptance: the tests cover the shared GitHub/GitLab Issue rendering path and fail first if output changes during migration.

### Implementation / 实现

- [ ] T015 更新 `internal/converter/renderer.go`，仅在 Phase 2/4 引发编译或输出回归时做最小修正，保证 Markdown 输出文本保持不变。
  - Update `internal/converter/renderer.go` only if Phase 2 or Phase 4 causes compile or output regressions, keeping Markdown output text unchanged.
  - Dependencies: T013, T014.
  - Acceptance: T013 和 T014 通过；不引入新的模板层、DSL 或 provider-specific 分支。
  - Acceptance: T013 and T014 pass; no new templating layer, DSL, or provider-specific branching is introduced.

- [ ] T016 运行 `go test ./internal/converter`，若失败，仅修复 `internal/converter` 自身文件。
  - Run `go test ./internal/converter` and, if it fails, fix only `internal/converter`-owned files.
  - Dependencies: T015.
  - Acceptance: converter 包测试通过。
  - Acceptance: the converter package tests pass.

---

## Phase 4: CLI Assembly / 命令行入口集成

### Tests First / 测试先行

- [ ] T017 [P] 更新 `internal/cli/app_test.go`，将测试替身改为 `map[model.Provider]fetchprovider.Provider`，覆盖 provider registry 分发、provider 未注册错误、unsupported-capability 错误包装、stdout 输出和 output file 行为。
  - Update `internal/cli/app_test.go` to replace concrete test doubles with `map[model.Provider]fetchprovider.Provider`, covering provider-registry dispatch, provider-not-registered errors, unsupported-capability error wrapping, stdout output, and output-file behavior.
  - Dependencies: T002.
  - Acceptance: 测试在 `internal/cli/app.go` 仍依赖具体 GitHub/GitLab 字段时先失败。
  - Acceptance: tests fail while `internal/cli/app.go` still depends on concrete GitHub/GitLab fields.

- [ ] T018 [P] 更新 `cmd/issue2md/main_test.go`，锁定 `main` 只负责一次性构造 provider registry 并注入 `App`，不重新引入 provider-specific 业务逻辑。
  - Update `cmd/issue2md/main_test.go` to lock `main` into constructing the provider registry once and injecting it into `App`, without reintroducing provider-specific business logic.
  - Dependencies: T002.
  - Acceptance: 测试在 `cmd/issue2md/main.go` 仍组装具体 fetcher 字段时先失败。
  - Acceptance: tests fail while `cmd/issue2md/main.go` still assembles concrete fetcher fields directly.

### Implementation / 实现

- [ ] T019 更新 `internal/cli/app.go`，将 `App` 收敛为只依赖 `map[model.Provider]fetchprovider.Provider`，按 `target.Provider` 查表调用 `Fetch(...)`，并以 `fmt.Errorf("fetch document via %s: %w", ...)` 统一包装 provider 错误。
  - Update `internal/cli/app.go` so `App` depends only on `map[model.Provider]fetchprovider.Provider`, looks providers up by `target.Provider`, calls `Fetch(...)`, and wraps provider errors uniformly with `fmt.Errorf("fetch document via %s: %w", ...)`.
  - Dependencies: T017, T007, T008.
  - Acceptance: T017 通过；`App` 中不再保留 GitHub/GitLab provider-specific 分支。
  - Acceptance: T017 passes, and `App` no longer contains GitHub/GitLab-specific branching.

- [ ] T020 更新 `cmd/issue2md/main.go`，从 `internal/fetchprovider/github` 与 `internal/fetchprovider/gitlab` 一次性构造 provider 实例，组装 `map[model.Provider]fetchprovider.Provider`，并注入 `App`。
  - Update `cmd/issue2md/main.go` to construct provider instances once from `internal/fetchprovider/github` and `internal/fetchprovider/gitlab`, assemble `map[model.Provider]fetchprovider.Provider`, and inject it into `App`.
  - Dependencies: T018, T019.
  - Acceptance: T018 通过；`main` 仍保持装配层，不承担 URL 解析、provider 能力判断或 Markdown 渲染职责。
  - Acceptance: T018 passes, and `main` remains a wiring layer without taking over URL parsing, provider capability checks, or Markdown rendering.

- [ ] T021 运行 `make test`，若失败，仅在每次定位到的单一责任文件内修复问题。
  - Run `make test` and, if it fails, fix each issue only in its single responsible file.
  - Dependencies: T009, T010, T011, T012, T016, T019, T020.
  - Acceptance: 全量测试通过。
  - Acceptance: the full test suite passes.

- [ ] T022 [P] 运行 `go build ./cmd/issue2md`，验证 CLI 入口在 provider 重构后仍可构建。
  - Run `go build ./cmd/issue2md` to verify the CLI entrypoint still builds after the provider refactor.
  - Dependencies: T021.
  - Acceptance: CLI 构建成功。
  - Acceptance: the CLI build succeeds.

- [ ] T023 [P] 运行 `go build ./cmd/issue2mdweb`，验证未来 Web 入口未被 provider 重构破坏。
  - Run `go build ./cmd/issue2mdweb` to verify the future web entrypoint is not broken by the provider refactor.
  - Dependencies: T021.
  - Acceptance: Web 入口构建成功。
  - Acceptance: the web entrypoint build succeeds.

---

## Final Validation Checklist / 最终验证清单

- [ ] 所有行为变更都遵循“测试先行，再最小实现”。
- [ ] All behavior changes follow “tests first, then minimal implementation.”
- [ ] `internal/fetchprovider/types.go` 是唯一共享 provider 抽象定义位置。
- [ ] `internal/fetchprovider/types.go` is the single shared location for provider abstraction definitions.
- [ ] `App` 只依赖 `map[model.Provider]fetchprovider.Provider`。
- [ ] `App` depends only on `map[model.Provider]fetchprovider.Provider`.
- [ ] `main` 只负责 provider registry 装配与注入。
- [ ] `main` only performs provider-registry assembly and injection.
- [ ] provider 未注册与 provider 不支持该 kind 可被调用方明确区分。
- [ ] Callers can clearly distinguish provider-not-registered from provider-does-not-support-kind.
- [ ] GitHub 与 GitLab 的运行时实现只存在于 `internal/fetchprovider/github` 与 `internal/fetchprovider/gitlab`。
- [ ] GitHub and GitLab runtime implementations exist only under `internal/fetchprovider/github` and `internal/fetchprovider/gitlab`.
- [ ] Markdown 输出与重构前保持一致。
- [ ] Markdown output remains identical to the pre-refactor behavior.
- [ ] 最终验证通过 `make test`、`go build ./cmd/issue2md`、`go build ./cmd/issue2mdweb` 完成。
- [ ] Final validation completes with `make test`, `go build ./cmd/issue2md`, and `go build ./cmd/issue2mdweb`.
