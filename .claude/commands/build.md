---
description: 使用 make build 构建 issue2md 项目 / Build the issue2md project using make build
allowed-tools:
  - Bash(make:build)
---

# 构建 issue2md 项目 / Build the issue2md Project

执行 `make build` 来编译 `issue2md` 和 `issue2mdweb` 两个二进制文件。  
Execute `make build` to compile the `issue2md` and `issue2mdweb` binaries.

如果构建失败，应基于实际错误输出分析原因，并提供对应的排查建议。  
If the build fails, analyze the actual error output and provide troubleshooting guidance based on the real failure.

## 步骤 / Steps
1. 首先执行 `make build`，构建 CLI 和 Web 两个应用。  
   First, run `make build` to compile both the CLI and web applications.
2. 如果构建成功，确认 Makefile 构建流程已经产出目标二进制文件。  
   If the build succeeds, confirm that the Makefile build flow produced the expected binaries.
3. 如果构建失败，先分析输出，再判断最可能的根因。  
   If the build fails, analyze the output first and then determine the most likely root cause.

## 错误分析指引 / Error Analysis Guidelines
当构建失败时，应尽量精确分类，不要泛泛而谈。常见问题包括：  
When the build fails, classify the error as precisely as possible rather than giving generic advice. Common categories include:

- Go 编译错误 / Go compilation errors
  - 语法错误 / syntax errors
  - 类型不匹配 / type mismatches
  - 未定义标识符 / undefined identifiers
  - 缺失导入 / missing imports
- 模块或依赖问题 / Module or dependency issues
  - `go.mod` 无效 / invalid `go.mod`
  - 依赖缺失或不一致 / missing or inconsistent dependencies
- 文件系统或权限问题 / Filesystem or permission issues
  - 无法创建输出目录 / cannot create output directory
  - 无法写入构建产物 / cannot write build artifacts
- 工具链或环境问题 / Toolchain or environment issues
  - Go 工具链缺失 / missing Go toolchain
  - Go 版本不受支持 / unsupported Go version
  - 平台相关环境配置错误 / invalid platform-specific environment
- Makefile 目标问题 / Makefile target issues
  - 目标缺失 / missing target
  - recipe 写法错误 / invalid recipe
  - `make build` 内部 shell 命令失败 / shell command failure inside `make build`

## 输出要求 / Output Requirements
- 如果构建成功 / If the build succeeds:
  - 简要确认成功 / briefly confirm success
  - 说明 CLI 和 Web 二进制已通过 `make build` 构建 / mention that the CLI and web binaries were built via `make build`
- 如果构建失败 / If the build fails:
  - 明确指出失败阶段 / identify the failing stage
  - 基于实际输出总结具体原因 / summarize the concrete reason based on the actual output
  - 给出简短、可执行的修复建议 / provide short, actionable remediation steps
  - 如果错误本身不明确，再说明还需要检查什么 / if the error is ambiguous, explain what to inspect next
