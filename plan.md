# 静态校验代码生成器（gengo + validate tag）实现计划

> 2026-08-22 更新：本计划以 `docs/spec/validation-gen.md` 为权威规格。
> 规则声明统一使用 `go-playground/validator` 的 `validate:"..."` struct tag
> （取代早期 `+validate:` 注释 tag 草案），默认值使用 `default:"..."`。

## 摘要

在 `github.com/formal-you/validation-gen` 实现一个基于 `k8s.io/gengo/v2` 的自研代码生成器：

- 输入 `validate:"..."` / `default:"..."` struct tag；
- 生成无反射、可审查的 `Validate() error` 与 `FillDefaults()`；
- 架构仿 k8s validation-gen（validator 注册表 + 规则 IR + 生成器 + 测试组织），
  但不引入 validation-gen 本体，不耦合 k8s API 语义；
- 静态不支持规则在生成期报错，不静默忽略；完整能力走显式 runtime API。

## 目录与依赖

```text
cmd/valgen/          CLI（-type 可重复，-output 默认 zz_generated.validation.go）
pkg/ir/              规则 IR（Rule / FieldRules / Default）
pkg/parser/          validate/default/json tag 解析、类型检查、规则约束
pkg/check/           纯校验 helper（email regex、UTF-8 长度）
pkg/errorx/          公共错误模型（FieldError / CollectFieldErrors）
pkg/gen/             代码生成（渲染 Validate/FillDefaults、gofmt、原子写入）
pkg/runtime/         runtime adapter（validator/v10 StructCtx + JSON 字段名映射）
example/dto/         DTO 示例 + go:generate + 静态/运行时一致性测试
example/http/        HTTP JSON bind -> FillDefaults -> Validate -> 400 结构化错误
example/grpc/        gRPC Validate -> status.Error(codes.InvalidArgument, ...)
docs/                spec / issue / goal / pr
```

直接依赖固定版本：`k8s.io/gengo/v2`、`go-playground/validator/v10`（不用 `@latest`）。

## 关键语义（与 validator/v10 一致）

- 规则：required / omitempty / min / max / len / gt / gte / lt / lte / oneof / email。
- string 长度按 UTF-8 字符数；pointer：nil 表示未提供，非 nil 才执行规则；
  `omitempty` 短路；named type 按其底层类型处理；数值边界按值比较。
- `required` 失败短路同字段其余规则；`required+omitempty` 生成期报错。
- `json:",omitempty"` 只影响序列化，不参与校验；`json:"-"` + validate 生成期报错。
- 默认值仅基础标量，生成期解析，非法/溢出/类型不匹配直接报错。

## 实施顺序（对应 Issue 1-5）

1. pkg/ir + pkg/parser + 单元测试；
2. pkg/errorx + pkg/check；
3. pkg/gen + cmd/valgen + example/dto（golden + 一致性测试）；
4. pkg/runtime + 测试；
5. example/http、example/grpc、README、docs。

## 验收门槛

```bash
go generate ./...
go test ./...
go test -race ./...
go vet ./...
go build ./...
git diff --check
```

并确认：生成文件再生成无 diff；静态/runtime 对照通过；不支持规则不静默忽略；
生成代码无反射；HTTP/gRPC 示例通过；README 与 tag 行为一致。
