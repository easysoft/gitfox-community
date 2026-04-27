# Artifact 本地迁移核心逻辑

> 适用仓库：`/root/repo/pangu/gitfox-community`
> 目的：说明当前分支把 `/root/repo/gitfox` 制品库模块迁入本项目后的核心链路、分层职责与已知风险。

---

## 1. 迁移目标与当前状态

这次迁移不是重写，而是把原有 Artifact 模块按本项目现有分层方式重新落位：

- HTTP 入口落到 `app/api/router/artifact.go` + `app/api/controller/artifact/*`
- 业务编排落到 `app/service/artifact/*`
- 制品格式/上传协议/树同步等核心逻辑保留在 `app/artifact/*`
- 数据访问落到 `app/store/artifact/*`
- 数据模型落到 `app/model/artifact/*`

当前分支的特点是：

1. **主链路已经能编译通过**，`go test ./app/service/artifact ./app/store/artifact ./app/artifact/... ./cmd/cli/operations/server/...` 可过。
2. **SQL 已扩成完整 Artifact 表集**，并新增了全局 / 空间 / 仓库三级作用域。
3. **功能总体仍以“原封迁移”为主**，行为兼容优先，很多边界语义还没有按本项目重新打磨。

---

## 2. 分层职责

### 2.1 Router / Controller

入口集中在：

- `app/api/router/artifact.go`
- `app/api/controller/artifact/artifact.go`
- `app/api/controller/artifact/permission.go`

职责：

- 挂载 HTTP 路由
- 解析 `spaceId` / `repoId` / `entityId` / `assetId` / `versionId` 等参数
- 调用 `app/service/artifact` 暴露的业务方法
- 输出统一响应

### 2.2 Service

核心服务在：

- `app/service/artifact/artifact.go`
- `app/service/artifact/permission.go`

职责：

- 解析作用域（global / space / repo）
- 选择上传器、组织上传与落库流程
- 处理 soft delete / GC / statistic / archive / permission
- 调度 Job Scheduler 跑异步 GC

### 2.3 Artifact Core

核心适配逻辑在：

- `app/artifact/adapter/*`
- `app/artifact/adapter/request/*`
- `app/artifact/model/tree_sync.go`
- `app/artifact/entity/parser.go`
- `app/artifact/authorizer/authorizer.go`

职责：

- 按制品格式解析请求（generic / raw / container / helm）
- 管理 blob 写入、asset 创建/更新
- 在 package / version / group / asset 之间建立关系
- 维护树节点与权限实体路径

### 2.4 Store / Model

数据访问与模型在：

- `app/store/artifact/*`
- `app/model/artifact/*`

职责：

- 提供 views / groups / packages / versions / assets / blobs / tree_nodes 等表的 CRUD
- 封装跨表查询（如包列表、容量统计、blob 历史）
- 定义 Artifact 相关枚举、请求体、响应体与树节点结构

---

## 3. 作用域模型：global / space / repo

当前迁移后的作用域核心是 `ops_artifact_views`。

表结构位置：

- `app/store/migrate/mysql/0034_create_artifact_tables.up.sql`

关键字段：

- `spaceID = 0`：全局级 Artifact View
- `spaceID > 0 && repoID = 0`：空间级 Artifact View
- `spaceID > 0 && repoID > 0`：仓库级 Artifact View

服务层通过 `ParseSpaceView` 解析请求作用域：

- `repoId > 0`：优先解析仓库级默认 view
- 否则 `spaceId > 0`：解析空间级默认 view
- 两者都没有：回落到全局默认 view

对应代码：

- `app/service/artifact/artifact.go:209`

这意味着当前所有制品数据，最终都要先归属到一个 **view**，后续 package / version / group / asset 都围绕 view 展开。

---

## 4. 上传主链路

### 4.1 入口

上传入口：

- `POST /artifacts/upload/:format`
- 控制器：`app/api/controller/artifact/artifact.go:254`

控制器先绑定 `SpaceRepoReq`，再调用：

