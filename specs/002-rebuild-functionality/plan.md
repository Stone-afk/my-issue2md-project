# issue2md 重构技术方案 / Rebuild Functionality Plan

**规格文档 / Spec**: `specs/002-rebuild-functionality/spec.md`  
**任务清单 / Tasks**: `specs/002-rebuild-functionality/tasks.md`  
**日期 / Date**: 2026-04-30

## 1. 背景与目标 / Context and Goals

当前 `issue2md` 已具备 GitHub Issue、Pull Request、Discussion 与 GitLab Issue 的抓取和 Markdown 渲染能力，但 `App` 层仍直接感知具体实现，导致调用层与 provider 实现耦合偏紧，不利于后续新增 provider、扩展认证方式或调整 API 版本策略。

`issue2md` already supports fetching and rendering for GitHub Issues, Pull Requests, Discussions, and GitLab Issues, but the `App` layer still knows too much about concrete implementations. That keeps the calling layer too tightly coupled to provider details and makes future provider additions, authentication variants, and API-version evolution harder.

本次重构的目标不是新增终端用户功能，而是在**不改变 CLI 输入输出契约和 Markdown 输出内容**的前提下，建立一个更稳定、更薄的抓取抽象边界。

The goal of this refactor is not to add end-user functionality, but to establish a thinner and more stable fetch abstraction boundary **without changing the CLI contract or Markdown output**.

## 2. 设计原则 / Design Principles

- **最小抽象 / Minimal abstraction**：统一接口只保留 `Fetch(...)`，不把 `GetIssue(...)`、`GetDiscussion(...)` 等 provider 细节暴露给上层。
- **显式依赖 / Explicit dependencies**：provider registry、HTTP client、token 与其他配置继续显式注入。
- **差异内聚 / Provider-local differences**：GitHub 与 GitLab 的能力差异、认证方式、API 版本细节保留在各自实现内部。
- **上层变薄 / Thin orchestration layer**：`App` 只做 target 解析、provider 查找、统一错误包装、渲染和输出。
- **TDD 优先 / Test-first**：所有会改变行为的重构都先通过失败测试锁定行为，再最小实现通过。

## 3. 目标架构 / Target Architecture

### 3.1 Provider 抽象 / Provider Abstraction

新增中立包：`internal/fetchprovider`。

A new neutral package will be introduced: `internal/fetchprovider`.

该包负责定义：

This package owns:

- `Provider` 接口
- provider-not-registered 错误
- unsupported-capability 错误类型与判断辅助逻辑

目标接口保持最小：

```go
type Provider interface {
    Fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error)
}
```

该接口是调用层唯一依赖的抓取抽象。`GetIssue(...)`、`GetPullRequest(...)`、`GetDiscussion(...)` 如果仍有必要，应只保留在 provider 实现内部，不进入共享接口。

This interface is the only fetch abstraction the calling layer depends on. If `GetIssue(...)`, `GetPullRequest(...)`, or `GetDiscussion(...)` remain useful, they stay internal to provider implementations and do not become part of the shared interface.

### 3.2 Provider 实现位置 / Provider Implementation Location

GitHub 与 GitLab 的实现迁移到：

- `internal/fetchprovider/github`
- `internal/fetchprovider/gitlab`

这样目录语义更清晰：它们是统一 fetchprovider 抽象下的两个具体实现，而不是平铺在 `internal/` 顶层的孤立客户端。

This makes the directory semantics clearer: they become concrete implementations beneath a shared fetch-provider abstraction rather than standalone top-level internal clients.

### 3.3 App 调度模型 / App Dispatch Model

`internal/cli/app.go` 中的 `App` 收敛为仅依赖：

```go
map[model.Provider]fetchprovider.Provider
```

运行流程固定为：

1. 解析 URL 为 `parser.Target`
2. 根据 `target.Provider` 查找 provider
3. 调用 `Provider.Fetch(...)`
4. 将结果交给 renderer
5. 输出到 stdout 或文件

`App` 不再包含 GitHub/GitLab 的具体字段，也不再根据平台做 if/switch 分支判断能力。

`App` will no longer hold concrete GitHub/GitLab fields and will not use provider-specific branching to infer capabilities.

### 3.4 Main 装配模型 / Main Composition Model

`cmd/issue2md/main.go` 负责在启动时一次性完成以下动作：

- 读取环境配置
- 构造 GitHub provider
- 构造 GitLab provider
- 组装 `map[model.Provider]fetchprovider.Provider`
- 构造 renderer 工厂
- 注入 `App`

`main` 保持为装配层，不承载业务逻辑。

`main` remains a composition root and does not take on business logic.

## 4. 错误模型 / Error Model

需要明确定义两类上层可识别错误：

