# 制品格式与版本控制模型

> 面向实现者 / 运维 / 调试方的内部设计文档。梳理七种制品格式在系统层的落库模型、版本控制语义、物理存储布局。

## 1. 总览：版本化 vs 非版本化

分类谓词在 `app/model/artifact/artifact.go` 的 `IsVersionedFormat`：

| Format | 版本化？ | 是否注册 uploader | 实现位置 |
|---|---|---|---|
| `container` | 是 | 是 | `app/artifact/adapter/container/` |
| `helm`      | 是 | 是 | `app/artifact/adapter/helm/` |
| `maven`     | 是 | 否（预留占位） | — |
| `pypi`      | 是 | 否（预留占位） | — |
| `npm`       | 是 | 否（预留占位） | — |
| `file`   | **否** | 是 | `app/artifact/adapter/generic/` |
| `raw`       | **否**（在 `IsVersionedFormat` 分类下为 false，但 **数据模型上实际落 version**） | 是 | `app/artifact/adapter/raw/` |

`IsVersionedFormat` 谓词用于一处控制语义：
1. **禁止重命名**：版本化格式的 package 名参与分发协议（OCI pull URL、helm index 路径），改名会破坏客户端依赖，所以 `move_rename.go:227 / :402` 对版本化格式直接拒绝 rename。

> ⚠️ 另一个相关逻辑：**创建视图时是否自动铸默认 group**。当前 container 已不再创建默认 group；container package 直接挂在 format 根节点下，raw/generic 仍按各自上传参数创建或选择 group。

> ⚠️ 注意：raw 在 `IsVersionedFormat` 上返回 false，但 raw uploader 会显式要求 `version` 字段并落 `ops_artifact_versions`。"是否版本化"在代码里其实是两个语义合一的谓词——**分发协议语义**（是否强绑包名）与 **数据模型语义**（是否落 version 记录）。raw 属于"有 version 但不是协议级版本化"的中间态。

## 2. 核心数据模型

五层实体（`app/model/artifact/artifact.go`）：

```
ops_artifact (view, 即"制品库")
  └── ops_artifact_groups (可嵌套，按 parentID 构树)
        └── ops_artifact_packages (包)
              └── ops_artifact_versions (版本)
                    └── ops_artifact_assets (文件；asset 可直接挂 group，绕过 version)
                          └── ops_artifact_blobs (物理 blob，多 asset 可共享)
```

### 挂载字段矩阵（`ArtifactAsset`）

| 字段 | 含义 | 生效场景 |
|---|---|---|
| `VersionID` | 挂到某个 version | 版本化格式（container manifest tag、helm chart、raw 上传） |
| `GroupID` | 直接挂到 group，**无 version** | `generic` 唯一路径 |
| `ArtifactID` | 挂到 view 级别（全库范围） | 全局隔离资产（容器 blob、format 级 index 文件） |

一条 asset 只取其一；另两个字段置 0 或 null。

### 三种挂载 Attr（`app/artifact/adapter/hosted.go:87`）

| Attr | 行为 | 用例 |
|---|---|---|
| `AttrAssetNormal` | 按 (versionID, path) 或 (groupID, path) 去重；存在即更新 blob 并软删旧 blob | helm chart、raw、generic |
| `AttrAssetExclusive` | 同 normal，但会把同一 version 下**除当前外**全部软删 | container manifest **tag**（一个 tag 只对应一个 manifest asset） |
| `AttrAssetIsolated` | 按 path 全局去重（通常 path=`sha256:xxx`），跨 view 共享 | container layer blob、container manifest by digest |

## 3. 各格式落库细节

### 3.1 container（OCI Distribution）

- **版本化协议**：包名参与 `docker pull {domain}/{artifact}/{image}:{tag}` URL，不可 rename。
- **包结构**：不再为 container view 创建默认 group。每个 OCI image 对应一个同作用域 `ArtifactPackage`，其 `Namespace = ""`、`GroupID = 0`，`ArtifactID`/`artifactLibID` 存创建该 package 时的作用域锚点制品库 ID。
- **版本**：tag 即 version。`ArtifactVersion.Version = tag`，`ArtifactVersion.ArtifactLibID` 决定该 tag 展示在哪个制品库下。
- **并发控制**：container 首次创建 package 时使用 `pkg/lock` 内存锁，锁粒度为 `spaceID + repoID + format + namespace + name`，用于避免单进程内并发 push 创建重复 package。
- **Asset 拆分**：
  | 资产类型 | Attr | 挂载字段 | path 形式 |
  |---|---|---|---|
  | manifest by tag | `AttrAssetExclusive` | `VersionID` | `sha256:xxx`（同 version 内去重） |
  | manifest by digest | `AttrAssetIsolated` | `ArtifactID` | `sha256:xxx` |
  | layer blob | `AttrAssetIsolated` | `ArtifactID` | `sha256:xxx` |
