package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
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

// ---- File models ----

type File struct {
	ID            string    `json:"id"`
	DirectoryPath string    `json:"directory_path"`
	Filename      string    `json:"filename"`
	FileType      string    `json:"file_type"`
	Size          int64     `json:"size"`
	Checksum      string    `json:"checksum"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateFileRequest struct {
	DirectoryPath string `json:"directory_path"`
	Filename      string `json:"filename"`
	FileType      string `json:"file_type"`
	Size          int64  `json:"size"`
	Checksum      string `json:"checksum"`
}

func main() {
	db, err := sql.Open("sqlite3", "files.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserRequest
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("POST /files", func(w http.ResponseWriter, r *http.Request) {
		var req CreateFileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		now := time.Now().UTC()
		file := File{
			ID:            uuid.NewString(),
			DirectoryPath: req.DirectoryPath,
			Filename:      req.Filename,
			FileType:      req.FileType,
			Size:          req.Size,
			Checksum:      req.Checksum,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		_, err := db.Exec(`
			INSERT INTO files (
				id, directory_path, filename, file_type,
				size, checksum, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			file.ID,
			file.DirectoryPath,
			file.Filename,
			file.FileType,
			file.Size,
			file.Checksum,
			file.CreatedAt,
			file.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "Failed to insert file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(file)
	})

	mux.HandleFunc("GET /files/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var file File
		err := db.QueryRow(`
			SELECT id, directory_path, filename, file_type,
			       size, checksum, created_at, updated_at
			FROM files WHERE id = ?`, id).
			Scan(
				&file.ID,
				&file.DirectoryPath,
				&file.Filename,
				&file.FileType,
				&file.Size,
				&file.Checksum,
				&file.CreatedAt,
				&file.UpdatedAt,
			)

		if err == sql.ErrNoRows {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(file)
	})

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func initDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS files (
			id TEXT PRIMARY KEY,
			directory_path TEXT NOT NULL,
			filename TEXT NOT NULL,
			file_type TEXT NOT NULL,
			size INTEGER NOT NULL,
			checksum TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	file := File{
		ID:            "b0320eab-57a6-4c45-ba6d-0b68a3501ef6",
		DirectoryPath: "cmd/server/",
		Filename:      "main.go",
		FileType:      "file",
		Size:          123,
		Checksum:      "1afb2837cb93eb1f3d68027adf777218",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	_, err = db.Exec(`
		INSERT OR REPLACE INTO files (
			id, directory_path, filename, file_type,
			size, checksum, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		file.ID,
		file.DirectoryPath,
		file.Filename,
		file.FileType,
		file.Size,
		file.Checksum,
		file.CreatedAt,
		file.UpdatedAt,
	)
	if err != nil {
		fmt.Printf("Error inserting a file: %s\n", err)
		return err
	}

	return nil
}
