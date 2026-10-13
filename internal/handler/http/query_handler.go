package http

import (
	"encoding/json"
	"net/http"

	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/usecase"
)

type QueryHandler struct {
	queryUseCase *usecase.QueryUseCase
}

func NewQueryHandler(queryUseCase *usecase.QueryUseCase) *QueryHandler {
	return &QueryHandler{
		queryUseCase: queryUseCase,
	}
}

// Request DTO
type QueryRequest struct {
	Question string `json:"question"`
	Limit    int    `json:"limit,omitempty"`
}

// Response DTOs
type SourceDTO struct {
	DocumentTitle string `json:"document_title"`
	SourceURI     string `json:"source_uri"`
	ChunkID       string `json:"chunk_id"`
}

type QueryResponse struct {
	Answer  string      `json:"answer"`
	Sources []SourceDTO `json:"sources"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *QueryHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "payload JSON inválido"})
		return
	}

	if req.Question == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "o campo 'question' é obrigatório"})
		return
	}

	// Executa o UseCase passando context, question e limit
	result, err := h.queryUseCase.Execute(r.Context(), req.Question, req.Limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "falha interna ao processar consulta RAG"})
		return
	}

	// Mapeia os MatchedChunks para o DTO de resposta
	sources := make([]SourceDTO, 0, len(result.MatchedChunks))
	for _, chunk := range result.MatchedChunks {
		sources = append(sources, SourceDTO{
			DocumentTitle: chunk.DocumentTitle,
			SourceURI:     chunk.SourceURI,
			ChunkID:       chunk.ID.String(),
		})
	}

	resp := QueryResponse{
		Answer:  result.Answer,
		Sources: sources,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}