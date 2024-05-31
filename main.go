package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/KMalkowski/chirpy/database"
	"github.com/joho/godotenv"
)

func healthzHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte("OK"))
}

type apiConfig struct {
	fileserverHits int
	db             *database.DB
	jwt_secret     string
}

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		a.fileserverHits++
		next.ServeHTTP(w, req)
	})
}

func (a *apiConfig) metricsHandler(w http.ResponseWriter, req *http.Request) {
	html, err := template.ParseFiles("admin/metrics.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/html")

	html.Execute(w, a.fileserverHits)
}

func (a *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
	a.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
}

type JsonBody struct {
	Body string `json:"body"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidResponse struct {
	database.Chirp
}

func cleanBadWords(body string) string {
	badWords := [3]string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(body, " ")
	for i, word := range words {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				words[i] = "****"
			}
		}
	}

	return strings.Join(words, " ")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5xx error %s", msg)
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	respondWithJson(w, code, errorResponse{Error: msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(dat)
}

func main() {
	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")

	mux := http.NewServeMux()
	db, err := database.NewDB("database.json")
	if err != nil {
		panic(err)
	}

	api := &apiConfig{db: db, jwt_secret: jwtSecret}

	mux.Handle("/app/*", api.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", api.metricsHandler)
	mux.HandleFunc("/api/reset", api.resetHandler)
	mux.HandleFunc("POST /api/chirps", api.AuthenticationMiddleware(api.chirpsPostHandler))
	mux.HandleFunc("GET /api/chirps", api.chirpsGetHandler)
	mux.HandleFunc("GET /api/chirps/{id}", api.ChirpsGetByIdHandler)
	mux.HandleFunc("POST /api/users", api.CreateUserHandler)
	mux.HandleFunc("POST /api/login", api.LoginHandler)
	mux.HandleFunc("PUT /api/users", api.UpdateUserHandler)
	mux.HandleFunc("POST /api/refresh", api.RefreshTokenHandler)
	mux.HandleFunc("POST /api/revoke", api.AuthenticationMiddleware(api.RevokeTokenHandler))
	mux.HandleFunc("DELETE /api/chirps/{id}", api.AuthenticationMiddleware(api.DeleteChirpHandler))
	mux.HandleFunc("POST /api/polka/webhooks", api.HandlePolkaWebhooks)

	server := http.Server{
		Handler: mux,
		Addr:    "localhost:8080",
	}

	log.Printf("Serving files from %s on port: %s\n", ".", "8080")
	log.Fatal(server.ListenAndServe())

	server.ListenAndServe()
}
