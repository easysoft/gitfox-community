# ops_* 50 张表逐表差异（真相版）

> 对比源
> - **A (ZenTao)**：`zentaopms/db/zentao.sql` 中所有 `ops_*` `CREATE TABLE` 定义（聚合终态）
> - **B (GitFox)**：`gitfox-community/app/store/migrate/mysql/0001..0038.up.sql`，按序号顺序回放 `CREATE TABLE` + `ALTER TABLE`（DDL 终态）

> 对比维度：列名、类型、`NOT NULL / NULL / DEFAULT`。
> 等价约定：`int(11) ≡ int`、`tinyint(1) ≡ tinyint`、`DEFAULT NULL ≡ NULL DEFAULT NULL`、`'0' ≡ 0`。COMMENT 文本、AFTER 位置不参与比较。

---

## 总揽

| 项 | 数量 |
|---|---:|
| ZenTao 侧 `ops_*` 表 | 50 |
| GitFox 侧 `ops_*` 表 | 50 |
| 共有表 | 50 |
| ZenTao 独有表 | 0 |
| GitFox 独有表 | 0 |

**差异密度前 10 名（按差异列数排序）**：

| # | 表名 | 差异列数 |
|---:|---|---:|
| 1 | `ops_ppm` | 18 |
| 2 | `ops_schedule_jobs` | 12 |
| 3 | `ops_scan_rules` | 9 |
| 4 | `ops_webhook_executions` | 9 |
| 5 | `ops_branch_ruleset` | 8 |
| 6 | `ops_stage_executions` | 8 |
| 7 | `ops_pipeline` | 7 |
| 8 | `ops_runner` | 7 |
| 9 | `ops_scan_tasks` | 7 |
| 10 | `ops_step_executions` | 7 |

---

## 主题一：text/mediumtext/longtext 列上 NOT NULL vs DEFAULT NULL 反向

> A 组：GitFox 更严格（`NOT NULL`），ZenTao 更宽松（`DEFAULT NULL`）。这是唯一 GitFox 比 ZenTao 严的方向。

### A 组：GitFox 严 / ZenTao 松

| 表 | 列 | ZenTao | GitFox |
|---|---|---|---|
| `ops_migrations` | `version` | `text NULL DEFAULT NULL` | `text NOT NULL` |
| `ops_pipeline_content` | `data` | `longtext NULL DEFAULT NULL` | `longtext NOT NULL` |
| `ops_pipeline_content` | `variables` | `text NULL DEFAULT NULL` | `text NOT NULL` |
| `ops_pipeline_executions` | `params` | `text NULL DEFAULT NULL` | `text NOT NULL` |
| `ops_plugins` | `desc` | `mediumtext NULL DEFAULT NULL` | `mediumtext NOT NULL` |
| `ops_plugins` | `json` | `longtext NULL DEFAULT NULL` | `longtext NOT NULL` |
| `ops_plugins` | `logo` | `mediumtext NULL DEFAULT NULL` | `mediumtext NOT NULL` |
| `ops_plugins` | `yaml` | `longtext NULL DEFAULT NULL` | `longtext NOT NULL` |
| `ops_ppm` | `desc` | `text NULL DEFAULT NULL` | `TEXT NOT NULL` |
| `ops_public_keys` | `content` | `text NULL DEFAULT NULL` | `text NOT NULL` |
| `ops_review_flow` | `definition` | `text NULL DEFAULT NULL` | `TEXT NOT NULL` |
| `ops_scan_rules` | `content` | `text NULL DEFAULT NULL` | `text NOT NULL` |
| `ops_scan_rules` | `contentEn` | `text NULL DEFAULT NULL` | `text NOT NULL` |
| `ops_schedule_jobs` | `data` | `text NULL DEFAULT NULL` | `text NOT NULL` |
| `ops_schedule_jobs` | `lastFailureError` | `text NULL DEFAULT NULL` | `text NOT NULL` |
| `ops_schedule_jobs` | `result` | `text NULL DEFAULT NULL` | `text NOT NULL` |
| `ops_stage_executions` | `errorLog` | `text NULL DEFAULT NULL` | `text NOT NULL` |
| `ops_tokens` | `token` | `text NULL DEFAULT NULL` | `text NOT NULL` |

