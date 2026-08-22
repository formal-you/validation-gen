# PR 004：GitOps 工作流落地与开源就绪

- 状态：merged（GitHub 合并，远程权威记录）
- 分支：`docs/004-gitops-workflow` -> `main`
- GitHub PR：[#2](https://github.com/formal-you/validation-gen/pull/2)
- 关联 Issue：[#1](https://github.com/formal-you/validation-gen/issues/1)（004 GitOps 工作流落地）

## 摘要

仓库已开源，但此前 gitOps 文档是「本地闭环 + 远程推送」：Issue/PR 只记录在本地 docs，
远程 GitHub 上没有 Issue/PR 记录，未形成真正的 GitOps 开发工作流。本 PR 把工作流改为
GitHub Issue/PR 为远程权威记录，并补齐开源项目应有的 GitHub Actions CI 与配套文档。

## 变更清单（已实现）

- `docs/gitops/README.md`：工作流改为 Issue -> 分支 -> GitHub PR -> GitHub 合并 -> 本地同步；
  本地 docs（issue/goal/pr）仅为镜像，`docs/review` 存评审报告；
- `AGENTS.md`：文档流（gitOps）约定同步更新；常用命令补充 GitHub Actions CI 说明；
- `.github/workflows/ci.yml`：push/PR 自动执行 generate / build / vet / test / race
  + golden 门禁（`git diff --exit-code`）；
- 开源配套文档：`CONTRIBUTING.md`、`CODE_OF_CONDUCT.md`、`SECURITY.md`、`CHANGELOG.md`、
  Issue/PR 模板；
- README：新增 CI 徽章、文档入口（CONTRIBUTING/CHANGELOG/SECURITY/CODE_OF_CONDUCT）、
  仓库结构与门禁说明更新；
- 本地镜像：`docs/issue/004`、`docs/goal/004`、`docs/pr/004`（关联 GitHub Issue #1 / PR #2）。

## 验收结果（2026-08-22）

- [x] `go generate ./...` 通过，生成文件再生成无 diff（golden 门禁）
- [x] `go build ./...` / `go vet ./...` 通过
- [x] `go test ./...` / `go test -race ./...` 全绿
- [x] `git diff --check` 无空白错误
- [x] GitHub Issue #1 与 PR #2 已在远程创建并合并，`gh issue list` / `gh pr list` 可见
