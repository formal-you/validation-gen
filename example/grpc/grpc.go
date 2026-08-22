// Package grpcexample 演示 gRPC request 校验集成：
//
//	request -> Validate -> status.Error(codes.InvalidArgument, ...)
//
// 不依赖 protobuf 代码生成：示例以拦截器 + 校验函数演示校验模式，
// 真实服务只需把生成的 Validate() 挂到对应请求类型上。
package grpcexample

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Validator 是静态校验请求的最小子集接口。
// 任何带有生成 Validate() 方法的请求类型都满足该接口。
type Validator interface {
	Validate() error
}

// ValidateRequest 校验 gRPC 请求，失败时返回 codes.InvalidArgument。
// 错误消息仅用于日志与调试；结构化信息通过 err.CollectFieldErrors 获取。
func ValidateRequest(ctx context.Context, req Validator) error {
	if err := req.Validate(); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}

// UnaryValidationInterceptor 返回 gRPC unary 拦截器：
// 对实现了 Validator 的请求自动执行静态校验，失败直接返回 InvalidArgument，
// 不进入业务 handler；非 Validator 请求原样放行。
func UnaryValidationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if v, ok := req.(Validator); ok {
			if err := ValidateRequest(ctx, v); err != nil {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}
