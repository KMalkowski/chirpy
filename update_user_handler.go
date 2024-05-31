package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (a *apiConfig) UpdateUserHandler(w http.ResponseWriter, req *http.Request) {
	bearer := req.Header.Get("Authorization")
	token := strings.Replace(bearer, "Bearer ", "", 1)

	if len(token) == 0 {
		respondWithError(w, 401, "you are not authorized")
		return
	}

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		respondWithError(w, 401, "you are not authorized")

		log.Printf("error parsing with claims", err.Error(), token)
		return
	}

	subject, err := claims.GetSubject()
	if err != nil {
		respondWithError(w, 401, "you are not authorized")
		log.Printf("error no subject")
		return
	}

	data := CreateUserBody{}
	err = json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		respondWithError(w, 400, "bad request")
		log.Printf("error decoding user body")
		return
	}

	intsub, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(w, 400, "bad request")
		log.Printf("error string to int")
		return
	}

	user, err := a.db.UpdateUser(intsub, data.Email, data.Password)

	respondWithJson(w, 200, user)
}
