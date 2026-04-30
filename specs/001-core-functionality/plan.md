# issue2md 核心功能技术实现方案 / Core Functionality Technical Implementation Plan

**规格文档 / Spec**: `specs/001-core-functionality/spec.md`  
**任务清单 / Tasks**: `specs/001-core-functionality/optimize-tasks.md`  
**实现方案 / Plan**: `specs/001-core-functionality/plan.md`  
**日期 / Date**: 2026-04-30  
**状态 / Status**: 架构优化与后续实现基线 / Baseline for architecture optimization and follow-up implementation

---

## 1. 文档目的 / Document Purpose

本文档用于统一 `issue2md` 核心功能的技术实现方向，明确：

- 现状评估 / current implementation baseline
- 宪法约束 / constitution constraints
- 包职责边界 / package responsibility boundaries
- 优化目标 / optimization goals
- 分阶段落地顺序 / phased execution order
- 验证与验收标准 / validation and acceptance criteria

This document aligns the technical direction for `issue2md` core functionality, including the current code baseline, architectural constraints, optimization targets, phased execution order, and validation criteria.

---

## 2. 输入依据 / Inputs and Governing Sources

本方案以以下内容为最高优先级输入：

1. `constitution.md`
2. `CLAUDE.md`
3. `specs/001-core-functionality/spec.md`
4. 当前仓库代码实现
5. 最近一次对 `internal/` 的架构审查结论

### 2.1 宪法强约束 / Constitution Hard Constraints

#### 第一条：简单性原则 / Simplicity First
- 只实现 `spec.md` 明确要求的功能。
- 优先标准库。
- 不为未来假设需求设计复杂抽象。

#### 第二条：测试先行铁律 / Test-First Imperative
- 所有新增行为或修复必须从失败测试开始。
- 单元测试优先表格驱动。
- 依赖优先真实依赖，避免 mock 框架。

#### 第三条：明确性原则 / Clarity and Explicitness
- 所有错误必须显式处理。
- 向上传递错误时必须使用 `fmt.Errorf("...: %w", err)`。
- 不使用全局变量传递运行时状态。

### 2.2 CLAUDE.md 项目协作约束 / CLAUDE.md Project Constraints

- Go 版本基线：`>= 1.24`
- 构建与测试优先通过 Makefile：
  - `make test`
  - `make web`
- 生成 spec / plan / tasks 等文档时，必须优先使用 **Markdown** 且使用 **中英双语**。
- 当涉及新功能时，应优先阅读 `internal/` 相关包并结合宪法思考。

---

## 3. 当前实现快照 / Current Implementation Snapshot

截至 2026-04-30，仓库已具备以下基础实现：

### 3.1 已存在的内部包 / Existing Internal Packages

- `internal/model`
- `internal/parser`
- `internal/config`
- `internal/github`
- `internal/gitlab`
- `internal/converter`
- `internal/cli`

### 3.2 已存在的入口 / Existing Entrypoints

- `cmd/issue2md/main.go`
- `cmd/issue2mdweb/main.go`

### 3.3 已具备的基础能力 / Existing Baseline Capabilities

- GitHub Issue 抓取基础路径已存在。
- GitLab Issue 抓取基础路径已存在。
- Markdown renderer 已实现基础 frontmatter、用户渲染、reaction 渲染和正文拼装。
- CLI 已实现 URL 解析、provider 分派、stdout / output file 输出。
- 基础测试与构建链路已存在。

### 3.4 最近审查识别出的优化点 / Optimization Points Identified in the Latest Review

1. `internal/parser/parser.go` 中存在多层重复错误包装，影响可读性和明确性。
2. `internal/cli/app.go` 中配置加载后未消费，存在“空操作”职责漂移。
3. `internal/github` 与 `internal/gitlab` 存在部分重复 HTTP/JSON 流程，需要控制重复增长。
4. `internal/converter/renderer.go` 中多类内容渲染结构高度相似，需要在不引入过度抽象的前提下保持可维护性。

---

## 4. 实现目标 / Implementation Goals

### 4.1 主目标 / Primary Goals

本轮技术方案的目标不是推翻现有实现，而是在**保持当前简单架构**的前提下，完成以下四类工作：

