# 制品库（Artifact）接口汇总

本文档列举制品库所有对外暴露的 HTTP 接口，分为两大模块：

1. **制品管理接口**（`/artifacts/*`）：通用制品包/版本/文件的 CRUD 与权限管理
2. **容器镜像接口**（`/v2/*`）：OCI Distribution Spec v2，兼容 Docker/OCI 客户端

---

## 一、制品管理接口 `/artifacts`

> 认证：Casbin Artifact 权限体系（pull / push / create / delete / admin）

### 1.1 包 & 格式

| 方法 | 路径 | 权限 | 接口名称 | 简介 |
|------|------|------|---------|------|
| POST | `/artifacts/packages` | pull | `ListPackages` | 列出指定空间/视图下的制品包，支持按格式和关键字过滤 |
| POST | `/artifacts/formats` | pull | `ListFormats` | 列出指定空间已使用的制品格式（maven/npm/generic/container 等） |

### 1.2 视图

| 方法 | 路径 | 权限 | 接口名称 | 简介 |
|------|------|------|---------|------|
| POST | `/artifacts/views` | create | `CreateView` | 创建制品库视图（测试/管理用途，与空间/仓库绑定） |

### 1.3 分组

| 方法 | 路径 | 权限 | 接口名称 | 简介 |
|------|------|------|---------|------|
| POST | `/artifacts/groups` | create | `CreateGroup` | 创建分组节点（generic 包按 `.` 分隔的目录结构） |
| GET  | `/artifacts/groups/:groupId/assets` | pull | `ListGroupAssets` | 列出指定分组下的所有制品文件 |
| GET  | `/artifacts/groups/:groupId/archive` | pull | `ArchiveGroup` | 将分组内所有文件打包为 ZIP 流下载 |

### 1.4 上传

| 方法 | 路径 | 权限 | 接口名称 | 简介 |
|------|------|------|---------|------|
| POST | `/artifacts/upload/:format` | push | `Upload` | 上传制品文件（multipart/form-data），`:format` 指定格式（如 `generic`、`maven`） |

### 1.5 版本

| 方法 | 路径 | 权限 | 接口名称 | 简介 |
|------|------|------|---------|------|
| POST | `/artifacts/versions/list` | pull | `ListVersions` | 列出指定包的所有版本 |
| GET  | `/artifacts/versions/:versionId/assets` | pull | `ListAssets` | 列出指定版本下的所有制品文件 |

### 1.6 文件（Asset）

| 方法 | 路径 | 权限 | 接口名称 | 简介 |
|------|------|------|---------|------|
| GET  | `/artifacts/assets/:assetId` | pull | `GetAsset` | 获取指定制品文件的详细元信息 |
| GET  | `/artifacts/assets/:assetId/history` | pull | `GetBlobHistory` | 获取指定制品文件的历史 Blob 列表（所有历史版本） |
| GET  | `/artifacts/assets/:assetId/download` | pull | `DownloadAsset` | 下载制品文件（octet-stream 流） |

### 1.7 删除

| 方法 | 路径 | 权限 | 接口名称 | 简介 |
|------|------|------|---------|------|
| DELETE | `/artifacts/entities` | delete | `SoftRemoveByEntity` | 按实体 ID 批量软删除，元素格式 `type.id`；支持 `asset` / `version` / `package` / `group`（级联）；以及 `artifact`（整库软删，仅允许在无活制品时执行，空分组允许保留） |

### 1.8 树节点

| 方法 | 路径 | 权限 | 接口名称 | 简介 |
|------|------|------|---------|------|
| GET  | `/artifacts/nodes/:nodeId/children` | pull | `ListTreeNodes` | 获取树节点的直接子节点列表（用于前端树形展示） |

### 1.9 统计

| 方法 | 路径 | 权限 | 接口名称 | 简介 |
|------|------|------|---------|------|
| GET | `/artifacts/capacity/all` | pull | `StatisticCapacity` | 非容器格式制品总容量统计 |
| GET | `/artifacts/capacity/container` | pull | `StatisticContainerCapacity` | 容器格式制品（container）容量统计 |
| GET | `/artifacts/statistic` | pull | `StatisticSummary` | 制品资产汇总统计（各格式数量、大小等） |

### 1.10 垃圾回收 & 运维

| 方法 | 路径 | 权限 | 接口名称 | 简介 |
|------|------|------|---------|------|
| GET  | `/artifacts/garbage/container` | delete | `GarbageCollectContainer` | 触发容器格式制品的垃圾回收后台任务 |
| GET  | `/artifacts/garbage/soft-remove` | delete | `GarbageCollectSoftRemove` | 触发已软删制品的物理清理后台任务 |
| POST | `/artifacts/tree/refresh` | create | `RefreshTree` | 重建所有制品树节点（数据修复/迁移用） |

