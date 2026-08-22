# 🛡️ validation-gen

<p align="center">
  <strong>让 Go 参数校验不再靠手写</strong><br>
  用 validator 的 <code>validate</code> tag 声明规则，生成无反射、可审查、可静态推理的校验代码。
</p>

<p align="center">
  <a href="https://go.dev/"><img alt="Go 1.26+" src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&logoColor=white"></a>
  <a href="https://github.com/go-playground/validator"><img alt="validator/v10" src="https://img.shields.io/badge/validator-v10.30.1-000000"></a>
  <a href="https://github.com/kubernetes/gengo"><img alt="gengo/v2" src="https://img.shields.io/badge/gengo-v2-326ce5"></a>
  <a href="https://github.com/formal-you/validation-gen/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/formal-you/validation-gen/actions/workflows/ci.yml/badge.svg"></a>
  <a href="LICENSE"><img alt="MIT License" src="https://img.shields.io/badge/License-MIT-blue.svg"></a>
</p>

<p align="center">
  <a href="#user-content-quick-start">🚀 快速开始</a> ·
  <a href="#user-content-rules">🧩 规则白名单</a> ·
  <a href="#user-content-runtime">🔌 runtime 兜底</a> ·
  <a href="#user-content-binding">🔗 绑定来源</a> ·
  <a href="#user-content-semantics">🧠 语义要点</a> ·
  <a href="#user-content-errors">🧨 错误模型</a> ·
  <a href="#user-content-packages">📦 包导航</a> ·
  <a href="#user-content-gates">✅ 门禁</a> ·
  <a href="#user-content-status">🏷️ 发布状态</a> ·
  <a href="#user-content-docs">📚 文档入口</a>
</p>

---

## 🎯 你的校验，可能“能跑”但不可审查

```text
😵 手写重复      每个 handler 都有一段 if/else，规则散落在业务代码里
🔮 反射黑盒      validator 运行时反射，规则与结构体声明分离，出问题难排查
🧨 错误漂移      字段名、错误码各写各的，前端和网关对不上
🤐 静默忽略      不支持的规则悄悄跳过，生产环境才暴露
🔗 绑定脱节      JSON、form、header、uri 各来一套，规则没法复用
```

> [!TIP]
> **validation-gen 是 `validate` tag 与无反射生成代码之间的语义层。**
> 规则声明统一、生成结果可审查、静态语义与 validator/v10 一致；
> 需要 dive、跨字段、自定义 validator 等完整能力时，显式走 runtime API。
> 它不替代 validator，而是把最容易失控的“规则声明”固化成可生成、可测试的 Go 代码。

---

## ✨ 一眼看懂它能做什么

| | 能力 | 你直接得到什么 |
|:---:|:---|:---|
| 🧩 | **规则声明统一** | `validate:"..."` 一处声明，静态与 runtime 共用同一语义 |
| ⚡ | **无反射执行** | 生成的 `Validate()` 是普通 Go 代码，可审查、可断点、零反射 |
| 📜 | **生成结果稳定** | 多次生成无 diff，规则一目了然，AI/同事都能读 |
| 🚫 | **生成期报错** | 不支持规则、非法组合直接失败，不静默忽略 |
| 🔌 | **runtime 可扩展** | dive、跨字段、自定义 validator 显式走 `go-playground/validator` |
| 🔗 | **绑定来源无关** | json/form/query/header/uri/param 一套规则全覆盖 |
| 🪞 | **静态/runtime 一致** | 对照测试逐字段、逐错误码验证语义与 validator/v10 一致 |
| 🎛️ | **默认值** | `FillDefaults()` 幂等填充，`Validate()` 不修改接收对象 |

> [!TIP]
> **30 秒抓住重点**
> 1. 在 DTO 上写 `validate:"..."` / `default:"..."` tag；
> 2. `go generate` 生成无反射的 `Validate()` / `FillDefaults()`；
> 3. 完整能力（dive/跨字段/自定义）显式调用 `runtime.Validate`。

---

## 👀 先看一眼结果

```bash
cd example/dto
go generate ./...
go test ./...
```

`CreateUserRequest` 声明：

```go
type CreateUserRequest struct {
    Name  string  `json:"name" validate:"required,min=3"`
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
    Age   int     `json:"age,omitempty" validate:"gte=0,lte=150"`
    Role  string  `json:"role,omitempty" validate:"oneof=admin user"`
}
```

生成后（`zz_generated.validation.go`）：

