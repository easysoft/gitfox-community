# 制品库（Artifact）迁移蓝图

> 源项目：`/root/repo/gitfox`（Harness/Gitfox Pro 架构：`api/controller + services + store/database`）
> 目标项目：`/root/repo/pangu/gitfox-community`（分层架构：`router → controller → service → store → model` + 核心流程放 `app/artifact`）
> 目标：**接口走 router/controller/service/store，核心处理逻辑放 app/artifact，Wire 可注入，正常启动**。

---

## 一、结论先行

| 项 | 结论 |
|---|---|
| 迁移基调 | **"搬 + 改"，不是"抄"**：源项目有两套制品架构叠加，只迁新套（`gitfox/app/artifact/*` + `services/artifact/*`），旧套（`api/controller/artifact`）丢弃 |
| 总代码量 | ≈ 150 个 Go 文件 / ≈ 8800 行（含测试），**建议分 5 阶段落地**，首阶段只上 generic + raw 格式打通链路 |
| 存储复用 | ✅ 目标项目已有 `pkg/storage`（filesystem + inmemory），**直接复用，不重搬 `gitfox/pkg/storage`** |
| 数据库迁移 | ✅ 目标项目已有 `0032_create_artifact_tables.up.sql`（views/groups/packages/versions），**后续补 assets/blobs/meta_assets/tree_nodes/repositories**；源项目 20 个迁移脚本需合并到当前项目迁移序号体系 |
| 架构重映射 | `gitfox.app/services/artifact` → 本项目 `app/service/artifact`（业务服务）<br>`gitfox.app/artifact/controller` → 本项目 `app/api/controller/artifact`（HTTP 控制器）<br>`gitfox.app/artifact/{adapter,authorizer,handler,storage,config}` → 本项目 `app/artifact/*`（核心处理逻辑，对齐 `app/pipeline` 范例）<br>`gitfox.app/store/database/artifacts` → 本项目 `app/store/artifact`<br>`gitfox.app/api/handler/artifact` → 合并进本项目 `app/api/controller/artifact`（当前项目 controller 直接写 gin handler，无独立 handler 层） |
| 路由风格 | 用本项目 gin 风格（`registerArtifactRoutes(r *gin.RouterGroup, controller *Controller)`），**不搬源项目的 go-chi** |
| Wire 注入 | 新增 `app/service/artifact/wire.go` + `app/store/artifact/wire.go` + `app/artifact/wire.go` + `app/api/controller/artifact/wire.go`，在 `cmd/cli/operations/server/wire.go` 的 `NewApp` 中组装；provider 侧增 `ArtifactWireSet` |

**风险前置**：
- 源项目依赖 `github.com/easysoft/gitfox/pkg/storage.ContentStorage`，目标项目 `pkg/storage` 目前只暴露 `driver.StorageDriver`，**未暴露 `ContentStorage` 高层抽象**。迁移前需先在 `pkg/storage` 增 `ContentStorage` + `CommonContentStore` 两个类型（见第四章第 0 步）。
- 源项目 `services/artifact` 依赖 `job.Scheduler` + `sse.Streamer`。目标项目已有 `app/service/jobs.Scheduler` 与 `app/service/sse.*`，直接复用，但构造函数签名有差异，需适配。
- 源项目 `artifact/controller` 依赖 `authz.Authorizer` + `authorizer.Authorizer`（两个 authz），目标项目 `pkg/auth/authz` 已有。RBAC 权限模型（`internal/rbac/artifact`）需评估是否必需，**首阶段建议先按空间粒度授权，不搬 node 级 RBAC**。

---

## 二、源码盘点（gitfox 侧）

### 2.1 `app/artifact/`（核心处理逻辑 — 对应本项目 `app/artifact/`）