- **Group 兜底**：container push/pull 不再查找或创建 group。路由只负责把 `domainid + artifactname` 解析成目标制品库，package 由 `imagename` 定位，version 由 tag 定位。

### 3.2 helm

- **版本化协议**：chart name + version 对应 `index.yaml` 条目，name 不可 rename。
- **包结构**：一个 chart 名对应一个 `ArtifactPackage`，`Namespace = ""`，`GroupID` 通常来自 view 根下的默认 group（非版本化视图建 view 时自动铸）。
- **版本**：从 chart.tgz 的 `Chart.yaml.metadata.version` 提取。
- **Asset 拆分**：
  - 主资产 path = `charts/{name}-{version}.tgz`，`Attr = AttrAssetNormal`，挂 `VersionID`。
  - `AssetMetadata` 保存 chart metadata JSON，用于后续重建 `index.yaml`。
- **索引**：`ops_artifact_meta_assets` 存 `index.yaml`（view 级，`ArtifactID` 挂载）。删 package 时 `softRemovePackage` 会重建索引（`artifact.go:1661-1663`）。

### 3.3 raw

- **协议**：自定义 multipart 上传，`name + version + group`（group 是点号分层 namespace）。
- **包结构**：`ArtifactPackage.Namespace = group`（例 `"com.example.libs"`），`GroupID` 由 `SyncRawGroup` 递归建组后回填（`tree_sync.go:195`）。
- **版本**：必填字段，直接落 `ops_artifact_versions`。
- **Asset**：`AttrAssetNormal`，挂 `VersionID`，path = 上传文件名。
- **关键**：raw 虽在 `IsVersionedFormat` 为 false，但它**确实有 version 记录**。不同于 generic。

### 3.4 file

- **协议**：自定义 multipart 上传，只要 `group` 或 `groupId` + 文件。
- **包结构**：**无 package、无 version**。asset 直接挂 group。
  - `createNormalAssetToGroup` 分支（`model_mgr.go:163`）：`VersionID = 0`、`GroupID = 指定 group`、`ArtifactID = view.ID`。
  - 同一 (groupID, path) 组合代表同一个 asset，上传相同 path 等于覆盖（旧 blob 软删）。
- **默认 group**：generic 本身不需要默认 group；`CreateView` 当前不提前创建 group。若 generic 视图新建时没有可用 group，首次上传要求客户端显式传 `group` 或 `groupId`。
- **树展示**：asset 作为 group 的子节点显示，不经 package/version 层级。

### 3.5 maven / pypi / npm

三者在 `IsVersionedFormat` 为 true，但当前代码库**未提供 uploader 实现**（`app/artifact/adapter/` 下无对应子包）。预期语义（推测，未落实现）：

- **maven**：path = `{groupId}/{artifactId}/{version}/{artifactId}-{version}.{ext}`，`ArtifactPackage.Namespace = groupId`、`Name = artifactId`。
- **pypi**：包名 + 版本 + wheel/sdist 文件。
- **npm**：scope@package + 版本 + tarball。

以上待后续实现时再补落库细节。`AllArtifactFormatList`（`artifact.go:41`）目前也只暴露 `generic / raw / maven / container / helm`——pypi / npm 甚至连 List 都未挂入。

## 4. 删除语义

### 4.1 软删 / 硬删分层

- **软删**（`SoftRemove` 系列）：`deleted` 字段置 1，`editedDate` 作为删除时间戳。
- **硬删** + 物理清理：由手动 GC（`GarbageCollect` / `GarbageCollectSoftRemove`）在保留期过后清理已软删的 asset/blob/物理文件，以及空的 package/version/group。GC 不物理删除 `ops_artifact` 制品库主表记录。

### 4.2 按实体层级的软删行为（`softRemove{Asset,Version,Package,Group,Artifact}`）

| 入口 | 对应 level | 级联动作 |
|---|---|---|
| `asset.{id}` | asset | 软删 asset + 关联 blob |
| `version.{id}` | version | 若无活 asset 则软删 version；**不动 package** |
| `package.{id}` | package | 非 container：确认 version 下无活 asset 后级联软删空 version → 软删 package；helm 会重建 `index.yaml`。container：只有该 package 在所有制品库下都没有 active version 时才允许软删 package |
| `group.{id}` | group | 遍历 group 下的 package + 直挂 group 的 asset 递归软删 → 软删 group |
| `artifact.{id}` | **整个视图** | 非 container：若存在未软删 group/package/version/asset 则拒绝。container：按 `version.artifactLibID` 判断当前制品库下是否仍有 tag；package 只作为同 scope 锚点，不阻塞单个制品库删除 |