```go
func (x *CreateUserRequest) Validate() error {
    var errs []error
    if x.Name == "" {
        errs = append(errs, &errorx.FieldError{Field: "name", Code: "required"})
    } else if check.Runes(string(x.Name)) < 3 {
        errs = append(errs, &errorx.FieldError{Field: "name", Code: "min"})
    }
    // ... 其余字段规则，全部是无反射的普通 if/else
    return errors.Join(errs...)
}
```

HTTP handler 校验失败返回：

```json
{"errors":[{"field":"name","code":"required"},{"field":"email","code":"email"}]}
```

> [!NOTE]
> 生成的代码只依赖标准库 `errors` 与公共包 `errorx`/`check`，不依赖反射；
> 同一字段只报告 tag 顺序中第一个失败的规则，与 validator/v10 完全一致。

---

## 🚀 三分钟跑起来

### 1️⃣ 克隆并生成

```bash
git clone https://github.com/formal-you/validation-gen.git
cd validation-gen
go generate ./...
go test ./...
```

### 2️⃣ 声明 DTO

```go
package dto

//go:generate go run ../../cmd/valgen -type=Settings -output=zz_generated.validation.go

type Settings struct {
    Role string `json:"role" validate:"required" default:"guest"`
    Page int    `json:"page,omitempty" validate:"omitempty,gte=1" default:"1"`
}
```

### 3️⃣ 接入 handler

```go
var req dto.Settings
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, "bad request", http.StatusBadRequest)
    return
}
req.FillDefaults() // 幂等填充默认值，不覆盖非零值
if err := req.Validate(); err != nil {
    writeFieldErrors(w, http.StatusBadRequest, errorx.CollectFieldErrors(err))
    return
}
```
> 只有声明了 `default` 字段的 DTO 才有 `FillDefaults()`；无 default 字段的 DTO 直接调用 `Validate()`。

完整可运行示例见 [`example/http`](example/http/)、[`example/grpc`](example/grpc/)。

> [!IMPORTANT]
> `Validate()` 不会调用 `FillDefaults()`，两者职责独立：
> 校验不修改对象，默认值只填零值字段。

---

## 🧩 静态规则白名单（v1）

| 规则 | 适用类型 | 语义（与 validator/v10 一致） |
| --- | --- | --- |
| `required` | 全部标量 | 零值（指针为 nil）失败 |
| `omitempty` | 全部标量 | 零值（指针为 nil）时跳过剩余规则 |
| `min=N` / `gte=N` | string / 数值 | string 按 UTF-8 字符数 `>= N`；数值按值 `>= N` |
| `max=N` / `lte=N` | string / 数值 | string 按字符数 `<= N`；数值按值 `<= N` |
| `gt=N` / `lt=N` | string / 数值 | 严格大于 / 严格小于 |
| `len=N` | string / 数值 | string 字符数 `== N`；数值值 `== N` |
| `oneof=A B C` | string / 整数 | 必须是候选项之一（整数按十进制字符串比较） |
| `email` | string | 与 validator 的 email 规则一致（net/mail 解析 + 正则） |

支持 string、bool、有符号/无符号整数、浮点，以及它们的 named type 与一级指针。
嵌套 struct、slice/map/array、跨字段与 `dive` 等静态暂不支持，声明即生成期报错。

**仅 runtime 模式**：`required_if`、`required_unless`、`required_with`、`required_without`、
`eqfield`、`nefield`、`gtfield`、`gtefield`、`ltfield`、`ltefield`、`dive`、`keys`、`endkeys`、
自定义 validator、struct-level validator、依赖数据库/外部服务的规则。

---

## 🔌 runtime 兜底：静态规则不够用时显式使用 go-playground/validator

v1 静态白名单是有边界的。需要 `dive`、跨字段（`eqfield`/`gtfield` 等）、条件规则
（`required_if`/`required_with` 等）、自定义 validator、struct-level validator、alias
或依赖外部服务的规则时，**生成期直接报错、绝不静默忽略**——这种场景显式走 runtime API：

```go
import (
    "context"
    "strings"

    "github.com/go-playground/validator/v10"

    "github.com/formal-you/validation-gen/pkg/errorx"
    "github.com/formal-you/validation-gen/pkg/runtime"
)

v := validator.New()
_ = v.RegisterValidation("is-upper", func(fl validator.FieldLevel) bool {
    s := fl.Field().String()
    return s != "" && s == strings.ToUpper(s)
})

type upperReq struct {
    Name string `json:"name" validate:"required,is-upper"`
}

err := runtime.Validate(context.Background(), v, &upperReq{Name: "abc"})
if err != nil {
    // 结构化错误与静态校验同构：字段路径已映射为 JSON 名称
    fields := errorx.CollectFieldErrors(err) // [{Field: "name", Code: "is-upper"}]
}
```

