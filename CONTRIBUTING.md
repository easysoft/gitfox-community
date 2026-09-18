# 贡献指南

感谢你参与 GitFox Community！本文档说明本地开发、提交规范与 PR 流程。

## 开发环境

1. 安装 Go 1.25+、Git 2.30+、[Task](https://taskfile.dev/) 3.0+
2. Fork 并克隆仓库
3. 编译与测试：

```bash
task build
go build ./...
go vet ./...
task test:service
```

## 分支与提交规范

- 分支命名：`feat/<简述>`、`fix/<简述>`、`docs/<简述>`、`refactor/<简述>`
- 提交信息：`[类型](模块): 描述`
  - 类型：`feat`（新功能）、`fix`（修复）、`docs`（文档）、`refactor`（重构）、`test`（测试）、`chore`（构建/依赖）
  - 示例：`feat(repo): 实现仓库创建接口`

## Pull Request 流程

1. 从 `main` 创建功能分支
2. 保持改动聚焦，附带必要的测试
3. 确保 `go build ./...`、`go vet ./...`、`task test:service` 通过
4. 提交 PR，描述变更动机与实现方式，关联相关 Issue
5. 等待评审并根据反馈修改

## 代码风格

- 遵循 Go 官方风格；提交前建议运行 `gofmt`/`goimports`
- 不要引入新的未使用依赖；新增依赖需说明理由
- 新增或修改配置项时，同步更新 `configs/` 示例与 README

## 许可证与贡献者协议

本项目采用 ZPL 1.2 / AGPL 3.0 双许可。提交贡献即表示你同意你的贡献在上述许可下发布。
