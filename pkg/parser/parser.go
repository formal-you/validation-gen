// Package parser 解析结构体字段的 validate/default/json struct tag，
// 进行类型检查与规则合法性检查，产出独立的规则 IR（pkg/ir）。
//
// 输入来自 k8s.io/gengo/v2 的 types.Member / types.Type；
// 输出不依赖 gengo，便于单独测试与演进。
package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	gotypes "go/types"

	"k8s.io/gengo/v2/types"

	"github.com/formal-you/validation-gen/pkg/ir"
)

// supportedRules 是 v1 静态生成支持的规则白名单。
var supportedRules = map[string]bool{
	"required":  true,
	"omitempty": true,
	"min":       true,
	"max":       true,
	"len":       true,
	"gt":        true,
	"gte":       true,
	"lt":        true,
	"lte":       true,
	"oneof":     true,
	"email":     true,
}

// runtimeOnlyRules 是 v1 仅支持 runtime 模式的规则；静态生成遇到时给出明确报错，
// 而不是静默忽略。自定义 validator 与 struct-level validator 不在 tag 层面出现，
// 由 README 说明走 runtime API。
var runtimeOnlyRules = map[string]bool{
	"required_if":      true,
	"required_unless":  true,
	"required_with":    true,
	"required_without": true,
	"eqfield":          true,
	"nefield":          true,
	"gtfield":          true,
	"gtefield":         true,
	"ltfield":          true,
	"ltefield":         true,
	"eqcsfield":        true,
	"necsfield":        true,
	"gtcsfield":        true,
	"gtecsfield":       true,
	"ltcsfield":        true,
	"ltecsfield":       true,
	"dive":             true,
	"keys":             true,
	"endkeys":          true,
	"eq":               true,
	"ne":               true,
	"excluded_if":      true,
	"excluded_unless":  true,
	"excluded_with":    true,
	"excluded_without": true,
}

// ParseType 解析一个结构体类型，返回该类型参与校验的字段 IR。
// 无 validate/default tag 的字段会被忽略；不支持的类型或非法规则返回错误。
func ParseType(pkgName string, t *types.Type) (ir.TypeRules, error) {
	out := ir.TypeRules{Name: t.Name.Name}
	if t.Kind != types.Struct {
		return out, fmt.Errorf("%s.%s: 不是结构体类型（Kind=%s），静态生成仅支持 struct",
			pkgName, t.Name.Name, t.Kind)
	}
	for _, m := range t.Members {
		fr, ok, err := ParseField(pkgName, t.Name.Name, m)
		if err != nil {
			return out, err
		}
		if ok {
			out.Fields = append(out.Fields, fr)
		}
	}
	return out, nil
}

// ParseField 解析单个结构体字段。
// 返回 ok=false 表示该字段不参与校验（无规则且无默认值），调用方应忽略。
func ParseField(pkgName, typeName string, m types.Member) (ir.FieldRules, bool, error) {
	fr := ir.FieldRules{GoName: m.Name}
	if m.Embedded {
		if hasTag(m.Tags, "validate") || hasTag(m.Tags, "default") {
			return fr, false, fmt.Errorf("%s.%s.%s: 嵌入字段不支持校验规则", pkgName, typeName, m.Name)
		}
		return fr, false, nil
	}

	st := structTag(m.Tags)
	validateTag, hasValidate := st.Lookup("validate")
	defaultTag, hasDefault := st.Lookup("default")

	// 未加 validate/default tag 的字段一律忽略（含 json:"-" 场景）。
	if !hasValidate && !hasDefault {
		return fr, false, nil
	}
	if hasValidate && strings.TrimSpace(validateTag) == "" {
		return fr, false, fmt.Errorf("%s.%s.%s: validate tag 为空，请删除或填写规则", pkgName, typeName, m.Name)
	}
	if hasDefault && strings.TrimSpace(defaultTag) == "" {
		return fr, false, fmt.Errorf("%s.%s.%s: default tag 为空，请删除或填写默认值", pkgName, typeName, m.Name)
	}

	jsonName, err := JSONName(m)
	if err != nil {
		return fr, false, fmt.Errorf("%s.%s.%s: %w", pkgName, typeName, m.Name, err)
	}
	fr.JSONName = jsonName

	kind, bitSize, isPointer, err := ResolveKind(m.Type)
	if err != nil {
		return fr, false, fmt.Errorf("%s.%s.%s: %w", pkgName, typeName, m.Name, err)
	}
	fr.Kind, fr.BitSize, fr.IsPointer = kind, bitSize, isPointer

	if hasValidate {
		// json:"-" 的字段不对外，禁止再声明校验规则。
		if jsonName == "-" {
			return fr, false, fmt.Errorf("%s.%s.%s: json:\"-\" 与 validate 规则冲突", pkgName, typeName, m.Name)
		}
		rules, err := ParseRules(validateTag, kind, bitSize, isPointer)
		if err != nil {
			return fr, false, fmt.Errorf("%s.%s.%s: %w", pkgName, typeName, m.Name, err)
		}
		fr.Rules = rules
	}

	if hasDefault {
		d, err := ParseDefault(defaultTag, kind, bitSize, isPointer)
		if err != nil {
			return fr, false, fmt.Errorf("%s.%s.%s: %w", pkgName, typeName, m.Name, err)
		}
		fr.Default = d
	}

	return fr, true, nil
}

