// Package httpexample 演示 HTTP request 校验集成：
//
//	JSON bind -> FillDefaults -> Validate -> 400 结构化错误
//
// 使用标准库 net/http，避免引入额外框架；中间件封装留待调用方按需扩展。
package httpexample

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/formal-you/validation-gen/example/dto"
	"github.com/formal-you/validation-gen/pkg/valerr"
)

// CreateUserHandler 处理 POST /users。
//
// 流程：解析 JSON -> FillDefaults -> Validate；校验失败返回
// 400 + {"errors":[{"field":"name","code":"required"},...]}。
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	req.FillDefaults()
	if err := req.Validate(); err != nil {
		writeFieldErrors(w, http.StatusBadRequest, valerr.CollectFieldErrors(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// WebRequestHandler 演示 form/header/query/uri 混合绑定的校验：
//
//	POST /users/{id}?page=N
//	Header: X-Token: <token>
//	Body (application/x-www-form-urlencoded): username=alice
//
// 绑定 -> FillDefaults -> Validate，失败返回 400 结构化错误。
// gin/echo 等框架用对应绑定 tag（form/header/uri/query/param）做同样的事，
// 这里用标准库演示，避免引入框架依赖。
func WebRequestHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.WebRequest
	req.Username = r.PostFormValue("username")
	req.Token = r.Header.Get("X-Token")
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	req.ID = id
	if p := r.URL.Query().Get("page"); p != "" {
		page, err := strconv.Atoi(p)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid page")
			return
		}
		req.Page = page
	}
	req.FillDefaults()
	if err := req.Validate(); err != nil {
		writeFieldErrors(w, http.StatusBadRequest, valerr.CollectFieldErrors(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SettingsHandler 演示 default 与校验的组合：
// 客户端只提交 Role，FillDefaults 填充其余字段后再校验。
func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.Settings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	req.FillDefaults()
	if err := req.Validate(); err != nil {
		writeFieldErrors(w, http.StatusBadRequest, valerr.CollectFieldErrors(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeFieldErrors 以 JSON 数组形式输出结构化校验错误。
func writeFieldErrors(w http.ResponseWriter, status int, fields []valerr.FieldError) {
	writeJSON(w, status, map[string]any{"errors": fields})
}

// writeError 输出纯文本错误。
func writeError(w http.ResponseWriter, status int, msg string) {
	http.Error(w, msg, status)
}

// writeJSON 输出 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
