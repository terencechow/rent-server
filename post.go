/*

TODOs:
3) handle the images array in EditOrCreatePost
4) Redirect ShowPost to a 'not found page' when no results are found

*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gocql/gocql"
	"github.com/rs/xmux"
	"golang.org/x/net/context"
)

// DeletePost Route to delete a post
func DeletePost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := requestUserFromContext(ctx)
	if user.ID.String() != xmux.Param(ctx, "user_id") && user.ID.String() != "" {
		fmt.Println("Can't delete post from a different user context")
		http.Error(w, "Can't delete post from a different user context", http.StatusInternalServerError)
		return
	}

	var postID = xmux.Param(ctx, "post_id")
	var category = xmux.Param(ctx, "category")
	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	if err := session.Query(`DELETE FROM posts_by_category WHERE category = ? and post_id = ?`,
		category, postID).Exec(); err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// ShowPost Route to show a post
func ShowPost(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	var postIDFromURL = xmux.Param(ctx, "post_id")
	var category = xmux.Param(ctx, "category")
	var userID, postID gocql.UUID
	var title, description, latlng string
	var price int
	var available bool
	var images []string

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	if err := session.Query(`SELECT
   user_id, post_id, title, description, price, available, category, images, latlng
   FROM posts_by_category WHERE category = ? and post_id = ? LIMIT 1`,
		category, postIDFromURL).Scan(&userID, &postID, &title, &description, &price, &available, &category, &images, &latlng); err != nil {
		// log.Fatal(err)
		// http.Error(w, err.Error(), http.StatusInternalServerError)

		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	post := Post{
		UserID:      userID,
		PostID:      postID,
		Title:       title,
		Description: description,
		Price:       price,
		Available:   available,
		Category:    category,
		Images:      images,
		Latlng:      latlng}

	postJSON, _ := json.Marshal(post)

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(postJSON)
}

// EditOrCreatePost route to create or update a post
func EditOrCreatePost(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	user := requestUserFromContext(ctx)

	if user.ID.String() != r.FormValue("user_id") && user.ID.String() != "" {
		fmt.Println("Can't edit or create a post from a different user context")
		http.Error(w, "Can't edit or create a post from a different user context", http.StatusInternalServerError)
		return
	}
	var title = r.FormValue("title")
	var description = r.FormValue("description")
	var category = r.FormValue("category")
	var latlng = r.FormValue("latlng")
	available, _ := strconv.ParseBool(r.FormValue("available"))
	price, _ := strconv.Atoi(r.FormValue("price"))
	var images []string

	var postID = r.FormValue("post_id")
	if postID == "" {
		postID = gocql.TimeUUID().String()
	}

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	// insert a post TTL 5184000 = 2 months in seconds
	if err := session.Query(`INSERT INTO posts_by_category
		(user_id, post_id, title, description, price, available, category, images, latlng)
		VALUES (?,?, ?, ?, ?, ?, ?, ?, ?) USING TTL 5184000`,
		user.ID, postID, title, description, price, available, category, images, latlng).Exec(); err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

//PostIndex route to show posts or a category of posts
func PostIndex(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var userID, postID gocql.UUID
	var title, description, category, latlng string
	var price int
	var available bool
	var images []string
	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	var searchCategory = xmux.Param(ctx, "category")

	var query *gocql.Query

	if searchCategory == "" {
		query = session.Query(`SELECT
  		user_id, post_id, title, description, price, available, category, images, latlng
  		FROM posts_by_category LIMIT 10`)
	} else {
		query = session.Query(`SELECT
      user_id, post_id, title, description, price, available, category, images, latlng
      FROM posts_by_category WHERE category = ? LIMIT 10`, searchCategory)
	}

	iter := query.Iter()

	var posts []Post

	for iter.Scan(&userID, &postID, &title, &description, &price, &available, &category, &images, &latlng) {
		post := Post{
			UserID:      userID,
			PostID:      postID,
			Title:       title,
			Description: description,
			Price:       price,
			Available:   available,
			Category:    category,
			Images:      images,
			Latlng:      latlng}
		posts = append(posts, post)
	}
	postsJSON, _ := json.Marshal(posts)

	if err := iter.Close(); err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(postsJSON)
}
