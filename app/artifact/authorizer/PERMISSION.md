# 制品库权限系统说明

> 源码位置：`app/artifact/authorizer/authorizer.go`  
> 权限管理入口：`app/service/artifact/permission.go`  
> Casbin 策略层：`pkg/rbac/artifact/`

---

## 一、核心概念

### 1.1 角色（Role）

系统在启动时通过 `RoleManager.EnsureDefaultRoles()` 初始化三个内置角色：

| 角色 | 值 | 拥有的操作权限 |
|------|----|---------------|
| **reader**（只读） | `artifact_reader` | `view`、`pull` |
| **writer**（读写） | `artifact_writer` | `view`、`pull`、`push` |
| **admin**（管理员） | `artifact_admin` | `view`、`pull`、`push`、`override`、`update`、`delete`、`create`、`permission` |

### 1.2 操作（Action）

| 操作 | 值 | 含义 |
|------|----|------|
| view | `view` | 查看制品列表、元数据 |
| pull | `pull` | 下载制品文件 |
| push | `push` | 上传制品 |
| override | `override` | 覆盖已有版本 |
| create | `create` | 创建分组、包 |
| update | `update` | 修改制品元数据 |
| delete | `delete` | 删除制品 |
| permission | `permission` | 修改权限配置 |

### 1.3 作用域（Scope）- 三级继承链

```
Repo 级（space:{spaceID}/repo:{repoID}）
    ↓ 无策略则向上继承
Space 级（space:{spaceID}）
    ↓ 无策略则向上继承
Global 级（space:0）← 系统默认规则在此
```

**规则**：某一级只要有策略，就在该级做决策，不再向上回落。

---

## 二、系统默认初始化

服务启动时（`pkg/rbac/artifact/policy_system.go`）自动写入一条全局默认规则：

```
sub=*  space=*  object=**  action=artifact_reader  effect=allow  priority=999
```

**含义**：所有用户（`*`），在任意 space，对所有资源（`**`），具有 reader 角色，默认允许。

> ⚠️ 注意：`*` 匹配的是"通配任意用户"，但代码层（`authorizer.go`）对**匿名用户**做了特殊处理：
> 在全局级无自定义策略时，匿名请求被代码提前拦截返回 `authentication required`，**不会真正到达 Casbin**。
> 因此默认模式下，登录用户有 reader 权限，匿名用户无权限。

---

## 三、三种权限模式

### 模式 1：默认（无任何自定义权限配置）

**触发条件**：从未调用过设置权限的 API。

**Casbin 策略状态**：只有系统默认规则（global 级），无 space/repo 级策略。

**行为**：

| 访问者 | 结果 |
|--------|------|
| 未登录（匿名） | ❌ 被 `authorizer.go` 代码层拦截，返回 401 |
| 已登录用户 | ✅ 命中全局默认规则，获得 reader 权限（view + pull） |
| 无 push 权限 | ❌ upload 操作被拒绝 |

**示例**：

```
用户 alice（已登录）→ GET /artifacts/packages → ✅ 允许（reader）
用户 alice（已登录）→ POST /artifacts/upload/helm → ❌ 拒绝（无 push）
匿名请求              → GET /artifacts/packages → ❌ 401 authentication required
```

---

### 模式 2：公开（所有人可访问，含匿名）

**触发条件**：调用 `POST /artifacts/permissions`，请求体 `private: false`。

**前端调用示例**：

```http
POST /artifacts/permissions?spaceId=1&repoId=2
Content-Type: application/json

{
  "private": false,
  "rules": []
}
```

**Casbin 写入的策略**：

```
sub=*  space=space:1/repo:2  object=**  action=artifact_reader  effect=allow  priority=798
```

（`*` 即 `WildcardIdentify`，匹配所有人包括匿名）

**行为**：

| 访问者 | 结果 |
|--------|------|
| 未登录（匿名） | ✅ 以 `WildcardIdentify(*)` 为 subject，命中 `* allow` 规则，允许 reader 操作 |
| 已登录用户 | ✅ 同上，reader 权限 |

**示例**：

```
匿名请求 → GET /artifacts/packages?spaceId=1&repoId=2 → ✅ 允许
匿名请求 → POST /artifacts/upload/helm?spaceId=1&repoId=2 → ❌ 拒绝（reader 无 push 权限）
```

---

### 模式 3：白名单（仅允许指定 account）

**触发条件**：调用 `POST /artifacts/permissions`，请求体 `private: true` + 指定 accounts。

**前端调用示例**：

```http
POST /artifacts/permissions?spaceId=1&repoId=2
Content-Type: application/json

{
  "private": true,
  "rules": [
    {
      "accounts": ["alice", "bob"],
      "role": "artifact_writer",
      "scope": "**"
    },
    {
      "accounts": ["carol"],
      "role": "artifact_admin",
      "scope": "**"
    }
  ]
}
```

