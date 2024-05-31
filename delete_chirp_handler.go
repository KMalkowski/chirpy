package main

import (
	"log"
	"net/http"
	"strconv"
)

func (a *apiConfig) DeleteChirpHandler(w http.ResponseWriter, req *http.Request) {
	chirpId := req.PathValue("id")
	intChirpId, err := strconv.Atoi(chirpId)
	if err != nil {
		respondWithError(w, 400, "invalid chirp id")
		return
	}

	userId := req.Context().Value("userId")
	err = a.db.DeleteChirp(intChirpId, userId.(int))
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, 403, "invalid chirp id")
		return
	}

	w.WriteHeader(204)
}
