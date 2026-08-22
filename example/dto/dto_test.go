package dto

import (
	"context"
	"reflect"
	"testing"

	"github.com/formal-you/validation-gen/pkg/runtime"
	"github.com/formal-you/validation-gen/pkg/valerr"
)

func strp(s string) *string { return &s }

// errs 收集校验错误。
func errs(err error) []valerr.FieldError {
	return valerr.CollectFieldErrors(err)
}

// assertConform 同时执行静态 Validate() 与 runtime.Validate，
// 断言两者与期望的 FieldError 列表完全一致（字段顺序 + 失败规则 code）。
func assertConform(t *testing.T, name string, v interface{ Validate() error }, want []valerr.FieldError) {
	t.Helper()
	gotStatic := errs(v.Validate())
	if !reflect.DeepEqual(gotStatic, want) {
		t.Fatalf("%s: 静态校验 = %+v, want %+v", name, gotStatic, want)
	}
	gotRuntime := errs(runtime.Validate(context.Background(), nil, v))
	if !reflect.DeepEqual(gotRuntime, want) {
		t.Fatalf("%s: runtime 校验 = %+v, want %+v", name, gotRuntime, want)
	}
}

func TestCreateUserRequestConform(t *testing.T) {
	cases := []struct {
		name string
		req  CreateUserRequest
		want []valerr.FieldError
	}{
		{name: "合法", req: CreateUserRequest{Name: "Alice", Email: strp("a@b.com"), Age: 30, Role: "admin", Nickname: "n", Active: true}},
		{name: "required 失败", req: CreateUserRequest{Name: "", Email: strp("a@b.com"), Age: 30, Role: "admin", Nickname: "n", Active: true},
			want: []valerr.FieldError{{Field: "name", Code: "required"}}},
		{name: "min 边界失败", req: CreateUserRequest{Name: "ab", Email: strp("a@b.com"), Age: 30, Role: "admin", Nickname: "n", Active: true},
			want: []valerr.FieldError{{Field: "name", Code: "min"}}},
		{name: "min 边界通过", req: CreateUserRequest{Name: "abc", Email: strp("a@b.com"), Age: 30, Role: "admin", Nickname: "n", Active: true}},
		{name: "max 失败", req: CreateUserRequest{Name: "abcdefghijklmnopqrstu", Email: strp("a@b.com"), Age: 30, Role: "admin", Nickname: "n", Active: true},
			want: []valerr.FieldError{{Field: "name", Code: "max"}}},
		{name: "UTF-8 按字符数", req: CreateUserRequest{Name: "你好", Email: strp("a@b.com"), Age: 30, Role: "admin", Nickname: "n", Active: true},
			want: []valerr.FieldError{{Field: "name", Code: "min"}}},
		{name: "UTF-8 通过", req: CreateUserRequest{Name: "你好啊", Email: strp("a@b.com"), Age: 30, Role: "admin", Nickname: "n", Active: true}},
		{name: "email 失败", req: CreateUserRequest{Name: "Alice", Email: strp("bad"), Age: 30, Role: "admin", Nickname: "n", Active: true},
			want: []valerr.FieldError{{Field: "email", Code: "email"}}},
		{name: "email 空串执行校验", req: CreateUserRequest{Name: "Alice", Email: strp(""), Age: 30, Role: "admin", Nickname: "n", Active: true},
			want: []valerr.FieldError{{Field: "email", Code: "email"}}},
		{name: "email nil 跳过", req: CreateUserRequest{Name: "Alice", Email: nil, Age: 30, Role: "admin", Nickname: "n", Active: true}},
		{name: "lte 失败", req: CreateUserRequest{Name: "Alice", Email: strp("a@b.com"), Age: 151, Role: "admin", Nickname: "n", Active: true},
			want: []valerr.FieldError{{Field: "age", Code: "lte"}}},
		{name: "gte 失败", req: CreateUserRequest{Name: "Alice", Email: strp("a@b.com"), Age: -1, Role: "admin", Nickname: "n", Active: true},
			want: []valerr.FieldError{{Field: "age", Code: "gte"}}},
		{name: "omitempty 零值跳过", req: CreateUserRequest{Name: "Alice", Email: strp("a@b.com"), Age: 0, Role: "admin", Nickname: "n", Active: true}},
		{name: "oneof 失败", req: CreateUserRequest{Name: "Alice", Email: strp("a@b.com"), Age: 30, Role: "superuser", Nickname: "n", Active: true},
			want: []valerr.FieldError{{Field: "role", Code: "oneof"}}},
		{name: "oneof 空串跳过", req: CreateUserRequest{Name: "Alice", Email: strp("a@b.com"), Age: 30, Role: "", Nickname: "n", Active: true}},
		{name: "required bool 失败", req: CreateUserRequest{Name: "Alice", Email: strp("a@b.com"), Age: 30, Role: "admin", Nickname: "n", Active: false},
			want: []valerr.FieldError{{Field: "active", Code: "required"}}},
		{name: "多字段错误按声明顺序", req: CreateUserRequest{Name: "", Email: strp("bad"), Age: 151, Role: "superuser", Nickname: "n", Active: false},
			want: []valerr.FieldError{
				{Field: "name", Code: "required"},
				{Field: "email", Code: "email"},
				{Field: "age", Code: "lte"},
				{Field: "role", Code: "oneof"},
				{Field: "active", Code: "required"},
			}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertConform(t, c.name, &c.req, c.want)
		})
	}
}

