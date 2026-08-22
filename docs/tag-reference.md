# validate / default struct tag 参考（v1）

> 本文档是 `validate:"..."` / `default:"..."` 写法的权威速查，与 `pkg/parser` 的实现一一对应。
> 完整语义与测试计划见 [docs/spec/validation-gen.md](spec/validation-gen.md)。

## 1. 一句话语法

规则写在结构体字段的 `validate:"..."` tag 里：

- **规则之间用英文逗号 `,` 分隔**；
- **规则名与参数用 `=` 连接**；
- **`oneof` 的多候选项用空格分隔**；
- **`default:"..."` 是独立 tag**（不属于 validate 规则）。

```go
type CreateUserRequest struct {
    Name  string  `json:"name" validate:"required,min=3,max=20"`
    Role  string  `json:"role,omitempty" validate:"omitempty,oneof=admin user guest"`
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
    Page  int     `json:"page,omitempty" validate:"omitempty,gte=1" default:"1"`
}
```

## 2. 分隔符速查表（重点）

| 位置 | 分隔符 | 示例 | 说明 |
| --- | --- | --- | --- |
| 规则之间 | 逗号 `,` | `required,min=3,max=20` | 多条规则按声明顺序执行 |
| 规则名与参数 | `=` | `min=3` | 无参数规则不要写 `=` |
| `oneof` 候选项之间 | 空格 | `oneof=admin user guest` | 候选项用空白分开 |
| `oneof` 含空格的候选项 | 单引号 `'...'` | `oneof='in progress' done` | 与 validator/v10 一致 |

⚠️ 常见误区：

- **不要用空格分隔规则**：`validate:"required min=3"` 会被当成一个名为 `required min=3` 的
  未知规则 → 生成期报错；
- **不要用逗号分隔 `oneof` 候选项**：`oneof=admin,user` 会把 `admin,user` 当成单个候选项，
  与预期不符（应写成 `oneof=admin user`）；
- 规则之间不要有多余逗号：`required,min=3,` 的末尾逗号会产生空规则 → 生成期报错。

## 3. 支持规则（v1 静态白名单）

| 规则 | 参数 | 适用类型 | 语义 | 示例 |
| --- | --- | --- | --- | --- |
| `required` | 无 | 所有类型 | 零值判失败：string `""`、数值 `0`、bool `false`、指针 `nil` | `validate:"required"` |
| `omitempty` | 无 | 所有类型 | 零值 / 指针 `nil` 时跳过其余规则（短路） | `validate:"omitempty,email"` |
| `min` / `max` | 数值 | string / 整数 / 浮点 | string 按 UTF-8 字符数；数值按值 | `min=3` `max=20` |
| `len` | 整数 | string / 整数 / 浮点 | string 按 UTF-8 字符数精确匹配 | `len=10` |
| `gt` / `gte` / `lt` / `lte` | 数值 | string / 整数 / 浮点 | string 按 UTF-8 字符数；数值按值 | `gte=0,lte=150` |
| `oneof` | 候选项（空格分隔） | string / 整数 | 精确匹配候选项之一；整数按十进制字符串比较（`01` ≠ `1`） | `oneof=admin user` |
| `email` | 无 | 仅 string | 合法邮箱（正则与 validator/v10 一致，要求域名含 TLD） | `validate:"email"` |

注意：

- bool 不支持数值规则（`min`/`max`/`len`/`gt`/`gte`/`lt`/`lte` 用于 bool 生成期报错）；
- `oneof` 不支持浮点；
- `min`/`max`/`gt`/`gte`/`lt`/`lte` 对 string 一律按 **UTF-8 字符数**（`check.Runes`），不是字节数。

## 4. 支持的字段类型

- `string`、`bool`、有符号整数（`int`/`int8`/…/`int64`）、无符号整数（`uint`/`uint8`/…/`uint64`）、
  浮点（`float32`/`float64`）；
- 以上类型的 named type（如 `type Username string`）；
- 以上类型的一级指针（如 `*string`）。

不支持：slice / map / struct 嵌套、多级指针、跨字段引用、`dive` 等集合规则。

## 5. 默认值 `default` tag

| 类型 | 写法 | 示例 |
| --- | --- | --- |
| string | 直接写值（生成时自动加引号） | `default:"guest"` |
| bool | `true` / `false` | `default:"true"` |
| 整数 | 十进制 / 十六进制 / 八进制均可，规范化为十进制 | `default:"3"` |
| 浮点 | 数字字面量 | `default:"0.5"` |

- 不支持 pointer / slice / map / struct / 复杂表达式；
- `FillDefaults()` 只填充零值字段、幂等、不覆盖非零值；
- `Validate()` 不调用 `FillDefaults()`，两者独立使用。

## 6. 生成期报错（常见错误写法）

| 写法 | 原因 |
| --- | --- |
| `validate:"required,omitempty"` | `required` 与 `omitempty` 冲突 |
| `validate:"min=3,min=5"` | 重复声明同名规则 |
| `validate:"required min=3"` | 用空格分隔规则 → 未知规则 |
| `validate:"min=3,"` | 空规则（多余逗号） |
| `validate:"oneof="` | `oneof` 缺少候选项 |
| `validate:"email=foo"` | `email` 不接受参数 |
| `validate:"dive"`、`eqfield`、`required_if` 等 | runtime-only 规则，静态生成不支持，请用 `runtime.Validate` |
| `json:"-"` + `validate:"..."` | 对外隐藏字段与校验规则冲突 |
| bool 字段 + `min=3` 等 | bool 不支持数值规则 |

## 7. 语义要点

1. 同一字段的规则按 tag 顺序**全部执行**，但**只报告第一个失败规则**（与 validator/v10 一致）；
2. pointer 的 `omitempty` 只在 **nil** 时短路；非 nil 指针（即使指向零值）继续执行规则；
3. string 长度一律按 **UTF-8 字符数**，不是字节数；
4. `json:",omitempty"` 只影响 JSON 序列化，**不参与校验语义**；presence 由 `validate:"omitempty"` 表达；
5. `Validate()` 不修改接收对象、不调用 `FillDefaults()`。

## 8. 完整示例

- 声明：`example/dto/dto.go`（含 named type、指针、oneof、email、default）；
- 生成结果：`example/dto/zz_generated.validation.go`（可直接审查）；
- 集成：`example/http/main.go`（JSON/form/header/uri 绑定）、`example/grpc/grpc.go`（拦截器）。
