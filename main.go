package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	database "github.com/rammyblog/rssagg/internal/database"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Port is not available")
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB is not available")
	}

	conn, dbConnectionErr := sql.Open("postgres", dbUrl)

	if dbConnectionErr != nil {
		log.Fatal("Cant connect to db", dbConnectionErr)
	}

	apiCfg := apiConfig{
		DB: database.New(conn),
	}

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	server := &http.Server{
		Handler: router,
		Addr:    ":" + port,
	}
	v1Router := chi.NewRouter()

	v1Router.Get("/health", handlerReadiness)
	v1Router.Get("/err", handlerErr)
	v1Router.Post("/users", apiCfg.handlerCreateUser)
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUserByApiKey))
	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)
	v1Router.Post("/feeds_follows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	v1Router.Get("/feeds_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollow))
	v1Router.Delete("/feeds_follows/{feedFollowId}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))


	router.Mount("/v1", v1Router)

	log.Printf("Server starting on port %v", port)
	err := server.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
}