// JSONName 返回字段对外使用的 JSON 名称：优先 json tag 的 name 部分，
// 否则使用 Go 字段名。json:"-" 返回 "-"。
func JSONName(m types.Member) (string, error) {
	raw, ok := structTag(m.Tags).Lookup("json")
	if !ok || raw == "" {
		return m.Name, nil
	}
	name := raw
	if i := strings.IndexByte(raw, ','); i >= 0 {
		name = raw[:i]
	}
	return name, nil
}

// ResolveKind 把 gengo 类型解析为 v1 支持的标量类型族。
// 支持：string/bool/有符号整数/无符号整数/浮点，以及它们的 named type 与一级指针。
func ResolveKind(t *types.Type) (kind ir.Kind, bitSize int, isPointer bool, err error) {
	for t != nil {
		switch t.Kind {
		case types.Pointer:
			if isPointer {
				return 0, 0, false, fmt.Errorf("不支持多级指针（%s）", t.Name)
			}
			isPointer = true
			t = t.Elem
		case types.Alias, types.DeclarationOf:
			t = t.Underlying
		default:
			b, ok := t.GoType.Underlying().(*gotypes.Basic)
			if !ok {
				return 0, 0, false, fmt.Errorf("v1 静态生成不支持类型 %q（仅支持标量及其指针）", t.Name)
			}
			switch rk := basicKind(b); rk {
			case reflect.String:
				return ir.KindString, 0, isPointer, nil
			case reflect.Bool:
				return ir.KindBool, 0, isPointer, nil
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return ir.KindSigned, intBits(rk), isPointer, nil
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				return ir.KindUnsigned, uintBits(rk), isPointer, nil
			case reflect.Float32, reflect.Float64:
				bits := 64
				if rk == reflect.Float32 {
					bits = 32
				}
				return ir.KindFloat, bits, isPointer, nil
			default:
				return 0, 0, false, fmt.Errorf("v1 静态生成不支持基础类型 %q", b.Name())
			}
		}
	}
	return 0, 0, false, fmt.Errorf("无法解析类型")
}

// basicKind 把 go/types 的 BasicKind 转换为 reflect.Kind。
func basicKind(b *gotypes.Basic) reflect.Kind {
	switch b.Kind() {
	case gotypes.String:
		return reflect.String
	case gotypes.Bool:
		return reflect.Bool
	case gotypes.Int:
		return reflect.Int
	case gotypes.Int8:
		return reflect.Int8
	case gotypes.Int16:
		return reflect.Int16
	case gotypes.Int32:
		return reflect.Int32
	case gotypes.Int64:
		return reflect.Int64
	case gotypes.Uint:
		return reflect.Uint
	case gotypes.Uint8:
		return reflect.Uint8
	case gotypes.Uint16:
		return reflect.Uint16
	case gotypes.Uint32:
		return reflect.Uint32
	case gotypes.Uint64:
		return reflect.Uint64
	case gotypes.Uintptr:
		return reflect.Uintptr
	case gotypes.Float32:
		return reflect.Float32
	case gotypes.Float64:
		return reflect.Float64
	default:
		return reflect.Invalid
	}
}

// intBits 返回有符号整数的位宽（int 按目标平台取 64）。
func intBits(k reflect.Kind) int {
	switch k {
	case reflect.Int8:
		return 8
	case reflect.Int16:
		return 16
	case reflect.Int32:
		return 32
	default:
		return 64
	}
}

