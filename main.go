package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Post struct matches your MySQL posts table
type Post struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Likes     int    `json:"likes"`
	Dislikes  int    `json:"dislikes"`
	CreatedAt string `json:"created_at"`
}

var db *sql.DB

// Initialize MySQL connection
func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:SJEC!23/priya@tcp(127.0.0.1:3306)/goblog")
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MySQL")
}

// Get all posts
func getPosts(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, content, likes, dislikes, created_at FROM posts ORDER BY id DESC")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Title, &p.Content, &p.Likes, &p.Dislikes, &p.CreatedAt)
		posts = append(posts, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// Create a new post
func createPost(w http.ResponseWriter, r *http.Request) {
	var p Post
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	_, err = db.Exec("INSERT INTO posts(title, content, likes, dislikes) VALUES(?,?,0,0)", p.Title, p.Content)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("Post Created"))
}

// Like / Dislike / Delete a post
func postActions(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	idStr = strings.TrimSuffix(idStr, "/like")
	idStr = strings.TrimSuffix(idStr, "/dislike")
	id, _ := strconv.Atoi(idStr)

	if strings.HasSuffix(r.URL.Path, "/like") && r.Method == "POST" {
		_, err := db.Exec("UPDATE posts SET likes = likes + 1 WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte("Liked"))
		return
	}

	if strings.HasSuffix(r.URL.Path, "/dislike") && r.Method == "POST" {
		_, err := db.Exec("UPDATE posts SET dislikes = dislikes + 1 WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte("Disliked"))
		return
	}

	if r.Method == "DELETE" {
		_, err := db.Exec("DELETE FROM posts WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte("Deleted"))
		return
	}

	http.Error(w, "Invalid request", 400)
}

func main() {
	// Initialize database
	initDB()

	// API routes
	http.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getPosts(w, r)
		} else if r.Method == "POST" {
			createPost(w, r)
		} else {
			http.Error(w, "Method not allowed", 405)
		}
	})

	http.HandleFunc("/api/posts/", postActions) // handles delete/like/dislike

	// Serve frontend
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}