共 **18** 处。

---

## 主题二：数值列 int unsigned vs int / 缺 default

| 表 | 列 | ZenTao | GitFox |
|---|---|---|---|
| `ops_artifact_groups` | `artifactLibID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_artifact_meta_assets` | `artifactLibID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_artifact_packages` | `artifactLibID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_artifact_tree_nodes` | `artifactLibID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_artifact_tree_nodes` | `linkRecord` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` |
| `ops_artifact_versions` | `packageID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_branch_ruleset` | `branchType` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_branch_ruleset` | `repo` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_branch_type` | `repo` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_pipeline` | `latestVersion` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` |
| `ops_pipeline` | `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_pipeline` | `spaceID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_pipeline_content` | `pipelineID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_pipeline_executions` | `duration` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` |
| `ops_pipeline_executions` | `pipelineID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_plugins` | `groupID` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL` |
| `ops_ppm` | `additions` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` |
| `ops_ppm` | `commitCount` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` |
| `ops_ppm` | `deletions` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` |
| `ops_ppm` | `fileCount` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` |
| `ops_ppm` | `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_ppm` | `sourceRepoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_ppm` | `targetRepoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_repo` | `forkID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` |
| `ops_repo` | `spaceID` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL` |
| `ops_runner` | `heartBeat` | `int unsigned NOT NULL DEFAULT 0` | `int NULL DEFAULT NULL` |
| `ops_scan_issue_task_binds` | `issueID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_issue_task_binds` | `taskID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_issues` | `ignored` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NULL` |
| `ops_scan_issues` | `line` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NOT NULL DEFAULT 0` |
| `ops_scan_issues` | `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_issues` | `ruleID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_plan_conditions` | `planID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_plan_conditions` | `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_plan_conditions` | `threshold` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NOT NULL` |
| `ops_scan_plan_solutions` | `planID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_plan_solutions` | `solutionID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_plans` | `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_ruleset_rules` | `ruleID` | `int unsigned NOT NULL DEFAULT 0` | `INT NOT NULL` |
| `ops_scan_ruleset_rules` | `rulesetID` | `int unsigned NOT NULL DEFAULT 0` | `INT NOT NULL` |
| `ops_scan_solution_rulesets` | `rulesetID` | `int unsigned NOT NULL DEFAULT 0` | `INT NOT NULL` |
| `ops_scan_solution_rulesets` | `solutionID` | `int unsigned NOT NULL DEFAULT 0` | `INT NOT NULL` |
| `ops_scan_tasks` | `executionID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_tasks` | `issueNumber` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NOT NULL DEFAULT 0` |
| `ops_scan_tasks` | `planID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_tasks` | `planNumber` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NOT NULL DEFAULT 0` |
| `ops_scan_tasks` | `repoNumber` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NOT NULL DEFAULT 0` |
| `ops_scan_tasks` | `triggerID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL` |
| `ops_scan_trigger_solutions` | `solutionID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_trigger_solutions` | `triggerID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_triggers` | `planID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_scan_triggers` | `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_schedule_jobs` | `consecutiveFailures` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` |
| `ops_schedule_jobs` | `maxDurationSeconds` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` |
| `ops_schedule_jobs` | `maxRetries` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` |
| `ops_schedule_jobs` | `priority` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` |
| `ops_schedule_jobs` | `runProgress` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` |
| `ops_schedule_jobs` | `totalExecutions` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` |
| `ops_stage_executions` | `duration` | `int unsigned NOT NULL DEFAULT 0` | `int NULL` |
| `ops_stage_executions` | `errorCode` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL` |
| `ops_stage_executions` | `executionID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_stage_executions` | `sn` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_step_executions` | `duration` | `int unsigned NOT NULL DEFAULT 0` | `int NULL` |
| `ops_step_executions` | `errorCode` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL` |
| `ops_step_executions` | `sn` | `tinyint unsigned NOT NULL DEFAULT 0` | `tinyint(1) unsigned NOT NULL` |
| `ops_step_executions` | `stageID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_triggers` | `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` |
| `ops_webhook_executions` | `duration` | `bigint unsigned NOT NULL DEFAULT 0` | `BIGINT NOT NULL DEFAULT '0'` |
| `ops_webhook_executions` | `respStatusCode` | `int unsigned NOT NULL DEFAULT 0` | `INT NOT NULL` |
| `ops_webhook_executions` | `retriggerID` | `int unsigned NOT NULL DEFAULT 0` | `INT unsigned NULL DEFAULT NULL` |
| `ops_webhook_executions` | `retriggerable` | `tinyint unsigned NOT NULL DEFAULT 0` | `TINYINT(1) NOT NULL` |
| `ops_webhook_executions` | `webhookID` | `int unsigned NOT NULL DEFAULT 0` | `INT unsigned NOT NULL` |
| `ops_webhooks` | `repoID` | `int unsigned NOT NULL DEFAULT 0` | `INT unsigned NULL DEFAULT NULL` |
| `ops_webhooks` | `spaceID` | `int unsigned NOT NULL DEFAULT 0` | `INT unsigned NULL DEFAULT NULL` |

共 **74** 处。ZenTao 侧多用 `int unsigned NOT NULL DEFAULT 0`；GitFox 侧或用 `int`、或缺 `DEFAULT 0`、或用 `NULL DEFAULT NULL`。

---

## 主题三：字符串列缺 default

共 **74** 处，模式统一为：ZenTao `NOT NULL DEFAULT ''` vs GitFox `NOT NULL`（无 DEFAULT 声明）。

前 30 行代表：

| 表 | 列 | ZenTao | GitFox |
|---|---|---|---|
| `ops_artifact_libs` | `name` | `varchar(100) NOT NULL DEFAULT ''` | `varchar(100) NOT NULL` |
| `ops_artifact_libs` | `type` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` |
| `ops_branch_ruleset` | `createUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` |
| `ops_branch_ruleset` | `deleteUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` |
| `ops_branch_ruleset` | `forcePushUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` |
| `ops_branch_ruleset` | `ppmCreateUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` |
| `ops_branch_ruleset` | `ppmHandleUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` |
| `ops_branch_ruleset` | `updateUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` |
| `ops_branch_type` | `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_pipeline` | `defaultBranch` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_pipeline` | `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_pipeline` | `scope` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NOT NULL` |
| `ops_pipeline` | `status` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NOT NULL` |
| `ops_pipeline_executions` | `createdBy` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NOT NULL` |
| `ops_pipeline_executions` | `trigger` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` |
| `ops_plugin_group` | `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_plugins` | `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_ppm` | `mergeCheckStatus` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_ppm` | `sourceBranch` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_ppm` | `sourceSHA` | `char(40) NOT NULL DEFAULT ''` | `char(40) NOT NULL` |
| `ops_ppm` | `status` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_ppm` | `targetBranch` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_ppm` | `title` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_public_keys` | `account` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NOT NULL` |
| `ops_public_keys` | `fingerprint` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_public_keys` | `identifier` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_public_keys` | `type` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` |
| `ops_repo` | `defaultBranch` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |
| `ops_repo` | `gitUID` | `char(42) NOT NULL DEFAULT ''` | `char(42) NOT NULL` |
| `ops_repo` | `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` |