1. `ParseSpaceView()` 解析视图
2. `Upload(view, format, req)` 执行实际上传

### 4.2 Service 选择上传器

`app/service/artifact/artifact.go:305`

当前 `Upload()` 根据格式选择不同 uploader：

- `raw.NewUploader(...)`
- `generic.NewUploader(...)`
- 其余格式直接走专门 adapter（container / helm 的相关逻辑也已迁入代码树）

### 4.3 Request 层提交模型

上传请求最终会进入：

- `app/artifact/adapter/request/request.go`

`ArtifactUploadRequest.Commit()` 的核心动作：

1. 从 session 提取 creator
2. 构造 `ModelManager`
3. 对主资产执行 `ModelManager.Sync()`
4. 对子资产继续执行 `Sync()`
5. 最后 commit 所有 storage writer

对应代码：

- `app/artifact/adapter/request/request.go:75`

### 4.4 ModelManager 的三类资产语义

`app/artifact/adapter/model_mgr.go`

当前迁移保留了原模块的三种资产挂载语义：

1. **isolated**
   - 不挂 version / group
   - 更像全局 blob 资产
   - 典型场景：container manifest digest 资产

2. **exclusive**
   - 同一 version 下只保留当前资产
   - 新资产写入后，其它资产会被 soft delete

3. **normal**
   - 常规资产
   - raw 走 version 维度
   - generic 走 group 维度

对应代码：

- `app/artifact/adapter/model_mgr.go:41`

### 4.5 package / version / group 自动建模

`ModelManager` 会在上传时自动补齐结构：

- 包不存在则建 package
- 版本不存在则建 version
- raw 包如果有 namespace，则自动拆成 group 树

raw 的 group 树同步入口：

- `app/artifact/model/tree_sync.go:185`

其核心逻辑是：

- `pkg.Namespace = a.b.c`
- 拆成 `a -> b -> c`
- 调用 `Groups().RecursiveCreate()` 建层级 group
- 把末级 group 回填给 package 的 `groupID`

---

## 5. 树节点同步逻辑

树节点模型表：

- `ops_artifact_tree_nodes`

主要代码：

- `app/artifact/model/tree_sync.go`

当前同步器职责：

- `SyncTreeNodeForFormat()`：建 format 根节点
- `SyncTreeNodeForGroup()`：建 group 节点
- `SyncTreeNodeForPackage()`：建 package 节点
- `SyncTreeNodeForVersion()`：建 version 节点

触发方式：

- 上传新 version 后，通过 `ModelManager` 的 defer 逻辑同步 version 节点
- raw package 建成后，先同步 group 链
- 手动接口 `POST /artifacts/tree/refresh` 会调用 `RefreshTree()`

对应代码：

- `app/service/artifact/artifact.go:935`

---

## 6. 查询 / 下载主链路

### 6.1 包列表

查询包列表入口：

- `POST /artifacts/packages`

底层由 `Store.FindPackages()` 完成：

- 先查每个 package 的最新 version
- 再回查 package 展示字段
- 支持 format / query 过滤

对应代码：

- `app/store/artifact/store.go:55`

### 6.2 版本 / 资产 / 历史

服务层提供：

- `ListVersions()`
- `ListAssets()`
- `ListGroupAssets()`
- `GetAsset()`
- `GetBlobHistory()`

资产详情与列表会联表补齐 blob 信息，再把 `checkSum` / `metadata` 从 JSON 字符串反序列化为结构化响应。

对应代码：

- `app/store/artifact/store.go:171`
- `app/store/artifact/store.go:207`
- `app/store/artifact/store.go:236`

### 6.3 下载

下载主链路：

1. 通过 `assetId` 取 asset
2. 再用 `blobID` 取 blob
3. 调 `fileStore.Open(adapter.BlobPath(blob.Ref))` 打开底层文件

对应代码：

- `app/service/artifact/artifact.go:288`

---

## 7. 删除与回收链路

### 7.1 软删除

当前分两类：

1. **单对象软删**
   - `DeleteVersion()`
   - `DeletePackage()`

