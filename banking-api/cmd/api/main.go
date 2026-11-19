package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"moneyflow/banking-api/internal/accounts"
	"moneyflow/banking-api/internal/database"
	"moneyflow/banking-api/internal/transfers"
)

func main() {
	ctx := context.Background()

	pool, err := database.Connect(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	accounts.NewHandler(accounts.NewRepository(pool)).Register(mux)
	transfers.NewHandler(transfers.NewRepository(pool)).Register(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}