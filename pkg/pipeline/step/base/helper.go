package base

const (
	// envMavenMirror Maven 仓库镜像地址环境变量名（maven/gradle 共用）
	envMavenMirror = "MAVEN_MIRROR_URL"
	// defaultMavenMirror 默认的 Maven 仓库镜像地址
	defaultMavenMirror = "https://maven.aliyun.com/repository/public"
)

func getOrDefaultTrue(b *bool) bool {
	if b == nil {
		return true
	}

	return *b
}
