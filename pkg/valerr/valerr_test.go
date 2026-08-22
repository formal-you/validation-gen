package valerr

import (
	"errors"
	"reflect"
	"testing"
)

func TestFieldErrorError(t *testing.T) {
	e := &FieldError{Field: "name", Code: "required"}
	if got, want := e.Error(), `field "name": required`; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestCollectFieldErrors(t *testing.T) {
	t.Run("nil 返回 nil", func(t *testing.T) {
		if got := CollectFieldErrors(nil); got != nil {
			t.Fatalf("CollectFieldErrors(nil) = %v, want nil", got)
		}
	})

	t.Run("单个 FieldError", func(t *testing.T) {
		got := CollectFieldErrors(&FieldError{Field: "name", Code: "required"})
		want := []FieldError{{Field: "name", Code: "required"}}
		if len(got) != 1 || got[0] != want[0] {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("errors.Join 多错误保持顺序", func(t *testing.T) {
		err := errors.Join(
			&FieldError{Field: "name", Code: "required"},
			&FieldError{Field: "age", Code: "lte"},
			errors.New("其他错误"),
		)
		got := CollectFieldErrors(err)
		want := []FieldError{{Field: "name", Code: "required"}, {Field: "age", Code: "lte"}}
		if len(got) != len(want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %+v, want %+v", got, want)
			}
		}
	})
}

func TestFieldName(t *testing.T) {
	cases := []struct {
		name string
		tags string
		want string
	}{
		{name: "json 优先", tags: `json:"name" form:"username"`, want: "name"},
		{name: "无 json 用 form", tags: `form:"username"`, want: "username"},
		{name: "form 优先于 header", tags: `form:"page" header:"X-Page"`, want: "page"},
		{name: "header", tags: `header:"X-Token"`, want: "X-Token"},
		{name: "uri", tags: `uri:"id"`, want: "id"},
		{name: "query", tags: `query:"q"`, want: "q"},
		{name: "param（echo 路径参数）", tags: `param:"slug"`, want: "slug"},
		{name: "json 无 name 回退", tags: `json:",omitempty"`, want: "GoName"},
		{name: "json:- 返回 -", tags: `json:"-"`, want: "-"},
		{name: "无 tag 用 Go 字段名", tags: `validate:"required"`, want: "GoName"},
		{name: "忽略其他 tag", tags: `xml:"name" binding:"required"`, want: "GoName"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := FieldName(structTag(c.tags), "GoName"); got != c.want {
				t.Fatalf("FieldName(%q) = %q, want %q", c.tags, got, c.want)
			}
		})
	}
}

func structTag(s string) reflect.StructTag {
	return reflect.StructTag(s)
}
