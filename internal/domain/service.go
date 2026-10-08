package domain

import "context"

type ChunkingOptions struct {
	MaxChunkSize int // Ex: 500 tokens ou caracteres
	Overlap      int // Ex: 50 tokens ou caracteres
}

type ChunkerService interface {
	SplitText(text string, opts ChunkingOptions) ([]string, error)
}

type EmbeddingService interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
}

type LLMService interface {
	GenerateAnswer(ctx context.Context, query string, contextChunks []DocumentChunk) (string, error)
}