1. **收紧边界**：明确 CLI、parser、fetcher、converter 的职责。
2. **消除歧义**：统一错误风格、移除无效配置读取、减少误导性代码。
3. **稳住扩展面**：确保后续实现 PR、Discussion、认证接入时，不需要大规模重构。
4. **继续合宪**：所有新增调整遵循 TDD、标准库优先和显式依赖注入。

The goal of this plan is not to replace the current implementation, but to tighten package boundaries, unify error behavior, reduce ambiguity, and keep future feature work compatible with the current simple architecture.

### 4.2 非目标 / Non-Goals

本方案明确不做以下事情：

- 不引入插件化 provider 系统。
- 不引入 ORM、缓存层、任务队列或数据库。
- 不引入第三方 Web framework。
- 不为了复用而提前做复杂抽象。
- 不在本轮方案中设计 AI 总结、全文检索或多仓库批量抓取。

---

## 5. 合宪性设计结论 / Constitution Compliance Decisions

### 5.1 简单性原则 / Simplicity First

**结论 / Decision**: 保持现有按包直分的结构，不引入额外层级。

- `parser` 只负责 URL 识别。
- `github` / `gitlab` 只负责远端获取与数据映射。
- `converter` 只负责 Markdown 输出。
- `cli` 只负责编排、参数和 I/O。

This preserves a direct and readable architecture with minimal abstraction.

### 5.2 测试先行 / Test-First

**结论 / Decision**: 后续每一个优化项都必须遵循 Red-Green-Refactor。

- Parser 错误风格调整先写失败测试。
- CLI 配置接线或删除空调用先写失败测试。
- 如提取公共 HTTP helper，也必须由已有重复行为测试先覆盖。

### 5.3 明确性原则 / Clarity and Explicitness

**结论 / Decision**: 错误信息必须做到“单层明确、上下文清楚、无多余嵌套”。

- 允许：`fmt.Errorf("parse url %q: %w", rawURL, ErrUnsupportedURL)`
- 允许：`fmt.Errorf("parse url %q: unsupported url", rawURL)`
- 不推荐：`fmt.Errorf("parse url %q: %w", rawURL, fmt.Errorf("unsupported url %q", rawURL))`

---

## 6. 目标架构 / Target Architecture

```text
cmd/
  issue2md/
    main.go                  # CLI 进程入口 / CLI process entrypoint
  issue2mdweb/
    main.go                  # Future web placeholder

internal/
  cli/
    app.go                   # 参数解析、配置装配、provider 分派、输出控制
    app_test.go

  config/
    config.go                # GITHUB_TOKEN 等运行时配置读取
    config_test.go

  parser/
    parser.go                # URL 识别与 Target 归一化
    parser_test.go

  model/
    document.go              # provider-neutral 文档模型
    document_test.go

  github/
    client.go                # GitHub 获取与映射
    client_test.go

  gitlab/
    client.go                # GitLab 获取与映射
    client_test.go

  converter/
    frontmatter.go           # frontmatter 渲染
    users.go                 # 用户名/链接渲染
    reactions.go             # reaction 渲染
    renderer.go              # 文档总渲染编排
    *_test.go
```

### 6.1 架构原则 / Architectural Principles

1. **一层编排，三层功能 / one orchestration layer, three functional layers**
   - CLI orchestration
   - parsing / fetching / rendering

2. **内部模型中立 / provider-neutral internal model**
   - `converter` 不依赖 GitHub / GitLab 原始响应结构。
   - `cli` 不依赖远端 API 细节。

3. **依赖显式注入 / explicit dependency injection**
   - HTTP client
   - base URL
   - env/config
   - stdout/stderr
   - renderer factory

4. **最小抽象 / minimum abstraction**
   - 只有在重复逻辑已经被测试稳定覆盖后，才允许做小范围提取。

---

## 7. 包职责细化 / Package Responsibility Details

### 7.1 `internal/parser`

**职责 / Responsibility**
- 解析原始 URL。
- 识别 provider 与 content kind。
- 输出规范化 `Target`。
- 不发网络请求。
- 不承担用户输出格式职责。

**优化要求 / Optimization Requirements**
- 简化错误表达。
- 保持路径拆分逻辑短小。
- 将“URL 语法无效”和“URL 类型不支持”区分清楚。

