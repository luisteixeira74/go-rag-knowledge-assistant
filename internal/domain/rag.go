package domain

// Citation representa a fonte original usada na geração da resposta.
type Citation struct {
	DocumentID    string  `json:"document_id"`
	DocumentTitle string  `json:"document_title"`
	SourceURI     string  `json:"source_uri"`
	ChunkIndex    int     `json:"chunk_index"`
	Content       string  `json:"content"`
	Score         float64 `json:"similarity_score"`
}

// RAGResponse é o retorno final entregue ao usuário ou API.
type RAGResponse struct {
	Query     string     `json:"query"`
	Answer    string     `json:"answer"`
	Citations []Citation `json:"citations"`
}
