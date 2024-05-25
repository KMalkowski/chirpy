package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func healthzHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte("OK"))
}

type apiConfig struct {
	fileserverHits int
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
	Valid       bool   `json:"valid"`
	CleanedBody string `json:"cleaned_body"`
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

func chirpsGetHandler(w http.ResponseWriter, req *http.Request) {

}

func chirpsPostHandler(w http.ResponseWriter, req *http.Request) {
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValidResponse{Valid: true, CleanedBody: cleanBadWords(body.Body)})
}

func main() {
	mux := http.NewServeMux()
	api := &apiConfig{}
	mux.Handle("/app/*", api.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", api.metricsHandler)
	mux.HandleFunc("/api/reset", api.resetHandler)
	mux.HandleFunc("POST /api/chirps", chirpsPostHandler)
	mux.HandleFunc("GET /api/chirps", chirpsGetHandler)

	server := http.Server{
		Handler: mux,
		Addr:    "localhost:8080",
	}

	log.Printf("Serving files from %s on port: %s\n", ".", "8080")
	log.Fatal(server.ListenAndServe())

	server.ListenAndServe()
}
