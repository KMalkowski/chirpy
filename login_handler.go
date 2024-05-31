package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/KMalkowski/chirpy/database"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (a *apiConfig) GenerateAccessToken(expirationTimeInSeconds int, userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(expirationTimeInSeconds)).UTC()),
		Subject:   userId,
	})

	signedToken, err := token.SignedString([]byte(a.jwt_secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (a *apiConfig) LoginHandler(w http.ResponseWriter, req *http.Request) {
	type LoginBody struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	body := LoginBody{}
	err := json.NewDecoder(req.Body).Decode(&body)

	if err != nil {
		respondWithError(w, 400, "bad request, we need email and password")
	}

	db, err := a.db.ReadDatabase()
	if err != nil {
		respondWithError(w, 500, "could not read the db")
	}

	user := database.User{}
	for _, u := range db.Users {
		if u.Email == body.Email {
			user = u
			break
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		respondWithError(w, 401, "wrong password")
	}

	type LoginResponse struct {
		Id           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
	}

	expirationTimeInSeconds := 0
	if body.ExpiresInSeconds > 0 && body.ExpiresInSeconds < 60 {
		expirationTimeInSeconds = body.ExpiresInSeconds
	} else {
		expirationTimeInSeconds = 60
	}

	signedToken, err := a.GenerateAccessToken(expirationTimeInSeconds, fmt.Sprintf("%v", user.Id))

	if err != nil {
		log.Println(err.Error())
		respondWithError(w, 500, "could not sign jwt")
	}

	c := 10
	b := make([]byte, c)
	_, err = rand.Read(b)

	if err != nil {
		respondWithError(w, 500, "could not sign refresh token")
		return
	}

	refreshToken, err := a.db.AddRefreshToken(hex.EncodeToString(b), time.Now().Add(time.Hour*24*60), user.Id)
	if err != nil {
		respondWithError(w, 500, "could not add token to the databse")
		return
	}

	respondWithJson(w, 200, LoginResponse{Id: user.Id, Email: user.Email, Token: signedToken, RefreshToken: refreshToken.Token, IsChirpyRed: user.IsChirpyRed})
}
