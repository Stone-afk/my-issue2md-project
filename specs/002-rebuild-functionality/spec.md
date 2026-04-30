# issue2md 重构功能规格 / Rebuild Functionality Specification

## 变更记录 / Change Log

### 2026-04-30: Provider 抽象重构 / Provider Abstraction Refactor

- 变更：将 GitHub 与 GitLab 抓取能力重构为统一的 provider 抽象，并在 App 层通过 `map[model.Provider]Provider` 进行装配与调度。
- Change: Refactor GitHub and GitLab fetching into a unified provider abstraction and assemble/dispatch them in the App layer via `map[model.Provider]Provider`.
- 目标：降低 `App` 对具体实现的耦合，为未来新增 provider、认证方式和 API 版本扩展预留稳定边界。
- Goal: Reduce `App` coupling to concrete implementations and create a stable boundary for future provider, authentication, and API version expansion.
- 范围：本次变更聚焦抓取层重构，不改变 CLI 输入输出契约，不新增最终用户可见功能。
- Scope: This change focuses on fetch-layer refactoring, does not change the CLI input/output contract, and does not add end-user-visible features.

## 概述 / Overview

`issue2md` 当前已经支持 GitHub Issue、Pull Request、Discussion，以及 GitLab Issue 的 Markdown 归档能力。本规格定义一次面向架构的重构：将具体平台抓取逻辑收敛到统一 provider 抽象后面，使上层应用仅依赖稳定接口，而不再感知 GitHub 或 GitLab 的具体实现差异。

`issue2md` already supports Markdown archiving for GitHub Issues, Pull Requests, Discussions, and GitLab Issues. This specification defines an architecture-oriented refactor that moves concrete platform fetching behind a unified provider abstraction so upper layers depend only on a stable interface and no longer care about GitHub or GitLab implementation details.

本次重构的核心原则是“最小抽象、显式依赖、保留 provider 内部差异”。`App` 只负责解析目标、查找 provider、调用 `Fetch(...)`、统一包装错误；provider 自身负责认证、请求细节、能力差异和内容适配。

The core principle of this refactor is “minimal abstraction, explicit dependencies, and preserving provider-local differences.” The `App` layer only parses the target, looks up the provider, calls `Fetch(...)`, and wraps errors consistently; each provider remains responsible for authentication, request details, capability differences, and content adaptation.

## 用户故事 / User Stories

### CLI 重构用户故事 / CLI Refactor User Stories

- 作为维护者，我希望 `App` 只依赖统一的 provider 接口，从而在不修改主流程的前提下替换或新增平台实现。
- As a maintainer, I want `App` to depend only on a unified provider interface so platform implementations can be replaced or added without changing the main flow.

- 作为开发者，我希望 `cmd/issue2md/main.go` 在启动时一次性构造 provider registry，并注入 `App`，从而让装配逻辑集中且可测试。
- As a developer, I want `cmd/issue2md/main.go` to build the provider registry once at startup and inject it into `App` so composition stays centralized and testable.

- 作为开发者，我希望新增 provider 时只需实现统一接口并注册到 `map[model.Provider]Provider`，从而避免修改 `App` 内部的 provider-specific 分支。
- As a developer, I want adding a new provider to require only implementing the unified interface and registering it in `map[model.Provider]Provider` so I can avoid provider-specific branching inside `App`.

- 作为维护者，我希望 provider 不支持某类内容时返回显式错误，而不是由 `App` 推断平台能力，从而保持职责边界清晰。
- As a maintainer, I want providers to return explicit unsupported-capability errors instead of having `App` infer platform capabilities, so responsibilities stay clear.

- 作为调试者，我希望最终错误同时包含统一的 `fetch document` 语义和 provider-specific 上下文，从而既能快速定位失败阶段，也能保留底层诊断信息。
- As a debugger, I want final errors to contain both a uniform `fetch document` semantic and provider-specific context so failures are easy to localize without losing low-level diagnostics.

### 未来 Web 版用户故事 / Future Web User Stories

以下用户故事用于指导未来 Web 版接入，但不要求本次实现 Web 功能。

The following stories guide future web integration but do not require implementing web functionality in this change.

- 作为 Web 服务开发者，我希望 Web handler 也能复用统一的 provider registry，从而与 CLI 共享抓取主流程。
- As a web service developer, I want web handlers to reuse the same provider registry so the fetch flow can be shared with the CLI.

- 作为产品维护者，我希望未来接入不同认证方式时只改具体 provider 配置，而不修改 Web 或 CLI 的主流程控制逻辑。
- As a product maintainer, I want future authentication variants to be implemented by changing provider-specific configuration only, without modifying the main CLI or web control flow.

