package postgres

import (
	"context"
	"fmt"

	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"
)

type ChunkRepository struct {
	db *pgxpool.Pool
}

func NewChunkRepository(db *pgxpool.Pool) *ChunkRepository {
	return &ChunkRepository{db: db}
}

func (r *ChunkRepository) SaveBatch(ctx context.Context, chunks []domain.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO document_chunks (id, document_id, chunk_index, content, token_count, embedding, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	for _, chunk := range chunks {
		vec := pgvector.NewVector(chunk.Embedding)
		batch.Queue(query,
			chunk.ID,
			chunk.DocumentID,
			chunk.ChunkIndex,
			chunk.Content,
			chunk.TokenCount,
			vec,
			chunk.CreatedAt,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(chunks); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("erro ao executar inserção em lote no chunk index %d: %w", i, err)
		}
	}

	return nil
}

func (r *ChunkRepository) SearchSimilar(ctx context.Context, embedding []float32, limit int) ([]domain.DocumentChunk, error) {
	query := `
		SELECT 
			c.id, 
			c.document_id, 
			c.chunk_index, 
			c.content, 
			c.token_count, 
			c.created_at,
			d.title AS document_title,
			d.source_uri,
			1 - (c.embedding <=> $1) AS similarity_score
		FROM document_chunks c
		JOIN documents d ON d.id = c.document_id
		ORDER BY c.embedding <=> $1
		LIMIT $2
	`

	vec := pgvector.NewVector(embedding)
	rows, err := r.db.Query(ctx, query, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("erro na busca vetorial por similaridade: %w", err)
	}
	defer rows.Close()

	var chunks []domain.DocumentChunk
	for rows.Next() {
		var chunk domain.DocumentChunk
		if err := rows.Scan(
			&chunk.ID,
			&chunk.DocumentID,
			&chunk.ChunkIndex,
			&chunk.Content,
			&chunk.TokenCount,
			&chunk.CreatedAt,
			&chunk.DocumentTitle,
			&chunk.SourceURI,
			&chunk.Score,
		); err != nil {
			return nil, fmt.Errorf("erro ao ler linha do chunk similar: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func (r *ChunkRepository) GetByDocumentID(ctx context.Context, documentID string) ([]domain.DocumentChunk, error) {
	query := `
		SELECT id, document_id, chunk_index, content, token_count, created_at
		FROM document_chunks
		WHERE document_id = $1
		ORDER BY chunk_index ASC
	`

	rows, err := r.db.Query(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar chunks por documento: %w", err)
	}
	defer rows.Close()

	var chunks []domain.DocumentChunk
	for rows.Next() {
		var chunk domain.DocumentChunk
		if err := rows.Scan(
			&chunk.ID,
			&chunk.DocumentID,
			&chunk.ChunkIndex,
			&chunk.Content,
			&chunk.TokenCount,
			&chunk.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao ler linha do chunk: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
