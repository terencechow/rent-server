package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/terencechow/rent/constants"
	"github.com/terencechow/rent/middleware"
	"github.com/terencechow/rent/models"
)

// DeletePost route
// requires url to include post_id, category, string and user_id
func DeletePost(c *gin.Context) {
	user := middleware.RequestUserFromContext(c)
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
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	batch := gocql.NewBatch(gocql.LoggedBatch)
	postsStatement := `DELETE FROM posts WHERE state = ? and post_id = ?`
	postsCategoryStatement := `DELETE FROM posts_by_category
		WHERE state = ? and category = ? and post_id = ?`
	postsUserStatement := `DELETE FROM posts_by_user WHERE user_id = ? and post_id = ?`
	batch.Query(postsStatement, state, postID)
	batch.Query(postsCategoryStatement, state, category, postID)
	batch.Query(postsUserStatement, user.ID, postID)
	if err := session.ExecuteBatch(batch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "successfully deleted post"})
}

// ShowPost route
// requires state and post_id in the url
func ShowPost(c *gin.Context) {

	var postIDFromURL = c.Param("post_id")
	var stateFromURL = c.Param("state")
	var userID, postID gocql.UUID
	var category, title, description, city, state string
	var price, deposit int
	var available bool
	var lastUpdateTime time.Time
	var imageUrls []string
	var latitude, longitude float64

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	if err := session.Query(`SELECT
		user_id, post_id, category, title, description, price, deposit, available,
 	 last_update_time, image_urls, city, state, latitude, longitude
   FROM posts WHERE state = ? and post_id = ? LIMIT 1`,
		stateFromURL, postIDFromURL).Scan(&userID, &postID, &category, &title, &description,
		&price, &deposit, &available, &lastUpdateTime, &imageUrls, &city, &state, &latitude, &longitude); err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	post := models.Post{
		UserID:         userID,
		PostID:         postID,
		Category:       category,
		Title:          title,
		Description:    description,
		Price:          price,
		Deposit:        deposit,
		Available:      available,
		LastUpdateTime: lastUpdateTime,
		ImageUrls:      imageUrls,
		City:           city,
		State:          state,
		Latitude:       latitude,
		Longitude:      longitude}

	c.JSON(http.StatusOK, post)
}

// EditOrCreatePost route
// requires a form with the following:
// post_id, category, title, description,
// price, deposit, available
// city, state, latitude, longitude
// user_id
func EditOrCreatePost(c *gin.Context) {

	user := middleware.RequestUserFromContext(c)

	if user.ID.String() != c.PostForm("user_id") && user.ID.String() != "" {
		fmt.Println("Can't edit or create a post from a different user context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't edit or create a post from a different user context"})
		return
	}
	fmt.Println("USER IS", user)

	var postID = c.PostForm("post_id")
	var category = c.PostForm("category")
	var title = c.PostForm("title")
	var description = c.PostForm("description")
	price, _ := strconv.Atoi(c.PostForm("price"))
	deposit, _ := strconv.Atoi(c.PostForm("deposit"))
	available, _ := strconv.ParseBool(c.PostForm("available"))
	lastUpdateTime := time.Now()
	var city = c.PostForm("city")
	var state = c.PostForm("state")
	latitude, _ := strconv.ParseFloat(c.PostForm("latitude"), 64)
	longitude, _ := strconv.ParseFloat(c.PostForm("longitude"), 64)

	var imageUrls []string

	if postID == "" {
		postID = gocql.TimeUUID().String()
	}

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	batch := gocql.NewBatch(gocql.LoggedBatch)
	postsStatement := `INSERT INTO posts
		(user_id, post_id, category, title, description, price, deposit, available,
			last_update_time, image_urls, city, state, latitude, longitude)
		VALUES (?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	postsCategoryStatement := `INSERT INTO posts_by_category
	(user_id, post_id, category, title, description, price, deposit, available,
		last_update_time, image_urls, city, state, latitude, longitude)
		VALUES (?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	postsUserStatement := `INSERT INTO posts_by_user
	(user_id, post_id, category, title, description, price, deposit, available,
		last_update_time, image_urls, city, state, latitude, longitude)
		VALUES (?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	batch.Query(postsStatement, user.ID, postID, category, title, description, price,
		deposit, available, lastUpdateTime, imageUrls, city, state,
		latitude, longitude)
	batch.Query(postsCategoryStatement, user.ID, postID, category, title, description, price,
		deposit, available, lastUpdateTime, imageUrls, city, state,
		latitude, longitude)
	batch.Query(postsUserStatement, user.ID, postID, category, title, description, price,
		deposit, available, lastUpdateTime, imageUrls, city, state,
		latitude, longitude)
	if err := session.ExecuteBatch(batch); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusSeeOther, "/?latitude="+c.PostForm("latitude")+"&longitude="+c.PostForm("longitude"))
}

//PostIndex route to show posts or a category of posts
func PostIndex(c *gin.Context) {

	var userID, postID gocql.UUID
	var category, title, description, city, state string
	var price, deposit int
	var available bool
	var lastUpdateTime time.Time
	var imageUrls []string
	var latitude, longitude float64

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	searchCategory := c.Param("category")
	userLatitude := c.Query("latitude")
	userLongitude := c.Query("longitude")

	if userLatitude == "" || userLongitude == "" {
		fmt.Println("Missing latitude or longitude!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Missing latitude or longitude"})
		return
	}

	var query *gocql.Query

	if searchCategory == "" {

		var luceneQuery = `{
			filter: { type: "match", field: "available", value: "true"},
			query: { type: "geo_distance", field: "place", latitude:` + userLatitude + `, longitude: ` + userLongitude +
			`, max_distance: "20km"}}`

		query = session.Query(`SELECT
			user_id, post_id, category, title, description, price, deposit,
			available, last_update_time, image_urls, city, state, latitude, longitude
  		FROM posts WHERE expr(posts_index, ?) LIMIT 100`, luceneQuery)
		fmt.Println("Search category is nil")
	} else {
		var luceneQuery = `{
			filter: {
				type: "boolean", must:[
					{type: "match", field: "available", value: "true"},
					{type: "match", field: "category", value:"` + searchCategory + `"}` +
			`]
			},
			query: {type: "geo_distance", field: "place", latitude:` + userLatitude +
			`, longitude: ` + userLongitude + `, max_distance: "20km" }}`
		query = session.Query(`SELECT
			user_id, post_id, category, title, description, price, deposit,
			available, last_update_time, image_urls, city, state, latitude, longitude
  		FROM posts_by_category WHERE expr(posts_category_index, ?) LIMIT 100`, luceneQuery)
	}

	iter := query.Iter()

	var posts []models.Post

	for iter.Scan(&userID, &postID, &category, &title, &description, &price, &deposit,
		&available, &lastUpdateTime, &imageUrls, &city, &state,
		&latitude, &longitude) {
		post := models.Post{
			UserID:         userID,
			PostID:         postID,
			Category:       category,
			Title:          title,
			Description:    description,
			Price:          price,
			Deposit:        deposit,
			Available:      available,
			LastUpdateTime: lastUpdateTime,
			ImageUrls:      imageUrls,
			City:           city,
			State:          state,
			Latitude:       latitude,
			Longitude:      longitude}
		posts = append(posts, post)
	}
	fmt.Println("Done iterating", posts)
	// postsJSON, _ := json.Marshal(posts)

	if err := iter.Close(); err != nil {
		fmt.Println("Error in iter.close", err.Error())
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}
