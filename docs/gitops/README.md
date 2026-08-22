# gitOps 工作流（本地闭环 + 远程推送）

本仓库所有变更都走 git 记录：**本地先闭环（issue → goal → PR 分支 → 实现与测试 → PR 文档 → 合并），
远程就绪后再推送**。文档与代码同仓，变更可追溯、可回看。

---

## 1. 本地闭环流程

```text
1. 建 Issue      docs/issue/NNN-<标题>.md      提交到 master（作为 base）
2. 建 Goal       docs/goal/NNN-<标题>.md      目标 / 范围 / 完成定义（DoD）/ 拆解
3. 开 PR 分支    git checkout -b feat/NNN-<标题>（或 fix:/refactor:/docs:）
4. 实现 + 测试   按「验收门禁」全绿（golden 测试强制生成文件与提交版一致）
5. 写 PR 文档    docs/pr/NNN-<标题>.md        摘要 / 变更清单 / 验收结果勾选
6. 合并          git checkout master && git merge --no-ff feat/NNN-<标题> -m "Merge PR NNN: <标题>"
7.（可选）评审    docs/review/<日期>-<主题>.md，按报告修复后再合并
```

关键点：

- **Issue 在 master（base）上建**，实现全部在 PR 分支上，合并后才回到 master；
- 合并用 `--no-ff`，保留 PR 分支历史，方便追溯每次变更的来龙去脉；
- 分支名与提交信息用 conventional 前缀：`feat:` / `fix:` / `refactor:` / `docs:` / `test:`；
- 生成文件（`zz_generated.validation.go`）是派生产物，必须与 DTO 声明**同一 commit** 提交，
  否则 `TestGoldenExampleDTO` 会失败。

## 2. 目录与命名约定

```text
docs/
├── issue/NNN-<中文短横线标题>.md    # 背景 / 需求 / 验收
├── goal/NNN-<中文短横线标题>.md     # 目标 / 范围 / DoD / 拆解
├── pr/NNN-<中文短横线标题>.md       # 摘要 / 变更清单 / 验收结果
├── review/<日期>-<主题>.md          # 代码评审报告（Standards + Spec 双轴）
├── spec/validation-gen.md           # 权威规格
└── gitops/README.md                 # 本文档
```

编号从仓库内最大编号递增（当前 001、002）。

## 3. 验收门禁

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

## 4. 远程推送（首次）

前置条件：

1. 远程仓库已建立（如 https://github.com/formal-you/validation-gen.git）；
2. 本机凭据可用（**有仓库权限的 HTTPS PAT** 或 **已注册到 GitHub 的 SSH key**，二选一）。

```bash
# 1) 添加远程（只执行一次）
git remote add origin https://github.com/formal-you/validation-gen.git

# 2) 本地默认分支 master 重命名为 main（只执行一次）
git branch -M main

# 3) 首次推送并设置 upstream
git push -u origin main
```

推送成功后，后续提交只需 `git push`（main 已绑定 origin/main）。

> 若远程**已经有代码**（例如已从其他机器推送过）：先 `git fetch origin` 确认本地与远程的关系。
> 本地领先（fast-forward 可推）时直接 `git push` 即可；本地落后时先 `git pull --rebase` 再推；
> 若两边历史不相关（如远程有 GitHub UI 初始提交），先对齐历史（`git merge --allow-unrelated-histories` 或 rebase）再推送。

### 4.1 凭据问题排查（本次实测遇到）

- **HTTPS 403**：`Permission to formal-you/validation-gen.git denied to FormalYou`
  —— 本机 `~/.git-credentials` 里存的是 OAuth token（`gho_…`），归属账号 `FormalYou`，对该仓库无权限；
- **SSH 拒绝**：`git@github.com: Permission denied (publickey)`
  —— `~/.ssh/id_ed25519.pub` 未注册到 GitHub。

两条修复路径二选一：

| 路径 | 操作 |
| --- | --- |
| **SSH（推荐）** | 把 `~/.ssh/id_ed25519.pub` 加到 `formal-you` 账号的 GitHub → Settings → SSH and GPG keys，然后 `git remote set-url origin git@github.com:formal-you/validation-gen.git`，再 `git push -u origin main` |
| **HTTPS** | 给 `formal-you` 账号生成带 `repo` 权限的 PAT，更新本机凭据后重试 `git push -u origin main` |

验证：`git remote -v` 看远程 URL；`git ls-remote origin` 能看到 refs 即连接成功。

## 5. 发布（后续，打 tag）

```bash
git tag v0.1.0
git push origin v0.1.0
```

发布前检查：整理 `CHANGELOG.md`；在全新模块里验证带版本号的 `go get`；
检查 pkg.go.dev 索引与外部文档链接。

## 6. 当前状态（2026-08-22）

- 已合并：PR 001（静态校验生成器）、PR 002（HTTP 绑定示例与 README 增强）；
- 已就位：`origin` 已添加、本地分支已重命名为 `main`（`git branch -M main`）；
- **已推送**：远程 `main` 已存在（`0e74624`，含 valerr→errorx 重构）；
- **待同步**：本地 `main` 领先 1 个提交（`a4370e4`，gitOps 文档），执行 `git push -u origin main` 同步；
- 未跟踪：`AGENTS.md`、`docs/review/2026-08-22-代码评审报告.md`（评审建议尽快提交 AGENTS.md）。
