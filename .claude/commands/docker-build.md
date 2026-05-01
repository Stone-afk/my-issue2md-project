---
description: 使用 make docker-build 构建 issue2md Docker 镜像 / Build Docker image for the issue2md project using make docker-build
allowed-tools:
  - Bash(docker:*)
  - Bash(make:docker-build)
parameters:
  - name: tag
    description: Docker 镜像 Tag，未提供时默认使用 latest / Docker image tag, defaults to latest when omitted
    required: false
    default: "latest"
---

# 构建 issue2md Docker 镜像 / Build Docker Image for issue2md Project

## 用法 / Usage
- `/docker-build`               - 使用默认 tag `latest` 构建 / Build image with tag `latest` (default)
- `/docker-build v1.0.0`        - 使用 tag `v1.0.0` 构建 / Build image with tag `v1.0.0`
- `/docker-build my-custom-tag` - 使用自定义 tag 构建 / Build image with tag `my-custom-tag`

## 步骤 / Steps
1. 检查容器运行时是否可通过 Makefile 中的 `docker-build` 流程访问。  
   Check whether the container runtime is available through the Makefile `docker-build` flow.
2. 确认当前仓库可以执行项目级 Docker 构建流程。  
   Verify that the project-level Docker build flow can run from the current repository.
3. 使用传入参数作为镜像 Tag。  
   Use the provided parameter as the image tag.
4. 如果未提供参数，则默认使用 `latest`。  
   If no parameter is provided, default to `latest`.
5. 执行 `DOCKER_TAG=${1:-latest} make docker-build`。  
   Execute `DOCKER_TAG=${1:-latest} make docker-build`.
6. 如果构建成功，明确报告最终镜像名和 Tag。  
   If the build succeeds, clearly report the final image name and tag.

## 错误处理 / Error Handling
如果构建失败，应基于实际输出分析，并给出针对性的修复建议。常见问题包括：  
If the build fails, analyze the actual output and provide concrete remediation guidance. Common categories include:

- Docker CLI 或容器运行时不可用 / Docker CLI or container runtime not available
- Docker daemon 未启动 / Docker daemon not running
- 缺少 `Dockerfile` / Missing `Dockerfile`
- build context 或 `.dockerignore` 配置问题 / Build context or `.dockerignore` issues
- 权限问题 / Permission problems
- 磁盘空间不足 / Insufficient disk space
- 拉取基础镜像或下载依赖时的网络问题 / Network failures while downloading base images or dependencies
- Docker 构建步骤内部执行失败 / Errors inside Docker build steps themselves

## 输出要求 / Output Requirements
- 构建成功时 / On success:
  - 确认镜像构建完成 / confirm the image build completed
  - 明确给出最终镜像引用 / state the resulting image reference clearly
- 构建失败时 / On failure:
  - 指出失败阶段 / identify the failing stage
  - 根据实际输出总结具体原因 / summarize the concrete cause based on actual output
  - 给出简短、可执行的下一步建议 / provide short, actionable next steps
