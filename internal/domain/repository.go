package domain

import "context"

type DocumentRepository interface {
	Save(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id string) (*Document, error)
	List(ctx context.Context) ([]Document, error)
	Delete(ctx context.Context, id string) error
}

type ChunkRepository interface {
	SaveBatch(ctx context.Context, chunks []DocumentChunk) error
	SearchSimilar(ctx context.Context, embedding []float32, limit int) ([]DocumentChunk, error)
	GetByDocumentID(ctx context.Context, documentID string) ([]DocumentChunk, error)
}
