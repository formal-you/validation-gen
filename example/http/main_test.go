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
}

func TestSettingsHandlerDefaults(t *testing.T) {
	// 只提交 role：FillDefaults 填充 page/ratio/flag/count/title 后应通过校验。
	body := `{"role":"admin"}`
	rec := doPost(t, SettingsHandler, body)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
}
