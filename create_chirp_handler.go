package main

import (
	"encoding/json"
	"net/http"
)

func (a *apiConfig) chirpsPostHandler(w http.ResponseWriter, req *http.Request) {
	body := JsonBody{}
	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil || body.Body == "" {
		w.WriteHeader(http.StatusBadRequest)
		errorJson := ErrorResponse{Error: "Something went wrong"}
		err := json.NewEncoder(w).Encode(errorJson)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}

	if len(body.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		dat, err := json.Marshal(ErrorResponse{Error: "Chirp is too long"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(dat)
		return
	}

	userId := req.Context().Value("userId").(int)
	if userId < 1 {
		respondWithError(w, 400, "unknown user")
		return
	}

	chirp, err := a.db.CreateChirp(cleanBadWords(body.Body), userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ValidResponse{chirp})
}
