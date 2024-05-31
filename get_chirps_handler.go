package main

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/KMalkowski/chirpy/database"
)

func (a *apiConfig) chirpsGetHandler(w http.ResponseWriter, req *http.Request) {
	authorId := req.URL.Query().Get("author_id")
	sortParam := req.URL.Query().Get("sort")

	chirps := []database.Chirp{}
	if len(authorId) > 0 {
		authorIdInt, err := strconv.Atoi(authorId)
		if err != nil {
			if err != nil {
				respondWithError(w, 400, "Invalid authorid")
				return
			}
		}

		c, err := a.db.GetChirps()
		if err != nil {
			respondWithError(w, 500, "Error getting chirps from the db")
			return
		}

		for _, chirp := range c {
			if chirp.AuthorId == authorIdInt {
				chirps = append(chirps, chirp)
			}
		}
	} else {
		c, err := a.db.GetChirps()

		if err != nil {
			respondWithError(w, 500, "Error getting chirps from the db")
			return
		}

		chirps = c
	}

	if len(sortParam) > 0 && sortParam != "asc" && sortParam != "desc" {
		respondWithError(w, 400, "invalid sort param")
		return
	} else if len(sortParam) > 0 && sortParam == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].Id < chirps[j].Id
		})
	} else if len(sortParam) > 0 && sortParam == "asc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].Id > chirps[j].Id
		})
	}

	respondWithJson(w, http.StatusOK, chirps)
}
