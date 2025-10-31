# 用户部署手册
# 研发流程
## 一、核心分支定义与规范
### 1.1 分支类型与职责
| 分支名称          | 分支角色                | 核心职责                                                                 | 生命周期                     | 命名规范                          |
|-------------------|-------------------------|--------------------------------------------------------------------------|------------------------------|-----------------------------------|
| `release`         | 主集成分支（稳定分支）  | 承载当前可发布的稳定代码，作为迭代分支的基准源，最终接收迭代分支的合并成果 | 长期存在（项目生命周期内）   | 固定名称 `release`                |
| `release-<版本号>`| 迭代集成分支（临时分支）| 某一迭代的集成分支，聚合该迭代所有开发任务的代码，用于迭代测试与验证     | 单个迭代周期（从启动到发布） | `release-<主版本>.<次版本>`，例：`release-2.0` |
| `dev-task<编号>`  | 开发分支（临时分支）    | 单个开发任务的独立开发分支，承载具体功能/修复的编码工作                   | 单个任务周期（开发到合并）   | `dev-task<任务ID>-<简要描述>`，例：`dev-task001-user-login` |
| `hotfix-<版本号>` | 线上紧急修复分支（临时）| 处理线上生产环境的紧急bug，不影响当前迭代进度                             | 修复周期（1-3天）            | `hotfix-<主版本>.<次版本>.<补丁版本>`，例：`hotfix-2.0.1` |

### 1.2 分支创建基准与权限规范
- 所有分支均基于 Git 仓库 `main` 分支的上游分支体系创建，禁止跨仓库分支依赖；
- 权限控制：
  - `release` 分支：仅允许管理员/架构师执行合并操作，禁止开发人员直接提交或推送；
  - `release-<版本号>` 分支：仅允许通过合并请求（MR）合入代码，禁止直接推送；
  - 开发分支/热修复分支：开发人员拥有创建、提交、推送权限，但需遵循提交规范。

### 1.3 开发规范执行
   - 基于任务需求文档编码，遵循团队代码规范；
   - 每完成一个功能点/模块，提交代码时需遵循「语义化提交规范」，提交信息格式：`[类型](模块): 描述`；
   - 类型说明：
     `feat`（新功能）`fix`（bug修复）、`docs`（文档更新）、`refactor`（重构）、`test`（测试代码）、`chore`（构建/依赖调整）；
     示例：`feat(repo): 实现仓库创建接口`；
- 开发过程中，建议每日同步 `release-2.0` 分支的最新代码到本地开发分支，避免长期偏离导致合并冲突；

## 二、技术文档
### 2.1 文档分类与核心定义

|文档类别|包含子类型|核心作用|编写时机|
|--|--|--|--|
|业务相关技术文档|数据库表设计文档|记录业务表结构、字段说明、索引设计、关联关系|需求评审后、开发前（或开发中同步编写）|
||接口文档设计|记录接口路径、请求参数、响应格式、错误码、调用示例，供前后端/测试协作|需求评审后、开发前|
|调研型技术文档|方案调研文档|分析业务功能的技术实现方案（如“空间权限控制方案”），对比优劣并确定最终方案|复杂功能开发前、技术方案未明确时|
||技术选型文档|对比引入新技术/组件的选型（如“缓存框架选型：Redis vs Memcached”），评估可行性|引入新依赖、替换旧技术前|

# 开发部署手册
## 一、环境准备

### 1.1 基础工具安装

