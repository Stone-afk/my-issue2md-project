# issue2md 核心功能任务清单 / Core Functionality Tasks

**规格文档 / Spec**: `specs/001-core-functionality/spec.md`  
**技术方案 / Plan**: `specs/001-core-functionality/plan.md`  
**任务清单 / Task**: `specs/001-core-functionality/optimize-tasks.md`  
**日期 / Date**: 2026-04-30  
**状态 / Status**: 面向现状优化的任务清单 / Optimization task list aligned to the current codebase

---

## 1. 说明 / Notes

这份任务清单**替代**旧的“从零实现”清单。当前仓库已经具备基础实现，因此本文件只跟踪与新版 `plan.md` 对齐的**优化与收敛任务**，不再重复记录已完成的脚手架搭建工作。

This task list **replaces** the earlier bootstrap-oriented list. The repository already contains a working baseline, so this file tracks only the **optimization and alignment tasks** required by the updated `plan.md`.

### 1.1 任务格式 / Task Format

- `[ ] T### [P] 任务描述 / Task description`
- `[P]` 表示该任务在依赖满足后可并行执行。
- `[P]` means the task can run in parallel after dependencies are satisfied.
- 任何会改变运行时行为的实现任务，都必须依赖一个更早的失败测试任务。
- Any implementation task that changes runtime behavior must depend on an earlier failing test task.
- 每个任务应只修改一个主要文件；验证任务除外。
- Each task should modify one primary file, except validation tasks.

### 1.2 当前基线 / Current Baseline

仓库中已存在以下主要文件：

- `internal/parser/parser.go`
- `internal/config/config.go`
- `internal/github/client.go`
- `internal/gitlab/client.go`
- `internal/converter/renderer.go`
- `internal/cli/app.go`
- `cmd/issue2md/main.go`
- `cmd/issue2mdweb/main.go`

本清单的目标是让这些文件与新版 `plan.md` 中的优化目标完全对齐。

---

## 2. 依赖总览 / Dependency Overview

- Phase 1 Parser Clarity 阻塞后续依赖 parser 错误语义的任务。
- Phase 1 Parser Clarity blocks later work that depends on parser error semantics.
- Phase 2 Config Wiring 依赖当前 `config`、`github client` 和 `main` 的现有实现。
- Phase 2 Config Wiring depends on the existing `config`, `github client`, and `main` implementation.
- Phase 3 Provider Alignment 依赖 Phase 2 的配置接线收敛。
- Phase 3 Provider Alignment depends on the Phase 2 config-wiring cleanup.
- Phase 4 Renderer Maintainability 依赖现有 converter 测试基线，但不依赖 provider 重构。
- Phase 4 Renderer Maintainability depends on the existing converter baseline, but not on provider refactoring.
- Phase 5 Final Validation 依赖所有前置阶段完成。
- Phase 5 Final Validation depends on all earlier phases.

---

## Phase 1: Parser Clarity Optimization / Parser 明确性优化

### Tests First / 测试先行

- [ ] T001 [P] 更新 `internal/parser/parser_test.go`，用表格驱动测试补齐并锁定以下语义：GitHub Issue、GitHub PR、GitHub Discussion、GitLab Issue、非法 scheme、缺失 path segment、非正整数 ID、unsupported URL，以及 query string 保留原始 source URL 的行为。
  - Update `internal/parser/parser_test.go` with table-driven tests that lock the supported URL cases and the invalid/unsupported semantics.
  - Dependencies: none.
  - Acceptance: 测试在 parser 错误风格调整前先失败；断言错误至少能区分 `invalid` 与 `unsupported` 两类语义。
  - Acceptance: tests fail before the parser error-style cleanup and distinguish at least `invalid` vs `unsupported` semantics.

### Implementation / 实现

