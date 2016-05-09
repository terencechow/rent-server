
package main

import "github.com/gocql/gocql"

type Post struct {
  User_id gocql.UUID `json:"user_id"`
	Post_id gocql.UUID `json:"post_id"`
	Title string `json:"title"`
	Description string `json:"description"`
	Price int `json:"price"`
  Available bool `json:"available"`
	Category string `json:"category"`
	Images []string `json:"images"`
	Latlng string `json:"latlng"`
}
