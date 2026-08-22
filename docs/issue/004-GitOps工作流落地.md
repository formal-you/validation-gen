# Issue 004：GitOps 工作流落地与开源就绪

- 状态：open
- GitHub Issue：[#1](https://github.com/formal-you/validation-gen/issues/1)
- 创建：2026-08-22

## 背景

仓库已开源（public）。此前 gitOps 文档描述的是「本地闭环 + 远程推送」：
Issue 与 PR 只记录在本地 docs，远程 GitHub 上没有 Issue/PR 记录，
未形成可被查看、评审、追溯的 GitOps 开发工作流；同时缺少开源项目应有的
GitHub Actions CI 与配套文档（CONTRIBUTING / CODE_OF_CONDUCT / SECURITY / CHANGELOG）。

## 需求

1. 修正 `docs/gitops/README.md`：GitHub Issue/PR 为远程权威记录，本地 docs 仅镜像。
2. 修正 `AGENTS.md`「文档流（gitOps）」约定，并补充 GitHub Actions CI 说明。
3. 新增 GitHub Actions CI（`.github/workflows/ci.yml`）：
   generate / build / vet / test / race + golden 门禁（`git diff --exit-code`）。
4. 新增开源配套文档：CONTRIBUTING / CODE_OF_CONDUCT / SECURITY / CHANGELOG、
   Issue/PR 模板，并更新 README（CI 徽章、文档入口、仓库结构）。
5. 本变更以 GitOps 方式落地：GitHub Issue + GitHub PR，合并后远程可见记录。

## 验收

- [ ] `docs/gitops/README.md` 与 `AGENTS.md` 描述 GitHub Issue/PR 权威工作流；
- [ ] GitHub Actions CI 已入库且覆盖全部验收门禁；
- [ ] 开源文档与 README 已更新且互相一致；
- [ ] 本变更通过 GitHub Issue #1 + GitHub PR 完成并合并；
- [ ] `gh issue list` / `gh pr list` 远程可见对应记录。