`package.{id}` 查询版本列表时，container 格式必须额外传当前 `artifactID`，服务端按 `packageID + artifactLibID` 过滤 tag，避免同 scope 下 dev/prod 的 tag 混在一起。

回收站与恢复链路中，container 的 version / 挂 version 的 asset 均按 `version.artifactLibID` 回填和恢复对应制品库；`package.artifactLibID` 只表示同 scope package 锚点。

`RefreshTree` 重建树时，container 不按 `package.artifactLibID` 预建 package 节点；只通过 version 同步按 `version.artifactLibID` 补建 package/version 节点，避免锚点制品库出现空 package。

单独恢复 container package 时只恢复 package 锚点记录，不恢复锚点制品库、不重建 package 树节点。group 链路遇到 container package 会跳过，不按 package 全量 version 做列表或级联删除。

> 关键不对称：**删 version 不清 package，删 package 不清 group**。这对所有格式统一，不是 container 专属。长期使用会积累"空壳" package / version / group，硬删 GC 才会真正清除；制品库主表记录只会保持软删状态。

### 4.3 `artifact.{id}` 判空口径

位置：`softRemoveArtifact`（`app/service/artifact/artifact.go`）。

**允许删除的条件**：仅当 view 下未软删的 `group`、`package`、`version`、`asset` 都不存在。

**允许保留的东西**：
- 已软删的 group/package/version/asset（由 GC 保留期后按资源层级物理清理）。
- 已软删的 view 自身（`ops_artifact` 记录不由 GC 物理删除）。

**多语言提示**：按阻塞资源区分为 `ArtifactSoftRemoveArtifactAssetExists`、`ArtifactSoftRemoveArtifactPackageExists`、`ArtifactSoftRemoveArtifactVersionExists`、`ArtifactSoftRemoveArtifactGroupExists`，同时中英双语。

**事务语义**：路 B（整批回滚）。批量软删一旦命中非空视图，整个请求回滚，不部分成功。

## 5. 物理存储

- Blob 文件由 `pkg/storage` 通过 `storageID + ref` 寻址，实际落地由 `storage.ContentStorage` 适配（本地 / OSS 等）。
- `adapter.NewRandomBlobWriter` 生成随机 ref，上传完成后回链到 `ArtifactBlob.Ref`。
- Blob 与 asset 是多对一关系：**多 asset 可共享同一 blob**（例如 container layer 被多 image 引用、generic 相同内容的文件）。
- 硬删 asset 时，仅当 asset.BlobID 仍活且 Ref 非空才触发 `fileStore.Delete`（`artifact.go:2042-2045`）——GC 阶段才会真正释放物理空间。

## 6. 已知落差与后续事项

1. **maven / pypi / npm uploader 缺失**：`IsVersionedFormat` 已将其纳入版本化，但上传链路未实现。
2. **`ops_artifact_treenode` 在 view 软删后仍存在**：GC 不再硬删 `ops_artifact` 制品库主表记录，回收站期间和软删后前端都需要用 `deleted` 过滤。
3. **`raw` 在 `IsVersionedFormat` 的归属模糊**：它实际有 version 记录，但 `IsVersionedFormat` 返回 false；目前这只影响 rename 约束（raw 允许 rename，版本化拒绝 rename）。若后续语义拆分，需要同步审视此分支。
4. **generic 无默认 group**：generic/raw 首次上传需客户端显式带 `group`/`groupId`。若要让 generic 开箱即用，可以扩展默认 group 创建逻辑。
5. **container package 锁仅限单进程**：当前同作用域 package 创建使用内存锁，只能防同一进程内的并发重复。多实例部署时需要替换为 Redis/分布式锁，或引入 owner 冗余字段后用数据库唯一索引约束。
6. **TODO：锁职责迁移**：`LockArtifactScopePackage` 当前暂放 store 层以减少改动；container 主链路和晋级闭环跑通后，应迁到 service 层或通过 adapter hook 注入，保持 store 只负责数据读写。

## 7. 相关文件索引

| 关注点 | 文件 |
|---|---|
| 实体模型 | `app/model/artifact/artifact.go`、`app/model/artifact/artifact_resp.go` |
| Uploader 统一入口 | `app/artifact/adapter/model_mgr.go`、`app/artifact/adapter/request/request.go` |
| 按格式的 uploader | `app/artifact/adapter/{container,helm,raw,file}/` |
| Store 层 | `app/store/artifact/{view,package,version,asset,blob,group,node}.go` |
| Service 编排 | `app/service/artifact/artifact.go`、`app/service/artifact/move_rename.go` |
| 控制器 | `app/api/controller/artifact/`、`app/artifact/handler/container.go` |