| 工具      | 版本要求  | 安装说明                                                                                            |
| ------- | ----- | ----------------------------------------------------------------------------------------------- |
| Git     | 2.30+ | 代码版本控制工具。Ubuntu：`sudo apt install git`；Windows/Mac：直接下载[Git 客户端](https://git-scm.com/downloads) |
| Golang  | 1.25  | 开发语言，需配置环境变量（见 1.2 小节）                                                                          |
| MySQL   | 8.0   | 数据存储服务，需初始化数据库（见 1.3 小节）                                                                        |
| VS Code | 最新稳定版 | 代码编辑器，需安装「Go」插件（Microsoft 官方）用于语法高亮与调试                                                          |
| Wire    | 最新版   | 依赖注入工具，安装命令：`go install github.com/google/wire/cmd/wire@latest`                                 |
| Task    | 3.0+  | 任务自动化工具，Ubuntu 安装见 1.4 小节                                                                       |                              

### 1.2 Golang 配置（Linux/macOS）

1. 下载 Go 1.25 安装包：
   
   访问[Go 官网下载页](https://go.dev/dl/)，选择对应系统版本（如 Linux amd64：`go1.25.linux-amd64.tar.gz`）。

2. 解压到系统目录：

```
sudo tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
```

1. 配置环境变量：
   
   编辑终端配置文件（`~/.bashrc` 或 `~/.zshrc`），添加以下内容：

```shell
# Go基础路径

export GOROOT=/usr/local/go

export GOPATH=\$HOME/go

export PATH=\$PATH:\$GOROOT/bin:\$GOPATH/bin

# 启用Go Modules

export GO111MODULE=on

# 国内代理（解决依赖下载慢）

export GOPROXY=https://goproxy.cn,direct
```

2. 生效配置并验证：

```shell
source ~/.bashrc  # 若用zsh则执行 source ~/.zshrc

go version        # 输出 "go version go1.25 xxx" 即成功
```
### 1.4 Task 安装（Ubuntu）

采用预编译二进制方式（无需依赖 Go 环境）：

1. 下载 Task 安装包：

```shell
curl -fsSL https://github.com/go-task/task/releases/latest/download/task\_linux\_amd64.tar.gz -o task.tar.gz
```

1. 解压并安装：

```shell
tar -xzf task.tar.gz  # 解压得到task二进制文件

chmod +x task         # 赋予执行权限

sudo mv task /usr/local/bin/  # 移动到系统可执行目录
```

1. 验证安装：

```shell
task --version  # 输出 "Task version: vX.X.X" 即成功
```

## 二、项目初始化

### 2.1 代码拉取

克隆项目仓库到本地：

```shell
git clone git@gitfox.zcorp.cc:Pangu/gitfox-community.git

```
### 2.2 安装依赖：

```go
go mod tidy  # 自动下载依赖并清理无用依赖
```

### 2.3 配置文件
> 按环境区分配置，本地开发使用`local`环境配置
1. 用户相关配置文件：configs/*.env,如日志路径，数据库连接
2. 工程相关配置文件：configs/*.yaml
3. dev,prod与环境相对应，dev为测试环境配置，prod为生产环境配置
4. 本地配置从configs/dev.env复制为configs/local.env,configs/dev.yaml复制为configs/local.yaml

### 2.4 vscode 本地调试文件 lanuch.json
```json
{
    "version": "0.2.0",
    "configurations": [
       {
            "name": "GitFox",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/cli/main.go",
            "args": ["server", "debug"]
        }
    ]
}
```
## 三、分支开发策略
开发分支策略与禅道保持一致

# 目录结构
以下是基于需求设计的完整代码目录结构，包含各目录职责边界及分支创建功能在不同版本的实现示例：
```plaintext
gitfox.zcorp.cc/git/Pangu/gitfox-community/
├─app                       # 按模块区分各业务实现              
│  ├─api
│  │  ├─controller
│  │  ├─middleware
│  │  └─router
│  ├─constants
│  ├─dao                    # 数据库层结构定义
│  ├─model                  # api层结构定义, request,response
│  ├─pkg
│  ├─service
│  │  └─codebase
│  │      └─branch
│  └─store
│      └─codebase
│          └─branch
├─cmd
│  └─cli                    # server服务入口
│      └─operations         # cli相关操作
├─configs                   # 配置文件           
├─scripts                   # 部署脚本
├─deploy                    # dockerfile等镜像构建脚本
├─locales                   # 多语言配置
│  ├─en-US
│  └─zh-CN
└─pkg                       # 公共工具库
    ├─config
    ├─database
    ├─errors
    ├─i18n
    ├─jwt
    ├─logger
    ├─queue
    ├─utils
    │  ├─cron
    │  └─datetime
    └─xgo
```