Two caller-visible error classes must be explicit:

1. **provider 未注册 / provider not registered**
   - 说明 registry 中不存在 `target.Provider` 对应实现。
2. **provider 不支持该 kind / provider does not support the requested kind**
   - 说明 provider 存在，但不支持当前 `target.Kind`。

在 provider 内部，仍然保留 provider-specific 错误上下文，例如：

- `get issue: unexpected status 418`
- `parse issue created_at: ...`
- `graphql discussion: ...`

在 `App` 层，对 provider 返回错误统一再包一层：

```go
fmt.Errorf("fetch document via %s: %w", target.Provider, err)
```

这样可以同时保留：

- 失败发生在“抓取文档”阶段的统一语义
- provider-specific 的底层诊断上下文

This gives both a unified “fetch document” semantic and the provider-specific low-level diagnostic context.

## 5. 配置策略 / Configuration Strategy

本次重构不抽象统一 provider 配置对象。

This refactor does not introduce a shared provider configuration object.

保留现状：

- GitHub 使用自己的 `Options`
- GitLab 使用自己的 `Options`
- token、HTTP client、base URL、未来的 API version 等都保留在具体 provider 内部

这样可以避免为还未出现的第二套真实认证/版本策略过早设计复杂抽象。

This avoids prematurely designing a broader abstraction before a second real authentication or versioning strategy exists.

## 6. 实施阶段 / Implementation Phases

### Phase 1: Foundation / 数据结构定义

目标：建立共享抽象与共享错误模型。

Key outputs:
- `internal/fetchprovider/types.go`
- 最小 `Provider` 接口
- provider-not-registered 与 unsupported-capability 错误定义

### Phase 2: GitHub Fetcher / API 交互逻辑，TDD

目标：完成 GitHub provider 迁移，并保持现有 Issue / Pull Request / Discussion 行为不变。

Key outputs:
- `internal/fetchprovider/github/client.go`
- 对应迁移后的测试
- 认证头行为保持由显式 token 注入驱动

### Phase 3: Markdown Converter / 转换逻辑，TDD

目标：锁定并保护 Markdown 输出兼容性，确保 provider 重构不外溢到渲染结果。

Key outputs:
- `internal/converter/frontmatter_test.go`
- `internal/converter/renderer_test.go`
- 如有必要，对 `internal/converter/renderer.go` 做最小修正

### Phase 4: CLI Assembly / 命令行入口集成

目标：让 `App` 与 `main` 完成对新抽象的切换。

Key outputs:
- `internal/cli/app.go`
- `internal/cli/app_test.go`
- `cmd/issue2md/main.go`
- `cmd/issue2md/main_test.go`

## 7. 关键文件 / Critical Files

- `internal/fetchprovider/types.go`
- `internal/fetchprovider/github/client.go`
- `internal/fetchprovider/gitlab/client.go`
- `internal/cli/app.go`
- `cmd/issue2md/main.go`
- `internal/converter/frontmatter_test.go`
- `internal/converter/renderer_test.go`
- `specs/002-rebuild-functionality/tasks.md`

## 8. 验证策略 / Verification Strategy

重构完成后应至少执行以下验证：

At minimum, run the following verification after the refactor:

```sh
go test ./internal/fetchprovider/...
go test ./internal/converter
go test ./internal/cli ./cmd/issue2md
make test
go build ./cmd/issue2md
go build ./cmd/issue2mdweb
```

验证目标：

- `App` 只依赖统一 provider 抽象
- `main` 只负责装配
- provider 未注册与 provider 不支持能力两类错误可区分
- GitHub / GitLab 现有抓取行为保持一致
- Markdown 输出完全兼容

## 9. 非目标 / Non-Goals

本次不做以下内容：

This refactor explicitly does not do the following:

- 不新增终端用户可见 CLI 功能
- 不改变 Markdown 文本结构或 frontmatter 契约
- 不新增 GitLab token 支持
- 不抽象统一 provider `Options`
- 不引入 factory 框架、插件系统或复杂 registry 机制
- 不为未来 API 版本策略设计额外抽象层

## 10. 风险与控制 / Risks and Controls

- **风险 / Risk**：迁移 provider 包路径时可能打破现有测试与导入关系。  
  **控制 / Control**：先迁移测试，再迁移实现，并在切换完成后删除旧文件。

- **风险 / Risk**：错误包装层级变化导致测试断言失效。  
  **控制 / Control**：在 Phase 1 先锁定错误分类，在 Phase 2/4 锁定最终错误文本语义。

- **风险 / Risk**：provider 重构意外影响 Markdown 输出。  
  **控制 / Control**：在 Phase 3 先强化 frontmatter 与 renderer 测试作为回归保护。
