package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
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

	log.Fatal(app.Listen(":8080"))
}
