package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInvalidRulesDoNotOverwrite 验证 spec「非法规则不写文件」：
// DTO 含 runtime-only 规则时生成失败，且不覆盖已有的生成文件。
func TestInvalidRulesDoNotOverwrite(t *testing.T) {
	fixture, err := filepath.Abs(filepath.Join("testdata", "invalid"))
	if err != nil {
		t.Fatal(err)
	}
	outFile := filepath.Join(fixture, "zz_generated.validation.go")
	sentinel := []byte("package invalid\n\n// 已有生成文件，非法规则不应覆盖它。\n")

	if err := os.WriteFile(outFile, sentinel, 0o644); err != nil {
		t.Fatalf("写入 sentinel 失败: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(outFile) })

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(fixture); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	var stdout, stderr bytes.Buffer
	code := run([]string{"-type=Request"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("非法规则应生成失败，实际成功；stdout=%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "dive") {
		t.Fatalf("stderr 应包含 dive 相关错误，实际: %s", stderr.String())
	}

	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("读取输出文件失败: %v", err)
	}
	if !bytes.Equal(got, sentinel) {
		t.Fatalf("已有生成文件被覆盖: got %q, want %q", got, sentinel)
	}
}
