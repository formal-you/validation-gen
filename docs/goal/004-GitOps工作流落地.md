# Goal 004：GitOps 工作流落地与开源就绪

- 状态：in progress
- GitHub Issue：[#1](https://github.com/formal-you/validation-gen/issues/1)
- 创建：2026-08-22

## 目标

让仓库形成真正的开源 GitOps 开发工作流：GitHub Issue/PR 为远程权威记录，
GitHub Actions CI 自动执行验收门禁，并补齐开源配套文档。

## 范围

- gitOps 文档与 `AGENTS.md` 约定；
- `.github/workflows/ci.yml` 与 Issue/PR 模板；
- CONTRIBUTING / CODE_OF_CONDUCT / SECURITY / CHANGELOG；
- README 徽章、文档入口与仓库结构更新；
- 本地镜像 `docs/issue/004`、`docs/goal/004`、`docs/pr/004`。

## 完成定义（DoD）

1. 远程存在 Issue #1 与合并后的 PR #1 记录；
2. CI 工作流入库并在 push/PR 时执行全套门禁；
3. 开源文档与 README 互相一致；
4. 验收门禁全绿；
5. 本地 `main` 与 `origin/main` 同步。