---

## 主题四：列缺失（一侧有、另一侧没有）

| 表 | 列 | 哪侧有 | 哪侧无 | 类型/定义 |
|---|---|---|---|---|
| `ops_migrations` | `id` | zentao | gitfox | `int unsigned NOT NULL` |
| `ops_pipeline_executions` | `commitBefore` | zentao | gitfox | `varchar(255) NOT NULL DEFAULT ''` |
| `ops_scan_rule_migrations` | `id` | zentao | gitfox | `int unsigned NOT NULL` |
| `ops_space` | `code` | zentao | gitfox | `varchar(50) NOT NULL DEFAULT ''` |

共 **4** 处。

---

## 主题五：其他类型/默认值差异

| 表 | 列 | ZenTao | GitFox | 差异说明 |
|---|---|---|---|---|
| `ops_ppm` | `mergeMethod` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NULL DEFAULT NULL` | null/default_value |
| `ops_ppm` | `mergeSHA` | `char(40) NOT NULL DEFAULT ''` | `char(40) NULL DEFAULT NULL` | null/default_value |
| `ops_ppm` | `mergeTargetSHA` | `char(40) NOT NULL DEFAULT ''` | `char(40) NULL DEFAULT NULL` | null/default_value |
| `ops_ppm` | `mergedBy` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NULL DEFAULT NULL` | null/default_value |
| `ops_runner` | `createdBy` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NULL DEFAULT NULL` | null/default_value |
| `ops_runner` | `createdDate` | `datetime NULL DEFAULT NULL` | `datetime NOT NULL` | null/default |
| `ops_runner` | `editedBy` | `varchar(30) NOT NULL DEFAULT ''` | `char(30) NULL DEFAULT NULL` | type/null/default_value |
| `ops_runner` | `editedDate` | `datetime NULL DEFAULT NULL` | `datetime NOT NULL` | null/default |
| `ops_scan_issues` | `resolution` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NULL DEFAULT NULL` | null/default_value |
| `ops_scan_rulesets` | `descEn` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NULL DEFAULT NULL` | null/default_value |
| `ops_scan_rulesets` | `lang` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NULL DEFAULT NULL` | null/default_value |
| `ops_scan_solutions` | `desc` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NULL DEFAULT NULL` | null/default_value |
| `ops_scan_solutions` | `descEn` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NULL DEFAULT NULL` | null/default_value |
| `ops_scan_solutions` | `tag` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NULL DEFAULT NULL` | null/default_value |
| `ops_scan_tasks` | `result` | `varchar(20) NOT NULL DEFAULT ''` | `varchar(20) NULL DEFAULT NULL` | null/default_value |
| `ops_webhooks` | `latestExecResult` | `varchar(255) NOT NULL DEFAULT ''` | `VARCHAR(255) NULL DEFAULT NULL` | null/default_value |

共 **16** 处。

---

## 逐表详情

> 只列出有差异的列。若某表两侧一致，写"✓ 一致"。

### `ops_artifact_assets`

✓ 一致

### `ops_artifact_blobs`

✓ 一致

### `ops_artifact_groups`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `artifactLibID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |

