package usecase

import (
	"context"
	"fmt"

	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/domain"
)

type QueryUseCase struct {
	chunkRepo    domain.ChunkRepository
	embedService domain.EmbeddingService
	llmService   domain.LLMService
}

func NewQueryUseCase(
	chunkRepo domain.ChunkRepository,
	embedService domain.EmbeddingService,
	llmService domain.LLMService,
) *QueryUseCase {
	return &QueryUseCase{
		chunkRepo:    chunkRepo,
		embedService: embedService,
		llmService:   llmService,
	}
}

type QueryResult struct {
	Answer        string
	MatchedChunks []domain.DocumentChunk
}

func (uc *QueryUseCase) Execute(ctx context.Context, query string, limit int) (*QueryResult, error) {
	if limit <= 0 {
		limit = 4 // limite padrão de chunks contextuais
	}

	// 1. Gera o embedding vetorial da pergunta do usuário
	queryEmbedding, err := uc.embedService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar embedding da pergunta: %w", err)
	}

	// 2. Busca os K chunks mais similares no PostgreSQL (pgvector)
	contextChunks, err := uc.chunkRepo.SearchSimilar(ctx, queryEmbedding, limit)
	if err != nil {
		return nil, fmt.Errorf("falha na busca de chunks similares: %w", err)
	}

	if len(contextChunks) == 0 {
		return &QueryResult{
			Answer:        "Não encontrei informações relevantes na base de conhecimento para responder a essa pergunta.",
			MatchedChunks: nil,
		}, nil
	}

	// 3. Envia o contexto recuperado + query para a LLM sintetizar a resposta
	answer, err := uc.llmService.GenerateAnswer(ctx, query, contextChunks)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar resposta com LLM: %w", err)
	}

	return &QueryResult{
		Answer:        answer,
		MatchedChunks: contextChunks,
	}, nil
}
