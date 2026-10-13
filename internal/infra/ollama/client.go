package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/domain"
)

// Garante em tempo de compilação que o Client atende às interfaces do domínio
var _ domain.EmbeddingService = (*Client)(nil)
var _ domain.LLMService = (*Client)(nil)

type Client struct {
	baseURL        string
	llmModel       string
	embeddingModel string
	httpClient     *http.Client
}

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = getEnvOrDefault("OLLAMA_HOST", "http://localhost:11434")
	}

	llmModel := getEnvOrDefault("LLM_MODEL_NAME", "llama3.2:1b")
	embeddingModel := getEnvOrDefault("EMBEDDING_MODEL_NAME", "nomic-embed-text")

	return &Client{
		baseURL:        baseURL,
		llmModel:       llmModel,
		embeddingModel: embeddingModel,
		httpClient:     &http.Client{Timeout: 90 * time.Second},
	}
}

// ============================================================================
// EMBEDDINGS SERVICE (/api/embed)
// ============================================================================

type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	reqBody, err := json.Marshal(embedRequest{
		Model: c.embeddingModel,
		Input: text,
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar request de embedding: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/embed", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar request HTTP para Ollama: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha na comunicação com container Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama respondeu com status HTTP %d ao chamar /api/embed", resp.StatusCode)
	}

	var res embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("falha ao deserializar resposta de embedding: %w", err)
	}

	if len(res.Embeddings) == 0 {
		return nil, fmt.Errorf("resposta de embedding do Ollama veio vazia")
	}

	return res.Embeddings[0], nil
}

func (c *Client) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := c.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("falha no lote de embeddings (índice %d): %w", i, err)
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

// ============================================================================
// LLM SERVICE (/api/chat)
// ============================================================================

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatResponse struct {
	Message chatMessage `json:"message"`
}

// GenerateAnswer satisfaz a interface domain.LLMService via endpoint /api/chat
func (c *Client) GenerateAnswer(ctx context.Context, query string, contextChunks []domain.DocumentChunk) (string, error) {
	var chunkTexts []string
	for _, chunk := range contextChunks {
		chunkTexts = append(chunkTexts, chunk.Content)
	}

	contextText := strings.Join(chunkTexts, "\n---\n")

	systemPrompt := "Você é um assistente prestativo especialista da AcmeShop. Responda à pergunta do usuário de forma clara, direta e objetiva utilizando APENAS as informações fornecidas nos contextos."
	userPrompt := fmt.Sprintf("Contextos de Conhecimento:\n%s\n\nPergunta:\n%s", contextText, query)

	reqBody, err := json.Marshal(chatRequest{
		Model: c.llmModel,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("falha ao serializar request de chat: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("falha ao criar request HTTP para Ollama: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("falha na comunicação com container Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama respondeu com status HTTP %d ao chamar /api/chat", resp.StatusCode)
	}

	var res chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("falha ao deserializar resposta do chat: %w", err)
	}

	return res.Message.Content, nil
}