### `ops_artifact_libs`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `name` | `varchar(100) NOT NULL DEFAULT ''` | `varchar(100) NOT NULL` | default |
| `type` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |

### `ops_artifact_meta_assets`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `artifactLibID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |

### `ops_artifact_packages`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `artifactLibID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |

### `ops_artifact_tree_nodes`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `artifactLibID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `linkRecord` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` | null/default_value |

### `ops_artifact_versions`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `packageID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |

### `ops_branch_ruleset`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `repo` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `branchType` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `createUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` | default |
| `deleteUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` | default |
| `updateUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` | default |
| `forcePushUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` | default |
| `ppmCreateUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` | default |
| `ppmHandleUser` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` | default |

### `ops_branch_type`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `repo` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |

### `ops_migrations`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `id` | `int unsigned NOT NULL` | _(无)_ | missing_gitfox |
| `version` | `text NULL DEFAULT NULL` | `text NOT NULL` | null/default |

### `ops_pipeline`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `scope` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NOT NULL` | default |
| `spaceID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `status` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NOT NULL` | default |
| `latestVersion` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` | type |
| `defaultBranch` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |

### `ops_pipeline_content`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `pipelineID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `data` | `longtext NULL DEFAULT NULL` | `longtext NOT NULL` | null/default |
| `variables` | `text NULL DEFAULT NULL` | `text NOT NULL` | null/default |

### `ops_pipeline_executions`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `pipelineID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `trigger` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |
| `params` | `text NULL DEFAULT NULL` | `text NOT NULL` | null/default |
| `duration` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` | type |
| `createdBy` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NOT NULL` | default |
| `commitBefore` | `varchar(255) NOT NULL DEFAULT ''` | _(无)_ | missing_gitfox |

