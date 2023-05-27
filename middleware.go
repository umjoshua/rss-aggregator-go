package main

import (
	"net/http"

	"github.com/umjoshua/rss-aggregator-go/internal/auth"
	"github.com/umjoshua/rss-aggregator-go/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetAPIKey(r.Header)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Couldn't get user")
			return
		}

		user, err := cfg.DB.GetUserByAPIKey(r.Context(), apiKey)

		if err != nil {
			respondWithError(w, http.StatusNotFound, "Couldn't get user")
			return
		}
		handler(w, r, user)
	}
}
