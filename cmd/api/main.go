package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/domain"
	handlerHTTP "github.com/luisteixeira74/go-rag-knowledge-assistant/internal/handler/http"
	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/infra/ollama"
	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/infra/openai"
	repoPostgres "github.com/luisteixeira74/go-rag-knowledge-assistant/internal/repository/postgres"
	"github.com/luisteixeira74/go-rag-knowledge-assistant/internal/usecase"
	pkgPostgres "github.com/luisteixeira74/go-rag-knowledge-assistant/pkg/postgres"
)

func getEnvOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func main() {
	_ = godotenv.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("==========================================================")
	fmt.Println("  🚀 AcmeShop Knowledge Assistant - REST API Server")
	fmt.Println("==========================================================")

	// Conexão PostgreSQL
	dbConfig := pkgPostgres.Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     getEnvOrDefault("DB_USER", "rag_user"),
		Password: getEnvOrDefault("DB_PASSWORD", "rag_password"),
		DBName:   getEnvOrDefault("DB_NAME", "rag_db"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	dbPool, err := pkgPostgres.NewPool(ctx, dbConfig)
	if err != nil {
		log.Fatalf("❌ Erro ao conectar ao PostgreSQL: %v", err)
	}
	defer dbPool.Close()

	chunkRepo := repoPostgres.NewChunkRepository(dbPool)

	// Provedor de IA
	provider := strings.ToLower(getEnvOrDefault("LLM_PROVIDER", "ollama"))
	embeddingSvc, llmSvc, err := createAIServices(provider)
	if err != nil {
		log.Fatalf("❌ Erro ao inicializar serviços de IA: %v", err)
	}

	// Injeta Repositórios e Serviços de IA no UseCase
	queryUseCase := usecase.NewQueryUseCase(chunkRepo, embeddingSvc, llmSvc)

	// Instancia o Handler HTTP
	queryHandler := handlerHTTP.NewQueryHandler(queryUseCase)

	// Router Chi
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	r.Post("/api/v1/query", queryHandler.HandleQuery)

	port := getEnvOrDefault("PORT", "8080")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		fmt.Printf("🌐 Servidor HTTP ouvindo na porta %s...\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Erro no servidor HTTP: %v", err)
		}
	}()

	<-ctx.Done()
	fmt.Println("\n🛑 Encerrando servidor HTTP graciosamente...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)

	fmt.Println("👋 Servidor HTTP finalizado.")
}

func createAIServices(provider string) (domain.EmbeddingService, domain.LLMService, error) {
	switch provider {
	case "ollama":
		host := getEnvOrDefault("OLLAMA_HOST", "http://localhost:11434")
		client := ollama.NewClient(host)
		return client, client, nil

	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, nil, fmt.Errorf("OPENAI_API_KEY não informada no .env")
		}
		client := openai.NewClient(apiKey)
		return client, client, nil

	default:
		return nil, nil, fmt.Errorf("provedor '%s' não suportado", provider)
	}
}