package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/domain"
)

type Client struct {
	apiKey          string
	embeddingModel  string
	completionModel string
	httpClient      *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:          apiKey,
		embeddingModel:  "text-embedding-3-small",
		completionModel: "gpt-4o-mini",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Garantindo conformidade de tempo de compilação com as interfaces do domain
var _ domain.EmbeddingService = (*Client)(nil)
var _ domain.LLMService = (*Client)(nil)

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := c.GenerateBatchEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("nenhum vetor retornado")
	}
	return embeddings[0], nil
}

func (c *Client) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := embeddingRequest{
		Model: c.embeddingModel,
		Input: texts,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar request de embedding: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisicao HTTP de embedding: %w", err)
	}
	defer resp.Body.Close()

	var res embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta de embedding: %w", err)
	}

	if res.Error != nil {
		return nil, fmt.Errorf("erro da API OpenAI: %s", res.Error.Message)
	}

	result := make([][]float32, len(res.Data))
	for _, item := range res.Data {
		result[item.Index] = item.Embedding
	}

	return result, nil
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) GenerateAnswer(ctx context.Context, query string, contextChunks []domain.DocumentChunk) (string, error) {
	var contextBuilder strings.Builder
	for i, chunk := range contextChunks {
		contextBuilder.WriteString(fmt.Sprintf("\n--- DOCUMENTO %d (%s) ---\n%s\n", i+1, chunk.DocumentTitle, chunk.Content))
	}

	systemPrompt := `Você é o assistente virtual corporativo da AcmeShop.
Responda às perguntas dos colaboradores utilizando EXCLUSIVAMENTE os documentos fornecidos no contexto.
Se a informação não estiver presente no contexto, responda educadamente: "Não encontrei informações suficientes na base de conhecimento para responder a essa pergunta."
Não invente nem assuma informações fora do contexto.`

	userPrompt := fmt.Sprintf("Contexto de Documentos:\n%s\n\nPergunta do Colaborador: %s", contextBuilder.String(), query)

	reqBody := chatRequest{
		Model: c.completionModel,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar request do chat: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro na requisicao HTTP de chat: %w", err)
	}
	defer resp.Body.Close()

	var res chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta de chat: %w", err)
	}

	if res.Error != nil {
		return "", fmt.Errorf("erro da API OpenAI: %s", res.Error.Message)
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("nenhuma resposta gerada")
	}

	return res.Choices[0].Message.Content, nil
}