### `ops_plugin_group`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |

### `ops_plugins`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `groupID` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL` | type/default |
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `desc` | `mediumtext NULL DEFAULT NULL` | `mediumtext NOT NULL` | null/default |
| `logo` | `mediumtext NULL DEFAULT NULL` | `mediumtext NOT NULL` | null/default |
| `yaml` | `longtext NULL DEFAULT NULL` | `longtext NOT NULL` | null/default |
| `json` | `longtext NULL DEFAULT NULL` | `longtext NOT NULL` | null/default |

### `ops_ppm`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `title` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `desc` | `text NULL DEFAULT NULL` | `TEXT NOT NULL` | null/default |
| `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `sourceRepoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `sourceBranch` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `sourceSHA` | `char(40) NOT NULL DEFAULT ''` | `char(40) NOT NULL` | default |
| `targetRepoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `targetBranch` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `mergedBy` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NULL DEFAULT NULL` | null/default_value |
| `mergeMethod` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NULL DEFAULT NULL` | null/default_value |
| `mergeTargetSHA` | `char(40) NOT NULL DEFAULT ''` | `char(40) NULL DEFAULT NULL` | null/default_value |
| `mergeSHA` | `char(40) NOT NULL DEFAULT ''` | `char(40) NULL DEFAULT NULL` | null/default_value |
| `mergeCheckStatus` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `commitCount` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` | null/default_value |
| `fileCount` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` | null/default_value |
| `additions` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` | null/default_value |
| `deletions` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` | null/default_value |
| `status` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |

### `ops_provider`

✓ 一致

### `ops_public_keys`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `account` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NOT NULL` | default |
| `identifier` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `fingerprint` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `content` | `text NULL DEFAULT NULL` | `text NOT NULL` | null/default |
| `type` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |

### `ops_repo`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `spaceID` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL` | type/default |
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `gitUID` | `char(42) NOT NULL DEFAULT ''` | `char(42) NOT NULL` | default |
| `forkID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL DEFAULT NULL` | null/default_value |
| `defaultBranch` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |

### `ops_repobranch`

✓ 一致

### `ops_repofiles`

✓ 一致

### `ops_repohistory`

✓ 一致

### `ops_repouser`

✓ 一致

### `ops_request_reviewers`

✓ 一致

### `ops_review_flow`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `definition` | `text NULL DEFAULT NULL` | `TEXT NOT NULL` | null/default |

### `ops_runner`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `name` | `varchar(200) NOT NULL DEFAULT ''` | `varchar(200) NOT NULL` | default |
| `token` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `heartBeat` | `int unsigned NOT NULL DEFAULT 0` | `int NULL DEFAULT NULL` | type/null/default_value |
| `createdBy` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NULL DEFAULT NULL` | null/default_value |
| `createdDate` | `datetime NULL DEFAULT NULL` | `datetime NOT NULL` | null/default |
| `editedBy` | `varchar(30) NOT NULL DEFAULT ''` | `char(30) NULL DEFAULT NULL` | type/null/default_value |
| `editedDate` | `datetime NULL DEFAULT NULL` | `datetime NOT NULL` | null/default |

### `ops_scan_issue_task_binds`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `taskID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `issueID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |

### `ops_scan_issues`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `issueKey` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `ruleID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `line` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NOT NULL DEFAULT 0` | type |
| `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `resolution` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NULL DEFAULT NULL` | null/default_value |
| `ignored` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NULL` | type/null/default_value |

