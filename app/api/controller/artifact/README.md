# 制品库（Artifact）接口汇总

本文档列举制品库当前对外暴露的 HTTP 接口，分为两大模块：

1. **制品管理接口**（`/artifacts/*`）：制品视图、分组、上传/下载、查询、移动、删除
2. **容器镜像接口**（`/v2/*`）：OCI Distribution Spec v2，兼容 Docker/OCI 客户端

> 鉴权说明：当前版本已下线 Casbin/制品库内置权限体系，不再走 `pull / push / create / delete / admin` 区分；统一走 JWT 会话级鉴权，由上层 `/api/v2` 路由组的 `Attempt(authn.Get(authn.AuthKindEasy))` 中间件提供。

---

## 一、制品管理接口 `/artifacts`

### 1.1 视图

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| POST | `/artifacts/views` | `CreateView` | 创建制品库视图（与空间/仓库/全局作用域绑定） |

### 1.2 分组

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| POST | `/artifacts/groups` | `CreateGroup` | 创建分组节点（generic 包按 `.` 分隔的目录结构） |

### 1.3 上传 / 下载

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| POST | `/artifacts/upload/:format` | `Upload` | 上传制品文件（multipart/form-data），`:format` 指定格式 |
| GET  | `/artifacts/assets/:assetID/download` | `DownloadAsset` | 下载制品文件（octet-stream 流） |

### 1.4 实体操作

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| POST   | `/artifacts/entities/move` | `MoveEntity` | 将 package / asset / group 移动到目标制品库或目标分组 |
| POST   | `/artifacts/entities/rename` | `RenameEntity` | 重命名 asset 或 group |
| POST   | `/artifacts/entities/relocate` | `RelocateEntity` | 同事务内同时支持移动 + 重命名 |
| DELETE | `/artifacts/entities` | `SoftRemoveByEntity` | 按 `kind.id` 批量软删除，支持 asset / version / package / group / artifact 级联 |

### 1.5 查询

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| POST | `/artifacts/nodes` | `ListTree` | 通用制品树查询，支持 path/format/level 聚合，`/**` 启用递归 |
| GET  | `/artifacts/assets/:assetID` | `GetAsset` | 获取指定制品文件的详细元信息 |
| POST | `/artifacts/assets/list` | `ListAssetPage` | 按 `entityID`（group/version/package/asset）分页查询资产 |

### 1.6 树重建

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| POST | `/artifacts/tree/refresh` | `RefreshTree` | 按 format（或 `all`）重建制品树节点（数据修复 / 迁移用） |

### 1.7 暂时下线（仅注释保留，可按需放开）

下列接口的 controller / service 实现仍在仓库中，但路由层已暂时注释。需要时直接在 `app/api/router/artifact.go` 解开注释即可。

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| GET | `/artifacts/groups/:groupID/archive` | `ArchiveGroup` | 将分组内所有文件打包为 ZIP 流下载 |
| GET | `/artifacts/capacity/all` | `StatisticCapacity` | 非容器格式制品总容量统计 |
| GET | `/artifacts/capacity/container` | `StatisticContainerCapacity` | 容器格式制品（container）容量统计 |
| GET | `/artifacts/statistic` | `StatisticSummary` | 制品资产汇总统计 |
| GET | `/artifacts/garbage/container` | `GarbageCollectContainer` | 触发容器格式制品的垃圾回收任务 |
| GET | `/artifacts/garbage/soft-remove` | `GarbageCollectSoftRemove` | 触发已软删制品的物理清理任务 |

---

## 二、容器镜像接口 `/v2`（OCI Distribution Spec）

> 认证：Bearer Token（先 `GET /v2/token` 换取 token，后续请求通过内存 TTLCache 校验）  
> 兼容 `docker login / pull / push` 标准流程

### 2.1 基础认证

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| GET | `/v2/` | `HandleAPIBase` | 探测端点；匿名请求返回 `401 + WWW-Authenticate: Bearer` 挑战，引导客户端去 `/v2/token` |
| GET | `/v2/token` | `HandleToken` | 使用 Basic Auth（用户名/密码）换取临时 Bearer Token |

### 2.2 Blob 操作

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| HEAD   | `/v2/:space/:name/blobs/:digest` | `HandleBlobHead` | 检查 blob 是否存在 |
| GET    | `/v2/:space/:name/blobs/:digest` | `HandleBlobGet` | 下载指定 digest 的 blob |
| DELETE | `/v2/:space/:name/blobs/:digest` | `HandleBlobDelete` | 删除指定 digest 的 blob |

### 2.3 Blob 分块上传

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| POST   | `/v2/:space/:name/blobs/uploads/` | `HandleBlobUploadStart` | 启动新的分块 blob 上传会话 |
| GET    | `/v2/:space/:name/blobs/uploads/:uuid` | `HandleBlobUploadStatus` | 查询指定上传会话的当前状态 |
| PATCH  | `/v2/:space/:name/blobs/uploads/:uuid` | `HandleBlobUploadPatch` | 追加数据块到上传会话 |
| PUT    | `/v2/:space/:name/blobs/uploads/:uuid` | `HandleBlobUploadFinish` | 完成分块上传（需携带 `digest`） |
| DELETE | `/v2/:space/:name/blobs/uploads/:uuid` | `HandleBlobUploadCancel` | 取消并清理上传会话 |

### 2.4 Manifest 操作

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| HEAD   | `/v2/:space/:name/manifests/:reference` | `HandleManifestHead` | 检查 manifest 是否存在 |
| GET    | `/v2/:space/:name/manifests/:reference` | `HandleManifestGet` | 获取 manifest 内容（支持 `If-None-Match`） |
| PUT    | `/v2/:space/:name/manifests/:reference` | `HandleManifestPut` | 推送 manifest |
| DELETE | `/v2/:space/:name/manifests/:reference` | `HandleManifestDelete` | 删除 manifest |

### 2.5 Tag 操作

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| GET | `/v2/:space/:name/tags/list` | `HandleListTag` | 列出镜像仓库的所有 tag |

---

## 接口路径前缀

| 前缀 | 挂载位置 | 说明 |
|------|---------|------|
| `/artifacts` | `app/api/router/artifact.go` | 制品管理 REST 接口 |
| `/v2` | `app/api/router/container.go` | OCI v2 容器镜像协议接口 |

## 核心文件索引

| 文件 | 说明 |
|------|------|
| `app/api/controller/artifact/artifact.go` | Controller 基座、视图、移动 / 重命名 / 重定位 |
| `app/api/controller/artifact/asset.go` | 资产查询、下载、分页列表 |
| `app/api/controller/artifact/group.go` | 分组创建（含归档接口注释保留段） |
| `app/api/controller/artifact/tree.go` | 通用树查询、树重建 |
| `app/api/controller/artifact/upload.go` | 制品上传入口 |
| `app/api/controller/artifact/admin.go` | 实体软删、暂时下线接口（注释段） |
| `app/artifact/handler/container.go` | OCI v2 容器镜像 Handler |
| `app/artifact/controller/container/controller.go` | 容器镜像核心业务控制器 |
| `app/api/router/artifact.go` | 制品管理路由注册 |
| `app/api/router/container.go` | OCI v2 路由注册及 Bearer Token 中间件 |
