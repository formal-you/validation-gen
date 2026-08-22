# gitOps 工作流（GitHub Issue/PR 为远程权威记录）

本仓库所有变更都走 git 记录，且 **Issue 与 PR 的权威记录在 GitHub（远程可见）**：
用 `gh issue create` / `gh pr create` 创建远程记录，本地 `docs/issue`、`docs/goal`、`docs/pr`
只作为镜像/归档，`docs/review` 存放评审报告。变更可被团队查看、评审、追溯。

---

## 1. 工作流（Issue -> 分支 -> GitHub PR -> 合并）

```text
1. 建 Issue      gh issue create（或 GitHub 网页）  远程 Issue：背景 / 需求 / 验收（权威记录）
                 （可选）镜像 docs/issue/NNN-<标题>.md
2. 建 Goal       在 Issue 中写目标 / 范围 / DoD，或镜像 docs/goal/NNN-<标题>.md
3. 开 PR 分支    git checkout -b <type>/<NNN>-<slug>（见下方分支命名规范）
4. 实现 + 测试   按「验收门禁」全绿（golden 测试强制生成文件与提交版一致）
5. 推送分支      git push -u origin feat/NNN-<标题>
6. 开 PR         gh pr create --base main --head feat/NNN-<标题>  远程 PR：摘要 / 变更清单 / 评审
                 （可选）镜像 docs/pr/NNN-<标题>.md
7. 合并          gh pr merge <编号> --rebase（GitHub 上 rebase 合并，main 保持线性）
                 git checkout main && git pull（本地同步）
8.（可选）评审    docs/review/<日期>-<主题>.md，按报告修复后再合并
```

关键点：

- **Issue 与 PR 必须在 GitHub 上创建**，这是远程可见的权威记录；本地 docs 只是镜像，不能替代；
- **`main` 保持线性**：合并一律用 `gh pr merge --rebase`（保留每个提交）或 `--squash`（合并为单个提交），
  **禁止 `--no-ff` / `--merge` 产生 merge commit**；CI 会校验 `main` 无 merge commit；
- 合并发生在 **GitHub**（`gh pr merge`），不是本地 `git merge`；本地随时 `git pull` 保持同步；
- **分支命名规范**：`<type>/<NNN>-<kebab-slug>` —— `type` ∈
  `feat` / `fix` / `refactor` / `docs` / `test` / `chore`（conventional 前缀），
  `NNN` 为关联 Issue 的本地编号，`kebab-slug` 为小写英文短标题；CI 的 `branch-name`
  作业会在 PR 时强制校验，例如 `feat/002-http-binding-examples`、`fix/003-review-fixes`；
- 提交信息用 conventional 前缀：`feat:` / `fix:` / `refactor:` / `docs:` / `test:`；
- 生成文件（`zz_generated.validation.go`）是派生产物，必须与 DTO 声明**同一 commit** 提交，
  否则 `TestGoldenExampleDTO` 会失败；
- PR 正文写 `Closes #<issue>` 可在合并时自动关闭关联 Issue。

## 2. 远程记录操作（gh CLI）

前置：`gh auth status` 显示已登录且有 `repo` 权限（或直接用 GitHub 网页操作）。

```bash
gh issue create --title "<标题>" --body "<背景/需求/验收>"
gh pr create --base main --head <分支> --title "<标题>" --body "<摘要/变更清单>"
gh pr merge <编号> --rebase        # GitHub 上 rebase 合并，main 保持线性（保留每个提交）
gh pr merge <编号> --squash        # 或 squash 合并为单个提交（同样线性）
gh issue list / gh pr list          # 查看远程记录
```

## 3. 目录与命名约定

```text
validation-gen/
├── .github/workflows/ci.yml        # GitHub Actions：自动执行验收门禁
├── .github/ISSUE_TEMPLATE.md       # Issue 模板（背景 / 需求 / 验收）
├── .github/PULL_REQUEST_TEMPLATE.md# PR 模板（摘要 / 变更清单 / 验收）
├── CONTRIBUTING.md                 # 参与贡献指南（开发环境 / 变更流程 / 代码约定）
├── CODE_OF_CONDUCT.md              # 行为准则
├── SECURITY.md                     # 安全策略与漏洞报告
├── CHANGELOG.md                    # 变更记录（Keep a Changelog）
└── docs/
    ├── issue/NNN-<中文短横线标题>.md    # 本地镜像（可选）：GitHub Issue 的归档
    ├── goal/NNN-<中文短横线标题>.md     # 本地镜像（可选）：目标 / 范围 / DoD
    ├── pr/NNN-<中文短横线标题>.md       # 本地镜像（可选）：GitHub PR 的归档
    ├── review/<日期>-<主题>.md          # 代码评审报告（Standards + Spec 双轴）
    ├── spec/validation-gen.md           # 权威规格
    └── gitops/README.md                 # 本文档
```

编号从仓库内最大编号递增（当前 001-004）。本地 `NNN` 是稳定序列，与 GitHub 平台分配的
Issue/PR 编号解耦；镜像文档中标注对应的 GitHub Issue/PR 编号与 URL 以便回溯。

## 4. 验收门禁

```bash
go generate ./...   # 重新生成 zz_generated.validation.go（golden 测试强制与提交版一致）
go build ./...      # 全量编译
go vet ./...        # 静态检查
go test ./...       # 单元 + 集成 + golden + 静态/runtime 一致性测试
go test -race ./... # 竞态检测
git diff --check    # 无空白错误
```

并确认：生成文件重新生成后无 diff；静态/runtime 对照测试通过；
不支持规则不会被静默忽略；生成代码无反射；HTTP/gRPC 示例通过。

## 5. 发布（后续，打 tag）

```bash
git tag v0.1.0
git push origin v0.1.0
```

发布前检查：整理 `CHANGELOG.md`；在全新模块里验证带版本号的 `go get`；
检查 pkg.go.dev 索引与外部文档链接。

## 6. 当前状态（2026-08-22）

- PR 001-003：本地闭环合并并推送（当时 Issue/PR 仅记录在本地 docs，远程无记录）；
- **自 004 起**：按本文档执行，GitHub Issue/PR 为远程权威记录；
- 已落地：Issue #1 与 PR #2（004 GitOps 工作流落地）已创建并在 GitHub 合并到远程 `main`；
- 已就位：GitHub Actions CI（`.github/workflows/ci.yml`）自动执行验收门禁与
  「main 无 merge commit」线性校验，开源配套文档（CONTRIBUTING / CODE_OF_CONDUCT / SECURITY / CHANGELOG）已入库；
- **已线性化**：PR 001-004 的历史已重写为线性（移除全部 merge commit，内容不变），
  此后 `main` 保持线性；
- 待办：按需打 tag 发布（见 §5）。