```
app/artifact/
├── config.go                # Config / ConfigStorage
├── storage.go               # LoadContentStorage
├── wire.go                  # ProvideArtifactLoadStorage
├── adapter/                 # 各格式适配器（核心逻辑核心）
│   ├── adapter.go / asset.go / hash.go / hosted.go / error.go / metadata.go
│   ├── storage_file_writer.go
│   ├── model_mgr.go
│   ├── raw/                 # raw（Maven 风格）
│   ├── generic/             # generic（文件/目录）
│   ├── container/           # OCI / Docker 镜像（含 schema、gc、token）
│   ├── helm/                # Helm Chart
│   ├── request/             # 请求抽象层
│   └── testsuite/           # 测试工具
├── authorizer/              # 鉴权抽象
│   ├── authorizer.go
│   └── wire.go
├── controller/              # 核心控制器（业务编排）
│   ├── controller.go
│   ├── container_*.go       # 容器镜像相关（8 个文件）
│   ├── helm.go
│   ├── generic.go
│   ├── archive.go
│   ├── list_*.go / create_group.go / soft_remove.go / refresh_tree.go
│   ├── garbagecollect.go / statistic*.go / permission_*.go / util.go / ...
└── handler/                 # 协议层 handler
    ├── handle_container.go / handle_container_tag.go / handle_container_token.go
    ├── generic.go / raw.go / helm.go / upload.go
    └── helm_test.go
```

### 2.2 `app/services/artifact/`（异步服务 — 对应本项目 `app/service/artifact/`）

```
services/artifact/
├── service.go               # Service 结构（包含 tx, artStore, fileStore, settings, scheduler, sseStreamer）
├── wire.go                  # ProvideArtifactSweepSvc
├── handle_container.go      # 容器镜像异步清理
├── handle_soft_remove.go    # 软删除后的异步清理
├── sync_tree_node.go        # 树节点同步
├── input.go / store_option.go
```

### 2.3 `app/api/controller/artifact/` + `app/api/handler/artifact/`

- `api/controller/artifact` 是**旧路径**，只有 `controller.go/wire.go`，ASO 路由用的也不是它。
- `api/handler/artifact` 的 22 个文件是**新路径**，每个 handler 一个 HTTP 入口，调用 `app/artifact/controller`。
- **两层合并**：目标项目 controller 直接承担 handler 职责，把这 22 个 handler 合进 `app/api/controller/artifact/*.go`。

### 2.4 `app/store/database/artifacts/` → 本项目 `app/store/artifact/`

```
artifact.go              # 聚合 Store
artifact_repository.go   # 制品仓库
artifact_group.go        # 分组
artifact_view.go         # 视图
artifact_package.go      # 包
artifact_version.go      # 版本
artifact_asset.go        # 资产
artifact_meta_asset.go   # 元资产
artifact_blob.go         # Blob（物理文件）
artifact_node.go         # 树节点
artifact_upstream.go     # 上游代理
artifact_option.go       # 查询选项
const.go / validator.go
```

### 2.5 `app/events/artifact/` → 本项目 `app/events/artifact/`

```
events.go reporter.go reader.go create.go update.go delete.go wire.go
```

### 2.6 类型与迁移

- `gitfox/types/artifact.go` + `artifact_opt.go` + `artifact_resp.go` + `enum/artifact.go` → 本项目 `app/model/artifact/*.go`（按本项目命名风格，分文件）
- 20 个 migration SQL → 合并到 `app/store/migrate/mysql/0033_*` 起步（当前项目 0032 已有部分表）

---

## 三、目标目录结构（本项目侧）

