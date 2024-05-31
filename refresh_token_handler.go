package main

import (
	"fmt"
	"net/http"
	"strings"
)

func (a *apiConfig) RefreshTokenHandler(w http.ResponseWriter, req *http.Request) {
	bearer := req.Header.Get("Authorization")
	authToken := strings.Replace(bearer, "Bearer ", "", 1)

	database, err := a.db.ReadDatabase()
	if err != nil {
		respondWithError(w, 400, "provide your refresh token")
		return
	}

	var userId int
	for _, token := range database.RefreshTokens {
		if token.Token == authToken {
			userId = token.UserId
			break
		}
	}

	if userId == 0 {
		respondWithError(w, 401, "you don't have access")
		return
	}

	accessToken, err := a.GenerateAccessToken(3600, fmt.Sprintf("%v", userId))
	if err != nil {
		respondWithError(w, 500, "could not generate accesst token")
		return
	}

	type RefreshResponse struct {
		Token string `json:"token"`
	}

	println(accessToken, "test ")

	respondWithJson(w, 200, RefreshResponse{Token: accessToken})
}
