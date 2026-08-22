package httpexample

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func doPost(t *testing.T, handler http.HandlerFunc, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestCreateUserHandler(t *testing.T) {
	t.Run("合法请求返回 204", func(t *testing.T) {
		body := `{"name":"Alice","email":"a@b.com","age":30,"role":"admin","nickname":"n","active":true}`
		rec := doPost(t, CreateUserHandler, body)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("校验失败返回 400 结构化错误", func(t *testing.T) {
		body := `{"name":"","email":"bad","age":151,"role":"superuser","nickname":"n","active":true}`
		rec := doPost(t, CreateUserHandler, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		var resp struct {
			Errors []struct {
				Field string `json:"field"`
				Code  string `json:"code"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("响应不是合法 JSON: %v", err)
		}
		want := []string{"name:required", "email:email", "age:lte", "role:oneof"}
		if len(resp.Errors) != len(want) {
			t.Fatalf("errors = %+v, want %v", resp.Errors, want)
		}
		for i, w := range want {
			got := resp.Errors[i].Field + ":" + resp.Errors[i].Code
			if got != w {
				t.Fatalf("errors[%d] = %s, want %s", i, got, w)
			}
		}
	})

	t.Run("JSON bind 失败返回 400", func(t *testing.T) {
		rec := doPost(t, CreateUserHandler, `{"name":`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("显式提供零值 nickname=空串 -> min 失败", func(t *testing.T) {
		// 显式提供的零值不会被当作未提供跳过：nickname 无 omitempty，min=1 校验失败。
		body := `{"name":"Alice","email":"a@b.com","age":30,"role":"admin","nickname":"","active":true}`
		rec := doPost(t, CreateUserHandler, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		var resp struct {
			Errors []struct {
				Field string `json:"field"`
				Code  string `json:"code"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("响应不是合法 JSON: %v", err)
		}
		if len(resp.Errors) != 1 || resp.Errors[0].Field != "nickname" || resp.Errors[0].Code != "min" {
			t.Fatalf("errors = %+v, want [{nickname min}]", resp.Errors)
		}
	})

	t.Run("显式提供零值 age=0 且规则允许 -> 204", func(t *testing.T) {
		// age 显式传 0（零值）：omitempty 只对 nil/空指针短路，值 0 仍执行 gte=0/lte=150，应通过。
		body := `{"name":"Alice","nickname":"n","active":true,"age":0}`
		rec := doPost(t, CreateUserHandler, body)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204; body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestSettingsHandlerDefaults(t *testing.T) {
	// 只提交 role：FillDefaults 填充 page/ratio/flag/count/title 后应通过校验。
	body := `{"role":"admin"}`
	rec := doPost(t, SettingsHandler, body)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
}

func TestWebRequestHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/{id}", WebRequestHandler)

	do := func(t *testing.T, target, token, formBody string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(formBody))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if token != "" {
			req.Header.Set("X-Token", token)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}
	fieldErrors := func(t *testing.T, rec *httptest.ResponseRecorder) []string {
		t.Helper()
		var resp struct {
			Errors []struct {
				Field string `json:"field"`
				Code  string `json:"code"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("响应不是合法 JSON: %v; body=%s", err, rec.Body.String())
		}
		out := make([]string, 0, len(resp.Errors))
		for _, e := range resp.Errors {
			out = append(out, e.Field+":"+e.Code)
		}
		return out
	}

	t.Run("form+header+uri+query 合法请求返回 204", func(t *testing.T) {
		rec := do(t, "/users/1?page=2", "1234567890", "username=alice")
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("form 字段缺失 -> username required", func(t *testing.T) {
		rec := do(t, "/users/1", "1234567890", "username=")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		got := fieldErrors(t, rec)
		if len(got) != 1 || got[0] != "username:required" {
			t.Fatalf("errors = %v, want [username:required]", got)
		}
	})

	t.Run("header 缺失 -> X-Token required", func(t *testing.T) {
		rec := do(t, "/users/1", "", "username=alice")
		got := fieldErrors(t, rec)
		if len(got) != 1 || got[0] != "X-Token:required" {
			t.Fatalf("errors = %v, want [X-Token:required]", got)
		}
	})

	t.Run("header 长度非法 -> X-Token len", func(t *testing.T) {
		rec := do(t, "/users/1", "short", "username=alice")
		got := fieldErrors(t, rec)
		if len(got) != 1 || got[0] != "X-Token:len" {
			t.Fatalf("errors = %v, want [X-Token:len]", got)
		}
	})

	t.Run("uri id 非法 -> 400 bind 错误", func(t *testing.T) {
		rec := do(t, "/users/abc", "1234567890", "username=alice")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("query page 非法 -> 400 bind 错误", func(t *testing.T) {
		rec := do(t, "/users/1?page=abc", "1234567890", "username=alice")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("page 缺省 -> FillDefaults 填充后通过", func(t *testing.T) {
		rec := do(t, "/users/1", "1234567890", "username=alice")
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204; body=%s", rec.Code, rec.Body.String())
		}
	})
}