```
app/
├── api/
│   ├── controller/
│   │   └── artifact/                  # ★ 新增，HTTP Controller（直接写 gin handler）
│   │       ├── controller.go          # 结构体 + 构造函数（从 gitfox.app/artifact/controller.go 改造）
│   │       ├── wire.go                # wire.NewSet(NewController)
│   │       ├── list_tree.go           # ← gitfox.api/handler/artifact/list_tree.go
│   │       ├── list_assets.go        # ← list_assets.go
│   │       ├── list_versions.go      # ← list_versions.go
│   │       ├── list_generic_assets.go # ← list_generic_assets.go
│   │       ├── list_asset_history.go  # ← list_asset_history.go
│   │       ├── list_container.go      # ← list_container.go
│   │       ├── list_repos.go          # ← list_repos.go
│   │       ├── list_node_info.go      # ← list_node_info.go
│   │       ├── generic_group_tree.go  # ← generic_group_tree.go
│   │       ├── create_generic_group.go
│   │       ├── archive_group.go
│   │       ├── refresh_tree.go
│   │       ├── soft_remove_by_node.go
│   │       ├── statistic_container.go
│   │       ├── statistic_summary.go
│   │       ├── statistic_capacity.go
│   │       ├── garbagecollect.go
│   │       ├── get_space_members.go
│   │       ├── get_space_permissions.go / set_space_permission.go
│   │       └── get_node_permissions.go / set_node_permission.go
│   └── router/
│       └── artifact.go                # ★ 新增 registerArtifactRoutes
│
├── artifact/                          # ★ 新增，核心处理逻辑（对齐 app/pipeline）
│   ├── config.go                      # ← gitfox.app/artifact/config.go（重写 import）
│   ├── storage.go                     # ← gitfox.app/artifact/storage.go
│   ├── wire.go                        # ProvideArtifactLoadStorage
│   ├── adapter/                       # ← gitfox.app/artifact/adapter/*（全量移植）
│   ├── authorizer/                    # ← gitfox.app/artifact/authorizer/*
│   ├── core/                          # ← gitfox.app/artifact/controller/*（改包名 controller → core，避开与 api/controller 混淆）
│   │   ├── controller.go / container_*.go / helm.go / generic.go / ...
│   └── handler/                       # ← gitfox.app/artifact/handler/*（保持协议 handler 层，被 api/controller 调用）
│
├── service/
│   └── artifact/                      # ★ 新增，业务服务
│       ├── interface.go               # Interface 定义（对齐 app/service/pipeline/interface.go 风格）
│       ├── service.go                 # ← gitfox.app/services/artifact/service.go
│       ├── wire.go                    # ProvideService → NewArtifactService
│       ├── handle_container.go        # ← handle_container.go
│       ├── handle_soft_remove.go      # ← handle_soft_remove.go
│       ├── sync_tree_node.go          # ← sync_tree_node.go
│       ├── input.go / store_option.go
│
├── store/
│   └── artifact/                      # ★ 新增，存储层
│       ├── interface.go               # Interface（聚合 ArtifactStore）
│       ├── wire.go                    # ProvideArtifactStore
│       ├── artifact.go                # Store 聚合体
│       ├── artifact_repository.go / artifact_group.go / artifact_view.go /
│       ├── artifact_package.go / artifact_version.go / artifact_asset.go /
│       ├── artifact_meta_asset.go / artifact_blob.go / artifact_node.go /
│       ├── artifact_upstream.go / artifact_option.go / const.go / validator.go
│       └── mock/
│
├── model/
│   └── artifact/                      # ★ 新增，数据模型
│       ├── artifact.go                # ← gitfox/types/artifact.go 拆分
│       ├── artifact_opt.go
│       ├── artifact_resp.go
│       └── enum.go                    # ← gitfox/types/enum/artifact.go
│
├── events/
│   └── artifact/                      # ★ 新增，事件总线
│       ├── events.go / reporter.go / reader.go
│       ├── create.go / update.go / delete.go
│       └── wire.go
│
└── constant/
    └── artifact.go                    # ★ 新增：路径参数常量（PathParamArtifactRepoRef 等）

app/store/migrate/mysql/               # 新增 SQL（0033_ 起步）
  0033_create_artifact_assets.up.sql
  0034_create_artifact_blobs.up.sql
  0035_create_artifact_meta_assets.up.sql
  0036_create_artifact_tree_nodes.up.sql
  0037_create_artifact_repositories.up.sql
  0038_create_artifact_upstream.up.sql
  0039_artifact_post_migrations.up.sql  # 聚合 gitfox 0005_*/0007_*/0080_* 的 alter

pkg/storage/                           # ★ 扩展现有包
  content.go                           # 新增 ContentStorage 接口
  content_option.go                    # 新增 Option 模式
  file_reader.go                       # 新增 FileReader
  common_store.go                      # 新增 CommonContentStore
```

