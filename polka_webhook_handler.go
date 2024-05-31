package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

func (a *apiConfig) HandlePolkaWebhooks(w http.ResponseWriter, req *http.Request) {
	apiKey := req.Header.Get("Authorization")
	if strings.Replace(apiKey, "ApiKey ", "", 1) != os.Getenv("POLKA_API_KEY") {
		respondWithError(w, 401, "you're not authorized to send this request")
		return
	}

	type WebhookBody struct {
		Event string `json:"event"`
		Data  struct {
			UserId int `json:"user_id"`
		} `json:"data"`
	}

	body := WebhookBody{}
	err := json.NewDecoder(req.Body).Decode(&body)

	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(204)
		return
	}

	if body.Event != "user.upgraded" {
		println("bad")
		w.WriteHeader(204)
		return
	}

	println("running")
	err = a.db.UpgradeUser(body.Data.UserId, true)
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, 404, "user not found")
		return
	}

	w.WriteHeader(204)
}
