## 关联 Issue

Closes #<issue>

## 摘要

<!-- 一句话说明这个 PR 改了什么、为什么。 -->

## 变更清单

- [ ] <!-- 变更点 1 -->
- [ ] <!-- 变更点 2 -->

## 验收

- [ ] `go generate ./...` 通过，生成文件再生成无 diff（golden 门禁）
- [ ] `go build ./...` / `go vet ./...` 通过
- [ ] `go test ./...` / `go test -race ./...` 全绿
- [ ] `git diff --check` 无空白错误
- [ ] 涉及 DTO/生成器时，`zz_generated.validation.go` 与声明同一 commit 提交

> 完整流程见 [docs/gitops/README.md](docs/gitops/README.md)。
