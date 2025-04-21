package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/przemekk6973/swift-code-app/app/internal/adapter/api"
	"github.com/przemekk6973/swift-code-app/app/internal/adapter/persistence"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/usecases"
	"github.com/przemekk6973/swift-code-app/app/internal/initializer"
	"github.com/przemekk6973/swift-code-app/app/internal/util"
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
	log.Printf("listening on :%s", port)
	router.Run(":" + port)
}
