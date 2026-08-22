# 安全策略

## 支持的版本

| 版本 | 支持情况 |
| --- | --- |
| 最新 main（开发中） | ✅ 活跃开发 |
| v0.1.0+ | ✅ 首个正式发布后支持 |

> 仓库当前处于公开预览阶段，首个正式 tag（v0.1.0）前公共 API 可能调整。

## 报告漏洞

请**不要**在公开 Issue 中披露安全漏洞。优先使用 GitHub 的私有漏洞报告功能：

- 访问 https://github.com/formal-you/validation-gen/security/advisories
- 选择 **New draft security advisory**，填写漏洞描述、影响与复现步骤。

## 期望响应

- 确认收到：3 个工作日内；
- 漏洞评估与修复计划：14 个工作日内；
- 修复合并后会同步发布说明。

## 安全相关关注点

本项目涉及校验代码生成与 validator 语义对齐，以下方面属于安全相关：

- 生成代码中规则语义与 validator/v10 的偏差（可能绕过校验）；
- 错误消息/字段路径中的信息泄露；
- runtime adapter 对自定义 validator 实例的误用；
- 依赖链（gengo、validator、grpc）中的已知漏洞。
