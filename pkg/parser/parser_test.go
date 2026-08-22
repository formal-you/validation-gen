package parser

import (
	"reflect"
	"strings"
	"testing"

	gotypes "go/types"

	"k8s.io/gengo/v2/types"

	"github.com/formal-you/validation-gen/pkg/ir"
)

// builtin 构造带 GoType 的基础类型（模拟 gengo 解析后的内置类型）。
func builtin(name string, k reflect.Kind) *types.Type {
	return &types.Type{Name: types.Name{Name: name}, Kind: types.Builtin, GoType: basic(k)}
}

func basic(k reflect.Kind) *gotypes.Basic {
	switch k {
	case reflect.String:
		return gotypes.Typ[gotypes.String]
	case reflect.Bool:
		return gotypes.Typ[gotypes.Bool]
	case reflect.Int:
		return gotypes.Typ[gotypes.Int]
	case reflect.Int8:
		return gotypes.Typ[gotypes.Int8]
	case reflect.Int64:
		return gotypes.Typ[gotypes.Int64]
	case reflect.Uint:
		return gotypes.Typ[gotypes.Uint]
	case reflect.Uint8:
		return gotypes.Typ[gotypes.Uint8]
	case reflect.Uint64:
		return gotypes.Typ[gotypes.Uint64]
	case reflect.Uintptr:
		return gotypes.Typ[gotypes.Uintptr]
	case reflect.Float32:
		return gotypes.Typ[gotypes.Float32]
	case reflect.Float64:
		return gotypes.Typ[gotypes.Float64]
	default:
		panic("unsupported test kind " + k.String())
	}
}

func ptrOf(t *types.Type) *types.Type {
	return &types.Type{Name: types.Name{Name: "*" + t.Name.Name}, Kind: types.Pointer, Elem: t, GoType: gotypes.NewPointer(t.GoType)}
}

func named(name string, underlying *types.Type) *types.Type {
	return &types.Type{Name: types.Name{Name: name}, Kind: types.Alias, Underlying: underlying}
}

func member(name, tags string, typ *types.Type) types.Member {
	return types.Member{Name: name, Tags: tags, Type: typ}
}

func TestResolveKind(t *testing.T) {
	cases := []struct {
		name     string
		typ      *types.Type
		wantKind ir.Kind
		wantBits int
		wantPtr  bool
		wantErr  string
	}{
		{name: "string", typ: builtin("string", reflect.String), wantKind: ir.KindString},
		{name: "bool", typ: builtin("bool", reflect.Bool), wantKind: ir.KindBool},
		{name: "int", typ: builtin("int", reflect.Int), wantKind: ir.KindSigned, wantBits: 64},
		{name: "int8", typ: builtin("int8", reflect.Int8), wantKind: ir.KindSigned, wantBits: 8},
		{name: "int64", typ: builtin("int64", reflect.Int64), wantKind: ir.KindSigned, wantBits: 64},
		{name: "uint", typ: builtin("uint", reflect.Uint), wantKind: ir.KindUnsigned, wantBits: 64},
		{name: "uint8", typ: builtin("uint8", reflect.Uint8), wantKind: ir.KindUnsigned, wantBits: 8},
		{name: "uintptr", typ: builtin("uintptr", reflect.Uintptr), wantKind: ir.KindUnsigned, wantBits: 64},
		{name: "float32", typ: builtin("float32", reflect.Float32), wantKind: ir.KindFloat, wantBits: 32},
		{name: "float64", typ: builtin("float64", reflect.Float64), wantKind: ir.KindFloat, wantBits: 64},
		{name: "named string", typ: named("Username", builtin("string", reflect.String)), wantKind: ir.KindString},
		{name: "named int", typ: named("Count", builtin("int8", reflect.Int8)), wantKind: ir.KindSigned, wantBits: 8},
		{name: "pointer", typ: ptrOf(builtin("string", reflect.String)), wantKind: ir.KindString, wantPtr: true},
		{name: "pointer to named", typ: ptrOf(named("Count", builtin("int", reflect.Int))), wantKind: ir.KindSigned, wantBits: 64, wantPtr: true},
		{name: "multi pointer", typ: ptrOf(ptrOf(builtin("string", reflect.String))), wantErr: "多级指针"},
		{name: "struct", typ: &types.Type{Name: types.Name{Name: "T"}, Kind: types.Struct, GoType: gotypes.NewStruct(nil, nil)}, wantErr: "仅支持标量"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			kind, bits, isPtr, err := ResolveKind(c.typ)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("ResolveKind() err = %v, want contains %q", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveKind() unexpected err = %v", err)
			}
			if kind != c.wantKind || bits != c.wantBits || isPtr != c.wantPtr {
				t.Fatalf("ResolveKind() = (%v, %d, %v), want (%v, %d, %v)",
					kind, bits, isPtr, c.wantKind, c.wantBits, c.wantPtr)
			}
		})
	}
}

