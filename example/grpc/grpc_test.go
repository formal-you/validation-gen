package grpcexample

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/formal-you/validation-gen/example/dto"
	"github.com/formal-you/validation-gen/pkg/errorx"
)

func TestValidateRequest(t *testing.T) {
	t.Run("合法请求返回 nil", func(t *testing.T) {
		req := &dto.Order{ID: 1, Code: "express", Quantity: 5}
		if err := ValidateRequest(context.Background(), req); err != nil {
			t.Fatalf("ValidateRequest() = %v, want nil", err)
		}
	})

	t.Run("非法请求返回 InvalidArgument", func(t *testing.T) {
		req := &dto.Order{ID: 0, Code: "priority", Quantity: 5}
		err := ValidateRequest(context.Background(), req)
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("status.Code = %v, want InvalidArgument", status.Code(err))
		}
		// status.Error 只保留文本；结构化错误需在转换前通过 req.Validate() 获取。
		if !strings.Contains(err.Error(), "id") || !strings.Contains(err.Error(), "required") {
			t.Fatalf("错误消息应包含字段与规则信息: %v", err)
		}
		fields := errorx.CollectFieldErrors(req.Validate())
		want := []errorx.FieldError{{Field: "id", Code: "required"}, {Field: "code", Code: "oneof"}}
		if !reflect.DeepEqual(fields, want) {
			t.Fatalf("req.Validate() 错误 = %+v, want %+v", fields, want)
		}
	})
}

func TestUnaryValidationInterceptor(t *testing.T) {
	interceptor := UnaryValidationInterceptor()
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	t.Run("非法请求被拦截", func(t *testing.T) {
		req := &dto.Order{ID: 0}
		_, err := interceptor(context.Background(), req, info, handler)
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("status.Code = %v, want InvalidArgument", status.Code(err))
		}
	})

	t.Run("合法请求进入 handler", func(t *testing.T) {
		req := &dto.Order{ID: 1, Code: "express", Quantity: 5}
		resp, err := interceptor(context.Background(), req, info, handler)
		if err != nil {
			t.Fatalf("interceptor err = %v, want nil", err)
		}
		if resp != "ok" {
			t.Fatalf("resp = %v, want ok", resp)
		}
	})

	t.Run("非 Validator 请求放行", func(t *testing.T) {
		_, err := interceptor(context.Background(), "plain", info, handler)
		if err != nil {
			t.Fatalf("非 Validator 请求应放行，err = %v", err)
		}
	})

	t.Run("拦截器错误可通过 errors.As 还原", func(t *testing.T) {
		req := &dto.Order{ID: 0}
		_, err := interceptor(context.Background(), req, info, handler)
		var st interface{ GRPCStatus() *status.Status }
		if !errors.As(err, &st) {
			t.Fatalf("err 应实现 GRPCStatus: %T", err)
		}
	})
}