2. **按实体批量软删**
   - `SoftRemove()`
   - 支持 `asset.xxx / version.xxx / package.xxx / group.xxx`

`SoftRemove()` 的级联策略：

- 删 asset：soft delete asset + blob
- 删 version：先删其下 assets，再删 version
- 删 package：先删 versions，再删 package
- 删 group：先删 packages，再删 group assets，再删 group

对应代码：

- `app/service/artifact/artifact.go:611`

### 7.2 手动垃圾回收

当前保留两类 GC Job：

- `artifact-container-gc`
- `artifact-soft-remove-gc`

触发入口：

- `/artifacts/garbage/container`
- `/artifacts/garbage/soft-remove`

底层通过 `JobScheduler.RunJob()` 提交后台任务。

对应代码：

- `app/service/artifact/artifact.go:544`

### 7.3 过期 View 清理

当前新增了第三类逻辑：**过期已软删 view 清理**。

相关配置：

- `pkg/config/config.go` 中的 `ArtifactConfig`
- `recycleBinRetention`
- `gcInterval`

注册位置：

- `cmd/cli/operations/server/server.go:109`

当前实现是：

- 服务启动时注册一个 recurring job
- 每天凌晨 3 点跑一次
- handler 调 `GarbageCollectExpiredViews()`
- 按回收站保留期筛掉过期 view，再执行资源清理

对应代码：

- `app/service/cron/handler.go:186`
- `app/service/artifact/artifact.go:945`

---

## 8. 权限模型

### 8.1 路由级权限门

路由层使用：

- `middleware.ArtifactAuthMiddleware(...)`

按动作区分：

- pull
- push
- create
- delete
- permission

路由位置：

- `app/api/router/artifact.go:10`

### 8.2 实体级权限路径

Artifact 的权限实体不是直接靠表主键判断，而是先把实体解析成一棵逻辑树，再转成策略路径。

核心逻辑：

- `app/artifact/entity/parser.go`
- `app/model/artifact/artifact.go` 中的 `ArtifactEntity.PolicyLocation()`

例如：

- format
- group
- package
- version
- asset

都会被拼成层级路径，再交给 Casbin / RBAC 判断。

### 8.3 Space / Entity 权限接口

当前已迁入两类权限接口：

- 空间级权限：`/artifacts/permissions`
- 实体级权限：`/artifacts/entities/:entityId/permissions`

服务逻辑在：

- `app/service/artifact/permission.go`

---

## 9. 当前审查发现的主要问题

下面几处，不是代码风格问题，而是**结构性风险**。

### 9.1 过期 view 的“立即重建”只删 view，不删关联资源

问题位置：

- `app/service/artifact/artifact.go:159`
- `app/store/artifact/view.go:23`

当前 `CreateView()` 在遇到“已过保留期的同名软删 view”时，会先直接执行 `Views().Delete(ctx, v.ID)`，然后创建新 view。

但这里**只删了 view 主记录**，没有同步清理：

- groups
- packages
- versions
- group 级 assets
- meta assets
- tree nodes
- 物理 blob 文件

结果：

- 新 view 会拿到新的 `viewID`
- 旧 view 下的所有资源会变成孤儿数据
- 后续统计、树同步、GC、排障都会越来越脏

### 9.2 定时 GC 只清 package/version 这条链，遗漏 generic/group/tree/meta 资源

问题位置：

- `app/service/artifact/artifact.go:984`

`hardDeleteViewResources()` 现在只做了：

- 查询 view 下 packages
- 删除 package 下 versions
- 删除 version 下 assets/blob/file
- 删除 package

没有清理：

- `ops_artifact_groups`
- group 直接挂载的 generic assets
- `ops_artifact_meta_assets`
- `ops_artifact_tree_nodes`

这意味着即使定时 GC 跑了，view 相关数据仍然可能残留半套尸体。

### 9.3 group 软删后没有真正的 restore/undelete 语义

问题位置：

- `app/store/artifact/group.go:122`
- `app/store/artifact/group.go:228`
- `app/artifact/model/tree_sync.go:185`

