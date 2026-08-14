package scan

type AIReviewRule struct {
	ID        uint     `json:"id"`
	Language  string   `json:"language"`
	Extention []string `json:"extention"`
	Rule      string   `json:"rule"`
}
