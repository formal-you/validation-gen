// Package runtime 提供基于 go-playground/validator/v10 的运行时校验入口。
//
// 静态生成的 Validate() 只覆盖 v1 白名单规则；需要完整能力
// （自定义 validator、dive、跨字段规则、alias 等）时，调用方显式使用
// runtime.Validate，传入自己的 *validator.Validate 实例（可注入自定义规则）。
// 本包不使用全局可变 validator，也不会修改调用方传入的实例配置。
package runtime

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/formal-you/validation-gen/pkg/valerr"
)

// Validate 使用 validator/v10 对 value 执行完整运行时校验。
//
// v 为 nil 时创建一次性实例（每次调用独立，不共享全局状态）；
// 传入自定义实例时，调用方已有的 RegisterValidation/RegisterAlias/StructCtx 配置保持不变。
//
// 返回的错误可通过 valerr.CollectFieldErrors 还原为 []FieldError，
// 字段路径已映射为 JSON 名称；非校验错误（如 InvalidValidationError）原样返回。
func Validate(ctx context.Context, v *validator.Validate, value any) error {
	if v == nil {
		v = validator.New()
	}
	err := v.StructCtx(ctx, value)
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		fields := make([]error, 0, len(verrs))
		for _, fe := range verrs {
			fields = append(fields, &valerr.FieldError{
				Field: FieldPath(fe, value),
				Code:  fe.Tag(),
			})
		}
		return errors.Join(fields...)
	}
	return err
}

// FieldPath 把 validator 的 StructNamespace（如 "CreateUserRequest.Profile.Email"）
// 映射为 JSON 字段路径（如 "profile.email"）。
//
// 通过反射遍历根值逐段解析 json tag；数组/切片索引（[0]）与 map key 原样保留。
// 解析失败时回退到 validator 原始分段，保证返回值始终可用。
func FieldPath(fe validator.FieldError, root any) string {
	segs := strings.Split(fe.StructNamespace(), ".")
	if len(segs) < 2 {
		return fe.Field()
	}
	var out []string
	cur := reflect.ValueOf(root)
	for _, seg := range segs[1:] {
		name, idx, hasIdx, key, hasKey := splitIndex(seg)

		cur = deref(cur)
		if !cur.IsValid() || cur.Kind() != reflect.Struct {
			out = append(out, seg)
			return strings.Join(out, ".")
		}
		f := cur.FieldByName(name)
		if !f.IsValid() {
			out = append(out, seg)
			return strings.Join(out, ".")
		}
		if sf, ok := cur.Type().FieldByName(name); ok {
			out = append(out, jsonFieldName(sf))
		} else {
			out = append(out, name)
		}

		cur = f
		if hasIdx {
			cur = deref(cur)
			if cur.IsValid() && (cur.Kind() == reflect.Array || cur.Kind() == reflect.Slice) {
				if idx >= 0 && idx < cur.Len() {
					cur = cur.Index(idx)
					out[len(out)-1] += "[" + strconv.Itoa(idx) + "]"
				} else {
					return strings.Join(out, ".")
				}
			} else {
				return strings.Join(out, ".")
			}
		}
		if hasKey {
			cur = deref(cur)
			if cur.IsValid() && cur.Kind() == reflect.Map {
				kv := reflect.ValueOf(key)
				if kv.IsValid() && kv.Type().AssignableTo(cur.Type().Key()) {
					cur = cur.MapIndex(kv)
					out[len(out)-1] += "[" + key + "]"
				} else {
					return strings.Join(out, ".")
				}
			} else {
				return strings.Join(out, ".")
			}
		}
	}
	return strings.Join(out, ".")
}

// deref 解引用指针/接口到具体值；nil 时返回零 Value。
func deref(v reflect.Value) reflect.Value {
	for v.IsValid() && (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

// splitIndex 拆分 "Field[0]" 或 "Field[key]" 形式的分段。
func splitIndex(seg string) (name string, idx int, hasIdx bool, key string, hasKey bool) {
	i := strings.IndexByte(seg, '[')
	if i < 0 {
		return seg, 0, false, "", false
	}
	name = seg[:i]
	if !strings.HasSuffix(seg, "]") {
		return seg, 0, false, "", false
	}
	inner := seg[i+1 : len(seg)-1]
	if n, err := strconv.Atoi(inner); err == nil {
		return name, n, true, "", false
	}
	return name, 0, false, inner, true
}

// jsonFieldName 返回结构体字段的对外名称（错误路径）。
// 与静态生成共用 valerr.FieldName：json > form > query > header > uri > param > Go 字段名。
func jsonFieldName(sf reflect.StructField) string {
	return valerr.FieldName(sf.Tag, sf.Name)
}
