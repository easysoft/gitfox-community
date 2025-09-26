
以下是基于需求设计的完整代码目录结构，包含各目录职责边界及分支创建功能在不同版本的实现示例：
```plaintext
gitfox.zcorp.cc/demo/gitfox-pms/
├── cmd/                     # 应用程序入口点
│   ├── api/                 # API服务入口
│   │   └── main.go          # 启动API服务，初始化依赖注入
│   └── worker/              # 异步任务执行节点
│       └── main.go          # 启动工作节点，处理流水线任务等
├── internal/                # 核心业务逻辑（所有版本共享）
│   ├── domain/              # 领域模型定义（实体和基础行为）
│   │   ├── codebase/        # 代码库领域
│   │   │   └── branch/      # 分支领域模型
│   │   │       ├── branch.go  # 分支实体定义（ID、名称、仓库ID等）
│   │   │       └── store.go  # 分支与仓库的关联接口
│   ├── service/             # 业务服务接口（定义通用能力）
│   │   └── codebase/        # 代码库服务
│   │       └── branch/      # 分支服务接口
│   │           ├── service.go  # 定义BranchService接口和通用方法
│   │           └── types.go    # 定义通用请求/响应结构体
│   ├── store/          # 数据访问层（通用数据操作）
│   │   └── codebase/        # 代码库数据访问
│   │       └── branch/      # 分支数据访问
│   │           └── repo.go   # 分支CRUD通用实现
│   └── api/                 # API处理（通用路由和参数解析）
│       └── handler/         # 请求处理器
│           └── codebase/    # 代码库API处理器
│               └── branch.go  # 分支API通用路由和参数校验
|          
├── modules/                # 版本特性实现（按版本隔离）
│   ├── pms/                # 免费版功能（基础功能）
│   │   └── codebase/        # 代码库免费版功能
│   │       └── branch/      # 分支免费版功能
│   │           ├── manual_creator.go  # 手动创建分支实现
│   │           └── service_binding.go  # 免费版服务绑定
│   ├── biz/          # 企业版功能（扩展免费版）
│   │   └── codebase/        # 代码库企业版功能
│   │       └── branch/      # 分支企业版功能
│   │           ├── auto_creator.go  # 自动创建分支实现
│   │           └── service_binding.go  # 企业版服务绑定
│   └── ipd/            # 旗舰版功能（扩展企业版）
│       └── codebase/        # 代码库旗舰版功能
│           └── branch/      # 分支旗舰版功能
│               ├── webhook_trigger.go  # Webhook触发分支实现
│               └── service_binding.go  # 旗舰版服务绑定
├── pkg/                     # 公共工具库（所有版本共享）
│   ├── auth/                # 认证工具
│   ├── config/              # 配置管理
│   └── validator/           # 数据验证工具
├── plugin/                  # 插件系统
├── configs/                 # 配置文件模板
├── scripts/                 # 部署脚本
└── test/                    # 测试代码
```

目录职责与边界说明

# 一级目录职责说明（按 Go 规范与功能划分）
# cmd/
* 职责：存放应用程序的入口点，每个子目录对应一个可执行程序
* 规范要求：必须包含 main.go，仅负责程序启动、依赖组装，不包含复杂业务逻辑
* 示例：cmd/api/main.go 启动 HTTP 服务，cmd/worker/main.go 启动任务处理节点
# internal/
* 职责：存放应用私有代码，仅本项目可使用，不可被外部导入（Go 规范强制）
* 边界：包含所有业务核心逻辑，但不包含版本差异代码，是各版本共用的 "骨架"
* 核心子目录：domain（领域模型）、service（服务接口）、repository（数据访问）等
# pkg/
* 职责：存放可导出的公共工具库，可被本项目和外部项目依赖
* 规范要求：代码必须通用、无业务耦合，如工具类、基础组件
* 示例：JWT 工具、配置解析、数据验证等
# modules/
* 职责：按版本存放功能实现，实现版本化差异（核心设计创新点）
* 层次关系：core（免费版）→ enterprise（企业版）→ ultimate（旗舰版），高版本依赖低版本
* 边界：仅包含版本特有功能，复用 internal/ 的基础能力
# api/
* 职责：存放 API 规范定义（符合 Go 规范）
* 内容：OpenAPI 文档、gRPC proto 文件等接口契约
configs/
* 职责：存放配置文件模板，按环境或版本划分
* 示例：默认配置、企业版专属配置、私有化部署配置等
# deploy/
* 职责：部署相关配置（符合 Go 规范）
* 内容：Dockerfile、Kubernetes manifests、部署脚本等
# docs/
* 职责：项目文档，包括用户手册、API 文档、开发指南等
# test/
* 职责：存放额外的测试代码（符合 Go 规范），补充 *_test.go 的不足
* 内容：端到端测试、集成测试、测试数据等