### `ops_scan_plan_conditions`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `planID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `priority` | `varchar(20) NOT NULL DEFAULT ''` | `varchar(20) NOT NULL` | default |
| `type` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |
| `unit` | `varchar(20) NOT NULL DEFAULT ''` | `varchar(20) NOT NULL` | default |
| `threshold` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NOT NULL` | type/default |

### `ops_scan_plan_solutions`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `planID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `solutionID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |

### `ops_scan_plans`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |

### `ops_scan_rule_migrations`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `id` | `int unsigned NOT NULL` | _(无)_ | missing_gitfox |
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `VARCHAR(255) NOT NULL` | default |
| `version` | `varchar(255) NOT NULL DEFAULT ''` | `VARCHAR(255) NOT NULL` | default |

### `ops_scan_rules`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `ruleKey` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `lang` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |
| `plugin` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |
| `tag` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |
| `priority` | `varchar(20) NOT NULL DEFAULT ''` | `varchar(20) NOT NULL` | default |
| `type` | `varchar(20) NOT NULL DEFAULT ''` | `varchar(20) NOT NULL` | default |
| `content` | `text NULL DEFAULT NULL` | `text NOT NULL` | null/default |
| `contentEn` | `text NULL DEFAULT NULL` | `text NOT NULL` | null/default |

### `ops_scan_ruleset_rules`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `rulesetID` | `int unsigned NOT NULL DEFAULT 0` | `INT NOT NULL` | type/default |
| `ruleID` | `int unsigned NOT NULL DEFAULT 0` | `INT NOT NULL` | type/default |

### `ops_scan_rulesets`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `descEn` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NULL DEFAULT NULL` | null/default_value |
| `type` | `varchar(20) NOT NULL DEFAULT ''` | `varchar(20) NOT NULL` | default |
| `lang` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NULL DEFAULT NULL` | null/default_value |
| `plugin` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |
| `tag` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |

### `ops_scan_solution_rulesets`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `solutionID` | `int unsigned NOT NULL DEFAULT 0` | `INT NOT NULL` | type/default |
| `rulesetID` | `int unsigned NOT NULL DEFAULT 0` | `INT NOT NULL` | type/default |

### `ops_scan_solutions`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `desc` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NULL DEFAULT NULL` | null/default_value |
| `descEn` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NULL DEFAULT NULL` | null/default_value |
| `tag` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NULL DEFAULT NULL` | null/default_value |

### `ops_scan_tasks`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `planID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `executionID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `repoNumber` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NOT NULL DEFAULT 0` | type |
| `planNumber` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NOT NULL DEFAULT 0` | type |
| `triggerID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NULL` | null/default_value |
| `result` | `varchar(20) NOT NULL DEFAULT ''` | `varchar(20) NULL DEFAULT NULL` | null/default_value |
| `issueNumber` | `bigint unsigned NOT NULL DEFAULT 0` | `bigint NOT NULL DEFAULT 0` | type |

### `ops_scan_trigger_solutions`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `triggerID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `solutionID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |

### `ops_scan_triggers`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `planID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `type` | `varchar(20) NOT NULL DEFAULT ''` | `varchar(20) NOT NULL` | default |
| `scanType` | `varchar(20) NOT NULL DEFAULT ''` | `varchar(20) NOT NULL` | default |

### `ops_schedule_jobs`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `name` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `type` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `priority` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` | type |
| `data` | `text NULL DEFAULT NULL` | `text NOT NULL` | null/default |
| `result` | `text NULL DEFAULT NULL` | `text NOT NULL` | null/default |
| `maxDurationSeconds` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` | type |
| `maxRetries` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` | type |
| `state` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `totalExecutions` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` | type |
| `runProgress` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` | type |
| `consecutiveFailures` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL DEFAULT 0` | type |
| `lastFailureError` | `text NULL DEFAULT NULL` | `text NOT NULL` | null/default |

### `ops_space`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `name` | `varchar(200) NOT NULL DEFAULT ''` | `varchar(200) NOT NULL` | default |
| `code` | `varchar(50) NOT NULL DEFAULT ''` | _(无)_ | missing_gitfox |

### `ops_spaceuser`

✓ 一致

### `ops_stage_executions`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `executionID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `code` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `sn` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `duration` | `int unsigned NOT NULL DEFAULT 0` | `int NULL` | type/null/default_value |
| `status` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |
| `errorType` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |
| `errorCode` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL` | type/default |
| `errorLog` | `text NULL DEFAULT NULL` | `text NOT NULL` | null/default |

### `ops_step_executions`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `stageID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |
| `code` | `varchar(100) NOT NULL DEFAULT ''` | `varchar(100) NOT NULL` | default |
| `sn` | `tinyint unsigned NOT NULL DEFAULT 0` | `tinyint(1) unsigned NOT NULL` | default |
| `duration` | `int unsigned NOT NULL DEFAULT 0` | `int NULL` | type/null/default_value |
| `status` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |
| `errorCode` | `int unsigned NOT NULL DEFAULT 0` | `int NOT NULL` | type/default |
| `errorLog` | `varchar(500) NOT NULL DEFAULT ''` | `varchar(500) NOT NULL` | default |

