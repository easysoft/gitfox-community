package render

type Response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	TraceID string      `json:"traceID"`
	Data    interface{} `json:"data,omitempty"`
}

type ListResponse struct {
	Response
	ListArgs
}

type ListArgs struct {
	Total    int64 `json:"total"`
	Page     int64 `json:"page"`
	PageSize int64 `json:"pageSize"`
}