**Casbin 写入的策略**：

```
# deny-all（最高优先级，覆盖所有人）
sub=*           space=space:1/repo:2  object=**  action=*  effect=deny   priority=800

# 白名单用户
sub=user:alice  space=space:1/repo:2  object=**  action=artifact_writer  effect=allow  priority=798
sub=user:bob    space=space:1/repo:2  object=**  action=artifact_writer  effect=allow  priority=798
sub=user:carol  space=space:1/repo:2  object=**  action=artifact_admin   effect=allow  priority=800（owner）
```

**行为**：

| 访问者 | 结果 |
|--------|------|
| 未登录（匿名） | ❌ 以 `*` 为 subject，命中 deny-all，拒绝 |
| 已登录但不在白名单 | ❌ 无匹配 allow 规则，命中 deny-all，拒绝 |
| alice（已登录） | ✅ 命中 allow 规则，writer 权限（view/pull/push） |
| carol（已登录） | ✅ admin 权限（全部操作） |

---

## 四、节点级（细粒度）权限

除了 Space/Repo 级权限，还可以对具体的 **package / version / group** 设置独立权限（节点级覆盖）。

**触发条件**：调用 `POST /artifacts/entities/{entityId}/permissions`。

`entityId` 格式：`{kind}.{id}`，如 `package.123`、`group.456`。

**前端调用示例**：

```http
POST /artifacts/entities/package.123/permissions?spaceId=1
Content-Type: application/json

{
  "private": true,
  "rules": [
    {
      "accounts": ["dave"],
      "role": "artifact_reader",
      "scope": "**"
    }
  ]
}
```

**效果**：package 123 只对 dave 开放 reader 权限，其他人被 deny。
父级（space/repo 级）的规则**不影响此节点**（节点级优先级高于 space 级）。

**查询时返回继承信息**：若节点未设置自己的策略，`GET` 接口会返回 `inherit` 字段，指出该节点从哪个父节点继承了权限：

```json
{
  "private": false,
  "rules": [],
  "inherit": {
    "inherited": true,
    "inherit_entity_id": "group.10",
    "inherit_from": "my-group",
    "inherit_parents": ["my-group"]
  }
}
```

---

## 五、权限决策流程图

```
前端请求到达
       │
       ▼
session 是否有效？
  ├─ 是（已登录） → sub = NewUser(session.User.Account)  例：user:alice
  └─ 否（匿名）   → sub = WildcardIdentify("*")
       │
       ▼
repoID > 0 且 repo 级有 Casbin 策略？
  ├─ 是 → 在 repo scope（space:1/repo:2）执行 Casbin.Enforce(sub, scope, obj, action)
  │        → 有 allow 规则 → ✅ 通过
  │        → 无 allow 或命中 deny → ❌ 拒绝
  └─ 否 ↓
       ▼
spaceID > 0 且 space 级有 Casbin 策略？
  ├─ 是 → 在 space scope（space:1）执行 Casbin.Enforce(sub, scope, obj, action)
  └─ 否 ↓
       ▼
匿名用户（sub == WildcardIdentify）？
  ├─ 是 → 直接返回 401 authentication required（不进入 Casbin）
  └─ 否 ↓
       ▼
在 global scope（space:0）执行 Casbin.Enforce
命中系统默认规则：* * ** artifact_reader allow priority=999
→ ✅ 已登录用户获得 reader 权限
```

---

## 六、前端权限配置接口汇总

| 操作 | Method | Path | 关键参数 |
|------|--------|------|---------|
| 设置 Space/Repo 级权限 | `POST` | `/artifacts/permissions` | `?spaceId=&repoId=` + body |
| 查询 Space/Repo 级权限 | `GET` | `/artifacts/permissions` | `?spaceId=&repoId=` |
| 设置节点级权限 | `POST` | `/artifacts/entities/{entityId}/permissions` | `?spaceId=` + body |
| 查询节点级权限 | `GET` | `/artifacts/entities/{entityId}/permissions` | `?spaceId=` |
| 查询成员列表 | `GET` | `/artifacts/members` | `?spaceId=&repoId=` |

### Request Body 结构（`model.PermissionRequest`）

```json
{
  "private": true,
  "rules": [
    {
      "accounts": ["alice", "bob"],
      "role": "artifact_reader | artifact_writer | artifact_admin",
      "scope": ". | * | **"
    }
  ]
}
```

`scope` 含义：

| scope | 值 | 生效范围 |
|-------|----|---------|
| single | `.` | 只对当前节点 |
| sub | `*` | 当前节点 + 直接子节点 |
| recursive | `**` | 当前节点 + 所有后代（**最常用**） |