### `ops_tokens`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `account` | `varchar(30) NOT NULL DEFAULT ''` | `varchar(30) NOT NULL` | default |
| `identifier` | `varchar(255) NOT NULL DEFAULT ''` | `varchar(255) NOT NULL` | default |
| `type` | `varchar(50) NOT NULL DEFAULT ''` | `varchar(50) NOT NULL` | default |
| `token` | `text NULL DEFAULT NULL` | `text NOT NULL` | null/default |

### `ops_triggers`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `repoID` | `int unsigned NOT NULL DEFAULT 0` | `int unsigned NOT NULL` | default |

### `ops_webhook_executions`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `webhookID` | `int unsigned NOT NULL DEFAULT 0` | `INT unsigned NOT NULL` | default |
| `triggerType` | `varchar(255) NOT NULL DEFAULT ''` | `VARCHAR(255) NOT NULL` | default |
| `eventID` | `varchar(255) NOT NULL DEFAULT ''` | `VARCHAR(255) NOT NULL` | default |
| `duration` | `bigint unsigned NOT NULL DEFAULT 0` | `BIGINT NOT NULL DEFAULT '0'` | type |
| `reqUrl` | `varchar(255) NOT NULL DEFAULT ''` | `VARCHAR(255) NOT NULL` | default |
| `respStatusCode` | `int unsigned NOT NULL DEFAULT 0` | `INT NOT NULL` | type/default |
| `respStatus` | `varchar(255) NOT NULL DEFAULT ''` | `VARCHAR(255) NOT NULL` | default |
| `retriggerID` | `int unsigned NOT NULL DEFAULT 0` | `INT unsigned NULL DEFAULT NULL` | null/default_value |
| `retriggerable` | `tinyint unsigned NOT NULL DEFAULT 0` | `TINYINT(1) NOT NULL` | default |

### `ops_webhooks`

| 列 | ZenTao | GitFox | 差异 |
|---|---|---|---|
| `spaceID` | `int unsigned NOT NULL DEFAULT 0` | `INT unsigned NULL DEFAULT NULL` | null/default_value |
| `repoID` | `int unsigned NOT NULL DEFAULT 0` | `INT unsigned NULL DEFAULT NULL` | null/default_value |
| `displayName` | `varchar(255) NOT NULL DEFAULT ''` | `VARCHAR(255) NOT NULL` | default |
| `url` | `varchar(255) NOT NULL DEFAULT ''` | `VARCHAR(255) NOT NULL` | default |
| `triggers` | `varchar(500) NOT NULL DEFAULT ''` | `VARCHAR(500) NOT NULL` | default |
| `latestExecResult` | `varchar(255) NOT NULL DEFAULT ''` | `VARCHAR(255) NULL DEFAULT NULL` | null/default_value |

---

## 附录：等价约定

以下 DDL 差异**视为等价，不计入差异**：

- `int(11) ≡ int`、`bigint(20) ≡ bigint`、`smallint(6) ≡ smallint`、`tinyint(4) ≡ tinyint`
- `tinyint(1) ≡ tinyint`（含布尔语义列，如 `deleted`、`isSSL`、`enabled` 等）
- `NULL DEFAULT NULL ≡ DEFAULT NULL`
- 数值默认 `'0' ≡ 0`、`'1' ≡ 1`
- 前导/尾随空白、COMMENT 文本差异
- 索引/外键、`AFTER col` 位置声明（不影响列本身定义）