// uintBits 返回无符号整数的位宽（uint/uintptr 按目标平台取 64）。
func uintBits(k reflect.Kind) int {
	switch k {
	case reflect.Uint8:
		return 8
	case reflect.Uint16:
		return 16
	case reflect.Uint32:
		return 32
	default:
		return 64
	}
}

// ParseRules 解析 validate tag 的规则列表并做类型/参数合法性检查。
//
// 语义与 go-playground/validator/v10 保持一致：
//   - 规则按声明顺序保留，omitempty 保留在列表中由生成器处理短路；
//   - min/max/len/gt/gte/lt/lte 参数按 validator 的解析规则（base 0）解析，
//     并规范化为十进制字面量，避免 "0x10"、"017"、" +5" 等写法在生成代码中的歧义；
//   - oneof 参数按 validator 的 '...' 或空白拆分规则拆分候选；
//   - 非法参数、未知规则、重复规则、required+omitempty、类型不兼容均报错。
func ParseRules(tag string, kind ir.Kind, bitSize int, isPointer bool) ([]ir.Rule, error) {
	parts := strings.Split(tag, ",")
	rules := make([]ir.Rule, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	hasRequired, hasOmitempty := false, false

	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("validate tag %q 含空规则", tag)
		}
		name := part
		raw := ""
		if i := strings.IndexByte(part, '='); i >= 0 {
			name, raw = part[:i], part[i+1:]
		}
		if !supportedRules[name] {
			if runtimeOnlyRules[name] {
				return nil, fmt.Errorf("规则 %q 当前仅支持 runtime 模式（静态生成不支持），请使用 pkg/runtime.Validate", name)
			}
			return nil, fmt.Errorf("未知规则 %q（validate tag: %q）", name, tag)
		}
		if seen[name] {
			return nil, fmt.Errorf("规则 %q 重复声明（validate tag: %q）", name, tag)
		}
		seen[name] = true
		if name == "required" {
			hasRequired = true
		}
		if name == "omitempty" {
			hasOmitempty = true
		}

		r := ir.Rule{Name: name, Raw: raw}
		if err := fillRule(&r, kind, bitSize, isPointer, tag); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}

	if hasRequired && hasOmitempty {
		return nil, fmt.Errorf("required 与 omitempty 冲突（validate tag: %q）", tag)
	}
	return rules, nil
}

// fillRule 按规则名填充参数、字面量与类型约束。
func fillRule(r *ir.Rule, kind ir.Kind, bitSize int, isPointer bool, tag string) error {
	switch r.Name {
	case "required", "omitempty":
		if r.Raw != "" {
			return fmt.Errorf("规则 %q 不接受参数（validate tag: %q）", r.Name, tag)
		}
		return nil
	case "email":
		if r.Raw != "" {
			return fmt.Errorf("规则 %q 不接受参数（validate tag: %q）", r.Name, tag)
		}
		if kind != ir.KindString {
			return fmt.Errorf("规则 %q 仅支持 string 类型，当前为 %s", r.Name, kind)
		}
		return nil
	case "min", "max", "len", "gt", "gte", "lt", "lte":
		if r.Raw == "" {
			return fmt.Errorf("规则 %q 缺少参数（validate tag: %q）", r.Name, tag)
		}
		if kind == ir.KindBool {
			return fmt.Errorf("规则 %q 不支持 bool 类型（validator 运行时同样会 panic）", r.Name)
		}
		lit, err := numericLiteral(r.Raw, kind, bitSize)
		if err != nil {
			return fmt.Errorf("规则 %q 参数 %q 非法：%w（validate tag: %q）", r.Name, r.Raw, err, tag)
		}
		r.Literal = lit
		return nil
	case "oneof":
		if r.Raw == "" {
			return fmt.Errorf("规则 oneof 缺少候选项（validate tag: %q）", tag)
		}
		switch kind {
		case ir.KindString, ir.KindSigned, ir.KindUnsigned:
			r.Args = splitOneOf(r.Raw)
			if len(r.Args) == 0 {
				return fmt.Errorf("规则 oneof 候选项为空（validate tag: %q）", tag)
			}
			return nil
		default:
			return fmt.Errorf("规则 oneof 仅支持 string/整数类型，当前为 %s", kind)
		}
	default:
		return fmt.Errorf("未知规则 %q", r.Name)
	}
}