- 作为架构设计者，我希望未来支持不同 API 版本时，版本差异留在具体 provider 中处理，从而避免上层接口污染。
- As an architect, I want future API-version differences to remain inside concrete providers so upper-layer interfaces stay clean.

## 功能性需求 / Functional Requirements

### Provider 抽象 / Provider Abstraction

- 系统必须定义一个 provider-neutral 的最小接口，放置在 `internal/fetchprovider/types.go`。
- The system must define a provider-neutral minimal interface located at `internal/fetchprovider/types.go`.

- 该接口必须只暴露一个抓取入口：
- The interface must expose only one fetch entrypoint:

```go
type Provider interface {
    Fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error)
}
```

- `App` 不得依赖 `GetIssue(...)`、`GetPullRequest(...)`、`GetDiscussion(...)` 等 provider-specific 细粒度方法。
- `App` must not depend on provider-specific fine-grained methods such as `GetIssue(...)`, `GetPullRequest(...)`, or `GetDiscussion(...)`.

- GitHub 与 GitLab 的具体实现必须迁移到 `internal/fetchprovider/github` 和 `internal/fetchprovider/gitlab`。
- Concrete GitHub and GitLab implementations must live under `internal/fetchprovider/github` and `internal/fetchprovider/gitlab`.

- provider 内部可以继续保留 `GetIssue(...)` 等细粒度方法作为实现细节，但这些方法不得进入统一抽象接口。
- Providers may continue to keep fine-grained methods like `GetIssue(...)` as implementation details, but those methods must not become part of the unified abstraction.

### App 装配与调度 / App Assembly and Dispatch

- `internal/cli/app.go` 中的 `App` 必须只依赖 `map[model.Provider]fetchprovider.Provider`。
- The `App` in `internal/cli/app.go` must depend only on `map[model.Provider]fetchprovider.Provider`.

- `cmd/issue2md/main.go` 必须在启动时一次性构造 provider 实例，并完成 registry 组装后注入 `App`。
- `cmd/issue2md/main.go` must construct provider instances once at startup, assemble the registry, and inject it into `App`.

- `App` 在运行时必须执行以下固定流程：
- At runtime, `App` must execute the following fixed flow:
  1. 解析 URL 为 `parser.Target`
  2. 使用 `target.Provider` 作为 key 从 registry 查找 provider
  3. 调用 `Provider.Fetch(...)`
  4. 将返回的 `DocumentData` 交给 renderer
  5. 统一处理输出到 stdout 或文件

- provider 查找必须使用 `model.Provider` 强类型 key，不得使用裸字符串 key。
- Provider lookup must use `model.Provider` as a strongly typed key and must not use raw string keys.

- 当 registry 中不存在目标 provider 时，`App` 必须返回显式错误。
- When the target provider does not exist in the registry, `App` must return an explicit error.

### Provider 配置 / Provider Configuration

- 每个 provider 必须继续使用各自独立的构造函数和 `Options` 类型。
- Each provider must continue to use its own constructor and `Options` type.

- 本次重构不得引入统一的跨-provider `Options` 结构。
- This refactor must not introduce a shared cross-provider `Options` structure.

- 认证方式、HTTP client、Base URL、未来 API 版本等配置必须保留在具体 provider 的 `Options` 中。
- Authentication mode, HTTP client, Base URL, future API versioning, and similar concerns must remain inside concrete provider `Options`.

- 未来如果出现第二种真实认证方式或第二套版本策略，再决定是否抽象，不得为假设场景提前设计。
- The system must defer abstraction for authentication or versioning until a second real strategy exists and must not pre-design for hypothetical cases.

### 能力差异处理 / Capability Differences

- provider 必须自行决定是否支持 `target.Kind` 对应的内容类型。
- Each provider must decide for itself whether it supports the content type represented by `target.Kind`.

- 当 provider 不支持某种内容类型时，provider 必须返回显式的“不支持能力”错误。
- When a provider does not support a content kind, it must return an explicit unsupported-capability error.

- `App` 不得包含 `if provider == github` 或 `if provider == gitlab` 之类的 provider-specific 分支。
- `App` must not contain provider-specific branching such as `if provider == github` or `if provider == gitlab`.

- `App` 也不得根据 `target.Kind` 预先推断某个 provider 是否支持某能力；该判断必须由 provider 本身负责。
- `App` must not infer provider support for a kind ahead of time based on `target.Kind`; that decision must be owned by the provider itself.

### 错误模型 / Error Model

- 系统必须定义一个通用的“不支持能力”错误类型或等价的显式错误标识，放置在 `internal/fetchprovider` 中。
- The system must define a shared unsupported-capability error type or equivalent explicit error marker in `internal/fetchprovider`.

- 该错误必须允许调用方区分以下至少两类情况：
- The error must allow callers to distinguish at least these cases:
  - provider 未注册 / provider not registered
  - provider 已注册但不支持该 kind / provider registered but does not support that kind

