package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type CreateUserResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

func main() {
	mux := http.NewServeMux()

	// 1. Health Check
	// Go 1.22+ allows "METHOD /path" syntax
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// 2. Create User
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserRequest

		// Decode JSON body
		// Equivalent to fiber's c.BodyParser
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		response := CreateUserResponse{
			ID:        uuid.NewString(),
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
		}

		// Set Headers and Encode JSON response
		// Equivalent to fiber's c.JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // Optional: explicit 201 Created
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
	})

	log.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