`runtime.Validate(ctx, v, value)` 三个要点：

| 参数/行为 | 说明 |
| --- | --- |
| `v == nil` | 创建一次性实例，每次调用独立，不共享全局状态 |
| 传入自定义实例 | 调用方已有的 `RegisterValidation` / `RegisterAlias` / `StructCtx` 配置保持不变，本包不修改传入实例 |
| 返回值 | 校验失败为 `errors.Join` 聚合的 `*FieldError`，用 `errorx.CollectFieldErrors` 还原；非校验错误（如 `InvalidValidationError`）原样返回 |

> [!IMPORTANT]
> 静态 `Validate()` 与 `runtime.Validate` 是两条**显式**路径：白名单内用生成代码（无反射、可审查），
> 白名单外统一走 runtime，绝不在生成代码里悄悄扩规则。两者错误模型与字段路径语义一致，
> 需要切换时只改一行调用。

---

## 🔗 绑定来源无关

字段来自 JSON、form 表单、query、header 或 URI 路径参数，都可以使用同样的规则：
生成器只关心规则与类型，不关心绑定机制。

```go
type WebRequest struct {
    Username string `form:"username" validate:"required,min=3"`
    Token    string `header:"X-Token" validate:"required,len=10"`
    ID       int64  `uri:"id" validate:"required,gt=0"`
    Page     int    `form:"page" validate:"omitempty,gte=1" default:"1"`
}
```

标准库 handler（[`example/http`](example/http/)）：

```go
// POST /users/{id}?page=N + X-Token header + form body
var req dto.WebRequest
req.Username = r.PostFormValue("username")
req.Token = r.Header.Get("X-Token")
req.ID, _ = strconv.ParseInt(r.PathValue("id"), 10, 64)
req.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
req.FillDefaults()
if err := req.Validate(); err != nil {
    writeFieldErrors(w, http.StatusBadRequest, errorx.CollectFieldErrors(err))
    return
}
```

gin/echo 等框架用对应绑定 tag（`form`/`header`/`uri`/`query`/`param`）做同样的事，
`validate`/`default` 声明无需改动。

---

## 🧠 五个刻意做出的语义选择

### 🧱 1. 同一字段只报第一个失败规则

与 validator/v10 一致：字段按 tag 顺序短路，`required` 失败自然跳过其余规则。
这让静态与 runtime 的对照测试可以逐字段、逐错误码精确相等。

### 🪞 2. pointer 的 `omitempty` 只在 nil 时短路

`*string` + `omitempty,min=3`：`nil` -> 跳过；`ptr("")` -> 执行 min 并失败；
`ptr("a")` -> 执行 min 并失败。非 nil 指针（即使指向零值）会执行后续规则。

### 📏 3. string 长度按 UTF-8 字符数

`你好` 计 2 个字符，不是 6 个字节；与 validator 的 `utf8.RuneCountInString` 一致。

### 🚫 4. 生成期报错优于运行时 panic

空 validate tag、未知/重复规则、`required + omitempty`、`json:"-"` + validate、
参数非法/溢出、类型不兼容（如 bool + min）、default 非法/溢出、default 用于指针，
全部在生成期失败且不写文件，绝不静默忽略。

### 🔗 5. 错误路径按绑定 tag 解析

优先级 `json` > `form` > `query` > `header` > `uri` > `param` > Go 字段名，
静态生成与 runtime adapter 共用 `err.FieldName`，保证两侧错误路径一致
（例如 `header:"X-Token"` 的字段错误路径是 `X-Token`）。

---

## 🧨 错误模型与数据流

```go
type FieldError struct {
    Field string `json:"field"` // 对外字段路径，如 "name"、"profile.email"、"items[0].sku"
    Code  string `json:"code"`  // 失败规则名，如 "required"、"min"
}

func CollectFieldErrors(err error) []FieldError
```

```text
DTO (validate/default tag)
  → go generate (cmd/valgen + gengo 解析)
  → zz_generated.validation.go（Validate / FillDefaults，无反射）
  → HTTP: bind → (FillDefaults) → Validate → 400 结构化错误
  → gRPC: request → Validate → status.Error(codes.InvalidArgument, ...)
  → 需要完整能力 → runtime.Validate(ctx, v, req)（validator/v10）
```

