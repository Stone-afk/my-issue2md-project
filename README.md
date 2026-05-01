# issue2md

issue2md 是一个将 Issue / Pull Request / Discussion 页面内容抓取并整理为 Markdown 文档的 Go 工具。它面向“归档、沉淀、转存、二次整理”这类场景，当前支持 GitHub 与 GitLab 的部分内容类型。

issue2md is a Go tool that fetches Issue / Pull Request / Discussion content and converts it into Markdown. It is designed for archiving, note taking, export, and follow-up editing workflows, and currently supports selected content types from GitHub and GitLab.

## Overview / 项目简介

当前版本提供一个 CLI 入口 `issue2md`，根据输入 URL 识别平台与内容类型，然后：

The current version provides a CLI entrypoint, `issue2md`, which identifies the provider and content type from the input URL and then:

- 抓取远程内容 / fetches remote content
- 统一映射为内部文档模型 / maps it into a shared internal document model
- 渲染为 Markdown，包含 YAML frontmatter、摘要区与原始归档区 / renders Markdown with YAML frontmatter, a summary section, and a raw archive section
- 可输出到标准输出或写入目标文件 / writes either to stdout or to an output file

当前仓库还包含第二个入口 `issue2mdweb`，以及双模式 Docker 入口脚本；但 `cmd/issue2mdweb/main.go` 目前仍是占位实现，尚未提供实际 Web 服务能力。

The repository also contains a second entrypoint, `issue2mdweb`, plus a dual-mode Docker entrypoint script; however, `cmd/issue2mdweb/main.go` is still a placeholder and does not yet provide a working web service.

## Features / 核心特性

### Supported inputs / 当前支持的输入

- GitHub Issue  
  `https://github.com/OWNER/REPO/issues/123`
- GitHub Pull Request  
  `https://github.com/OWNER/REPO/pull/123`
- GitHub Discussion  
  `https://github.com/OWNER/REPO/discussions/123`
- GitLab Issue  
  `https://gitlab.com/GROUP/PROJECT/-/issues/123`

### Output structure / 输出结构

生成的 Markdown 当前包含这些结构：

Generated Markdown currently includes these sections:

- YAML frontmatter
  - `title`
  - `url`
  - `author`
  - `created_at`
- `# Title`
- `## Summary / 摘要`
- `## Structured Notes / 结构化笔记`
- `## Raw Archive / 原始归档`
- 评论区 / comment sections
- Pull Request 的 review comments / pull request review comments
- Discussion 的 accepted answer / accepted answer for discussions

### Optional rendering features / 可选渲染能力

- `-enable-reactions`：包含 reactions 摘要 / include reaction summaries
- `-enable-user-links`：将作者名渲染为 Markdown 链接（当平台返回用户 URL 时）/ render author names as Markdown links when user URLs are available

### Provider behavior / Provider 行为

- GitHub provider 当前支持：Issue、Pull Request、Discussion
- GitLab provider 当前仅支持：Issue
- 不支持的 provider / kind 会返回显式错误

The GitHub provider currently supports issues, pull requests, and discussions. The GitLab provider currently supports issues only. Unsupported provider/kind combinations return explicit errors.

## Installation / 安装指南

### Prerequisites / 前置要求

- Go `1.24.7` 或更高兼容版本 / Go `1.24.7` or a compatible newer version
- 可访问 GitHub / GitLab API 的网络环境 / network access to GitHub / GitLab APIs

### Clone and build / 克隆并构建

```bash
git clone https://github.com/stoneafk/issue2md.git
cd issue2md
make build
```

构建完成后会在 `./bin` 下生成：

After the build completes, the following binaries are produced under `./bin`:

- `bin/issue2md`
- `bin/issue2mdweb`

### Optional GitHub authentication / 可选的 GitHub 认证

当前代码只读取一个环境变量：

The current code reads exactly one environment variable:

```bash
export GITHUB_TOKEN="your-token"
```

说明：

Notes:

- `GITHUB_TOKEN` 会传给 GitHub provider 的 `Authorization: Bearer ...` 请求头
- 对公开仓库，未设置 token 时部分请求仍可能可用，但更容易受到速率限制影响
- GitLab provider 当前没有实现额外的 token 配置入口

`GITHUB_TOKEN` is passed to the GitHub provider as a `Bearer` token. Public GitHub content may still work without it, but you are more likely to hit rate limits. The GitLab provider does not currently expose additional token configuration.

## Usage / 使用方法

### Command syntax / 命令格式

```bash
issue2md [flags] <url> [output_file]
```

### Flags / 命令行参数

#### `-enable-reactions`

包含文档主体和评论中的 reaction 摘要。

Include reaction summaries from the main content and comments.

默认值 / Default: `false`

示例 / Example:

```bash
issue2md -enable-reactions https://github.com/OWNER/REPO/issues/123
```

#### `-enable-user-links`

将作者名渲染为 Markdown 链接；如果上游数据没有用户 URL，则仍会退化为纯文本用户名。

Render author names as Markdown links. If the upstream payload does not include a user URL, the renderer falls back to plain text usernames.

默认值 / Default: `false`

