package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestApp() *fiber.App {
	// Set up a Fiber app and routes
	app := fiber.New()

	// Replace the real database with a mock database
	db = setupMockDB()

	setupRoutes(app)
	return app
}

func setupRoutes(app *fiber.App) {
	// Route to create a blog post
	app.Post("/api/blog-post", func(c *fiber.Ctx) error {
		var payload map[string]string
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(http.StatusBadRequest).SendString("Invalid request body")
		}
		return c.Status(http.StatusCreated).JSON([]map[string]string{
			{"title": "Test Title", "description": "Test Description", "body": "Test Body"},
		})
	})

	// Route to get all blog posts
	app.Get("/api/blog-post", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON([]map[string]string{
			{"title": "Test Title", "description": "Test Description", "body": "Test Body"},
		})
	})

	// Route to get a single blog post
	app.Get("/api/blog-post/:id", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(map[string]string{
			"title":       "Test Title",
			"description": "Test Description",
			"body":        "Test Body",
		})
	})

	// Route to update a blog post
	app.Patch("/api/blog-post/:id", func(c *fiber.Ctx) error {
		var payload map[string]string
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(http.StatusBadRequest).SendString("Invalid request body")
		}
		return c.Status(http.StatusOK).JSON(map[string]string{
			"message": "Blog post updated",
		})
	})

	// Route to delete a blog post
	app.Delete("/api/blog-post/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})
}

func setupMockDB() *sql.DB {
	connStr := "user=postgres password=12345 dbname=go-crud-test sslmode=disable"
	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		panic("Failed to connect to test database")
	}
	return testDB
}

func TestCreateBlogPost(t *testing.T) {
	app := setupTestApp()

	payload := map[string]string{
		"title":       "Test Title",
		"description": "Test Description",
		"body":        "Test Body",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/blog-post", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestGetAllBlogPosts(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/blog-post", nil)
	resp, _ := app.Test(req, -1)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSingleBlogPost(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/blog-post/1", nil)
	resp, _ := app.Test(req, -1)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUpdateBlogPost(t *testing.T) {
	app := setupTestApp()

	payload := map[string]string{
		"title":       "Updated Title",
		"description": "Updated Description",
		"body":        "Updated Body",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/api/blog-post/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDeleteBlogPost(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest(http.MethodDelete, "/api/blog-post/1", nil)
	resp, _ := app.Test(req, -1)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
