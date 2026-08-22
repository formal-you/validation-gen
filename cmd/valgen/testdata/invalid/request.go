// Package invalid 是「非法规则」fixture：用于集成测试验证
// 生成失败时不会覆盖已有的生成文件。
package invalid

// Request 的 validate tag 含 runtime-only 规则 dive，生成期必须报错。
type Request struct {
	Name string `json:"name" validate:"required,dive"`
}
