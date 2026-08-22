package runtime

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"

	"github.com/formal-you/validation-gen/pkg/valerr"
)

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Code     string `json:"code" validate:"omitempty,oneof=ios android"`
}

type diveRequest struct {
	Items []string `json:"items" validate:"required,dive,min=2"`
}

type crossRequest struct {
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
}

func collect(err error) []valerr.FieldError {
	return valerr.CollectFieldErrors(err)
}

func TestValidateBuiltinRules(t *testing.T) {
	t.Run("通过", func(t *testing.T) {
		req := &loginRequest{Email: "a@b.com", Password: "12345678"}
		if err := Validate(context.Background(), nil, req); err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
	})

	t.Run("多字段错误按 JSON 名映射", func(t *testing.T) {
		req := &loginRequest{Email: "bad", Password: "short"}
		err := Validate(context.Background(), nil, req)
		got := collect(err)
		want := []valerr.FieldError{
			{Field: "email", Code: "email"},
			{Field: "password", Code: "min"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("omitempty 短路", func(t *testing.T) {
		req := &loginRequest{Email: "a@b.com", Password: "12345678", Code: "windows"}
		err := Validate(context.Background(), nil, req)
		got := collect(err)
		want := []valerr.FieldError{{Field: "code", Code: "oneof"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
}

func TestValidateDive(t *testing.T) {
	req := &diveRequest{Items: []string{"ok", "x"}}
	err := Validate(context.Background(), nil, req)
	got := collect(err)
	if len(got) != 1 || got[0].Field != "items[1]" || got[0].Code != "min" {
		t.Fatalf("got %+v, want items[1] min", got)
	}
}

func TestValidateCrossField(t *testing.T) {
	req := &crossRequest{Password: "abc", ConfirmPassword: "abd"}
	err := Validate(context.Background(), nil, req)
	got := collect(err)
	if len(got) != 1 || got[0].Field != "confirm_password" || got[0].Code != "eqfield" {
		t.Fatalf("got %+v, want confirm_password eqfield", got)
	}
}

func TestValidateCustomValidator(t *testing.T) {
	v := validator.New()
	if err := v.RegisterValidation("is-upper", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		return s != "" && s == strings.ToUpper(s)
	}); err != nil {
		t.Fatal(err)
	}

	req := &loginRequest{Email: "a@b.com", Password: "12345678", Code: "ios"}
	_ = req // 自定义规则作用于独立 DTO

	type upperReq struct {
		Name string `json:"name" validate:"required,is-upper"`
	}
	if err := Validate(context.Background(), v, &upperReq{Name: "abc"}); err == nil {
		t.Fatal("want error for non-upper name")
	}
	if err := Validate(context.Background(), v, &upperReq{Name: "ABC"}); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
	// 实例未被修改：注册的规则仍然存在。
	if err := Validate(context.Background(), v, &upperReq{Name: "abc"}); err == nil {
		t.Fatal("custom validator 配置应保持不变")
	}
}

func TestValidateStructCtx(t *testing.T) {
	type ctxReq struct {
		N int `json:"n" validate:"required,even-mul"`
	}
	v := validator.New()
	if err := v.RegisterValidationCtx("even-mul", func(ctx context.Context, fl validator.FieldLevel) bool {
		mul, ok := ctx.Value("mul").(int)
		if !ok || mul == 0 {
			return false
		}
		return fl.Field().Int()%int64(mul) == 0
	}); err != nil {
		t.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), "mul", 3)
	if err := Validate(ctx, v, &ctxReq{N: 9}); err != nil {
		t.Fatalf("9 应为 3 的倍数，got %v", err)
	}
	if err := Validate(ctx, v, &ctxReq{N: 10}); err == nil {
		t.Fatal("10 不应是 3 的倍数")
	}
	// 缺少 context 值 -> 校验失败。
	if err := Validate(context.Background(), v, &ctxReq{N: 9}); err == nil {
		t.Fatal("缺少 ctx 值时应失败")
	}
}

func TestValidateInvalidValue(t *testing.T) {
	err := Validate(context.Background(), nil, "not-a-struct")
	var inv *validator.InvalidValidationError
	if !errors.As(err, &inv) {
		t.Fatalf("want InvalidValidationError, got %T: %v", err, err)
	}
}

func TestFieldPathJSONMapping(t *testing.T) {
	type profile struct {
		Email string `json:"email" validate:"email"`
	}
	type req struct {
		Name    string   `json:"name" validate:"min=2"`
		Profile profile  `json:"profile"`
		Tags    []string `json:"tags" validate:"dive,min=2"`
	}

	v := validator.New()
	err := Validate(context.Background(), v, &req{
		Name:    "x",
		Profile: profile{Email: "bad"},
		Tags:    []string{"a"},
	})
	got := collect(err)
	wantFields := []string{"name", "profile.email", "tags[0]"}
	if len(got) != len(wantFields) {
		t.Fatalf("got %+v, want %v", got, wantFields)
	}
	for i, wf := range wantFields {
		if got[i].Field != wf {
			t.Fatalf("got %+v, want fields %v", got, wantFields)
		}
	}
}
