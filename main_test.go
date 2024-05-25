package main

import "testing"

func TestCleanBadWords(t *testing.T) {
	body := "This is a kerfuffle opinion I need to share with the world"
	cleanedBody := cleanBadWords(body)

	if cleanedBody != "This is a **** opinion I need to share with the world" {
		t.Errorf("Expected 'This is a **** opinion I need to share with the world', got %s", cleanedBody)
	}
}
