package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentRepository struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Save(ctx context.Context, doc *domain.Document) error {
	query := `
		INSERT INTO documents (id, title, source_uri, mime_type, total_chunks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			source_uri = EXCLUDED.source_uri,
			total_chunks = EXCLUDED.total_chunks,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		doc.ID,
		doc.Title,
		doc.SourceURI,
		doc.MimeType,
		doc.TotalChunks,
		doc.CreatedAt,
		doc.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("erro ao salvar documento: %w", err)
	}

	return nil
}

func (r *DocumentRepository) GetByID(ctx context.Context, id string) (*domain.Document, error) {
	query := `
		SELECT id, title, source_uri, mime_type, total_chunks, created_at, updated_at
		FROM documents
		WHERE id = $1
	`

	var doc domain.Document
	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID,
		&doc.Title,
		&doc.SourceURI,
		&doc.MimeType,
		&doc.TotalChunks,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar documento por ID: %w", err)
	}

	return &doc, nil
}

func (r *DocumentRepository) List(ctx context.Context) ([]domain.Document, error) {
	query := `
		SELECT id, title, source_uri, mime_type, total_chunks, created_at, updated_at
		FROM documents
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar documentos: %w", err)
	}
	defer rows.Close()

	var docs []domain.Document
	for rows.Next() {
		var doc domain.Document
		if err := rows.Scan(
			&doc.ID,
			&doc.Title,
			&doc.SourceURI,
			&doc.MimeType,
			&doc.TotalChunks,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao ler linha do documento: %w", err)
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM documents WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erro ao deletar documento: %w", err)
	}
	return nil
}