当前 `group` 的行为是：

- `SoftDeleteById()` 只把 `deleted = 1`
- `RecursiveCreate()` 查到已存在 group 就直接返回
- 但**不会把已软删 group 恢复为 deleted = 0**

结果：

- raw namespace 再次上传时，可能复用到一个 `deleted=1` 的 group
- generic `CreateGroup()` 也可能“创建成功”，实则拿到的是软删 group
- package 会继续挂到这个已软删 group 上

这会让 group 层处于“逻辑已删、业务又继续复用”的撕裂状态。

### 9.4 asset 软删后，按原路径重传可能撞唯一索引

问题位置：

- SQL：`app/store/migrate/mysql/0034_create_artifact_tables.up.sql:110`
- 代码：`app/artifact/adapter/model_mgr.go:136`
- 代码：`app/store/artifact/asset.go:49`

当前 `asset` 查重逻辑只查 `deleted = 0` 的记录；
但 SQL 上 `ops_artifact_assets` 仍有：

- `UNIQUE(versionID, path)`

这意味着：

- 旧 asset 被 soft delete 后
- 同一路径再重新创建 asset
- 业务层会以为“不存在，可创建”
- 但数据库层可能直接打出 duplicate key

这在“软删后重传同版本同路径文件”的链路上风险很高。

### 9.5 权限中间件目前只做了粗粒度校验，且 GET 场景 spaceId 取值不完整

问题位置：

- `app/api/middleware/artifact_auth.go:18`
- `app/artifact/authorizer/authorizer.go:37`

当前中间件存在两个问题：

1. 它构造的是一个固定的 mock entity：`format/repository`
   - 没有解析真实的 package / version / asset / group 实体
   - 所以大多数接口没有真正走到实体级权限判断

2. 它只从 JSON body 或 form 里取 `spaceId`
   - GET 路由上的 query 参数 `spaceId` 没有统一解析
   - 很多请求会以 `spaceID = 0` 去鉴权

所以当前权限更接近“路由级粗门禁”，不是完整的节点级 RBAC 落地。

### 9.6 `artifact.gcInterval` 配置目前没有真正生效

问题位置：

- 配置定义：`pkg/config/config.go:294`
- 实际注册：`cmd/cli/operations/server/server.go:112`

配置里已经有：

- `recycleBinRetention`
- `gcInterval`

但服务启动时注册定时任务仍然写死为：

- `0 3 * * *`

也就是说：

- `gcInterval` 当前只被 `Sanitize()` 校验
- 还没有参与 cron 表达式生成
- 现在属于“看起来可配，实际上未生效”的状态

### 9.7 `RefreshTree()` 不是全量重建全树，只重放了 group 链

问题位置：

- `app/service/artifact/artifact.go:935`
- `app/artifact/model/tree_sync.go:24`

`RefreshTree()` 调的是 `BuildTreeNode()`，而 `BuildTreeNode()` 现在只遍历 `Groups().All()`。

它不会主动全量扫：

- packages
n- versions
- assets

因此“刷新制品树”这个接口，当前更接近“重放 group/format 节点”，不是彻底 rebuild 全量树结构。

---

## 10. 建议的后续修正顺序

建议按下面顺序继续收口：

1. **先修 view GC 与 orphan 资源问题**
   - 否则数据一旦开始积累，后面越修越脏

2. **再修 group / asset 的 soft delete 与 restore 语义**
   - 这是上传、重建、回收站恢复的主病灶

3. **再补权限中间件的真实实体解析**
   - 至少先把 `spaceId` 获取和 entity 解析补齐

4. **最后再统一 tree rebuild 与 GC 配置化**
   - 这两项偏收口与可运维性

---

## 11. 一句话总结

当前本地迁移已经把 **Artifact 的表、路由、服务、上传/下载/删除主链路** 搬进了本项目，也已经能编译通过；但它仍然属于“骨架已立、回收/恢复/权限边界还没磨平”的状态，后续改动应优先围绕 **view 生命周期、soft delete 语义、权限落点** 三条链继续收口。