func TestUserProfileConform(t *testing.T) {
	handle := "handle"
	nick := "nick"
	shortNick := "ab"
	cases := []struct {
		name string
		req  UserProfile
		want []valerr.FieldError
	}{
		{name: "合法", req: UserProfile{Username: "alice", Weight: 50, Level: 2, Avatar: nil, Handle: &handle, Nick: &nick, Score: 90}},
		{name: "named string required", req: UserProfile{Username: "", Weight: 50, Level: 2, Avatar: nil, Handle: &handle, Nick: &nick, Score: 90},
			want: []valerr.FieldError{{Field: "username", Code: "required"}}},
		{name: "named string min", req: UserProfile{Username: "ab", Weight: 50, Level: 2, Avatar: nil, Handle: &handle, Nick: &nick, Score: 90},
			want: []valerr.FieldError{{Field: "username", Code: "min"}}},
		{name: "named float32 gt", req: UserProfile{Username: "alice", Weight: -1, Level: 2, Avatar: nil, Handle: &handle, Nick: &nick, Score: 90},
			want: []valerr.FieldError{{Field: "weight", Code: "gt"}}},
		{name: "named float32 lte", req: UserProfile{Username: "alice", Weight: 201, Level: 2, Avatar: nil, Handle: &handle, Nick: &nick, Score: 90},
			want: []valerr.FieldError{{Field: "weight", Code: "lte"}}},
		{name: "named float32 零值跳过", req: UserProfile{Username: "alice", Weight: 0, Level: 2, Avatar: nil, Handle: &handle, Nick: &nick, Score: 90}},
		{name: "named uint8 lte", req: UserProfile{Username: "alice", Weight: 50, Level: 6, Avatar: nil, Handle: &handle, Nick: &nick, Score: 90},
			want: []valerr.FieldError{{Field: "level", Code: "lte"}}},
		{name: "len 失败", req: UserProfile{Username: "alice", Weight: 50, Level: 2, Avatar: strp("123"), Handle: &handle, Nick: &nick, Score: 90},
			want: []valerr.FieldError{{Field: "avatar", Code: "len"}}},
		{name: "len 通过", req: UserProfile{Username: "alice", Weight: 50, Level: 2, Avatar: strp("1234567890"), Handle: &handle, Nick: &nick, Score: 90}},
		{name: "指针无 omitempty nil 失败", req: UserProfile{Username: "alice", Weight: 50, Level: 2, Avatar: nil, Handle: nil, Nick: &nick, Score: 90},
			want: []valerr.FieldError{{Field: "handle", Code: "min"}}},
		{name: "指针 required nil 失败", req: UserProfile{Username: "alice", Weight: 50, Level: 2, Avatar: nil, Handle: &handle, Nick: nil, Score: 90},
			want: []valerr.FieldError{{Field: "nick", Code: "required"}}},
		{name: "指针 required 通过但 min 失败", req: UserProfile{Username: "alice", Weight: 50, Level: 2, Avatar: nil, Handle: &handle, Nick: &shortNick, Score: 90},
			want: []valerr.FieldError{{Field: "nick", Code: "min"}}},
		{name: "float64 gte 失败", req: UserProfile{Username: "alice", Weight: 50, Level: 2, Avatar: nil, Handle: &handle, Nick: &nick, Score: -0.5},
			want: []valerr.FieldError{{Field: "score", Code: "gte"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertConform(t, c.name, &c.req, c.want)
		})
	}
}

func TestOrderConform(t *testing.T) {
	cases := []struct {
		name string
		req  Order
		want []valerr.FieldError
	}{
		{name: "合法", req: Order{ID: 1, Code: "express", Quantity: 5}},
		{name: "required 优先于 gt", req: Order{ID: 0, Code: "express", Quantity: 5},
			want: []valerr.FieldError{{Field: "id", Code: "required"}}},
		{name: "int64 gt 失败", req: Order{ID: -5, Code: "express", Quantity: 5},
			want: []valerr.FieldError{{Field: "id", Code: "gt"}}},
		{name: "oneof 失败", req: Order{ID: 1, Code: "priority", Quantity: 5},
			want: []valerr.FieldError{{Field: "code", Code: "oneof"}}},
		{name: "uint lte 失败", req: Order{ID: 1, Code: "express", Quantity: 101},
			want: []valerr.FieldError{{Field: "quantity", Code: "lte"}}},
		{name: "uint 零值跳过", req: Order{ID: 1, Code: "express", Quantity: 0}},
		{name: "note max 失败", req: Order{ID: 1, Code: "express", Quantity: 5, Note: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" + "aa"},
			want: []valerr.FieldError{{Field: "note", Code: "max"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertConform(t, c.name, &c.req, c.want)
		})
	}
}

func TestSettingsConform(t *testing.T) {
	cases := []struct {
		name string
		req  Settings
		want []valerr.FieldError
	}{
		{name: "合法（已填值）", req: Settings{Role: "admin", Page: 2, Ratio: 0.5, Flag: true, Count: 5, Title: "t"}},
		{name: "required 失败", req: Settings{Role: "", Page: 2, Ratio: 0.5, Flag: true, Count: 5, Title: "t"},
			want: []valerr.FieldError{{Field: "role", Code: "required"}}},
		{name: "float lte 失败", req: Settings{Role: "admin", Page: 2, Ratio: 1.5, Flag: true, Count: 5, Title: "t"},
			want: []valerr.FieldError{{Field: "ratio", Code: "lte"}}},
		{name: "named uint lte 失败", req: Settings{Role: "admin", Page: 2, Ratio: 0.5, Flag: true, Count: 11, Title: "t"},
			want: []valerr.FieldError{{Field: "count", Code: "lte"}}},
		{name: "title max 失败", req: Settings{Role: "admin", Page: 2, Ratio: 0.5, Flag: true, Count: 5, Title: "aaaaaaaaaaa"},
			want: []valerr.FieldError{{Field: "title", Code: "max"}}},
		{name: "零值经过 omitempty 跳过", req: Settings{Role: "admin", Page: 0, Ratio: 0, Flag: true, Count: 0, Title: ""}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertConform(t, c.name, &c.req, c.want)
		})
	}
}

func TestNoValidate(t *testing.T) {
	req := &NoValidate{A: "", B: 0}
	if err := req.Validate(); err != nil {
		t.Fatalf("NoValidate.Validate() = %v, want nil", err)
	}
}

func TestFillDefaults(t *testing.T) {
	s := &Settings{}
	s.FillDefaults()
	want := Settings{Role: "guest", Page: 1, Ratio: 0.5, Flag: true, Count: 3, Title: "untitled"}
	if *s != want {
		t.Fatalf("FillDefaults() = %+v, want %+v", *s, want)
	}

	// 幂等：二次调用结果不变。
	before := *s
	s.FillDefaults()
	if *s != before {
		t.Fatalf("FillDefaults 应幂等：%+v -> %+v", before, *s)
	}

	// 不覆盖非零值。
	orig := Settings{Role: "admin", Page: 2, Ratio: 1, Flag: true, Count: 7, Title: "custom"}
	copyOrig := orig
	orig.FillDefaults()
	if orig != copyOrig {
		t.Fatalf("FillDefaults 不应覆盖非零值：%+v -> %+v", copyOrig, orig)
	}
}

func TestValidateDoesNotMutate(t *testing.T) {
	// Validate 不调用 FillDefaults、不修改接收对象。
	s := Settings{} // Role 为空，Validate 应失败且不填充。
	if err := (&s).Validate(); err == nil {
		t.Fatal("Validate() 应失败（Role required）")
	}
	if s.Role != "" {
		t.Fatalf("Validate 不应修改接收对象，Role = %q", s.Role)
	}
}

func TestWebRequestConform(t *testing.T) {
	cases := []struct {
		name string
		req  WebRequest
		want []valerr.FieldError
	}{
		{name: "合法", req: WebRequest{Username: "alice", Token: "1234567890", ID: 1, Page: 2}},
		{name: "form required 失败", req: WebRequest{Username: "", Token: "1234567890", ID: 1, Page: 2},
			want: []valerr.FieldError{{Field: "username", Code: "required"}}},
		{name: "form min 失败", req: WebRequest{Username: "ab", Token: "1234567890", ID: 1, Page: 2},
			want: []valerr.FieldError{{Field: "username", Code: "min"}}},
		{name: "header required 失败", req: WebRequest{Username: "alice", Token: "", ID: 1, Page: 2},
			want: []valerr.FieldError{{Field: "X-Token", Code: "required"}}},
		{name: "header len 失败", req: WebRequest{Username: "alice", Token: "short", ID: 1, Page: 2},
			want: []valerr.FieldError{{Field: "X-Token", Code: "len"}}},
		{name: "uri required 失败", req: WebRequest{Username: "alice", Token: "1234567890", ID: 0, Page: 2},
			want: []valerr.FieldError{{Field: "id", Code: "required"}}},
		{name: "uri gt 失败", req: WebRequest{Username: "alice", Token: "1234567890", ID: -1, Page: 2},
			want: []valerr.FieldError{{Field: "id", Code: "gt"}}},
		{name: "form 零值跳过", req: WebRequest{Username: "alice", Token: "1234567890", ID: 1, Page: 0}},
		{name: "多字段错误按声明顺序", req: WebRequest{Username: "", Token: "", ID: 0},
			want: []valerr.FieldError{
				{Field: "username", Code: "required"},
				{Field: "X-Token", Code: "required"},
				{Field: "id", Code: "required"},
			}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertConform(t, c.name, &c.req, c.want)
		})
	}
}

func TestWebRequestFillDefaults(t *testing.T) {
	req := &WebRequest{}
	req.FillDefaults()
	if req.Page != 1 {
		t.Fatalf("FillDefaults() 后 Page = %d, want 1", req.Page)
	}
}