- [ ] T002 更新 `internal/parser/parser.go`，统一错误构造为单层明确上下文，移除重复包装，并保持 `Parse`、GitHub 分支、GitLab 分支的职责边界清晰。
  - Update `internal/parser/parser.go` to use single-layer contextual errors, remove duplicate wrapping, and keep helper responsibilities clear.
  - Dependencies: T001.
  - Acceptance: T001 通过；错误信息不再出现“包装一个刚创建的字符串错误再 `%w` 传递”的模式。
  - Acceptance: T001 passes; errors no longer wrap a freshly created string error via `%w`.

- [ ] T003 运行 `go test ./internal/parser`，如有失败，仅修复 `internal/parser/parser.go` 或 `internal/parser/parser_test.go` 中的对应问题。
  - Run `go test ./internal/parser` and fix only parser-owned failures.
  - Dependencies: T002.
  - Acceptance: `internal/parser` 测试通过。
  - Acceptance: parser package tests pass.

---

## Phase 2: Config Wiring Clarification / 配置接线明确化

### Tests First / 测试先行

- [ ] T004 [P] 更新 `internal/github/client_test.go`，增加 token-aware 请求测试：当提供 `GITHUB_TOKEN` 时请求携带认证头；未提供时保持未认证访问。
  - Update `internal/github/client_test.go` with token-aware request tests covering authenticated and unauthenticated GitHub access.
  - Dependencies: none.
  - Acceptance: 在 GitHub client 尚未消费 token 配置前先失败。
  - Acceptance: tests fail before the GitHub client consumes token configuration.

- [ ] T005 [P] 更新 `internal/config/config_test.go`，锁定运行时只读取 `GITHUB_TOKEN`，且不引入 GitLab token 配置。
  - Update `internal/config/config_test.go` to lock `GITHUB_TOKEN` as the only supported runtime credential input.
  - Dependencies: none.
  - Acceptance: 测试覆盖 environment map 与 lookup function 两条读取路径。
  - Acceptance: tests cover both environment-map and lookup-function loading paths.

### Implementation / 实现

- [ ] T006 更新 `internal/github/client.go`，在不改变默认未认证行为的前提下，新增显式 token 配置接入，并在请求阶段写入 GitHub 认证头。
  - Update `internal/github/client.go` to accept explicit token configuration while preserving unauthenticated fallback behavior.
  - Dependencies: T004.
  - Acceptance: T004 通过；token 只通过显式配置注入，不通过全局状态或 CLI flag 读取。
  - Acceptance: T004 passes; token is injected explicitly and not read from globals or CLI flags.

- [ ] T007 更新 `cmd/issue2md/main.go`，将 `config.LoadFromEnv` 读取结果显式注入 GitHub client，并保持 `main` 只承担运行时装配职责。
  - Update `cmd/issue2md/main.go` to load config explicitly and pass it into the GitHub client while keeping `main` as a wiring layer only.
  - Dependencies: T005, T006.
  - Acceptance: `main` 仍然只负责 env、fetcher、renderer 和 `app.Run` 的装配，不接收新的业务逻辑。
  - Acceptance: `main` remains a thin wiring layer with no new business logic.

- [ ] T008 更新 `internal/cli/app.go`，移除无效的空配置加载，确保 CLI 只负责编排、分派和输出。
  - Update `internal/cli/app.go` to remove the no-op config loading and keep CLI focused on orchestration only.
  - Dependencies: T007.
  - Acceptance: 文件中不再存在“读取但不消费”的配置调用；现有 CLI 行为测试仍保持通过。
  - Acceptance: the file no longer contains config loading with unused results, and existing CLI behavior tests still pass.

- [ ] T009 运行 `go test ./internal/config ./internal/github ./internal/cli ./cmd/issue2md`，如有失败，仅修复对应责任文件。
  - Run `go test ./internal/config ./internal/github ./internal/cli ./cmd/issue2md` and fix only responsible-file failures.
  - Dependencies: T008.
  - Acceptance: 相关包测试通过。
  - Acceptance: the affected package tests pass.

---

## Phase 3: Provider Behavior Alignment / Provider 行为对齐

### Tests First / 测试先行