### 7.2 `internal/config`

**职责 / Responsibility**
- 从环境或 lookup function 中读取运行时配置。
- 当前 MVP 只承载 `GITHUB_TOKEN`。
- 不做副作用操作。

**优化要求 / Optimization Requirements**
- 任何加载出的配置都必须被消费；禁止“读了但不用”。
- 如果当前阶段不接认证，就不要让 CLI 做空加载。

### 7.3 `internal/github`

**职责 / Responsibility**
- 获取 GitHub 内容。
- 将远端 JSON/HTTP 数据映射为 `model.DocumentData`。
- 不渲染 Markdown。
- 不处理 CLI 参数。

**优化要求 / Optimization Requirements**
- 保持请求、响应映射、排序逻辑可测试。
- 如果引入 token，必须通过显式配置注入。
- 继续保持 comments 时间升序。

### 7.4 `internal/gitlab`

**职责 / Responsibility**
- 获取公开 GitLab Issue 与 notes。
- 映射到统一文档模型。
- 不引入 GitLab token 支持，除非 spec 变更。

**优化要求 / Optimization Requirements**
- 与 GitHub 一样保持升序 comment/note 语义。
- 与 GitHub client 保持相似的错误上下文风格。

### 7.5 `internal/converter`

**职责 / Responsibility**
- 只关心 `model.DocumentData -> Markdown`。
- 保持 GitHub Flavored Markdown 输出。
- 控制 frontmatter、用户链接、reactions 以及三类文档正文结构。

**优化要求 / Optimization Requirements**
- 在不引入模板系统的前提下控制重复。
- 提取的 helper 只能是稳定且重复的片段。
- 不把 provider 特殊逻辑漏进 renderer。

### 7.6 `internal/cli`

**职责 / Responsibility**
- 解析 flags 和位置参数。
- 调用 parser。
- 分派 fetcher。
- 构造 renderer。
- 决定 stdout / file / stderr / exit code。

**优化要求 / Optimization Requirements**
- 不做空配置读取。
- 不包含 provider-specific 业务分支细节。
- 错误对用户可见，但内部上下文仍需保留。

---

## 8. 详细优化方案 / Detailed Optimization Plan

## 8.1 优化项 A：统一 Parser 错误模型 / Unify Parser Error Model

### 背景 / Background
`internal/parser/parser.go` 当前能够工作，但错误构造中存在重复包装与语义重复，例如把一个即时创建的字符串错误再 `%w` 包装一次。

### 目标 / Goal
- 保持错误信息明确。
- 减少不必要嵌套。
- 让测试断言更稳定。

### 实现策略 / Implementation Strategy
1. 先补充 parser 表格驱动测试，覆盖：
   - invalid URL
   - unsupported URL
   - non-positive number
   - GitHub/GitLab 正常路径
2. 统一错误风格为单层上下文。
3. 保持 `Parse`、`parseGitHubTarget`、`parseGitLabTarget` 的职责边界清晰。
4. 如有必要，仅提取极小的 path validation helper。

### 预期结果 / Expected Result
- 错误消息更易读。
- parser 更容易维护。
- 符合第三条明确性原则。

---

## 8.2 优化项 B：修复 CLI 配置空加载 / Remove No-Op Config Loading in CLI

### 背景 / Background
`internal/cli/app.go` 当前加载 `config.LoadFromEnv(a.Env)`，但没有消费返回值。这是一个典型的“边界看似存在，行为实际不存在”的问题。

### 目标 / Goal
在以下两条路径中二选一，并通过测试锁定行为：

1. **短期最简路径 / short-term simplest path**
   - 删除空加载。
   - 让 CLI 只在真正需要认证时装配配置。

2. **面向 spec 的显式路径 / spec-aligned explicit path**
   - 把 `GITHUB_TOKEN` 显式注入 GitHub client。
   - 让配置加载成为真实依赖装配的一部分。

### 推荐决策 / Recommended Decision
**推荐采用第 2 条**，因为 `spec.md` 已明确声明 GitHub token 使用 `GITHUB_TOKEN` 环境变量，且这是实际需求，不是过度设计。

