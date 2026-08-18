package scan

type AIReviewRule struct {
	ID        uint     `json:"id"`
	Language  string   `json:"language"`
	Extension []string `json:"extension"`
	Rule      string   `json:"rule"`
}