- [ ] T010 [P] 更新 `internal/gitlab/client_test.go`，锁定 GitLab 路径的 API 错误上下文、无效时间格式处理和 notes 时间升序行为，使其与 GitHub 路径的可观察语义保持一致。
  - Update `internal/gitlab/client_test.go` to lock API error context, invalid-time handling, and chronological note ordering in a way that matches the observable GitHub behavior.
  - Dependencies: none.
  - Acceptance: 在 GitLab 错误上下文尚未完全对齐前，新增断言先失败。
  - Acceptance: new assertions fail before the GitLab error-context alignment is implemented.

- [ ] T011 [P] 更新 `internal/github/client_test.go`，补充 GitHub 路径对非 200 响应、无效 `created_at` 和 comments 时间排序的上下文断言。
  - Update `internal/github/client_test.go` with stronger assertions around non-200 responses, invalid `created_at`, and comment ordering semantics.
  - Dependencies: none.
  - Acceptance: 测试锁定 GitHub 与 GitLab 应共享的错误上下文风格。
  - Acceptance: tests lock the error-context style that GitHub and GitLab should share.

### Implementation / 实现

- [ ] T012 更新 `internal/gitlab/client.go`，统一错误上下文风格，保持 notes 映射和时间升序逻辑显式且可读。
  - Update `internal/gitlab/client.go` to align error context style and keep note mapping and chronological ordering explicit.
  - Dependencies: T010.
  - Acceptance: T010 通过；不引入 GitLab token 或额外 provider 抽象。
  - Acceptance: T010 passes; no GitLab token or additional provider abstraction is introduced.

- [ ] T013 更新 `internal/github/client.go`，在已接入 token 的基础上保持与 GitLab 对齐的错误上下文与排序可读性。
  - Update `internal/github/client.go` to keep error context and ordering logic readable and aligned with GitLab after token wiring.
  - Dependencies: T011, T006.
  - Acceptance: T011 通过；代码不因对齐而引入额外层级抽象。
  - Acceptance: T011 passes without introducing extra abstraction layers.

- [ ] T014 运行 `go test ./internal/github ./internal/gitlab`，如有失败，只修复 provider 对应单一责任文件。
  - Run `go test ./internal/github ./internal/gitlab` and fix only provider-owned failures.
  - Dependencies: T012, T013.
  - Acceptance: GitHub 与 GitLab 包测试全部通过。
  - Acceptance: GitHub and GitLab package tests pass.

---

## Phase 4: Renderer Maintainability Optimization / Renderer 可维护性优化

### Tests First / 测试先行

- [ ] T015 [P] 更新 `internal/converter/renderer_test.go`，锁定 Issue、Pull Request、Discussion 三类输出的核心段落结构：title、Summary / 摘要、Structured Notes / 结构化笔记、Raw Archive / 原始归档、comments/review comments/accepted answer 的文本位置。
  - Update `internal/converter/renderer_test.go` to lock the core output structure for Issue, Pull Request, and Discussion rendering.
  - Dependencies: none.
  - Acceptance: 在 renderer 小重构前，测试先覆盖当前输出结构并提供回归保护。
  - Acceptance: tests provide regression protection for the current output structure before renderer cleanup.

- [ ] T016 [P] 更新 `internal/converter/frontmatter_test.go`，锁定 renderer 优化前后 frontmatter 字段和值保持不变。
  - Update `internal/converter/frontmatter_test.go` to lock frontmatter field stability across renderer cleanup.
  - Dependencies: none.
  - Acceptance: 覆盖至少 `title`、`url`、`author`、`created_at` 的稳定输出。
  - Acceptance: covers stable output for at least `title`, `url`, `author`, and `created_at`.

### Implementation / 实现

- [ ] T017 更新 `internal/converter/renderer.go`，仅提取稳定且重复的渲染片段，降低重复度，同时保持输出文本完全不变。
  - Update `internal/converter/renderer.go` by extracting only stable repeated rendering fragments while keeping the final output text unchanged.
  - Dependencies: T015, T016.
  - Acceptance: T015 和 T016 通过；不引入模板系统、DSL 或 provider registry。
  - Acceptance: T015 and T016 pass; no template system, DSL, or provider registry is introduced.