---

## 📦 仓库里有什么

```text
validation-gen/
├── 🤖 .github/workflows/ci.yml  # GitHub Actions：generate/build/vet/test/race 门禁
├── 🛠️ cmd/valgen/             # CLI：-type 可重复，-output 默认 zz_generated.validation.go
├── 🧩 pkg/ir/                 # 规则 IR（解析器与生成器的稳定契约）
├── 🔎 pkg/parser/             # tag 解析、类型检查、规则约束
├── 🧨 pkg/errorx/             # FieldError / CollectFieldErrors / FieldName
├── 🧰 pkg/check/              # 纯校验 helper（email、UTF-8 长度、oneof）
├── ⚙️ pkg/gen/                # 代码生成与原子写入（失败不覆盖已有文件）
├── 🔌 pkg/runtime/            # runtime adapter（validator/v10 StructCtx）
├── 🧪 example/dto/            # DTO + go:generate + 静态/runtime 一致性测试
├── 🧪 example/http/           # JSON/form/header/uri 绑定 + 400 结构化错误
├── 🧪 example/grpc/           # InvalidArgument + unary 拦截器
└── 📚 docs/                   # spec / issue / goal / pr / review / gitops
```

---

## ✅ 每次提交都过哪些门禁

```text
go generate ./...   # 生成文件与声明保持同步
go build ./...      # 全量编译
go vet ./...        # 静态检查
go test ./...       # 单元 + 集成 + golden + 一致性测试
go test -race ./... # 竞态检测
```

并确认：生成文件重新生成后无 diff（golden 测试强制）；静态/runtime 对照测试通过；
不支持规则不会静默忽略；生成代码无反射；HTTP/gRPC 示例通过。

> GitHub Actions（[`.github/workflows/ci.yml`](.github/workflows/ci.yml)）在 push/PR 到 `main`
> 时自动执行同一套门禁，`git diff --exit-code` 强制生成文件与提交版一致。

---

## 🏷️ 当前发布状态

> [!WARNING]
> 仓库处于**公开预览阶段**，公共 API 在首个 tag 前仍可能调整。生产接入请固定版本，并先完成兼容性验证。

首次正式发布将完成：

- 📝 持续维护 [`CHANGELOG.md`](CHANGELOG.md)；
- 🏷️ 创建并推送 `v0.1.0`；
- 📦 在全新模块中验证带版本号的 `go get`；
- 🔎 检查 pkg.go.dev 索引。

---

## 📚 文档入口

| | 文档 | 适合什么时候看 |
|:---:|:---|:---|
| 📐 | [规格说明](docs/spec/validation-gen.md) | 完整需求、测试计划与验收门槛 |
| 🧭 | [实现计划](plan.md) | 了解目录、语义与实施顺序 |
| 🧪 | [example/dto](example/dto/dto.go) | 看规则声明与一致性测试 |
| 🧪 | [example/http](example/http/main.go) | JSON/form/header/uri 绑定示例 |
| 🧪 | [example/grpc](example/grpc/grpc.go) | gRPC 校验与拦截器示例 |
| 🗂️ | [issue / goal / pr](docs/) | gitOps 变更记录 |
| 🤝 | [参与贡献](CONTRIBUTING.md) | 开发环境、变更流程与代码约定 |
| 📝 | [变更记录](CHANGELOG.md) | 版本间的重要变更 |
| 🛡️ | [安全策略](SECURITY.md) | 报告安全漏洞 |
| 💬 | [行为准则](CODE_OF_CONDUCT.md) | 社区行为规范 |

---

## 🤝 参与贡献

欢迎提交 Issue 和 PR，完整指南见 [CONTRIBUTING.md](CONTRIBUTING.md)。
新规则、框架适配、文档案例与互操作验证都很有价值，
但请保持核心包的通用边界，并为行为变化补充测试。

```text
发现问题 🐛 → 提交 Issue → 对齐行为边界 → 补测试与实现 → PR + 门禁 ✅
```

---

## 📄 License

[MIT License](LICENSE)，可自由使用、修改和分发。

---

<p align="center">
  <strong>🛡️ 让每一条规则都有稳定语义，让每一次校验都有可审查的实现。</strong><br>
  <sub>validation-gen · static validation code generator for Go</sub>
</p>
