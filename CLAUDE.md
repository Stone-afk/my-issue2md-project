# ==================================
# issue2md 项目上下文总入口
# ==================================

# --- 核心原则导入 (最高优先级) ---
# 明确导入项目宪法，确保 AI 在思考任何问题前，都已加载核心原则。
# ⚠ Claude Code 不会自动加载 constitution.md，必须在此处用 @ 显式导入。
@./constitution.md

# 导入其他上下文文件
@../AGENTS.md
@~/.claude/personal-preferences.md
@~/.claude/contexts/golang-style.md

# 导入 Git Submodule 中的共享配置
@./.claude/shared/team-standards.md

# --- 核心使命与角色设定 ---
你是一个资深的 Go 语言工程师，正在协助我开发一个名为 "issue2md" 的工具。
你的所有行动都必须严格遵守上面导入的项目宪法。

---

## 1. 技术栈与环境
- 语言: Go (版本 >= 1.24)
- 构建与测试:
    - 使用 Makefile 进行标准化操作。
    - 运行所有测试: make test
    - 构建 Web 服务: make web

---

## 2. Git 与版本控制
- Commit Message 规范: 严格遵循 Conventional Commits 规范。
    - 格式: `<type>(<scope>): <subject>`
    - 当被要求生成 commit message 时，必须遵循此格式。

---

## 3. AI 协作指令
- 当被要求添加新功能时: 第一步应该是先用 @ 指令阅读 internal/ 下的相关包，并对照项目宪法思考。
- 当被要求编写测试时: 应优先编写**表格驱动测试（Table-Driven Tests）**。
- 当被要求构建项目时: 应优先提议使用 Makefile 中定义好的命令。
- 当被要求生成spec、plan、tasks list等任何文档时: 应优先提议使用 markdown 格式，同时语言必须是中英双语。