- [ ] T018 运行 `go test ./internal/converter`，如有失败，仅修复 converter 自身文件。
  - Run `go test ./internal/converter` and fix only converter-owned failures.
  - Dependencies: T017.
  - Acceptance: converter 包测试通过。
  - Acceptance: converter package tests pass.

---

## Phase 5: Main Thinness and Final Validation / Main 装配层与最终验证

### Tests First / 测试先行

- [ ] T019 [P] 更新 `cmd/issue2md/main_test.go`，保持对 main package“可构建、低逻辑、仅装配”的回归保护。
  - Update `cmd/issue2md/main_test.go` to preserve regression protection for a buildable, low-logic, wiring-only main package.
  - Dependencies: none.
  - Acceptance: main package 的验证仍聚焦“装配层”而非业务行为。
  - Acceptance: validation remains focused on main as a wiring layer rather than a business-logic layer.

### Implementation / 实现

- [ ] T020 更新 `cmd/issue2md/main.go`，如前置改动导致 main 职责膨胀，则收缩回装配层，只保留 env、client、renderer 和 `app.Run` 的连接。
  - Update `cmd/issue2md/main.go` only if prior work causes main to grow beyond a wiring layer; keep only env, client, renderer, and `app.Run` composition.
  - Dependencies: T019, T007.
  - Acceptance: T019 通过；main 中不重新引入 URL 解析、provider 分派或 Markdown 渲染逻辑。
  - Acceptance: T019 passes; main does not re-introduce URL parsing, provider dispatch, or Markdown rendering logic.

- [ ] T021 运行 `make test`，只修复每个失败对应的单一责任文件。
  - Run `make test` and fix each failure in its single responsible file.
  - Dependencies: T003, T009, T014, T018, T020.
  - Acceptance: 所有测试通过。
  - Acceptance: all tests pass.

- [ ] T022 运行 `go build ./cmd/issue2md`，验证 CLI 二进制可构建。
  - Run `go build ./cmd/issue2md` to verify the CLI binary builds successfully.
  - Dependencies: T021.
  - Acceptance: CLI 构建成功。
  - Acceptance: the CLI build succeeds.

- [ ] T023 运行 `go build ./cmd/issue2mdweb`，验证未来 Web 入口仍可构建且未引入第三方 Web framework。
  - Run `go build ./cmd/issue2mdweb` to verify the future web entrypoint remains buildable without introducing a third-party web framework.
  - Dependencies: T021.
  - Acceptance: Web 入口构建成功。
  - Acceptance: the web entrypoint build succeeds.

---

## Final Validation Checklist / 最终验证清单

- [ ] 所有会改变运行时行为的实现任务都位于对应失败测试之后。
- [ ] All behavior-changing implementation tasks come after their failing tests.
- [ ] `internal/parser` 的错误语义能明确区分 `invalid` 与 `unsupported`。
- [ ] `internal/parser` clearly distinguishes `invalid` from `unsupported` semantics.
- [ ] `internal/cli/app.go` 中不再存在无效配置读取。
- [ ] `internal/cli/app.go` no longer contains no-op config loading.
- [ ] `GITHUB_TOKEN` 通过显式配置接入 GitHub client。
- [ ] `GITHUB_TOKEN` is wired into the GitHub client via explicit configuration.
- [ ] 不引入 GitLab token 支持。
- [ ] No GitLab token support is introduced.
- [ ] GitHub 与 GitLab 的错误上下文风格保持一致。
- [ ] GitHub and GitLab keep a consistent error-context style.
- [ ] renderer 优化不改变最终 Markdown 输出文本。
- [ ] renderer cleanup does not change the final Markdown output text.
- [ ] `cmd/issue2md/main.go` 仍保持装配层职责。
- [ ] `cmd/issue2md/main.go` remains a wiring layer.
- [ ] 使用 Makefile 完成最终测试验证。
- [ ] Final test validation is executed through the Makefile.
