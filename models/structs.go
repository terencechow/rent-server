package models

import (
	"time"

	"github.com/gocql/gocql"
)

// Post json structure
type Post struct {
	UserID         gocql.UUID `json:"userID"`
	PostID         gocql.UUID `json:"postID"`
	Category       string     `json:"category"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Price          int        `json:"price"`
	Deposit        int        `json:"deposit"`
	Available      bool       `json:"available"`
	LastUpdateTime time.Time  `json:"lastUpdateTime"`
	ImageUrls      []string   `json:"imageUrls"`
	City           string     `json:"city"`
	State          string     `json:"state"`
	Latitude       float64    `json:"latitude"`
	Longitude      float64    `json:"longitude"`
}

// User json structure
type User struct {
	ID         gocql.UUID `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	SessionKey string     `json:"session_key"`
}