package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
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

func main() {
	// SQLite setup
	db, err := sql.Open("sqlite3", "./files.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS files (
			id TEXT PRIMARY KEY,
			directory_path TEXT NOT NULL,
			filename TEXT NOT NULL,
			file_type TEXT NOT NULL,
			size INTEGER NOT NULL,
			checksum TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
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
	}

	app := fiber.New()

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Post("/users", func(c *fiber.Ctx) error {
		var request CreateUserRequest
		if err := c.BodyParser(&request); err != nil {
			return err
		}

		response := CreateUserResponse{
			ID:        uuid.New().String(),
			FirstName: request.FirstName,
			LastName:  request.LastName,
			Email:     request.Email,
		}

		return c.JSON(response)
	})

	app.Post("/files", func(c *fiber.Ctx) error {
		var file File
		if err := c.BodyParser(&file); err != nil {
			return err
		}

		now := time.Now().UTC()
		file.ID = uuid.New().String()
		file.CreatedAt = now
		file.UpdatedAt = now

		_, err := db.Exec(`
			INSERT INTO files (
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
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(file)
	})

	// Get file metadata by ID
	app.Get("/files/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		var file File
		err := db.QueryRow(`
			SELECT id, directory_path, filename, file_type,
			       size, checksum, created_at, updated_at
			FROM files WHERE id = ?
		`, id).Scan(
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
			return c.SendStatus(fiber.StatusNotFound)
		}
		if err != nil {
			return err
		}

		return c.JSON(file)
	})

	log.Fatal(app.Listen(":8080"))
}
