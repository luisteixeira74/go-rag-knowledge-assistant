package domain

import (
	"time"

	"github.com/google/uuid"
)

type DocumentChunk struct {
	ID         uuid.UUID `json:"id"`
	DocumentID uuid.UUID `json:"document_id"`
	ChunkIndex int       `json:"chunk_index"`
	Content    string    `json:"content"`
	TokenCount int       `json:"token_count"`
	Embedding  []float32 `json:"embedding,omitempty"`
	Score      float64   `json:"score,omitempty"` // Relevância / Similaridade em buscas
	CreatedAt  time.Time `json:"created_at"`

	// Rastreabilidade (Join leve / Metadados para citação)
	DocumentTitle string `json:"document_title,omitempty"`
	SourceURI     string `json:"source_uri,omitempty"`
}

func NewDocumentChunk(docID uuid.UUID, index int, content string, tokenCount int, embedding []float32) *DocumentChunk {
	return &DocumentChunk{
		ID:         uuid.New(),
		DocumentID: docID,
		ChunkIndex: index,
		Content:    content,
		TokenCount: tokenCount,
		Embedding:  embedding,
		CreatedAt:  time.Now().UTC(),
	}
}
