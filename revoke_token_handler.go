package main

import (
	"net/http"
	"strings"
)

func (a *apiConfig) RevokeTokenHandler(w http.ResponseWriter, req *http.Request) {
	bearer := req.Header.Get("Authorization")
	authToken := strings.Replace(bearer, "Bearer ", "", 1)

	err := a.db.RevokeRefreshToken(authToken)
	if err != nil {
		respondWithError(w, 400, "could not delete token from the db")
		return
	}

	w.WriteHeader(204)
}