示例 / Example:

```bash
issue2md -enable-user-links https://github.com/OWNER/REPO/pull/123
```

### Positional arguments / 位置参数

#### `<url>`

必填。必须是 `https` URL，并且符合当前支持的平台格式。

Required. Must be an `https` URL matching one of the currently supported provider patterns.

#### `[output_file]`

可选。提供时，生成的 Markdown 会写入该文件；未提供时，输出到标准输出。

Optional. If provided, the generated Markdown is written to that file; otherwise it is printed to stdout.

### Examples / 使用示例

#### 1) Export a GitHub issue to stdout / 导出 GitHub Issue 到标准输出

```bash
./bin/issue2md https://github.com/OWNER/REPO/issues/123
```

#### 2) Export a GitHub pull request with reactions / 导出带 reactions 的 GitHub Pull Request

```bash
./bin/issue2md -enable-reactions https://github.com/OWNER/REPO/pull/123
```

#### 3) Export a GitHub discussion to a file / 导出 GitHub Discussion 到文件

```bash
./bin/issue2md https://github.com/OWNER/REPO/discussions/123 out.md
```

#### 4) Render user links / 渲染用户链接

```bash
./bin/issue2md -enable-user-links https://github.com/OWNER/REPO/issues/123 out.md
```

#### 5) Export a GitLab issue / 导出 GitLab Issue

```bash
./bin/issue2md https://gitlab.com/GROUP/PROJECT/-/issues/123
```

### Example output shape / 输出示例结构

```md
---
title: "Issue title"
url: "https://github.com/OWNER/REPO/issues/123"
author: "octocat"
created_at: "2026-04-28T10:30:00Z"
---

# Issue title

## Summary / 摘要

- Author: octocat
- State: open

## Structured Notes / 结构化笔记

Issue body

## Raw Archive / 原始归档

Issue body

### Comment by reviewer

First comment
```

## Building from Source / 从源码构建

项目优先通过 Makefile 进行构建、测试和本地开发。

This project is intended to be built, tested, and operated locally through the Makefile.

### Common targets / 常用目标

```bash
make build
make test
make lint
make format
make verify
make clean
make web
```

含义如下：

These targets do the following:

- `make build`：构建 `issue2md` 与 `issue2mdweb` 到 `./bin`
- `make test`：运行所有 Go 测试
- `make lint`：当本机安装了 `golangci-lint` 时运行 lint
- `make format`：运行 `go fmt`，并在可用时执行 `goimports`
- `make verify`：执行 `format` 和 `test`
- `make clean`：清理本地构建产物
- `make web`：单独构建 `cmd/issue2mdweb`

### Developer tool setup / 开发工具安装

```bash
make dev-setup
```

该命令会安装可选本地工具：

This installs optional local development tools:

- `golangci-lint`
- `goimports`

## Docker / 容器构建

仓库包含一个双模式 Docker 镜像构建方案：

The repository includes a dual-mode Docker image setup:

- 默认模式运行 CLI / default mode runs the CLI
- 传入首个参数 `web` 时切换到 `issue2mdweb` / passing `web` as the first argument dispatches to `issue2mdweb`

### Build the image / 构建镜像

```bash
make docker-build
```

指定 tag：

With a custom tag:

```bash
DOCKER_TAG=latest make docker-build
```

### Run CLI mode / 运行 CLI 模式

```bash
make docker-run-cli URL=https://github.com/OWNER/REPO/issues/123
```

将结果写回当前目录挂载文件：

Write the output to a mounted file in the current directory:

```bash
make docker-run-cli URL=https://github.com/OWNER/REPO/issues/123 OUTPUT=/work/out.md
```

### Run web placeholder mode / 运行 web 占位模式

```bash
make docker-run-web
```

注意：当前 `issue2mdweb` 仍是占位实现，因此该模式只代表镜像分发 wiring 已就绪，不代表 Web 服务已完成。

Note: `issue2mdweb` is still a placeholder, so this only verifies the image dispatch wiring, not a finished web application.

## CI / 持续集成

仓库包含 GitHub Actions 工作流 `.github/workflows/ci.yml`，当前会在以下场景触发：

The repository includes a GitHub Actions workflow at `.github/workflows/ci.yml`, triggered on:

- push 到 `main`
- Pull Request 事件

流水线当前执行：

The pipeline currently runs:

- `make test`
- `make lint`
- `make build`

当 CI 失败时，还会触发 Claude Headless 诊断步骤读取日志并给出最小修复建议。

When CI fails, a Claude Headless diagnostic step analyzes the workflow logs and suggests the smallest next fix.

## Current limitations / 当前限制

- 仅支持 `https` 输入 URL / only `https` input URLs are supported
- GitLab 当前仅支持 Issue，不支持 Merge Request / Discussion / current GitLab support is limited to issues
- `issue2mdweb` 目前仍未实现实际 Web 服务 / `issue2mdweb` is not yet a working web service
- CLI 当前只读取 `GITHUB_TOKEN`，未提供更多 provider 配置入口 / the CLI currently reads only `GITHUB_TOKEN` and does not expose further provider configuration
