// Package errorx 提供校验错误模型：FieldError 与 CollectFieldErrors。
//
// 该包是生成代码与 runtime adapter 共同依赖的公共错误包，
// 生成代码只依赖 errorx 与纯校验 helper（pkg/check），不依赖反射。
package errorx

import (
	"fmt"
	"reflect"
	"strings"
)

// FieldError 表示单个字段的单个校验失败。
//
// Field 是对外字段路径，优先使用 JSON 名称（如 "name"、"profile.email"）；
// Code 是失败规则名（如 "required"、"min"），与 validator/v10 的 tag 一致。
//
// 错误字符串仅用于日志和调试，不作为 HTTP/gRPC 协议；
// 协议层应使用 CollectFieldErrors 获取结构化错误。
type FieldError struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}

// Error 实现 error 接口，返回用于日志与调试的文本。
func (e *FieldError) Error() string {
	return fmt.Sprintf("field %q: %s", e.Field, e.Code)
}

// CollectFieldErrors 递归收集错误树中的所有 *FieldError，保持出现顺序。
//
// 兼容 errors.Join 聚合的错误：同一字段在同一规则链中只报告
// tag 顺序里第一个失败的规则（与 validator/v10 一致）。
func CollectFieldErrors(err error) []FieldError {
	if err == nil {
		return nil
	}
	var out []FieldError
	collect(err, &out)
	return out
}

// collect 深度优先遍历错误树，提取 *FieldError。
func collect(err error, out *[]FieldError) {
	if err == nil {
		return
	}
	// errors.Join 返回的 joinError 实现了 Unwrap() []error。
	if u, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range u.Unwrap() {
			collect(e, out)
		}
		return
	}
	if u, ok := err.(interface{ Unwrap() error }); ok {
		collect(u.Unwrap(), out)
		return
	}
	if fe, ok := err.(*FieldError); ok {
		*out = append(*out, *fe)
	}
}

// fieldNameTagPriority 是解析对外字段路径时依次尝试的绑定 tag，
// 覆盖 JSON/form/query/header/URI 等常见 HTTP 绑定来源。
var fieldNameTagPriority = []string{"json", "form", "query", "header", "uri", "param"}

// FieldName 解析字段对外使用的名称（错误路径）。
//
// 优先级：json > form > query > header > uri > param > Go 字段名；
// tag 值取 name 部分（逗号前的部分），形如 `json:",omitempty"`（无 name）时
// 回退到下一优先级；`json:"-"` 返回 "-" 表示该字段不对外。
//
// 静态生成（pkg/parser）与 runtime adapter（pkg/runtime）共用该函数，
// 保证错误路径一致。
func FieldName(tags reflect.StructTag, goName string) string {
	for _, key := range fieldNameTagPriority {
		raw, ok := tags.Lookup(key)
		if !ok || raw == "" {
			continue
		}
		name := raw
		if i := strings.IndexByte(raw, ','); i >= 0 {
			name = raw[:i]
		}
		if name == "" {
			continue
		}
		return name
	}
	return goName
}
