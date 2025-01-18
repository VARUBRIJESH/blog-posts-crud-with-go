package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"

	fiberSwagger "github.com/swaggo/fiber-swagger" // Swagger handler

	_ "blog-crud-go/docs" // Import the generated Swagger docs
)

type BlogPost struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BlogPostRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Body        string `json:"body"`
}

type DeleteResponse struct {
	Message string `json:"message"`
}

var db *sql.DB

func initDB() {
	var err error
	connStr := "user=postgres password=12345 dbname=go-crud-test sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Unable to verify connection to database:", err)
	}

	log.Println("Connected to the database!")
}

// @Summary Get all blog posts
// @Description Retrieve a list of all blog posts
// @Accept  json
// @Produce  json
// @Success 200 {object} BlogPost "List of blog posts" // Use BlogPost for the response
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/blog-post [get]
func getBlogPosts(c *fiber.Ctx) error {
	rows, err := db.Query("SELECT id, title, description, body, created_at, updated_at FROM blog_posts")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch posts"})
	}
	defer rows.Close()

	var posts []BlogPost
	for rows.Next() {
		var post BlogPost
		if err := rows.Scan(&post.ID, &post.Title, &post.Description, &post.Body, &post.CreatedAt, &post.UpdatedAt); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse post"})
		}
		posts = append(posts, post)
	}

	return c.JSON(posts)
}

// @Summary Get a single blog post
// @Description Retrieve a blog post by its ID
// @Accept  json
// @Produce  json
// @Param id path int true "Blog Post ID"
// @Success 200 {object} BlogPost "List of blog posts" // Use BlogPost for the response
// @Failure 404 {object} map[string]string "Blog post not found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/blog-post/{id} [get]
func getBlogPost(c *fiber.Ctx) error {
	id := c.Params("id")
	var post BlogPost
	row := db.QueryRow("SELECT id, title, description, body, created_at, updated_at FROM blog_posts WHERE id = $1", id)
	if err := row.Scan(&post.ID, &post.Title, &post.Description, &post.Body, &post.CreatedAt, &post.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch post"})
	}

	return c.JSON(post)
}

// @Summary Create a new blog post
// @Description Create a new blog post with title, description, and body
// @Accept  json
// @Produce  json
// @Param blogPost body BlogPostRequest true "Create Blog Post" // Use BlogPostRequest for the request body
// @Success 201 {object} BlogPost "Successfully created blog post" // Use BlogPost for the response
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/blog-post [post]
func createBlogPost(c *fiber.Ctx) error {
	post := new(BlogPost)
	if err := c.BodyParser(post); err != nil {
		log.Printf("Error inserting post: %v", err) // Log the actual error
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	query := "INSERT INTO blog_posts (title, description, body, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err := db.QueryRow(query, post.Title, post.Description, post.Body, time.Now(), time.Now()).Scan(&post.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create post"})
	}

	return c.Status(http.StatusCreated).JSON(post)
}

// @Summary Update an existing blog post
// @Description Update the title, description, and/or body of an existing blog post by its ID
// @Accept  json
// @Produce  json
// @Param id path int true "Blog Post ID" // The ID of the blog post to update
// @Param blogPost body BlogPostRequest true "Updated blog post data" // Request body containing updated data
// @Success 200 {object} BlogPost "Successfully updated blog post" // Return the updated blog post details
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 404 {object} map[string]string "Blog post not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/blog-post/{id} [patch]
func updateBlogPost(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "ID is required"})
	}

	post := new(BlogPost)
	if err := c.BodyParser(post); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	var existingPost BlogPost
	query := "SELECT id, title, description, body, created_at, updated_at FROM blog_posts WHERE id = $1"
	row := db.QueryRow(query, id)
	err := row.Scan(&existingPost.ID, &existingPost.Title, &existingPost.Description, &existingPost.Body, &existingPost.CreatedAt, &existingPost.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve post"})
	}

	if post.Title != "" {
		existingPost.Title = post.Title
	}
	if post.Description != "" {
		existingPost.Description = post.Description
	}
	if post.Body != "" {
		existingPost.Body = post.Body
	}

	query = "UPDATE blog_posts SET title = $1, description = $2, body = $3, updated_at = $4 WHERE id = $5"
	_, err = db.Exec(query, existingPost.Title, existingPost.Description, existingPost.Body, time.Now(), id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update post"})
	}

	existingPost.UpdatedAt = time.Now()
	return c.JSON(existingPost)
}

// @Summary Delete a blog post
// @Description Delete a blog post by ID
// @Accept  json
// @Produce  json
// @Param id path int true "Blog Post ID"
// @Success 200 {object} DeleteResponse
// @Failure 400 {object} DeleteResponse
// @Failure 404 {object} DeleteResponse
// @Router /api/blog-post/{id} [delete]
func deleteBlogPost(c *fiber.Ctx) error {
	id := c.Params("id")
	query := "DELETE FROM blog_posts WHERE id = $1"
	res, err := db.Exec(query, id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete post"})
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	return c.JSON(fiber.Map{"message": "Post deleted"})
}

// @title Blog Post APIs
// @version 1.0
// @description This is a simple CRUD API for blog posts.

func main() {
	initDB()

	app := fiber.New()

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	app.Get("/api/blog-post", getBlogPosts)
	app.Get("/api/blog-post/:id", getBlogPost)
	app.Post("/api/blog-post", createBlogPost)
	app.Patch("/api/blog-post/:id", updateBlogPost)
	app.Delete("/api/blog-post/:id", deleteBlogPost)

	log.Fatal(app.Listen(":5000"))
}
