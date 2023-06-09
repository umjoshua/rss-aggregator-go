package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/umjoshua/rss-aggregator-go/internal/database"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	godotenv.Load(".env")
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("PORT is not defined")
	}
	dbURL := os.Getenv("DB_URL")

	conn, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("Postgress connection error", err)
	}

	dbConfig := database.New(conn)

	apiCfg := apiConfig{
		DB: dbConfig,
	}

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1router := chi.NewRouter()

	v1router.Get("/healthz", handlerReadiness)
	v1router.Get("/err", handleErr)
	v1router.Post("/users", apiCfg.handlerCreateUser)
	v1router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
	v1router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerFeedCreate))
	v1router.Get("/feeds", apiCfg.handlerGetFeeds)

	v1router.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	v1router.Post("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	v1router.Delete("/feed_follows/{feedFollowId}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))
	v1router.Get("/posts", apiCfg.middlewareAuth(apiCfg.handlerGetPostsForUser))

	router.Mount("/v1", v1router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + port,
	}

	fmt.Println("Server is starting on port", port)
	if err != nil {
		log.Fatal("Server couldn't start")
	}

	go startScraping(dbConfig, 10, time.Minute)

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal("Server couldn't start")
	}
}
