# GitFox Community

GitFox 是一个用 Go 编写的一体化代码托管与 DevOps 平台，提供 Git 仓库托管、代码评审、CI/CD 流水线、制品库、代码扫描等能力，可与禅道（ZenTao）等项目管理/研发流程系统集成。

> 本仓库为 **社区版（Community）**。AI 代码评审、分支归档、基于 ztflow 的代码推送请求属于商业版功能，不在本仓库中；社区版保留其接口与空实现。

## 功能特性

- **代码托管**：仓库、分支、标签、提交、合并请求（Pull Request）、Webhook、SSH/HTTP 协议访问
- **代码评审**：合并请求评审流程、分支保护规则、Code Owner
- **CI/CD**：内置流水线（pipeline）引擎与 Runner（支持 Docker），支持 Cron 定时任务
- **制品库**：容器镜像、Helm、Raw 等制品存储与回收站/GC
- **代码扫描**：规则/规则集/扫描方案/扫描计划管理与扫描任务
- **空间与权限**：多空间（Space）、RBAC 权限、SSO/Token 认证
- **仓库导入**：从其他平台导入/镜像同步仓库
- **国际化**：中英文界面（zh-CN / en-US）

## 技术栈

- 语言：Go 1.25
- HTTP：Gin
- 依赖注入：Google Wire
- ORM：GORM（MySQL / GaussDB / SQLite / 达梦）
- 任务编排：内置 Job Scheduler
- 构建：Taskfile（`task`）

## 快速开始

### 1. 环境要求

| 工具 | 版本 | 说明 |
| --- | --- | --- |
| Go | 1.25+ | 编译与开发 |
| Git | 2.30+ | 运行时依赖 |
| Task | 3.0+ | 构建任务（`go install github.com/go-task/task/v3/cmd/task@latest`） |
| MySQL | 8.0+ | 数据存储（也支持 SQLite/GaussDB/达梦） |
| Docker | 最新稳定版 | Runner 执行流水线时需要 |

### 2. 编译

```bash
git clone <本仓库地址>
cd gitfox-community
task build          # 产物位于 bin/gitfox_linux
```

### 3. 配置

配置支持 `yaml` 与 `env` 两种形式，示例见 `configs/prod.yaml`、`configs/prod.env`。

关键环境变量：

| 变量 | 说明 |
| --- | --- |
| `GITFOX_SERVER_HOST` | 对外访问域名/IP |
| `GITFOX_HTTP_PORT` / `GITFOX_SSH_PORT` | HTTP / SSH 端口 |
| `GITFOX_DB_DRIVER` | 数据库驱动：`mysql` / `sqlite` / `gaussdb` / `dm` |
| `GITFOX_DB_HOST` / `GITFOX_DB_PORT` / `GITFOX_DB_USER` / `GITFOX_DB_PASSWORD` / `GITFOX_DB_NAME` | 数据库连接 |
| `GITFOX_PIPELINE_AUTH_PASSWORD` | 流水线内部认证密码，**必须显式配置** |
| `GITFOX_CI_PLUGINS_ZIP_URL` | CI 插件包下载地址 |

> 首次启动会自动执行数据库迁移。请勿使用示例中的占位密码。

### 4. 启动

```bash
./bin/gitfox_linux server prod
```

### 5. Docker

```bash
# 需要先 task build，并准备 resources/ciPlugins/plugins.tar.gz
docker build -f deploy/Dockerfile -t gitfox-community .
docker compose -f deploy/docker-compose.yaml up -d
```

> `deploy/Dockerfile` 与流水线/扫描相关的容器镜像默认指向内部镜像仓库，公开部署时请替换为可访问的镜像地址（见下节）。

## 容器镜像说明

流水线执行与代码扫描会拉取一批容器镜像。当前代码中的默认镜像地址为内部仓库，社区部署时需按需替换：

- 语言/构建基础镜像（golang、node、gradle、alpine、docker-cli 等）可替换为 Docker Hub 官方镜像
- 插件类镜像（git/scanner/uploader/pipeline-helper 等）需替换为你自建的镜像地址

相关位置：`pkg/utils/common/const.go`、`pkg/pipeline/step/base/*.go`、`pkg/runner/docker/engine2/compiler/*.go`、`resources/pipeline/codescan*.yaml`。

## 开发指南

```bash
task test:service     # 运行 Service 层单测并生成覆盖率
go build ./...        # 全量编译
go vet ./...          # 静态检查
```

- 代码结构：`app/`（业务：api/controller、service、store、model、pipeline 等）、`pkg/`（通用组件）、`cmd/`（入口）、`resources/`（内置资源）
- 依赖注入：修改 provider 后运行 `wire ./cmd/cli/operations/server` 重新生成
- 提交规范：`[类型](模块): 描述`，类型见 `CONTRIBUTING.md`

## 许可证

本项目采用 **ZPL 1.2 / AGPL 3.0 双许可**，你可任选其一。详见 [LICENSE](./LICENSE)。

Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)

### 第三方开源代码

本项目部分代码派生自其他开源项目，相关文件**保留其原始版权与许可头**，具体如下：

| 来源 | 许可 | 主要位置 |
| --- | --- | --- |
| Harness, Inc. | Apache License 2.0 | `pkg/git/`、`pkg/livelog/`、`pkg/stream/`、`pkg/ssh/` 等 |
| Drone.IO Inc. | Polyform License（以源文件头为准） | `pkg/runner/docker/engine2/`、`pkg/pipeline/spec/parse/` 等 |
| Docker, Inc. | Apache License 2.0 | `pkg/runner/docker/internal/docker/stdcopy/` |

本项目还依赖 `go.mod` 中列出的第三方 Go 模块，各模块遵循其各自的许可协议。分发本项目时，请一并遵守上述第三方许可条款。

## 贡献

欢迎提交 Issue 与 Pull Request，请先阅读 [CONTRIBUTING.md](./CONTRIBUTING.md)。