func TestParseRules(t *testing.T) {
	cases := []struct {
		name    string
		tag     string
		kind    ir.Kind
		bitSize int
		isPtr   bool
		want    []ir.Rule
		wantErr string
	}{
		{
			name: "单规则 required",
			tag:  "required", kind: ir.KindString,
			want: []ir.Rule{{Name: "required"}},
		},
		{
			name: "多规则 min,max",
			tag:  "required,min=3,max=20", kind: ir.KindString,
			want: []ir.Rule{
				{Name: "required"},
				{Name: "min", Raw: "3", Literal: "3"},
				{Name: "max", Raw: "20", Literal: "20"},
			},
		},
		{
			name: "oneof 参数保留空格", tag: "oneof=admin user", kind: ir.KindString,
			want: []ir.Rule{{Name: "oneof", Raw: "admin user", Args: []string{"admin", "user"}}},
		},
		{
			name: "oneof 单引号候选", tag: "oneof='a b' c", kind: ir.KindString,
			want: []ir.Rule{{Name: "oneof", Raw: "'a b' c", Args: []string{"a b", "c"}}},
		},
		{
			name: "oneof 整数", tag: "oneof=1 2", kind: ir.KindSigned,
			want: []ir.Rule{{Name: "oneof", Raw: "1 2", Args: []string{"1", "2"}}},
		},
		{
			name: "十六进制参数规范化", tag: "min=0x10", kind: ir.KindString,
			want: []ir.Rule{{Name: "min", Raw: "0x10", Literal: "16"}},
		},
		{
			name: "uint 参数", tag: "gte=0", kind: ir.KindUnsigned,
			want: []ir.Rule{{Name: "gte", Raw: "0", Literal: "0"}},
		},
		{
			name: "float 参数", tag: "lte=0.5", kind: ir.KindFloat, bitSize: 64,
			want: []ir.Rule{{Name: "lte", Raw: "0.5", Literal: "0.5"}},
		},
		{name: "空 tag", tag: "", kind: ir.KindString, wantErr: "空规则"},
		{name: "中间空规则", tag: "required,,min=3", kind: ir.KindString, wantErr: "空规则"},
		{name: "未知规则", tag: "foo", kind: ir.KindString, wantErr: "未知规则"},
		{name: "runtime-only 规则", tag: "dive", kind: ir.KindString, wantErr: "仅支持 runtime"},
		{name: "重复规则", tag: "min=1,min=2", kind: ir.KindString, wantErr: "重复声明"},
		{name: "required+omitempty", tag: "required,omitempty", kind: ir.KindString, wantErr: "冲突"},
		{name: "min 缺参数", tag: "min", kind: ir.KindString, wantErr: "缺少参数"},
		{name: "min 非法参数", tag: "min=abc", kind: ir.KindString, wantErr: "非法"},
		{name: "min 负数 uint", tag: "min=-1", kind: ir.KindUnsigned, wantErr: "非负"},
		{name: "email 带参数", tag: "email=x", kind: ir.KindString, wantErr: "不接受参数"},
		{name: "email 非 string", tag: "email", kind: ir.KindSigned, wantErr: "仅支持 string"},
		{name: "oneof float", tag: "oneof=1 2", kind: ir.KindFloat, wantErr: "仅支持 string/整数"},
		{name: "min bool", tag: "min=1", kind: ir.KindBool, wantErr: "不支持 bool"},
		{name: "oneof 空候选", tag: "oneof=", kind: ir.KindString, wantErr: "缺少候选项"},
		{name: "required 带参数", tag: "required=1", kind: ir.KindString, wantErr: "不接受参数"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseRules(c.tag, c.kind, c.bitSize, c.isPtr)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("ParseRules() err = %v, want contains %q", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRules() unexpected err = %v", err)
			}
			if len(got) != len(c.want) {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
			for i := range c.want {
				if !reflect.DeepEqual(got[i], c.want[i]) {
					t.Fatalf("rule[%d] = %+v, want %+v", i, got[i], c.want[i])
				}
			}
		})
	}
}

