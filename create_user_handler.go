package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type CreateUserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *apiConfig) CreateUserHandler(w http.ResponseWriter, req *http.Request) {
	body := CreateUserBody{}
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		respondWithError(w, 400, "Could not read the body of request")
		return
	}

	log.Println(body)
	user, err := a.db.CreateUser(body.Email, body.Password)
	if err != nil {
		log.Fatalln(err.Error())
		respondWithError(w, 400, "error writing to the database")
		return
	}

	respondWithJson(w, 201, user)
}
