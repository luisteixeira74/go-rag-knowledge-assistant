package usecase

import (
	"context"
	"fmt"

	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/domain"
	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/service/chunker"
)

type IngestUseCase struct {
	docRepo      domain.DocumentRepository
	chunkRepo    domain.ChunkRepository
	embedService domain.EmbeddingService
	chunker      *chunker.TextChunker
}

func NewIngestUseCase(
	docRepo domain.DocumentRepository,
	chunkRepo domain.ChunkRepository,
	embedService domain.EmbeddingService,
	chunker *chunker.TextChunker,
) *IngestUseCase {
	return &IngestUseCase{
		docRepo:      docRepo,
		chunkRepo:    chunkRepo,
		embedService: embedService,
		chunker:      chunker,
	}
}

func (uc *IngestUseCase) Execute(ctx context.Context, doc *domain.Document, rawMarkdown string) error {
	// 1. Salva o registro do documento principal
	if err := uc.docRepo.Save(ctx, doc); err != nil {
		return fmt.Errorf("falha ao salvar documento: %w", err)
	}

	// 2. Fatiar o Markdown usando o TextChunker
	chunks := uc.chunker.SplitMarkdown(doc.ID, doc.Title, doc.SourceURI, rawMarkdown)
	if len(chunks) == 0 {
		return nil
	}

	// 3. Extrai apenas os conteúdos de texto para enviar em lote para a API de Embeddings
	texts := make([]string, len(chunks))
	for i, chk := range chunks {
		texts[i] = chk.Content
	}

	// 4. Gera os vetores em lote (Batch Embedding)
	embeddings, err := uc.embedService.GenerateBatchEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("falha ao gerar embeddings em lote: %w", err)
	}

	// 5. Atualiza os chunks com seus respectivos vetores
	for i, emb := range embeddings {
		chunks[i].Embedding = emb
	}

	// 6. Atualiza total de chunks no documento e salva a contagem final
	doc.TotalChunks = len(chunks)
	if err := uc.docRepo.Save(ctx, doc); err != nil {
		return fmt.Errorf("falha ao atualizar metadados do documento: %w", err)
	}

	// 7. Salva a batch de chunks e seus vetores no PostgreSQL via pgvector
	if err := uc.chunkRepo.SaveBatch(ctx, chunks); err != nil {
		return fmt.Errorf("falha ao salvar batch de chunks no postgres: %w", err)
	}

	return nil
}