### 实现策略 / Implementation Strategy
1. 为 config -> GitHub client wiring 增加测试。
2. 在 `cmd/issue2md/main.go` 或 `internal/cli/app.go` 中明确配置注入责任。
3. 保持 GitLab 路径不引入 token。
4. 不允许通过 CLI flag 传 token。

### 预期结果 / Expected Result
- 配置读取不再是空操作。
- 认证行为有清晰来源。
- 责任边界仍然简单。

---

## 8.3 优化项 C：控制 GitHub/GitLab 重复逻辑增长 / Control Cross-Provider Duplication Growth

### 背景 / Background
`internal/github/client.go` 与 `internal/gitlab/client.go` 中都存在 `getJSON`、时间解析、comments/notes 排序等相似逻辑。

### 目标 / Goal
- 只在确有必要时收敛重复。
- 避免“为了抽象而抽象”。

### 实现策略 / Implementation Strategy
采用“**先复制，后审查，再小提取**”的策略：

1. 当前规模下，允许 provider 级少量重复。
2. 如果后续新增 PR / Discussion / more endpoints 导致重复显著增加，再提取以下候选 helper：
   - 公共 JSON GET helper
   - RFC3339 parse helper
   - chronological sort helper
3. 所有提取都必须满足：
   - 已有重复至少两处以上。
   - 抽取后不会引入 provider-agnostic 虚假抽象。
   - 抽取前后测试行为保持一致。

### 预期结果 / Expected Result
- 当前代码依然直观。
- 后续扩展时不至于复制失控。
- 保持第一条简单性原则。

---

## 8.4 优化项 D：稳定 Renderer 结构，但避免模板化过度设计 / Stabilize Renderer Without Over-Templating

### 背景 / Background
`renderer.go` 中 Issue / PR / Discussion 的渲染流程形状相似，但不完全一致。

### 目标 / Goal
- 控制重复。
- 保持阅读直观。
- 不引入复杂模板系统。

### 实现策略 / Implementation Strategy
1. 保持三类渲染函数独立。
2. 仅提取以下重复度高且稳定的片段：
   - summary block
   - reaction block write helper
   - comment section write helper
3. 不提取 provider registry、template engine 或 DSL。
4. 继续通过表格驱动测试和固定字符串断言保护输出格式。

### 预期结果 / Expected Result
- renderer 仍然一眼可读。
- 未来增加字段时不容易漏改。

---

## 8.5 优化项 E：把 Main 保持为装配层 / Keep `main` as a Pure Wiring Layer

### 背景 / Background
`cmd/issue2md/main.go` 当前已经较薄，但随着 token wiring 或更多 fetcher 行为加入，main 可能变胖。

### 目标 / Goal
- `main` 只做装配。
- 业务逻辑仍留在 `internal/cli` 与 `internal/*`。

### 实现策略 / Implementation Strategy
1. `main` 只负责：
   - 读取 env
   - 构造 fetchers
   - 构造 renderer factory
   - 调用 `app.Run`
2. 不把 URL 解析、错误渲染、provider dispatch 写回 `main`。
3. 通过 `cmd/issue2md/main_test.go` 继续保证其“可构建、低逻辑”属性。

---

## 9. 分阶段实施顺序 / Phased Execution Order

> 以下顺序用于指导后续真实改造。每一步都应遵循 Red-Green-Refactor。

### Phase 1: Parser 明确性优化 / Parser Clarity Optimization
1. 增加失败测试。
2. 统一 invalid / unsupported 错误表达。
3. 缩减重复包装。
4. 跑 `make test`。

### Phase 2: Config Wiring 明确化 / Config Wiring Clarification
1. 先写 CLI/config 行为测试。
2. 决定删除空加载或接入 `GITHUB_TOKEN`。
3. 推荐完成 GitHub token 注入。
4. 跑 `make test`。

### Phase 3: Provider 行为对齐 / Provider Behavior Alignment
1. 统一 GitHub/GitLab 错误上下文风格。
2. 检查 comments/notes 排序行为。
3. 如重复已明显增加，再做最小 helper 抽取。
4. 跑 `make test`。

### Phase 4: Renderer 可维护性优化 / Renderer Maintainability Optimization
1. 先补或收紧渲染输出测试。
2. 只提取稳定片段。
3. 保持输出内容完全不变。
4. 跑 `make test`。

