package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (a *apiConfig) AuthenticationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		bearer := req.Header.Get("Authorization")
		token := strings.Replace(bearer, "Bearer ", "", 1)

		if len(token) == 0 {
			respondWithError(w, 401, "you are not authenticated")
			return
		}

		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			respondWithError(w, 401, "you are not authenticated")
			log.Printf("error parsing with claims", err.Error(), token)
			return
		}

		subject, err := claims.GetSubject()
		if err != nil {
			respondWithError(w, 401, "you are not authenticated")
			log.Printf("error no subject")
			return
		}

		val, _ := strconv.Atoi(subject)
		ctx := context.WithValue(req.Context(), "userId", val)

		next.ServeHTTP(w, req.WithContext(ctx))
	})
}
