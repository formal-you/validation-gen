// Package dto 演示 validate/default struct tag 的声明方式，并承载
// 静态生成代码与 runtime validator 的一致性测试。
//
// 重新生成：
//
//	go generate ./example/dto
package dto

//go:generate go run ../../cmd/valgen -type=CreateUserRequest -type=UserProfile -type=Order -type=Settings -type=NoValidate -type=WebRequest -output=zz_generated.validation.go

// Username 是 string 的 named type，验证 named type 支持。
type Username string

// Weight 是 float32 的 named type。
type Weight float32

// Level 是 uint8 的 named type。
type Level uint8

// CreateUserRequest 覆盖常见标量、指针、oneof、email 与 required 短路语义。
type CreateUserRequest struct {
	Name     string  `json:"name" validate:"required,min=3,max=20"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Age      int     `json:"age,omitempty" validate:"omitempty,gte=0,lte=150"`
	Role     string  `json:"role,omitempty" validate:"omitempty,oneof=admin user guest"`
	Nickname string  `json:"nickname" validate:"min=1,max=30"`
	Active   bool    `json:"active" validate:"required"`
}

// UserProfile 覆盖 named type、float、无 omitempty 的指针与 len 规则。
type UserProfile struct {
	Username Username `json:"username" validate:"required,min=3,max=16"`
	Bio      string   `json:"bio" validate:"max=200"`
	Weight   Weight   `json:"weight,omitempty" validate:"omitempty,gt=0,lte=200"`
	Level    Level    `json:"level,omitempty" validate:"omitempty,gte=1,lte=5"`
	Avatar   *string  `json:"avatar,omitempty" validate:"omitempty,len=10"`
	Handle   *string  `json:"handle,omitempty" validate:"min=3"`
	Nick     *string  `json:"nick,omitempty" validate:"required,min=3"`
	Score    float64  `json:"score,omitempty" validate:"omitempty,gte=0,lte=100"`
}

// Order 覆盖无符号整数、int64 与 oneof。
type Order struct {
	ID       int64  `json:"id" validate:"required,gt=0"`
	Code     string `json:"code,omitempty" validate:"omitempty,oneof=express normal"`
	Quantity uint   `json:"quantity,omitempty" validate:"omitempty,gt=0,lte=100"`
	Note     string `json:"note,omitempty" validate:"omitempty,max=50"`
}

// Settings 演示 default 与 validate 的组合：Validate 不调用 FillDefaults，
// 两者独立使用。
type Settings struct {
	Role  string  `json:"role" validate:"required" default:"guest"`
	Page  int     `json:"page,omitempty" validate:"omitempty,gte=1" default:"1"`
	Ratio float64 `json:"ratio,omitempty" validate:"omitempty,gte=0,lte=1" default:"0.5"`
	Flag  bool    `json:"flag" validate:"required" default:"true"`
	Count Level   `json:"count,omitempty" validate:"omitempty,lte=10" default:"3"`
	Title string  `json:"title,omitempty" validate:"max=10" default:"untitled"`
}

// NoValidate 没有任何校验规则也没有 default 字段：
// Validate() 直接返回 nil，不生成 FillDefaults()。
type NoValidate struct {
	A string `json:"a"`
	B int    `json:"b"`
}

// WebRequest 演示 form/header/uri 绑定来源的字段同样使用 validate/default tag：
// 校验与默认值不关心字段从哪个来源绑定，错误路径按绑定名输出。
type WebRequest struct {
	Username string `form:"username" validate:"required,min=3"`
	Token    string `header:"X-Token" validate:"required,len=10"`
	ID       int64  `uri:"id" validate:"required,gt=0"`
	Page     int    `form:"page" validate:"omitempty,gte=1" default:"1"`
}
