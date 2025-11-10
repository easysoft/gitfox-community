package constant

type SettingsKey string

type SettingsValue string

var (
	SettingsValueTrue  SettingsValue = "true"
	SettingsValueFalse SettingsValue = "false"
)

var (
	// KeySecretScanningEnabled [bool] enables secret scanning if set to true.
	KeySecretScanningEnabled     SettingsKey = "secret_scanning_enabled"
	DefaultSecretScanningEnabled             = false
	KeyAIReviewEnabled           SettingsKey = "ai_review_enabled"
	DefaultAIReviewEnabled                   = false
	KeyAIReviewPrompt            SettingsKey = "ai_custom_prompt"
	DefaultAIReviewPrompt                    = "" // 空字符串表示使用系统默认模板
	KeyZtFlowEnabled             SettingsKey = "ztflow_enabled"
	DefaultZtFlowEnabled                     = false
	KeyArchiveEnabled            SettingsKey = "archive_enabled"
	DefaultArchiveEnabled                    = false
	KeyFileSizeLimit             SettingsKey = "file_size_limit"
	DefaultFileSizeLimit                     = int64(5e+8) // 500MB
	KeyCodeOwnerPushEnabled      SettingsKey = "code_owner_push_enabled"
	DefaultCodeOwnerPushEnabled              = false

	// KeyCommitLimitEnabled [bool] enabled git commit lines limit
	KeyCommitLimitEnabled         SettingsKey = "commit_limit_enabled"
	DefaultCommitLimitEnabled                 = false
	KeyCommitLimitTotalLines      SettingsKey = "commit_limit_total_lines"
	DefaultCommitLimitTotalLines              = 50
	KeyCommitLimitAddLines        SettingsKey = "commit_limit_add_lines"
	DefaultCommitLimitAddLines                = 30
	KeyCommitLimitDeleteLines     SettingsKey = "commit_limit_delete_lines"
	DefaultCommitLimitDeleteLines             = 30
	KeyCommitLimitUpdateLines     SettingsKey = "commit_limit_update_lines"
	DefaultCommitLimitUpdateLines             = 30

	// KeyPushLimitEnabled [bool] enable git push lines limit
	KeyPushLimitEnabled         SettingsKey = "push_limit_enabled"
	DefaultPushLimitEnabled                 = false
	KeyPushLimitTotalLines      SettingsKey = "push_limit_total_lines"
	DefaultPushLimitTotalLines              = 200
	KeyPushLimitAddLines        SettingsKey = "push_limit_add_lines"
	DefaultPushLimitAddLines                = 100
	KeyPushLimitDeleteLines     SettingsKey = "push_limit_delete_lines"
	DefaultPushLimitDeleteLines             = 100
	KeyPushLimitUpdateLines     SettingsKey = "push_limit_update_lines"
	DefaultPushLimitUpdateLines             = 100

	// KeyRequireFitClient [bool] will check the git client
	KeyRequireFitClient     SettingsKey = "require_fit_client"
	DefaultRequireFitClient             = false

	// ContainerReadOnly [bool] disable write operates, used for gc
	ContainerReadOnly SettingsKey = "container_read_only"

	KeyInstallID     SettingsKey = "install_id"
	DefaultInstallID             = string("")
)