- provider 内部错误必须保留 provider-specific 上下文，例如请求阶段、状态码、解析失败点等。
- Provider-internal errors must preserve provider-specific context such as request stage, status code, or parsing failure point.

- `App` 在捕获 provider 错误后，必须再包装一层统一语义，至少体现 `fetch document` 和 provider 标识。
- After catching provider errors, `App` must wrap them with a unified semantic layer that includes at least `fetch document` and the provider identity.

- 错误向上传递时必须使用 `fmt.Errorf("...: %w", err)`。
- Propagated errors must use `fmt.Errorf("...: %w", err)`.

### CLI 与输出兼容性 / CLI and Output Compatibility

- 本次重构不得改变现有 CLI 命令格式：
- This refactor must not change the current CLI command shape:

```sh
issue2md [flags] <url> [output_file]
```

- 本次重构不得改变现有 renderer 行为、frontmatter 字段、评论排序、reaction flag、user links flag 的外部表现。
- This refactor must not change external behavior for rendering, frontmatter fields, comment ordering, reaction flags, or user-links flags.

- 对同一输入，重构前后的 Markdown 输出必须保持一致。
- For the same input, Markdown output must stay identical before and after the refactor.

## 非功能性需求 / Non-Functional Requirements

- 架构必须保持解耦：URL 解析、provider 抓取、内容模型、renderer、CLI 输出必须边界清晰。
- The architecture must remain decoupled: URL parsing, provider fetching, content modeling, rendering, and CLI output must have clear boundaries.

- `App` 必须变得更薄，只承担 orchestration，不承担 provider 选择之外的 provider-specific 知识。
- `App` must become thinner and serve only as orchestration, without owning provider-specific knowledge beyond provider selection.

- 统一抽象必须保持最小化，避免为了未来 provider、认证方式、API 版本而提前引入复杂接口或工厂层级。
- The unified abstraction must remain minimal and avoid introducing complex interfaces or factory hierarchies for speculative future provider, authentication, or API-version needs.

- 实现必须优先使用 Go 标准库。
- The implementation must prefer the Go standard library.

- 依赖必须显式注入，不得通过全局变量隐藏 provider registry、HTTP client 或配置。
- Dependencies must be injected explicitly and must not hide provider registries, HTTP clients, or configuration behind globals.

- 所有新实现必须遵循测试先行，优先使用表格驱动测试。
- All new implementation work must follow test-first development and prefer table-driven tests.

- 重构后代码必须保持可扩展，但扩展点应建立在真实需求基础上，而不是猜测未来场景。
- The refactored code must remain extensible, but extension points should be introduced only in response to real needs rather than guessed future scenarios.

## 验收标准 / Acceptance Criteria

### Provider Registry / Provider Registry

- 给定 `App` 已注入 `map[model.Provider]Provider`，当 `target.Provider` 为 `model.ProviderGitHub` 时，`App` 会从 map 中取出 GitHub provider 并调用其 `Fetch(...)`。
- Given an `App` injected with `map[model.Provider]Provider`, when `target.Provider` is `model.ProviderGitHub`, `App` retrieves the GitHub provider from the map and calls its `Fetch(...)`.

- 给定 `App` 已注入 `map[model.Provider]Provider`，当 `target.Provider` 为 `model.ProviderGitLab` 时，`App` 会从 map 中取出 GitLab provider 并调用其 `Fetch(...)`。
- Given an `App` injected with `map[model.Provider]Provider`, when `target.Provider` is `model.ProviderGitLab`, `App` retrieves the GitLab provider from the map and calls its `Fetch(...)`.

- 给定 registry 中缺少目标 provider，当用户运行受支持 URL 时，`App` 返回“provider 未注册”错误并以非零状态退出。
- Given the registry is missing the target provider, when the user runs a supported URL, `App` returns a “provider not registered” error and exits non-zero.

### Provider Capability Errors / Provider Capability Errors

- 给定 provider 已注册但不支持 `target.Kind`，当 `Fetch(...)` 被调用时，provider 返回可识别的“不支持能力”错误。
- Given a provider is registered but does not support `target.Kind`, when `Fetch(...)` is called, the provider returns a recognizable unsupported-capability error.

- 给定 provider 返回“不支持能力”错误，当错误传递到 `App` 时，最终错误同时包含 `fetch document` 语义与原始 provider 上下文。
- Given a provider returns an unsupported-capability error, when the error reaches `App`, the final error contains both `fetch document` semantics and the original provider context.

### App Orchestration / App Orchestration

- 给定一个合法 GitHub Issue URL，`App` 必须完成“parse target → lookup provider → fetch → render → output”的固定流程，且流程中不出现 provider-specific 条件分支。
- Given a valid GitHub Issue URL, `App` must execute the fixed flow “parse target → lookup provider → fetch → render → output” with no provider-specific conditional branches.

