# 参与贡献

感谢你对 validation-gen 的兴趣。本仓库已开源，所有变更都走 gitOps 工作流
（GitHub Issue/PR 为远程权威记录），详见 [docs/gitops/README.md](docs/gitops/README.md)。

## 开发环境

- Go 1.26+（与 `go.mod` 一致）；
- 可选：`gh` CLI（`gh auth status` 已登录且有 `repo` 权限），用于创建 Issue/PR；
- 无需额外生成步骤：`go generate ./...` 会重新生成 `zz_generated.validation.go`。

## 常用命令（验收门禁）

```bash
go generate ./...   # 重新生成 zz_generated.validation.go
go build ./...      # 全量编译
go vet ./...        # 静态检查
go test ./...       # 单元 + 集成 + golden + 一致性测试
go test -race ./... # 竞态检测
git diff --check    # 无空白错误
```

> GitHub Actions 会在 push/PR 时自动执行同一套门禁；
> `git diff --exit-code` 强制生成文件与提交版一致。

## 变更流程（Issue -> 分支 -> PR -> 合并）

```text
1. 建 Issue      gh issue create（背景 / 需求 / 验收）
2. 建 Goal       在 Issue 中写 DoD，或 docs/goal/NNN-<标题>.md 镜像
3. 开 PR 分支    git checkout -b <type>/<NNN>-<slug>（分支命名规范见 docs/gitops/README.md §1）
4. 实现 + 测试   按验收门禁全绿
5. 推送分支      git push -u origin fix/NNN-<标题>
6. 开 PR         gh pr create --base main --head fix/NNN-<标题>，正文写 Closes #<issue>
7. 合并          gh pr merge <编号> --merge，然后 git checkout main && git pull
```

## 代码约定

- **规则声明**：canonical source 是 DTO 上的 `validate:"..."` / `default:"..."` struct tag，
  不使用 `// +validate:` 或 `+k8s:` 注释 tag；
- **静态白名单（v1）**：`required`、`omitempty`、`min/max`、`len`、`gt/gte/lt/lte`、`oneof`、`email`；
  白名单外的规则在生成期报错，绝不静默忽略；需要完整能力时走 `runtime.Validate`；
- **生成文件**：`zz_generated.validation.go` 是派生产物，修改生成器或 DTO 声明后必须
  `go generate ./...` 并与声明**同一 commit** 提交，否则 golden 门禁失败；
- **错误模型**：`errorx.FieldError{Field, Code}` + `errorx.CollectFieldErrors`，对外字段路径
  按 `json > form > query > header > uri > param` 优先级解析；
- **注释**：导出 API 与非显然的边界（生成期报错、nil/零值语义、default 幂等）用简洁中文注释；
- **提交信息**：conventional 前缀（`feat:` / `fix:` / `refactor:` / `docs:` / `test:`），正文可用中文。
- **分支命名**：`<type>/<NNN>-<kebab-slug>`（type ∈ `feat`/`fix`/`refactor`/`docs`/`test`/`chore`，
  NNN 为关联 Issue 编号），CI 的 `branch-name` 作业在 PR 时强制校验。

## 建议的贡献方向

- 新静态规则（先与 validator/v10 语义对齐，补 parser/生成器/一致性测试）；
- 框架绑定示例（gin、echo、chi 等）与互操作验证；
- 文档案例、性能与可审查性改进。

如果对行为边界有疑问，先开 Issue 对齐再动手，避免 PR 反复。
