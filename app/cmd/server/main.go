package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/przemekk6973/swift-code-app/app/internal/adapter/api"
	"github.com/przemekk6973/swift-code-app/app/internal/adapter/persistence"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/usecases"
	"github.com/przemekk6973/swift-code-app/app/internal/initializer"
	"github.com/przemekk6973/swift-code-app/app/internal/util"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	_ = godotenv.Load(".env")

	// 1) Init Mongo repo
	uri := os.Getenv("MONGO_URI")
	db := os.Getenv("MONGO_DB")
	coll := os.Getenv("MONGO_COLLECTION")
	repo, err := persistence.NewMongoRepository(uri, db, coll)
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}

	// 2) Wczytaj mapę krajów (jeśli podana)
	countriesPath := os.Getenv("COUNTRIES_CSV")
	var countries map[string]string
	if countriesPath != "" {
		countries, err = util.LoadCountryMap(countriesPath)
		if err != nil {
			log.Fatalf("failed to load countries map: %v", err)
		}
	} else {
		log.Println("COUNTRIES_CSV not set, country‑name validation disabled")
		countries = map[string]string{}
	}

	// 3) Import CSV (jeśli ścieżka ustawiona w .env)
	if csvPath := os.Getenv("CSV_PATH"); csvPath != "" {
		if _, err := initializer.ImportCSV(repo, csvPath, countries); err != nil {
			log.Fatalf("CSV import failed: %v", err)
		}
	} else {
		log.Println("CSV_PATH not set, skipping import")
	}

	// 4) Wire up service & API
	svc := usecases.NewSwiftService(repo)
	router := api.SetupRouter(svc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// start server
	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// czekaj na SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// kontekst z timeoutem na graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// najpierw HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	// potem zamknięcie Mongo
	if closer, ok := repo.(interface{ Close(context.Context) error }); ok {
		if err := closer.Close(ctx); err != nil {
			log.Printf("error closing Mongo connection: %v", err)
		}
	}

	log.Println("server exited cleanly")
}