- 给定一个合法 GitLab Issue URL，`App` 必须通过相同的主流程完成抓取和渲染，而不引入第二套执行路径。
- Given a valid GitLab Issue URL, `App` must complete fetching and rendering via the same main flow without introducing a second execution path.

### Constructor and Configuration / Constructor and Configuration

- 给定 GitHub provider 需要 token，`cmd/issue2md/main.go` 在启动时使用 GitHub 自身的 `Options` 构造实例，并注册到 map 中。
- Given the GitHub provider needs a token, `cmd/issue2md/main.go` constructs it at startup using GitHub-specific `Options` and registers it in the map.

- 给定 GitLab provider 不需要 token，`cmd/issue2md/main.go` 仍使用 GitLab 自身的 `Options` 构造实例，而不要求统一配置结构。
- Given the GitLab provider does not need a token, `cmd/issue2md/main.go` still constructs it using GitLab-specific `Options` and does not require a unified configuration structure.

### CLI Compatibility / CLI Compatibility

- 给定当前已有 CLI 用例，当完成重构后，原有命令行测试必须继续通过。
- Given the current CLI test cases, all existing command-line tests must continue to pass after the refactor.

- 给定当前已有 provider 与 renderer 测试，当完成重构后，已有抓取与渲染行为测试必须继续通过。
- Given the current provider and renderer tests, the existing fetch and render behavior tests must continue to pass after the refactor.

### 具体测试 Case / Concrete Test Cases

- `App` 使用 `model.Provider` 作为 key 从 registry 正确分发到 GitHub provider。
- `App` dispatches to the GitHub provider correctly using `model.Provider` as the registry key.

- `App` 使用 `model.Provider` 作为 key 从 registry 正确分发到 GitLab provider。
- `App` dispatches to the GitLab provider correctly using `model.Provider` as the registry key.

- 当 registry 中不存在某 provider 时，返回可断言的未注册错误。
- When a provider does not exist in the registry, the app returns an assertable not-registered error.

- 当 provider 不支持某 `ContentKind` 时，返回可断言的 unsupported-capability 错误。
- When a provider does not support a `ContentKind`, it returns an assertable unsupported-capability error.

- `App` 对 provider 返回的错误执行统一包装，错误文本包含 provider 名称与 `fetch document`。
- `App` wraps provider errors consistently, and the error text includes both the provider name and `fetch document`.

- `cmd/issue2md/main.go` 在启动时完成 provider registry 组装并成功注入 `App`。
- `cmd/issue2md/main.go` assembles the provider registry at startup and injects it into `App` successfully.

- GitHub 与 GitLab provider 继续保留各自构造函数与 `Options`，测试无需共享统一配置结构。
- GitHub and GitLab providers keep their own constructors and `Options`, and tests do not require a shared configuration structure.

- 重构前后同一输入的 Markdown 输出一致。
- Markdown output remains identical for the same input before and after the refactor.

## 输出格式示例 / Output Format Examples

### Provider 接口示例 / Provider Interface Example

```go
package fetchprovider

import (
    "context"

    "github.com/stoneafk/issue2md/internal/model"
    "github.com/stoneafk/issue2md/internal/parser"
)

type Provider interface {
    Fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error)
}
```

### App Registry 组装示例 / App Registry Assembly Example

```go
providers := map[model.Provider]fetchprovider.Provider{
    model.ProviderGitHub: github.NewClient(github.Options{Token: cfg.GitHubToken}),
    model.ProviderGitLab: gitlab.NewClient(gitlab.Options{}),
}

app := cli.App{
    Providers: providers,
    Stdout:    os.Stdout,
    Stderr:    os.Stderr,
}
```

### 统一错误包装示例 / Unified Error Wrapping Example

```go
provider, ok := a.Providers[target.Provider]
if !ok {
    return fmt.Errorf("fetch document via %s: %w", target.Provider, fetchprovider.ErrProviderNotRegistered)
}

doc, err := provider.Fetch(ctx, target, opts)
if err != nil {
    return fmt.Errorf("fetch document via %s: %w", target.Provider, err)
}
```

### CLI 输出兼容性示例 / CLI Output Compatibility Example

```sh
issue2md https://github.com/OWNER/REPO/issues/123
issue2md -enable-reactions https://gitlab.com/GROUP/PROJECT/-/issues/123 output.md
```

```md
---
title: "Example Issue"
url: "https://github.com/OWNER/REPO/issues/123"
author: "octocat"
created_at: "2026-04-28T10:30:00Z"
---

# Example Issue

## Summary / 摘要

- Author: octocat
- State: open

## Structured Notes / 结构化笔记

Issue body

## Raw Archive / 原始归档

Issue body

### Comment by hubot

First comment
```
