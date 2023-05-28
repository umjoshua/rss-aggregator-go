package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/umjoshua/rss-aggregator-go/internal/database"
)

func (apiCg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Name string
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 500, "Invalid user name passed")
		return
	}

	user, err := apiCg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
	})

	if err != nil {
		respondWithError(w, 500, err.Error())
	}

	respondWithJSON(w, 200, user)
}

func (apiCg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, user)
}

func (apiCg *apiConfig) handlerGetPostsForUser(w http.ResponseWriter, r *http.Request, user database.User) {
	posts, err := apiCg.DB.GetPostsForUser(
		r.Context(), database.GetPostsForUserParams{
			UserID: user.ID,
			Limit:  10,
		})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "get posts faild")
		return
	}
	respondWithJSON(w, http.StatusOK, posts)
}
