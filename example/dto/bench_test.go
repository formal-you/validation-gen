package dto

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"

	"github.com/formal-you/validation-gen/pkg/runtime"
)

// 基准测试：静态生成 Validate() vs go-playground/validator/v10。
//
// 对比三条路径：
//   - Static：生成的 Validate()，零反射、零缓存查找；
//   - ValidatorWarm：复用已预热（map 缓存 + 已构建校验函数）的 validator 实例，
//     代表「validator 使用缓存后」的最优形态；
//   - RuntimeFallback：runtime.Validate(ctx, nil, v)，每次调用新建实例，
//     是项目文档化的显式扩展入口。
//
// 运行：go test -bench . -benchmem ./example/dto/
// 结果受 Go 版本与机器影响，README「性能对比」引用时标注运行环境。

var (
	benchValidUser = CreateUserRequest{
		Name: "Alice", Email: strp("a@b.com"), Age: 30, Role: "admin", Nickname: "n", Active: true,
	}
	benchInvalidUser = CreateUserRequest{
		Name: "", Email: strp("bad"), Age: 151, Role: "superuser", Nickname: "n", Active: false,
	}
	benchValidSettings = Settings{
		Role: "admin", Page: 2, Ratio: 0.5, Flag: true, Count: 3, Title: "untitled",
	}
)

// warmValidator 创建并预热一个 validator 实例：
// 首次校验会触发反射解析 struct 并写入 map 缓存、构建校验函数，预热后即为缓存态。
func warmValidator(v any) *validator.Validate {
	val := validator.New()
	_ = val.StructCtx(context.Background(), v)
	return val
}

// benchStatic 基准生成代码路径：合法输入必须通过，非法输入忽略错误（正确性由一致性测试保证）。
func benchStatic(b *testing.B, v interface{ Validate() error }, mustPass bool) {
	b.Helper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := v.Validate()
		if mustPass && err != nil {
			b.Fatal(err)
		}
	}
}

// benchValidatorWarm 基准复用预热实例路径（缓存态）。
func benchValidatorWarm(b *testing.B, val *validator.Validate, v any, mustPass bool) {
	b.Helper()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := val.StructCtx(context.Background(), v)
		if mustPass && err != nil {
			b.Fatal(err)
		}
	}
}

// benchRuntimeFallback 基准 runtime.Validate(ctx, nil, v)：每次调用新建 validator 实例，
// 并额外做 JSON 字段路径映射。
func benchRuntimeFallback(b *testing.B, v any, mustPass bool) {
	b.Helper()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := runtime.Validate(ctx, nil, v)
		if mustPass && err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateUserRequestValid(b *testing.B) {
	ctx := context.Background()
	_ = ctx
	b.Run("Static", func(b *testing.B) { benchStatic(b, &benchValidUser, true) })
	b.Run("ValidatorWarm", func(b *testing.B) {
		benchValidatorWarm(b, warmValidator(&benchValidUser), &benchValidUser, true)
	})
	b.Run("RuntimeFallback", func(b *testing.B) { benchRuntimeFallback(b, &benchValidUser, true) })
}

func BenchmarkCreateUserRequestInvalid(b *testing.B) {
	b.Run("Static", func(b *testing.B) { benchStatic(b, &benchInvalidUser, false) })
	b.Run("ValidatorWarm", func(b *testing.B) {
		benchValidatorWarm(b, warmValidator(&benchInvalidUser), &benchInvalidUser, false)
	})
	b.Run("RuntimeFallback", func(b *testing.B) { benchRuntimeFallback(b, &benchInvalidUser, false) })
}

func BenchmarkSettingsValid(b *testing.B) {
	b.Run("Static", func(b *testing.B) { benchStatic(b, &benchValidSettings, true) })
	b.Run("ValidatorWarm", func(b *testing.B) {
		benchValidatorWarm(b, warmValidator(&benchValidSettings), &benchValidSettings, true)
	})
}
