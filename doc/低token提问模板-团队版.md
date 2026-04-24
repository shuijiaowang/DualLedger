# 低 Token 提问模板（团队版）

## 目标
让 AI 在最小上下文下完成可验证交付，减少每轮 token 消耗。

## 标准任务模板（团队统一）
```md
# Task Brief (Low-Token Standard)

## Scope
- Goal: <one sentence>
- In scope files: <explicit file list>
- Out of scope: <what must not be touched>

## Context (minimal)
- Current behavior: <1-3 lines>
- Expected behavior: <1-3 lines>
- Evidence: <error/log/endpoint/test case, <=10 lines>

## Constraints
- No full-repo scan unless approved
- No generated files / dist / node_modules analysis
- Keep diff minimal and localized
- Ask before architecture-level refactor

## Deliverable
- 1) Root cause
- 2) Exact files changed
- 3) Verification steps (commands or manual checks)
```

## 团队执行规则（建议纳入开发流程）
- 新任务默认新会话，不复用超长历史会话。
- 未指定 `In scope files` 时，AI 只做定位，不直接改代码。
- 大文档仅给章节和片段，不整篇投喂。
- 任务拆分为三步：定位 -> 修改 -> 验证。
- 前后端问题分开提问：前端仅 `web/src/**`，后端仅 `service/**`。

## 团队验收清单
- 是否限制了文件范围（最多 3-8 个）？
- 是否明确了非目标（Out of scope）？
- 是否提供了最小证据（报错/日志）？
- 是否要求了验证步骤？
