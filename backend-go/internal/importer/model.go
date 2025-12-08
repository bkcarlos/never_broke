package importer

// ImportPreviewItem 导入预览项
type ImportPreviewItem struct {
	Date          string  `json:"date"`
	CategoryGuess string  `json:"categoryGuess"`
	Amount        float64 `json:"amount"`
	Remark        string  `json:"remark"`
}

// ImportPreviewResponse 导入预览响应
type ImportPreviewResponse struct {
	Count int                 `json:"count"`
	Items []ImportPreviewItem `json:"items"`
}

// ImportPreviewRequest 导入预览请求
type ImportPreviewRequest struct {
	Raw string `json:"raw"`
}
