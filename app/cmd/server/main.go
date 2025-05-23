// @title        SWIFT Codes API
// @version      1.0
// @description  REST API that manages SWIFT codes (HQ and branches)
// @termsOfService http://swagger.io/terms/

// @contact.name   Przemyslaw Kukla
// @contact.email  przemek.kukla0703@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host     localhost:8080
// @BasePath /

// @schemes http https

package main

import (
	"context"
	"github.com/joho/godotenv"
	_ "github.com/przemekk6973/swift-code-app/app/docs"
	"github.com/przemekk6973/swift-code-app/app/internal/adapter/api"
	"github.com/przemekk6973/swift-code-app/app/internal/adapter/persistence"
	"github.com/przemekk6973/swift-code-app/app/internal/domain/usecases"
	"github.com/przemekk6973/swift-code-app/app/internal/initializer"
	"github.com/przemekk6973/swift-code-app/app/internal/util"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	_ = godotenv.Load(".env")

	// Init Mongo repo
	uri := os.Getenv("MONGO_URI")
	db := os.Getenv("MONGO_DB")
	coll := os.Getenv("MONGO_COLLECTION")
	repo, err := persistence.NewMongoRepository(uri, db, coll)
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}

	// Load countries map
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

	// Import CSV
	if csvPath := os.Getenv("CSV_PATH"); csvPath != "" {
		if _, err := initializer.ImportCSV(repo, csvPath, countries); err != nil {
			log.Fatalf("CSV import failed: %v", err)
		}
	} else {
		log.Println("CSV_PATH not set, skipping import")
	}

	// Wire up service & API
	svc := usecases.NewSwiftService(repo)
	router := api.SetupRouter(svc)

	// Route for Swagger API
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

	// wait for SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	// shutdown Mongo
	if closer, ok := repo.(interface{ Close(context.Context) error }); ok {
		if err := closer.Close(ctx); err != nil {
			log.Printf("error closing Mongo connection: %v", err)
		}
	}

	log.Println("server exited")
}
