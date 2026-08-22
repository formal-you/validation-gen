# Issue 002：HTTP 绑定来源示例与 README 增强

- 状态：open
- 创建：2026-08-22
- 关联：PR 001（绑定来源无关的错误路径）、PR 002

## 背景

PR 001 已支持 form/header/uri/query 等绑定来源字段的校验与错误路径，
但 `example/http` 只演示了 JSON 绑定；README 也是偏平铺直叙的用法说明，
吸引力与信息层级不足。需要补齐示例，并按 `go-observability` 的 README 风格重写。

## 需求

1. `example/http` 增加 form/header/query/uri 混合绑定示例：
   - `POST /users/{id}?page=N` + `X-Token` header + form body；
   - 绑定 -> `FillDefaults` -> `Validate` -> 400 结构化错误；
   - 覆盖未提供字段、header 缺失、uri 参数非法、form 默认值等场景。
2. 按 [go-observability README](https://github.com/formal-you/go-observability) 风格重写 README：
   - 标题/标语/徽章/导航；
   - 痛点区与定位 callout；
   - 「一眼看懂它能做什么」价值表；
   - 「先看一眼结果」与「三分钟跑起来」；
   - 静态规则白名单、绑定来源无关、语义要点、错误模型与数据流；
   - 仓库结构、提交门禁、发布状态、文档入口、贡献与 License。
3. 所有示例与 README 内容必须与代码行为一致。

## 验收

- `go generate ./...`、`go build ./...`、`go vet ./...`、`go test ./...`、`go test -race ./...` 全绿；
- 生成文件再生成无 diff；
- 新增 HTTP 绑定示例的 httptest 通过；
- README 中每个代码片段与仓库实际行为一致。