func TestParseDefault(t *testing.T) {
	cases := []struct {
		name    string
		tag     string
		kind    ir.Kind
		bitSize int
		isPtr   bool
		want    string
		wantErr string
	}{
		{name: "string", tag: "guest", kind: ir.KindString, want: `"guest"`},
		{name: "string 转义", tag: `a"b`, kind: ir.KindString, want: `"a\"b"`},
		{name: "bool true", tag: "true", kind: ir.KindBool, want: "true"},
		{name: "bool 1", tag: "1", kind: ir.KindBool, want: "true"},
		{name: "int", tag: "5", kind: ir.KindSigned, bitSize: 64, want: "5"},
		{name: "int 十六进制", tag: "0x10", kind: ir.KindSigned, bitSize: 64, want: "16"},
		{name: "uint", tag: "10", kind: ir.KindUnsigned, bitSize: 64, want: "10"},
		{name: "float", tag: "0.5", kind: ir.KindFloat, bitSize: 64, want: "0.5"},
		{name: "int8 溢出", tag: "300", kind: ir.KindSigned, bitSize: 8, wantErr: "溢出"},
		{name: "bool 非法", tag: "yes", kind: ir.KindBool, wantErr: "不是合法 bool"},
		{name: "uint 负数", tag: "-1", kind: ir.KindUnsigned, bitSize: 64, wantErr: "非法或溢出"},
		{name: "float 非法", tag: "abc", kind: ir.KindFloat, bitSize: 64, wantErr: "非法或溢出"},
		{name: "pointer", tag: "1", kind: ir.KindSigned, isPtr: true, wantErr: "不支持 pointer"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d, err := ParseDefault(c.tag, c.kind, c.bitSize, c.isPtr)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("ParseDefault() err = %v, want contains %q", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDefault() unexpected err = %v", err)
			}
			if d.Literal != c.want {
				t.Fatalf("Literal = %q, want %q", d.Literal, c.want)
			}
		})
	}
}

func TestParseField(t *testing.T) {
	str := builtin("string", reflect.String)
	ptrStr := ptrOf(str)
	intT := builtin("int", reflect.Int)

	cases := []struct {
		name    string
		m       types.Member
		want    ir.FieldRules
		ok      bool
		wantErr string
	}{
		{
			name: "基本字段",
			m:    member("Name", `json:"name" validate:"required,min=3"`, str),
			want: ir.FieldRules{
				GoName: "Name", JSONName: "name", Kind: ir.KindString,
				Rules: []ir.Rule{{Name: "required"}, {Name: "min", Raw: "3", Literal: "3"}},
			},
			ok: true,
		},
		{
			name: "指针 + omitempty",
			m:    member("Email", `json:"email,omitempty" validate:"omitempty,email"`, ptrStr),
			want: ir.FieldRules{
				GoName: "Email", JSONName: "email", Kind: ir.KindString, IsPointer: true,
				Rules: []ir.Rule{{Name: "omitempty"}, {Name: "email"}},
			},
			ok: true,
		},
		{
			name: "header 绑定名作为错误路径",
			m:    member("Token", `header:"X-Token" validate:"required,len=10"`, str),
			want: ir.FieldRules{
				GoName: "Token", JSONName: "X-Token", Kind: ir.KindString,
				Rules: []ir.Rule{{Name: "required"}, {Name: "len", Raw: "10", Literal: "10"}},
			},
			ok: true,
		},
		{
			name: "form 优先于 uri",
			m:    member("Page", `form:"page" uri:"page_id" validate:"omitempty,gte=1"`, intT),
			want: ir.FieldRules{
				GoName: "Page", JSONName: "page", Kind: ir.KindSigned, BitSize: 64,
				Rules: []ir.Rule{{Name: "omitempty"}, {Name: "gte", Raw: "1", Literal: "1"}},
			},
			ok: true,
		},
		{
			name: "无 json tag 用 Go 字段名",
			m:    member("Age", `validate:"gte=0"`, intT),
			want: ir.FieldRules{
				GoName: "Age", JSONName: "Age", Kind: ir.KindSigned, BitSize: 64,
				Rules: []ir.Rule{{Name: "gte", Raw: "0", Literal: "0"}},
			},
			ok: true,
		},
		{
			name: "default",
			m:    member("Role", `json:"role" validate:"required" default:"guest"`, str),
			want: ir.FieldRules{
				GoName: "Role", JSONName: "role", Kind: ir.KindString,
				Rules:   []ir.Rule{{Name: "required"}},
				Default: &ir.Default{Literal: `"guest"`, Raw: "guest"},
			},
			ok: true,
		},
		{
			name: "无规则字段忽略",
			m:    member("NoTag", `json:"x"`, str),
			ok:   false,
		},
		{
			name:    "json:- 与 validate 冲突",
			m:       member("Skip", `json:"-" validate:"required"`, str),
			wantErr: `绑定 tag 为 "-" 与 validate 规则冲突`,
		},
		{
			name:    "空 validate tag",
			m:       member("E", `validate:""`, str),
			wantErr: "validate tag 为空",
		},
		{
			name:    "嵌入字段带规则",
			m:       types.Member{Name: "Base", Embedded: true, Tags: `validate:"required"`, Type: str},
			wantErr: "嵌入字段不支持",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fr, ok, err := ParseField("pkg", "T", c.m)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("ParseField() err = %v, want contains %q", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseField() unexpected err = %v", err)
			}
			if ok != c.ok {
				t.Fatalf("ok = %v, want %v", ok, c.ok)
			}
			if !c.ok {
				return
			}
			if !reflect.DeepEqual(fr, c.want) {
				t.Fatalf("got %+v, want %+v", fr, c.want)
			}
		})
	}
}
