# PR 002：HTTP 绑定来源示例与 README 增强

- 状态：merged（本地 gitOps 闭环）
- 分支：`feat/002-http-binding-examples` -> `master`
- 关联 Issue：`docs/issue/002-HTTP绑定示例与README.md`
- 关联 PR：PR 001（绑定来源无关的错误路径）

## 摘要

在 PR 001 的 form/header/uri/query 绑定支持之上补齐示例与文档：

- `example/http` 新增 `WebRequestHandler`：标准库演示
  `POST /users/{id}?page=N` + `X-Token` header + form body 的混合绑定，
  绑定 -> `FillDefaults` -> `Validate` -> 400 结构化错误；
- README 按 `go-observability` 的 README 风格重写
  （痛点区、价值表、快速开始、规则白名单、绑定来源、语义选择、
  错误模型与数据流、仓库结构、提交门禁、发布状态、文档入口）；
- 新增 MIT LICENSE。

## 验收结果（2026-08-22）

- [x] `go generate ./...` 通过
- [x] `go build ./...` / `go vet ./...` 通过
- [x] `go test ./...` / `go test -race ./...` 全绿
- [x] `git diff --check` 无空白错误
- [x] 新增 `TestWebRequestHandler` 覆盖 form 缺失/header 缺失与长度/uri 非法/query 非法/page 默认值
- [x] 生成文件再生成无 diff（golden 测试强制）