# cmd/
* 职责：应用程序入口点，负责初始化依赖、启动服务
* 边界：不包含业务逻辑，仅负责组装和启动系统
* 示例：cmd/api/main.go 启动 HTTP 服务，cmd/worker/main.go 启动任务处理节点
# internal/（核心共享逻辑）
## domain/
* 职责：定义业务实体（如 Codebase、Branch、Pipeline）及核心属性、行为，不依赖外部组件
* 边界：不包含具体业务规则，仅定义 "是什么"
## service/
* 职责：定义业务服务接口（如 BranchService、PipelineService），抽象业务能力，不包含具体实现
* 边界：不包含版本差异逻辑，仅定义通用接口
* 示例：service.go定义CreateBranch接口，不关心具体实现
## repository/
* 职责：提供通用数据访问能力，封装数据库操作
* 边界：与具体业务规则无关，仅负责数据存储和查询
* 示例：repo.go实现分支的数据库 CRUD 操作
## api/
* 职责：提供通用 API 处理能力，包括路由和参数校验
* 边界：不包含版本差异的业务逻辑，仅处理 HTTP 层通用逻辑
# modules/（版本特性）
## pms/（免费版）
* 职责：实现基础功能，作为高版本的基础
* 边界：仅包含免费版可用功能，不依赖其他版本
* 核心文件：manual_creator.go实现手动创建分支逻辑
## biz/（企业版）
* 职责：在免费版基础上扩展功能，复用免费版能力
* 边界：仅包含企业版新增功能，依赖免费版但不修改其代码
* 核心文件：auto_creator.go实现自动创建逻辑，复用手动创建能力
## ipd/（旗舰版）
* 职责：在企业版基础上增强功能，复用低版本能力
* 边界：仅包含旗舰版特有功能，依赖企业版和免费版
* 核心文件：webhook_trigger.go实现 Webhook 触发逻辑
# pkg/
* 职责：提供通用工具和组件，被所有模块共享
* 边界：与业务逻辑无关，仅提供技术能力（如认证、配置）



二级目录核心职责（以关键目录为例）
# internal/ 子目录
## internal/domain/
定义业务实体（如 Codebase、Branch、Pipeline）及核心属性、行为，不依赖外部组件
## internal/service/
定义业务服务接口（如 BranchService、PipelineService），抽象业务能力，不包含具体实现
## internal/repository/
定义数据访问接口（如 BranchRepository），隔离业务逻辑与存储实现
## internal/api/
实现 API 层逻辑，包括路由（router/）、处理器（handler/）、中间件（middleware/）
## internal/pkg/
内部公共功能，如权限校验（auth/）、错误处理（error/），被 internal/ 和 modules/ 

# modules/ 子目录
## modules/pms/
实现免费版核心功能，如分支手动创建（pms/codebase/branch/manual.go）
## modules/biz/
仅实现企业版新增功能（如分支自动创建），通过组合 pms 模块复用基础能力
## modules/ipd/
仅实现旗舰版新增功能（如 Webhook 触发分支），依赖 biz 和 pms 模块


严格遵循 Go 规范：保留 cmd/、internal/、pkg/ 等强制目录，符合社区最佳实践，降低团队协作成本
* 版本化功能隔离：通过 modules/ 实现版本差异，高版本自然复用低版本功能，避免代码重复
* 清晰的依赖关系：modules/ 依赖 internal/，internal/ 依赖 pkg/，单向依赖无循环
* 业务与技术分离：domain/ 纯业务模型，repository/ 处理数据存储，service/ 协调业务流程，符合领域驱动设计思想