---

## 四、分阶段落地清单

### 阶段 0：**基建补齐**（前置，必须先做）

> 目标：让 `pkg/storage` 暴露 `ContentStorage` 高层接口，避免 `services/artifact` 直接用底层 driver。

1. 复制 `gitfox/pkg/storage/content.go` → `pkg/storage/content.go`
2. 复制 `gitfox/pkg/storage/content_option.go` → `pkg/storage/content_option.go`
3. 复制 `gitfox/pkg/storage/file_reader.go` → `pkg/storage/file_reader.go`
4. 扩展 `pkg/storage/wire.go`：新增 `ProvideContentStorage(cfg *config.Config, driver driver.StorageDriver) (ContentStorage, error)`，加入 `WireSet`
5. 本地编译验证：`go build ./pkg/storage/...`

**验收**：`pkg/storage.ContentStorage` 能被 `wire.Build` 成功注入。

---

### 阶段 1：**数据层 + 模型**（无业务依赖，可独立验证）

1. **Model 迁移**
   - 创建 `app/model/artifact/`，从 `gitfox/types/artifact*.go` + `enum/artifact.go` 拆分（按本项目命名：gorm tag 保持源项目 snake_case，JSON tag 保持源项目风格）
   - 字段校验：与当前 `0032_create_artifact_tables.up.sql` 表结构对齐；发现 gitfox 用 `ops_artifact_views.id` 是 int unsigned、`null.Int`，本项目 `model/pipeline` 惯用 `int64` + `*int64`，**统一用 int64**

2. **Migration 迁移**
   - 复制 gitfox 20 个 sql 的**净效果**（去掉过渡性 0005_*、0080_*），合并为 7 个 `0033_–0039_` up/down 迁移
   - 差集：当前 0032 有 views/groups/packages/versions，本阶段补 assets/blobs/meta_assets/tree_nodes/repositories/upstream
   - 与 `app/store/migrate/migrate.go` 注册器对齐

3. **Store 迁移**
   - 建 `app/store/artifact/`，抄 `gitfox/app/store/database/artifacts/*.go`
   - 改造点：
     - import：`github.com/easysoft/gitfox/*` → `gitfox.zcorp.cc/git/Pangu/gitfox-community/*`
     - `types.*` → `app/model/artifact.*`
     - `database.ProcessSQLErrorf` 当前项目已有 `pkg/database.ProcessSQLErrorf`，直接对应
     - `dbtx.Transactor`：当前项目无此抽象，用 `*gorm.DB` + 在 service 层 `db.Transaction(func(tx *gorm.DB) error{...})` 替代
   - 写 `interface.go`：`ArtifactStore` 聚合所有子接口（`Repositories()/Groups()/Views()/Packages()/Versions()/Assets()/MetaAssets()/Blobs()/Nodes()/Upstream()`）
   - 写 `wire.go`：`ProvideArtifactStore(db *gorm.DB) Interface`

**验收**：
```bash
go build ./app/model/artifact/... ./app/store/artifact/...
go test ./app/store/artifact/...  # 可移植源项目的 *_test.go
```

---

### 阶段 2：**app/artifact 核心**（核心处理逻辑）

