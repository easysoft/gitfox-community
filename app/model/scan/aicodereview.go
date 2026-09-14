package scan

type AIReviewRule struct {
	ID        uint     `json:"id"`
	Language  string   `json:"language"`
	Extension []string `json:"extension"`
	Rule      string   `json:"rule"`
}

// AIReviewConfig AI 代码评审（OCR LLM）配置。
// aireview 插件在非内嵌流水线（INNER_PIPELINE 非 "true"）中通过
// GET /scan/aicodereview/ai-config 拉取，六字段均必填非空。
type AIReviewConfig struct {
	LLMProvider string `json:"llmProvider"`
	LLMUrl      string `json:"llmUrl"`
	LLMProtocol string `json:"llmProtocol"`
	LLMApiKey   string `json:"llmApiKey"`
	LLMModel    string `json:"llmModel"`
	LLMLanguage string `json:"llmLanguage"`
}

// AIReviewRange AI 评审范围（提交区间）。
// 非内嵌流水线（INNER_PIPELINE 非 "true"）中由插件通过
// GET /scan/aicodereview/review-range 拉取：取该分支最新两笔提交，
// commitBefore 为次新、commitAfter 为最新；分支只有一笔提交时两者相同，
// ocr 会以 skipped 空报告返回（不会失败）。插件据此写入
// GITFOX_COMMIT_BEFORE / GITFOX_COMMIT_AFTER（--from / --to）。
type AIReviewRange struct {
	CommitBefore string `json:"commitBefore"`
	CommitAfter  string `json:"commitAfter"`
}