### 1.11 权限管理（需系统管理员）

> 认证：`ArtifactAdminMiddleware`（zentao App 或 DB 管理员），避免权限死锁问题

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| POST | `/artifacts/entities/:entityId/permissions` | `SetEntityPermission` | 设置指定实体节点（`kind.id` 格式，如 `package.123`）的权限，支持私有/公开和用户角色规则 |
| GET  | `/artifacts/entities/:entityId/permissions` | `GetEntityPermissions` | 获取指定实体节点的权限配置（含父级继承信息） |
| POST | `/artifacts/permissions` | `SetPermissions` | 设置空间级权限（不传 spaceId → 全局权限，传 spaceId+repoId → 仓库级权限） |
| GET  | `/artifacts/permissions` | `GetPermissions` | 获取空间/仓库的制品库权限配置 |
| GET  | `/artifacts/members` | `GetMembers` | 获取空间/仓库制品库的成员列表（含权限角色） |

---

## 二、容器镜像接口 `/v2`（OCI Distribution Spec）

> 认证：Bearer Token（先 `GET /v2/token` 换取 token，后续请求通过内存 TTLCache 校验）  
> 兼容 `docker login / pull / push` 标准流程

### 2.1 基础认证

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| GET | `/v2/` | `HandleAPIBase` | 探测端点；匿名请求返回 `401 + WWW-Authenticate: Bearer` 挑战，引导客户端去 `/v2/token` |
| GET | `/v2/token` | `HandleToken` | 使用 Basic Auth（用户名/密码）换取临时 Bearer Token（存入内存 TTLCache） |

### 2.2 Blob 操作

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| HEAD   | `/v2/:space/:name/blobs/:digest` | `HandleBlobHead` | 检查 blob 是否存在（只返回头部，不传输内容） |
| GET    | `/v2/:space/:name/blobs/:digest` | `HandleBlobGet` | 下载指定 digest 的 blob 内容 |
| DELETE | `/v2/:space/:name/blobs/:digest` | `HandleBlobDelete` | 删除指定 digest 的 blob |

### 2.3 Blob 分块上传

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| POST  | `/v2/:space/:name/blobs/uploads/` | `HandleBlobUploadStart` | 启动一个新的分块 blob 上传会话，返回 `Location` 和 `Docker-Upload-UUID` |
| GET   | `/v2/:space/:name/blobs/uploads/:uuid` | `HandleBlobUploadStatus` | 查询指定上传会话的当前状态（已上传字节范围） |
| PATCH | `/v2/:space/:name/blobs/uploads/:uuid` | `HandleBlobUploadPatch` | 追加数据块到上传会话 |
| PUT   | `/v2/:space/:name/blobs/uploads/:uuid` | `HandleBlobUploadFinish` | 完成分块上传，需携带 `digest` 参数做校验 |
| DELETE| `/v2/:space/:name/blobs/uploads/:uuid` | `HandleBlobUploadCancel` | 取消并清理指定上传会话 |

### 2.4 Manifest 操作

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| HEAD   | `/v2/:space/:name/manifests/:reference` | `HandleManifestHead` | 检查 manifest 是否存在（`:reference` 可为 tag 或 digest） |
| GET    | `/v2/:space/:name/manifests/:reference` | `HandleManifestGet` | 获取 manifest 内容（支持 `If-None-Match` 协商缓存） |
| PUT    | `/v2/:space/:name/manifests/:reference` | `HandleManifestPut` | 推送 manifest（`:reference` 为 tag 时绑定 tag，为 digest 时直接存储） |
| DELETE | `/v2/:space/:name/manifests/:reference` | `HandleManifestDelete` | 删除 manifest |

### 2.5 Tag 操作

| 方法 | 路径 | 接口名称 | 简介 |
|------|------|---------|------|
| GET | `/v2/:space/:name/tags/list` | `HandleListTag` | 列出指定镜像仓库的所有 tag |

---

## 接口路径前缀说明

| 前缀 | 挂载位置 | 说明 |
|------|---------|------|
| `/artifacts` | `app/api/router/artifact.go` | 通用制品管理 REST 接口 |
| `/v2` | `app/api/router/container.go` | OCI v2 容器镜像协议接口 |

## 核心文件索引

| 文件 | 说明 |
|------|------|
| `app/api/controller/artifact/artifact.go` | 制品管理接口 Handler（包/版本/文件/上传/统计等） |
| `app/api/controller/artifact/permission.go` | 权限管理接口 Handler |
| `app/artifact/handler/container.go` | OCI v2 容器镜像接口 Handler |
| `app/artifact/controller/container/controller.go` | 容器镜像核心业务控制器（Blob/Manifest/Tag 操作） |
| `app/api/router/artifact.go` | 制品管理路由注册 |
| `app/api/router/container.go` | OCI v2 路由注册及 Bearer Token 中间件 |
