package main

import "github.com/gocql/gocql"

// Post json structure
type Post struct {
	UserID      gocql.UUID `json:"userID"`
	PostID      gocql.UUID `json:"postID"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Price       int        `json:"price"`
	Available   bool       `json:"available"`
	Category    string     `json:"category"`
	Images      []string   `json:"images"`
	Latlng      string     `json:"latlng"`
}

// User json structure
type User struct {
	ID         gocql.UUID `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Latlng     string     `json:"latlng"`
	SessionKey string     `json:"session_key"`
}