### Phase 5: 总体验证 / Final Validation
1. `make test`
2. `go build ./cmd/issue2md`
3. `go build ./cmd/issue2mdweb`
4. 如需要，再执行 `go test ./...`

---

## 10. 测试策略 / Testing Strategy

### 10.1 总原则 / General Rules
- 新行为先写失败测试。
- 单元测试优先表格驱动。
- HTTP 行为优先 `httptest.Server`。
- 不使用 mock framework。

### 10.2 Parser Tests
覆盖：
- GitHub Issue URL
- GitHub PR URL
- GitHub Discussion URL
- GitLab Issue URL
- 非法 scheme
- 缺失 path segment
- unsupported URL
- query string 保留 source URL 的行为

### 10.3 Config Tests
覆盖：
- 读取 `GITHUB_TOKEN`
- lookup function 行为
- 不引入 GitLab token
- CLI 装配后配置被实际消费

### 10.4 Fetcher Tests
覆盖：
- 成功响应映射
- comments/notes 升序
- 非 200 响应
- 无效时间格式
- 认证配置接线（如启用）

### 10.5 Converter Tests
覆盖：
- frontmatter 字段完整性
- user links 开关
- reactions 开关
- Issue / PR / Discussion 各自输出结构
- accepted answer 显著标记

### 10.6 CLI Tests
覆盖：
- 缺少 URL
- 无效 URL
- 不支持 URL
- stdout 输出
- output file 输出
- flag 透传
- GitHub / GitLab provider 分派

---

## 11. 风险与权衡 / Risks and Trade-Offs

### 11.1 风险：为了消除重复而过度抽象
**应对 / Mitigation**: 只有当重复已稳定且超过两处时，才允许做小范围提取。

### 11.2 风险：配置接线导致 main 或 cli 职责膨胀
**应对 / Mitigation**: 保持 `main` 只装配，`cli` 只编排，不把 token 解析逻辑散落到 fetcher 之外。

### 11.3 风险：错误文案调整导致测试大面积变动
**应对 / Mitigation**: 测试优先断言关键信号词和语义，而不是脆弱的整段字符串。

### 11.4 风险：renderer 小重构破坏 Markdown 稳定性
**应对 / Mitigation**: 所有重构都要以固定输出测试保护，确保输出文本不变。

---

## 12. 验收标准 / Acceptance Criteria

本方案落地后，应满足以下条件：

1. `internal/parser` 错误输出更简洁，无重复包装噪音。
2. `internal/cli` 不再存在“读取但不使用”的配置行为。
3. GitHub 与 GitLab fetcher 的错误上下文风格一致。
4. `internal/converter` 在不引入复杂模板系统的前提下保持可维护。
5. `cmd/issue2md/main.go` 继续保持装配层职责。
6. 所有新增或调整行为均由测试先保护。
7. `make test` 通过。
8. `go build ./cmd/issue2md` 通过。
9. `go build ./cmd/issue2mdweb` 通过。

---

## 13. 推荐执行命令 / Recommended Commands

根据 `CLAUDE.md`，优先使用 Makefile：

```sh
make test
```

如需要做构建验证：

```sh
go build ./cmd/issue2md
go build ./cmd/issue2mdweb
```

如需要针对单包快速验证：

```sh
go test ./internal/parser
ugo test ./internal/config
go test ./internal/github
go test ./internal/gitlab
go test ./internal/converter
go test ./internal/cli
```

---

## 14. 最终决策摘要 / Final Decision Summary

本方案选择：

- **保留当前简单架构 / keep the current simple architecture**
- **优先修复明确性问题 / fix clarity issues first**
- **把配置接线从“存在但无效”改为“存在且被消费” / make config loading real rather than decorative**
- **只在重复真正成为维护负担时做最小提取 / extract only when duplication becomes a real maintenance cost**
- **所有改动继续遵循 TDD 与中英双语文档规范 / continue under TDD and bilingual documentation rules**

这份 `plan.md` 作为后续 `optimize-tasks.md` 调整、实现和验证的统一基线。

This `plan.md` becomes the baseline for future `optimize-tasks.md` refinement, implementation work, and validation.
