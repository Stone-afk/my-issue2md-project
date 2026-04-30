---
description: 审查当前分支相对于 master/main 的 diff，并根据项目宪法给出反馈。
model: opus
allowed-tools: Read, Grep, Glob, Bash(git diff:*), Bash(git log:*)
---
请读取 `!git diff master...HEAD` 的输出，这包含了所有变更。
然后对变更的每一个文件进行审查，依据 @./constitution.md 的原则，
输出结构化审查报告（总体评价 / 优点 / 待改进项按优先级排序）。