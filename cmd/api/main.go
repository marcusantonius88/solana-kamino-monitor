package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"kamino-simulador/internal/handler"
	"kamino-simulador/internal/repository"
	"kamino-simulador/internal/service"
	solanapkg "kamino-simulador/pkg/solana"
)

func main() {
	logger := log.New(os.Stdout, "[kamino-simulador] ", log.LstdFlags|log.Lshortfile)

	rpcEndpoint := strings.TrimSpace(os.Getenv("SOLANA_RPC_URL"))
	if rpcEndpoint == "" {
		rpcEndpoint = "https://api.mainnet-beta.solana.com"
	}
	solanaClient := solanapkg.NewRPCClient(rpcEndpoint)
	solanaRepository := repository.NewSolanaRepository(solanaClient)
	positionService := service.NewPositionService(solanaRepository)
	positionHandler := handler.NewPositionHandler(positionService, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /positions/{wallet}", positionHandler.GetPositions)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Printf("HTTP server listening on %s", server.Addr)
	logger.Printf("using Solana RPC endpoint: %s", rpcEndpoint)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("failed to start HTTP server: %v", err)
	}
}
