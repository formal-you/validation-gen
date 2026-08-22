# AGENTS.md

validation-gen：基于 `validate:"..."` struct tag 生成无反射、可审查的 Go 校验代码（gengo/v2 解析 + 自研生成器），
同时保留 `go-playground/validator` runtime 入口作为显式扩展路径。

## 常用命令

```bash
go generate ./...   # 重新生成 zz_generated.validation.go（golden 测试会强制与提交版一致）
go build ./...      # 全量编译
go vet ./...        # 静态检查
go test ./...       # 单元 + 集成 + golden + 静态/runtime 一致性测试
go test -race ./... # 竞态检测
```

> GitHub Actions（`.github/workflows/ci.yml`）在 push/PR 到 `main` 时自动执行上述门禁，
> 其中 `git diff --exit-code` 强制生成文件与提交版一致。

> `pkg/gen` 的 `TestGoldenExampleDTO` 会比对重新生成结果与提交的 `example/dto/zz_generated.validation.go`；
> 修改生成器或 DTO 声明后，必须运行 `go generate ./...` 并把生成的派生文件一起提交，否则门禁失败。

## 生成器工作流（重要）

1. 规则声明是 canonical source：DTO 上的 `validate:"..."` / `default:"..."` struct tag。
2. 生成入口：`cmd/valgen`（`-type` 可重复、`-output` 默认 `zz_generated.validation.go`），
   通过 `k8s.io/gengo/v2` 解析类型/字段/tag；`//go:generate` 见 `example/dto/dto.go`。
3. 流水线：`pkg/parser`（tag 解析、类型检查、规则约束）→ `pkg/ir`（稳定 IR）→ `pkg/gen`（代码生成与原子写入）。
4. 生成代码只依赖标准库 `errors` 与公共包 `errorx`/`check`，无反射；失败时不覆盖已有文件。

## 包结构

| 包 | 职责 |
| --- | --- |
| `cmd/valgen` | CLI 入口（`-type` 可重复，`-output` 默认 zz_generated.validation.go） |
| `pkg/ir` | 规则 IR（parser 与 gen 的稳定契约） |
| `pkg/parser` | tag 解析、类型检查、规则约束 |
| `pkg/errorx` | 公共错误模型：`FieldError` / `CollectFieldErrors` / `FieldName` |
| `pkg/check` | 纯校验 helper（email、UTF-8 长度、oneof 等） |
| `pkg/gen` | 代码生成 + 原子写入 |
| `pkg/runtime` | runtime adapter（validator/v10 `StructCtx`），`runtime.Validate` |
| `example/*` | dto（一致性测试）、http（绑定示例）、grpc（拦截器示例） |

## 关键约定

- **静态白名单（v1）**：`required`、`omitempty`、`min/max`、`len`、`gt/gte/lt/lte`、`oneof`、`email`；
  白名单外（`dive`、跨字段、`required_if`、自定义 validator 等）在生成期直接报错，绝不静默忽略。
- **runtime 兜底**：需要完整能力时显式调用 `runtime.Validate(ctx, v, value)`（传入自己的 `*validator.Validate`
  实例，可注册自定义规则/alias）；不要在生成代码里悄悄扩规则。两条路径错误模型一致。
- **错误模型**：`errorx.FieldError{Field, Code}`，`errorx.CollectFieldErrors` 还原；对外字段路径按
  `json > form > query > header > uri > param` 优先级解析（与 runtime adapter 共用 `errorx.FieldName`）。
- **语义要点**：同一字段只报 tag 顺序第一个失败规则；pointer `omitempty` 只在 nil 时短路；string 长度按
  UTF-8 字符数；`Validate()` 不修改接收对象、不调用 `FillDefaults()`。
- **注释**：导出 API 与非显然的边界（生成期报错、nil/零值语义、default 幂等、事务/生命周期约束）用简洁中文注释。
- **命名**：公共错误包叫 `errorx`；包名避免与局部变量（如 `err`）冲突。
- **文档流（gitOps）**：GitHub Issue/PR 是远程权威记录（`gh issue create` / `gh pr create`，
  合并用 `gh pr merge --rebase`，`main` 保持线性、禁止 merge commit）；本地 `docs/issue`、
  `docs/goal`、`docs/pr` 仅作镜像归档，
  `docs/review` 存评审报告。每个变更先建 Issue（含需求/验收），实现后开 PR 到远程并在 GitHub
  合并，本地 `git pull` 同步；提交信息用 conventional 前缀（`feat:` / `fix:` / `refactor:` /
  `docs:` / `test:`），正文可用中文。详见 `docs/gitops/README.md`。
