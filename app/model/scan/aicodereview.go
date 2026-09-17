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