1. 复制 `gitfox/app/artifact/{config.go,storage.go,wire.go}` → `app/artifact/`
2. 复制 `gitfox/app/artifact/adapter/` → `app/artifact/adapter/`（全量）
3. 复制 `gitfox/app/artifact/authorizer/` → `app/artifact/authorizer/`
4. 复制 `gitfox/app/artifact/controller/` → `app/artifact/core/`（**改包名为 `core`**，避免与 `app/api/controller` 命名冲突）
5. 复制 `gitfox/app/artifact/handler/` → `app/artifact/handler/`
6. 批量 import 改写：
   ```
   github.com/easysoft/gitfox/app/store       → .../app/store/artifact
   github.com/easysoft/gitfox/app/artifact    → .../app/artifact
   github.com/easysoft/gitfox/app/services    → .../app/service
   github.com/easysoft/gitfox/pkg/storage     → .../pkg/storage
   github.com/easysoft/gitfox/types           → .../app/model/artifact
   github.com/easysoft/gitfox/types/enum      → .../app/model/artifact（合并到 enum.go）
   github.com/easysoft/gitfox/store/database/dbtx → 移除，替换为 gorm 原生事务
   github.com/easysoft/gitfox/internal/rbac/artifact → 首阶段砍掉（用空间级权限兜底），第 5 阶段补
   ```
7. 写 `app/artifact/wire.go`：聚合 adapter/authorizer/core/handler 的 ProvideXxx 到 `WireSet`
8. 首阶段可砍的模块：`adapter/container/*`（OCI 协议复杂，放阶段 4）、`adapter/helm/*`（放阶段 4），只保留 `raw` + `generic`

**验收**：`go build ./app/artifact/...` 通过。

---

### 阶段 3：**service + controller + router**（打通 HTTP 链路）

1. **Service 层**：`app/service/artifact/`
   - 复制 `gitfox/app/services/artifact/` 全量
   - 改造 `service.go`：依赖改为本项目 `app/service/jobs.Scheduler` + `app/service/sse.*` + `pkg/storage.ContentStorage` + `app/store/artifact.Interface`
   - 写 `interface.go`：对齐 `app/service/pipeline/interface.go` 的定义风格，暴露 `StatisticContainer/SoftRemove/SyncTreeNode/...` 等方法
   - 写 `wire.go`：`var WireSet = wire.NewSet(NewArtifactService, wire.Bind(new(Interface), new(*Service)))`

2. **Controller 层**：`app/api/controller/artifact/`
   - 把 `gitfox/app/api/handler/artifact/*.go` 的 22 个 handler **合并改写**为本项目 gin 风格：
     ```go
     // 原 handler (chi/net-http):
     //   func HandListAssets(artCtl *artctl.Controller, format types.ArtifactFormat) http.HandlerFunc {...}
     // 改为:
     func (c *Controller) ListAssets(g *gin.Context) {
         format := types.ArtifactFormat(g.Param("format"))
         ... c.svc.ListAssets(...) ...
         g.JSON(http.StatusOK, data)
     }
     ```
   - `controller.go` 结构体依赖：`artifactSvc artifact.Interface` + `urlProvider` + `authorizer` + `coreCtl *core.Controller`
   - 写 `wire.go`：`ProvideController(...)`

3. **Router 层**：`app/api/router/artifact.go`
   - 参照 `pipeline.go` 的 `registerPipelineRoutes` 格式：
     ```go
     func registerArtifactRoutes(r *gin.RouterGroup, controller *Controller) {
         art := r.Group("/spaces/:spaceRef/artifacts")
         {
             art.GET("/tree", controller.Artifact.ListTree)
             art.POST("/nodeInfo", controller.Artifact.ListNodeInfo)
             art.POST("/refresh_nodes", controller.Artifact.RefreshTree)

             format := art.Group("/formats/:format")
             {
                 format.GET("/versions", controller.Artifact.ListVersions)
                 format.GET("/assets", controller.Artifact.ListAssets)
             }
             // ... 其余路由按 gitfox/app/router/api.go#L280-330 映射
         }
     }
     ```
   - 在 `app/api/router/controller.go` 的 `Controller` 结构体中追加：`Artifact *artifact.Controller`
   - 在 `app/api/router/router.go` 合适位置调用 `registerArtifactRoutes(authGroup, controller)`

