package valerr

import (
	"errors"
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
