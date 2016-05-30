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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

// DeletePost Route to delete a post
func DeletePost(c *gin.Context) {
	user := requestUserFromContext(c)
	if user.ID.String() != c.Param("user_id") && user.ID.String() != "" {
		fmt.Println("Can't delete post from a different user context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't delete post from a different user context"})
		return
	}

	var postID = c.Param("post_id")
	var category = c.Param("category")
	var state = c.Param("state")

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	batch := gocql.NewBatch(gocql.LoggedBatch)
	postsStatement := `DELETE FROM posts WHERE state = ? and post_id = ?`
	postsCategoryStatement := `DELETE FROM posts_by_category
		WHERE state = ? and category = ? and post_id = ?`
	postsUserStatement := `DELETE FROM posts_by_user WHERE user_id = ? and post_id = ?`
	batch.Query(postsStatement, state, postID)
	batch.Query(postsCategoryStatement, state, category)
	batch.Query(postsUserStatement, user.ID, postID)
	if err := session.ExecuteBatch(batch); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, "/")
}

// ShowPost Route to show a post
func ShowPost(c *gin.Context) {

	var postIDFromURL = c.Param("post_id")
	var stateFromURL = c.Param("state")
	var userID, postID gocql.UUID
	var category, name, description, city, state string
	var price, deposit, minimumRentalDays int
	var nextAvailableDate string
	var imageUrls []string
	var latitude, longitude float64

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	if err := session.Query(`SELECT
		user_id, post_id, category, name, description, price, deposit, minimum_rental_days,
 	 next_available_date, image_urls, city, state, latitude, longitude
   FROM posts WHERE state = ? and post_id = ? LIMIT 1`,
		stateFromURL, postIDFromURL).Scan(&userID, &postID, &category, &name, &description,
		&price, &deposit, &minimumRentalDays, &nextAvailableDate, &imageUrls, &latitude, &longitude); err != nil {

		c.Redirect(http.StatusFound, "/")
		return
	}

	shortForm, _ := time.Parse(nextAvailableDate, "2013-Feb-03")
	post := Post{
		UserID:            userID,
		PostID:            postID,
		Category:          category,
		Name:              name,
		Description:       description,
		Price:             price,
		Deposit:           deposit,
		MinimumRentalDays: minimumRentalDays,
		NextAvailableDate: shortForm,
		ImageUrls:         imageUrls,
		City:              city,
		State:             state,
		Latitude:          latitude,
		Longitude:         longitude}

	postJSON, _ := json.Marshal(post)

	c.JSON(http.StatusOK, postJSON)
}

// EditOrCreatePost route to create or update a post
func EditOrCreatePost(c *gin.Context) {

	user := requestUserFromContext(c)

	if user.ID.String() != c.PostForm("user_id") && user.ID.String() != "" {
		fmt.Println("Can't edit or create a post from a different user context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't edit or create a post from a different user context"})
		return
	}

	var postID = c.PostForm("post_id")
	var category = c.PostForm("category")
	var name = c.PostForm("name")
	var description = c.PostForm("description")
	price, _ := strconv.Atoi(c.PostForm("price"))
	deposit, _ := strconv.Atoi(c.PostForm("deposit"))
	minimumRentalDays, _ := strconv.Atoi(c.PostForm("minimumRentalDays"))
	nextAvailableDate := c.PostForm("nextAvailableDate")
	var city = c.PostForm("city")
	var state = c.PostForm("state")
	var latitude = c.PostForm("latitude")
	var longitude = c.PostForm("longitude")

	var imageUrls []string

	if postID == "" {
		postID = gocql.TimeUUID().String()
	}

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	batch := gocql.NewBatch(gocql.LoggedBatch)
	postsStatement := `INSERT INTO posts
		(user_id, post_id, category, name, description, price, deposit, minimum_rental_days,
			next_available_date, image_urls, city, state, latitude, longitude)
		VALUES (?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	postsCategoryStatement := `INSERT INTO posts_by_category
		(user_id, post_id, category, name, description, price, deposit, minimum_rental_days,
			next_available_date, image_urls, city, state, latitude, longitude)
		VALUES (?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	postsUserStatement := `INSERT INTO posts_by_user
		(user_id, post_id, category, name, description, price, deposit, minimum_rental_days,
			next_available_date, image_urls, city, state, latitude, longitude)
		VALUES (?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	batch.Query(postsStatement, user.ID, postID, category, name, description, price,
		deposit, minimumRentalDays, nextAvailableDate, imageUrls, city, state,
		latitude, longitude)
	batch.Query(postsCategoryStatement, user.ID, postID, category, name, description, price,
		deposit, minimumRentalDays, nextAvailableDate, imageUrls, city, state,
		latitude, longitude)
	batch.Query(postsUserStatement, user.ID, postID, category, name, description, price,
		deposit, minimumRentalDays, nextAvailableDate, imageUrls, city, state,
		latitude, longitude)
	if err := session.ExecuteBatch(batch); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, "/")
}

//PostIndex route to show posts or a category of posts
func PostIndex(c *gin.Context) {

	var userID, postID gocql.UUID
	var category, name, description, city, state string
	var price, deposit, minimumRentalDays int
	var nextAvailableDate string
	var imageUrls []string
	var latitude, longitude float64

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	searchCategory := c.Param("category")
	userLatitude := c.Query("latitude")
	userLongitude := c.Query("longitude")

	var query *gocql.Query

	if searchCategory == "" {
		query = session.Query(`SELECT
			user_id, post_id, category, name, description, price, deposit, minimum_rental_days,
			next_available_date, image_urls, city, state, latitude, longitude
  		FROM posts WHERE expr(posts_index,'{
				query:{
					type: "geo_distance",
					field: "place",
					latitude: ?,
					longitude: ?,
					max_distance: "20km"
				}
			}') LIMIT 100`, userLatitude, userLongitude)
	} else {
		query = session.Query(`SELECT
			user_id, post_id, category, name, description, price, deposit, minimum_rental_days,
			next_available_date, image_urls, city, state, latitude, longitude
  		FROM posts_by_category WHERE expr(posts_category_index,'{
				filter: {
					type: "match",
					field: "category",
					value: ? },
				query:{
					type: "geo_distance",
					field: "place",
					latitude: ?,
					longitude: ?,
					max_distance: "20km"
				}
			}') LIMIT 100`, searchCategory, userLatitude, userLongitude)
	}

	iter := query.Iter()

	var posts []Post

	for iter.Scan(&userID, &postID, &category, &name, &description, &price, &deposit,
		&minimumRentalDays, &nextAvailableDate, &imageUrls, &city, &state,
		&latitude, &longitude) {

		shortForm, _ := time.Parse(nextAvailableDate, "2013-Feb-03")
		post := Post{
			UserID:            userID,
			PostID:            postID,
			Category:          category,
			Name:              name,
			Description:       description,
			Price:             price,
			Deposit:           deposit,
			MinimumRentalDays: minimumRentalDays,
			NextAvailableDate: shortForm,
			ImageUrls:         imageUrls,
			City:              city,
			State:             state,
			Latitude:          latitude,
			Longitude:         longitude}
		posts = append(posts, post)
	}
	postsJSON, _ := json.Marshal(posts)

	if err := iter.Close(); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, postsJSON)
}