4. **Constant 层**：`app/constant/artifact.go`：定义 `PathParamArtifactRepoRef`、`PathParamArtifactFormat`、`PathParamArtifactNodeID` 等

**验收**：`go build ./...` 通过；启动服务能看到 `/api/v1/spaces/.../artifacts/tree` 路由已挂载。

---

### 阶段 4：**事件 + 异步 + 容器/Helm 协议补齐**

1. 迁移 `app/events/artifact/`：对齐本项目 `app/events/pipeline/` 写法（pubsub 基于 `pkg/pubsub`）
2. 迁移 `adapter/container/` + `adapter/helm/`：依赖 OCI/Helm 解析
3. 异步清理 job：对齐 `app/service/jobs` 的 Job 注册方式，注册 `artifact-sweep` 定时任务

---

### 阶段 5：**RBAC 精细化（可选）**

- 迁移 `gitfox/internal/rbac/artifact` → `app/artifact/rbac/`
- 评估是否接入现有 `pkg/auth/authz`，或独立 node 级权限表
- 非关键路径，**可延后**

---

## 五、Wire 注入完整蓝图

### 5.1 各模块 WireSet

```go
// pkg/storage/wire.go（扩展）
var WireSet = wire.NewSet(
    ProvideStorageDriver,
    ProvideContentStorage,  // ★ 新增
)

// app/store/artifact/wire.go
var WireSet = wire.NewSet(
    ProvideArtifactStore,   // func(db *gorm.DB) Interface
)

// app/artifact/wire.go
var WireSet = wire.NewSet(
    ProvideConfig,          // func(*config.Config) Config
    ProvideArtifactLoadStorage, // func(ctx, Config, storage.Config) (ContentStorage, error)
    authorizer.WireSet,
    core.WireSet,           // NewController
    handler.WireSet,
    adapter.WireSet,
)

// app/service/artifact/wire.go
var WireSet = wire.NewSet(
    NewArtifactService,
    wire.Bind(new(Interface), new(*Service)),
)

// app/events/artifact/wire.go
var WireSet = wire.NewSet(
    ProvideReporter,
    ProvideReader,
)

// app/api/controller/artifact/wire.go
var WireSet = wire.NewSet(
    ProvideController,      // func(svc Interface, core *core.Controller, ...) *Controller
)
```

### 5.2 Provider 聚合

在 `cmd/cli/operations/provider/` 新增 `artifact.go`：

```go
package provider

import (
    "github.com/google/wire"

    artifactApp "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact"
    artifactCtrl "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/artifact"
    artifactEvents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/artifact"
    artifactSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/artifact"
    artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
)

var ArtifactWireSet = wire.NewSet(
    artifactStore.WireSet,
    artifactApp.WireSet,
    artifactSvc.WireSet,
    artifactEvents.WireSet,
    artifactCtrl.WireSet,
)
```

### 5.3 `cmd/cli/operations/server/wire.go` 改动

```go
type App struct {
    // ... 已有字段
    ArtifactService artifactSvc.Interface  // ★ 新增
    ArtifactStore   artifactStore.Interface // ★ 新增
}

func NewApp(mode string) (*App, func(), error) {
    panic(wire.Build(
        // ... 已有 WireSet
        provider.ArtifactWireSet,  // ★ 新增
        wire.Struct(new(App), "*"),
    ))
}
```

同步修改 `app/api/router/controller.go`：

```go
type Controller struct {
    // ... 已有字段
    Artifact *artifact.Controller  // ★ 新增
}
```

### 5.4 生成 wire_gen.go

```bash
cd cmd/cli/operations/server
wire
# 或 go generate ./...
```

---

## 六、配置文件 & Config 结构

在 `pkg/config/config.go`（或对应 Config 定义处）新增：

```go
type ArtifactConfig struct {
    Storage ArtifactStorageConfig `yaml:"storage"`
}

type ArtifactStorageConfig struct {
    Provider string `yaml:"provider"`  // "local" or "s3"
    Prefix   string `yaml:"prefix"`    // 存储目录前缀
}

type Config struct {
    // ... 已有字段
    Artifact ArtifactConfig `yaml:"artifact"`  // ★ 新增
}
```

