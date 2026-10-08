package chunker

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/domain"
)

// Garante conformidade em tempo de compilação com a interface do domínio.
var _ domain.ChunkerService = (*TextChunker)(nil)

type TextChunker struct {
	defaultChunkSize int
	defaultOverlap   int
}

func NewTextChunker(chunkSize, overlap int) *TextChunker {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	if overlap < 0 {
		overlap = 100
	}
	return &TextChunker{
		defaultChunkSize: chunkSize,
		defaultOverlap:   overlap,
	}
}

// SplitText implementa domain.ChunkerService dividindo texto em fatias com overlap.
func (c *TextChunker) SplitText(text string, opts domain.ChunkingOptions) ([]string, error) {
	chunkSize := opts.MaxChunkSize
	if chunkSize <= 0 {
		chunkSize = c.defaultChunkSize
	}
	overlap := opts.Overlap
	if overlap < 0 {
		overlap = c.defaultOverlap
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	paragraphs := strings.Split(text, "\n\n")

	var chunks []string
	var currentChunk strings.Builder

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if len(p) == 0 {
			continue
		}

		if currentChunk.Len()+len(p) > chunkSize && currentChunk.Len() > 0 {
			prevContent := strings.TrimSpace(currentChunk.String())
			chunks = append(chunks, prevContent)

			currentChunk.Reset()

			if len(prevContent) > overlap {
				// Pega o trecho final tentando respeitar limites de palavras/espaços
				overlapStart := len(prevContent) - overlap
				if spaceIdx := strings.IndexByte(prevContent[overlapStart:], ' '); spaceIdx != -1 {
					overlapStart += spaceIdx + 1
				}
				currentChunk.WriteString(prevContent[overlapStart:])
				currentChunk.WriteString("\n\n")
			}
		}

		currentChunk.WriteString(p)
		currentChunk.WriteString("\n\n")
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	return chunks, nil
}

// BuildDocumentChunks realiza a conversão direta de texto para domain.DocumentChunk.
func (c *TextChunker) BuildDocumentChunks(
	ctx context.Context,
	docID uuid.UUID,
	docTitle string,
	sourceURI string,
	content string,
) ([]domain.DocumentChunk, error) {
	rawChunks, err := c.SplitText(content, domain.ChunkingOptions{
		MaxChunkSize: c.defaultChunkSize,
		Overlap:      c.defaultOverlap,
	})
	if err != nil {
		return nil, err
	}

	documentChunks := make([]domain.DocumentChunk, 0, len(rawChunks))
	for i, rawContent := range rawChunks {
		tokenEstimate := len(rawContent) / 4

		chunk := *domain.NewDocumentChunk(docID, i, rawContent, tokenEstimate, nil)
		chunk.DocumentTitle = docTitle
		chunk.SourceURI = sourceURI

		documentChunks = append(documentChunks, chunk)
	}

	return documentChunks, nil
}

// SplitMarkdown expõe alias conveniente para use cases sem necessidade de passar contexto.
func (c *TextChunker) SplitMarkdown(
	docID uuid.UUID,
	docTitle string,
	sourceURI string,
	content string,
) []domain.DocumentChunk {
	chunks, err := c.BuildDocumentChunks(context.Background(), docID, docTitle, sourceURI, content)
	if err != nil {
		return nil
	}
	return chunks
}
