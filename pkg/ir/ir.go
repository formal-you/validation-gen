// Package ir 定义校验规则解析后的中间表示（IR）。
//
// IR 是生成器与解析器之间的稳定契约：解析器把 gengo 的 *types.Type 和
// validate/default struct tag 转换成独立的规则 IR，生成器只依赖 IR 输出代码，
// 从而让规则解析、类型检查和代码生成可以分别测试与演进。
package ir

// Kind 表示 v1 支持的标量类型族，与 Go 基础类型一一对应。
type Kind int

const (
	// KindString 对应 string 及其 named type。
	KindString Kind = iota
	// KindBool 对应 bool 及其 named type。
	KindBool
	// KindSigned 对应有符号整数族（int/int8/int16/int32/int64）及其 named type。
	KindSigned
	// KindUnsigned 对应无符号整数族（uint/uint8/uint16/uint32/uint64/uintptr）及其 named type。
	KindUnsigned
	// KindFloat 对应浮点族（float32/float64）及其 named type。
	KindFloat
)

// String 返回 Kind 的可读名称，用于生成期错误信息。
func (k Kind) String() string {
	switch k {
	case KindString:
		return "string"
	case KindBool:
		return "bool"
	case KindSigned:
		return "signed integer"
	case KindUnsigned:
		return "unsigned integer"
	case KindFloat:
		return "float"
	default:
		return "unknown"
	}
}

// Rule 表示 validate tag 中的一条规则。
//
// 语义与固定版本 go-playground/validator/v10 保持一致：
//   - Name  规则名，如 required、min、oneof、email；
//   - Raw   '=' 后的原始参数串（不含引号），如 "3"、"admin user"；
//   - Args  按 validator 的拆分规则解析后的参数：oneof 为候选项列表，
//     其余规则不使用（nil）。
//   - Literal  数值/长度规则参数规范化后的 Go 字面量，如 "3"、"0.5"；
//     生成器直接嵌入该字面量，避免 "0x10"、"017" 等写法导致的编译歧义。
type Rule struct {
	Name    string
	Raw     string
	Args    []string
	Literal string
}

// Default 表示 default tag 解析后的默认值。
type Default struct {
	// Literal 是规范化后的 Go 字面量，直接嵌入 FillDefaults。
	Literal string
	// Raw 保留原始参数串，用于生成期错误上下文。
	Raw string
}

// FieldRules 表示结构体单个字段的校验规则 IR。
type FieldRules struct {
	// GoName 是字段在结构体中的声明名。
	GoName string
	// JSONName 是对外使用的字段路径（错误路径）：
	// 按 err.FieldName 的优先级取 json/form/query/header/uri/param 绑定名，否则取 Go 字段名。
	JSONName string
	// Kind 是字段解析后的标量类型族。
	Kind Kind
	// BitSize 是数值类型的位宽（int/uint 按目标平台取 64）。
	BitSize int
	// IsPointer 表示字段是否为对应标量类型的一级指针。
	IsPointer bool
	// Rules 是 validate tag 中的规则列表，保持声明顺序；
	// omitempty 也保留在列表中，由生成器按其位置处理短路语义。
	Rules []Rule
	// Default 为 nil 表示没有 default tag。
	Default *Default
}

// TypeRules 表示一个结构体类型的完整校验 IR。
type TypeRules struct {
	// Name 是结构体类型名。
	Name string
	// Fields 是参与校验的字段（按 struct 声明顺序，忽略无规则字段）。
	Fields []FieldRules
}
