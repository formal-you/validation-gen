# Changelog

本文件记录 validation-gen 的重要变更，格式遵循
[Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### Added

- 静态校验代码生成器：gengo/v2 解析 + 自研生成器，生成无反射、可审查的
  `Validate()` / `FillDefaults()`（PR 001）。
- 公共错误模型 `errorx.FieldError` / `CollectFieldErrors` 与 runtime adapter
  （validator/v10 `StructCtx`，支持自定义 validator、dive、跨字段规则）。
- HTTP 绑定来源示例（form/header/query/uri）与 go-observability 风格 README（PR 002）。
- 代码评审问题修复：生成器只为含 `default` 字段的类型生成 `FillDefaults()`、
  「非法规则不写文件」集成测试、HTTP 显式零值用例（PR 003）。
- GitOps 工作流落地：GitHub Issue/PR 为远程权威记录、GitHub Actions CI
  与开源配套文档（CONTRIBUTING / CODE_OF_CONDUCT / SECURITY）（PR 004）。

### Changed

- 公共错误包 `valerr` 重命名为 `errorx`（避免与局部变量 `err` 冲突）。
- `main` 历史重写为线性：移除 PR 001-004 的 merge commit（内容不变），
  此后合并一律 `gh pr merge --rebase` / `--squash`，禁止 merge commit（PR 004）。

### Fixed

- `example/grpc` 注释：说明结构化错误需在转换前从 `req.Validate()` 获取。