// numericLiteral 把 validator 风格的数值参数解析并规范化为 Go 字面量。
func numericLiteral(raw string, kind ir.Kind, bitSize int) (string, error) {
	switch kind {
	case ir.KindString, ir.KindSigned:
		v, err := strconv.ParseInt(raw, 0, 64)
		if err != nil {
			return "", fmt.Errorf("参数必须为整数：%w", err)
		}
		return strconv.FormatInt(v, 10), nil
	case ir.KindUnsigned:
		v, err := strconv.ParseUint(raw, 0, 64)
		if err != nil {
			return "", fmt.Errorf("参数必须为非负整数：%w", err)
		}
		return strconv.FormatUint(v, 10), nil
	case ir.KindFloat:
		bits := bitSize
		if bits == 0 {
			bits = 64
		}
		v, err := strconv.ParseFloat(raw, bits)
		if err != nil {
			return "", fmt.Errorf("参数必须为浮点数：%w", err)
		}
		return strconv.FormatFloat(v, 'g', -1, bits), nil
	default:
		return "", fmt.Errorf("类型不支持数值规则")
	}
}

// ParseDefault 解析 default tag，产出规范化后的 Go 字面量。
// v1 仅支持基础标量默认值；指针、slice、map、struct、复杂表达式直接报错。
func ParseDefault(tag string, kind ir.Kind, bitSize int, isPointer bool) (*ir.Default, error) {
	if isPointer {
		return nil, fmt.Errorf("default 不支持 pointer 字段（v1 仅支持标量）")
	}
	switch kind {
	case ir.KindString:
		return &ir.Default{Literal: strconv.Quote(tag), Raw: tag}, nil
	case ir.KindBool:
		v, err := strconv.ParseBool(tag)
		if err != nil {
			return nil, fmt.Errorf("default 值 %q 不是合法 bool：%w", tag, err)
		}
		return &ir.Default{Literal: strconv.FormatBool(v), Raw: tag}, nil
	case ir.KindSigned:
		v, err := strconv.ParseInt(tag, 0, bitSize)
		if err != nil {
			return nil, fmt.Errorf("default 值 %q 非法或溢出 int%d：%w", tag, bitSize, err)
		}
		return &ir.Default{Literal: strconv.FormatInt(v, 10), Raw: tag}, nil
	case ir.KindUnsigned:
		v, err := strconv.ParseUint(tag, 0, bitSize)
		if err != nil {
			return nil, fmt.Errorf("default 值 %q 非法或溢出 uint%d：%w", tag, bitSize, err)
		}
		return &ir.Default{Literal: strconv.FormatUint(v, 10), Raw: tag}, nil
	case ir.KindFloat:
		bits := bitSize
		if bits == 0 {
			bits = 64
		}
		v, err := strconv.ParseFloat(tag, bits)
		if err != nil {
			return nil, fmt.Errorf("default 值 %q 非法或溢出 float%d：%w", tag, bits, err)
		}
		return &ir.Default{Literal: strconv.FormatFloat(v, 'g', -1, bits), Raw: tag}, nil
	default:
		return nil, fmt.Errorf("default 仅支持 string/bool/整数/浮点标量")
	}
}

// splitOneOf 按 validator/v10 的拆分规则拆分 oneof 候选项：
// 单引号包裹的 token（可含空格）或非空白 token，并去掉单引号。
func splitOneOf(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		for i < len(s) && isSpace(s[i]) {
			i++
		}
		if i >= len(s) {
			break
		}
		if s[i] == '\'' {
			j := i + 1
			for j < len(s) && s[j] != '\'' {
				j++
			}
			if j >= len(s) {
				// 未闭合引号：按 validator 的 \S+ 语义取剩余部分并去引号。
				out = append(out, strings.ReplaceAll(s[i+1:], "'", ""))
				break
			}
			out = append(out, s[i+1:j])
			i = j + 1
			continue
		}
		j := i
		for j < len(s) && !isSpace(s[j]) {
			j++
		}
		out = append(out, s[i:j])
		i = j
	}
	return out
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func structTag(raw string) reflect.StructTag {
	return reflect.StructTag(raw)
}

func hasTag(raw, key string) bool {
	_, ok := structTag(raw).Lookup(key)
	return ok
}
