package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (a *apiConfig) ChirpsGetByIdHandler(w http.ResponseWriter, req *http.Request) {
	chirps, err := a.db.GetChirps()
	if err != nil {
		respondWithError(w, 500, "coudn't get the chiprs from the databse")
	}

	pathvalue := req.PathValue("id")
	id, err := strconv.Atoi(pathvalue)

	if err != nil {
		respondWithError(w, 400, "Chirp id has to be a number")
		return
	}

	if len(chirps) < id {
		respondWithError(w, 404, "Chirp has not been found in the database")
	}

	var index int
	for i, chirp := range chirps {
		if chirp.Id == id {
			index = i
			break
		}
	}

	data, err := json.Marshal(chirps[index])

	if err != nil {
		respondWithError(w, 500, "Coulnd't encode chirp properly")
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