`configs/*.yaml` 样例新增：

```yaml
artifact:
  storage:
    provider: local
    prefix: artifact
```

---

## 七、迁移执行顺序（按天切分）

| 天 | 阶段 | 关键产出 | 验收命令 |
|---|---|---|---|
| D1 | 阶段 0 | `pkg/storage.ContentStorage` 可注入 | `go build ./pkg/storage/...` |
| D2 | 阶段 1（Model + Migration） | 表建起来，model 能映射 | `go build ./app/model/... && make migrate-up` |
| D3 | 阶段 1（Store） | `ArtifactStore` 可独立注入 | `go test ./app/store/artifact/...` |
| D4–D5 | 阶段 2 | `app/artifact/core` 能编译 | `go build ./app/artifact/...` |
| D6 | 阶段 3（Service） | service 可注入 | `go build ./app/service/artifact/...` |
| D7 | 阶段 3（Controller + Router） | HTTP 路由可通 | `curl /api/v1/spaces/xxx/artifacts/tree` |
| D8 | 阶段 3（Wire 全量通关） | `wire` 生成成功 | `go run ./cmd/cli/operations/server` 能起 |
| D9 | 阶段 4 | 事件 + 容器/Helm | 镜像 push/pull E2E |
| D10 | 阶段 5（可选） | RBAC | 权限校验 E2E |

---

## 八、回滚策略

- 每阶段一个 feature branch：`feat/artifact-migration-p0` … `p5`
- 阶段 3 未完成前，`provider.ArtifactWireSet` 不注册进 `NewApp`，互不影响启动
- Migration 全部准备 down 脚本
- 启动前置开关：`config.Artifact.Enabled=false` 时，router 层跳过 `registerArtifactRoutes`

---

## 九、Checklist（开干前自检）

- [ ] 阶段 0 产出：`pkg/storage.ContentStorage` 已落地并能被 wire 注入
- [ ] 阶段 1 产出：所有 model 已在本项目，所有 store 单测通过
- [ ] 阶段 2 产出：`app/artifact/` 编译通过，首批只含 raw/generic
- [ ] 阶段 3 产出：`registerArtifactRoutes` 在 `router.go` 中被调用；`Controller` 结构体含 `Artifact`；`NewApp` 含 `ArtifactWireSet`；`wire_gen.go` 重新生成无冲突
- [ ] 启动验证：`./bin/gitfox server` 无 panic，`/healthz` 正常，`/api/v1/spaces/.../artifacts/tree` 返回 200（即使空数据）
- [ ] 单测：至少 store 层单测全过
- [ ] 集成测：generic 格式上传/下载/列表 E2E 通过

---

## 十、风险 & 下一步

**风险**：
1. `dbtx.Transactor` 抽象差异，阶段 1 Store 层会涉及大面积事务写法改造
2. `authz.Authorizer` 签名差异：gitfox 用 `Check(ctx, session, scope, resource, permission)`，本项目 `pkg/auth/authz.Authorizer.Authorize(...)` 接口签名需对齐
3. OpenAPI / Swagger 定义：`gitfox/app/api/openapi/artifact.go` 未迁移，接口文档需另行补（阶段 3 完工后）
4. RBAC `internal/rbac/artifact` 有 Manager + 权限策略，砍掉后首阶段仅按 space 级鉴权，**需与产品确认**

**下一步**（动工前要魔尊拍板）：
1. 首阶段是否先只上 **raw + generic**？（推荐 ✅）
2. RBAC node 级权限是否进首阶段？（推荐 ❌，放阶段 5）
3. S3 驱动是否进首阶段？（推荐 ❌，源项目目前也只支持 local）
4. 容器（Docker/OCI）与 Helm 协议是放阶段 3 还是阶段 4？（推荐阶段 4）
