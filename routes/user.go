// TODO:
// 1. double check create user and login user methods. Not entirely sure the methods are proper
// 2. pass the session key through the http header, not as a variable in the param

package routes

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gebi/scryptauth"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/terencechow/rent/constants"
	"github.com/terencechow/rent/middleware"
)

// HmacSecretKey is the secret key that hmac is generated on
var HmacSecretKey = os.Getenv("HMAC_SECRET_KEY")

//DeleteUser route to delete user
func DeleteUser(c *gin.Context) {
	user := middleware.RequestUserFromContext(c)

	if user.ID.String() != c.Param("user_id") && user.ID.String() != "" || user.ID.String() == "" {
		fmt.Println("Can't delete user from a different user context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't delete user from a different user context"})
		return
	}

	// connect to the cluster
	cluster := gocql.NewCluster(constants.IPAddress)
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	//TODO: Remove the user's posts when they delete themselves
	// var postIds []gocql.UUID
	// var postID gocql.UUID
	// var state string
	// iter := session.Query(`SELECT post_id, state FROM posts_by_user WHERE user_id = ?`, user.ID).Iter()
	// for iter.Scan(&postID, &state) {
	// 	postIds = append(postID)
	// }
	// if err := iter.Close(); err != nil {
	// 	fmt.Println(err)
	// 	c.JSON(500, gin.H{"error": err.Error()})
	// 	return
	// }

	batch := gocql.NewBatch(gocql.LoggedBatch)
	usersStatement := `DELETE FROM users WHERE id = ?`
	usersEmailStatement := `DELETE FROM users_by_email WHERE email = ?`
	batch.Query(usersStatement, user.ID)
	batch.Query(usersEmailStatement, user.Email)
	if err := session.ExecuteBatch(batch); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//delete user posts

	// name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie(constants.RentSessionCookie, "", -1, "", "", true, true)
	c.Redirect(http.StatusSeeOther, "/")
}

// LogoutUser route
// requires a post with users email
func LogoutUser(c *gin.Context) {
	// check the user logging out matches user's session key
	user := middleware.RequestUserFromContext(c)

	if user.Email != c.PostForm("email") && user.Email != "" || user.Email == "" {
		fmt.Println("Can't logout user from a different user context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't logout user from a different user context"})
		return
	}

	// connect to the cluster
	cluster := gocql.NewCluster(constants.IPAddress)
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	batch := gocql.NewBatch(gocql.LoggedBatch)
	usersStatement := `DELETE session_key FROM users WHERE id = ?`
	usersEmailStatement := `DELETE session_key FROM users_by_email WHERE email = ?`
	batch.Query(usersStatement, user.ID)
	batch.Query(usersEmailStatement, user.Email)
	if err := session.ExecuteBatch(batch); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie(constants.RentSessionCookie, "", -1, "", "", true, true)

	c.Redirect(http.StatusSeeOther, "/")
}

// LoginUser route
// requires a postform with the email and password
func LoginUser(c *gin.Context) {

	var email = c.PostForm("email")
	user := middleware.RequestUserFromContext(c)

	// if we alreay have a user from context, it means we have a valid session_key,
	if user.Email != "" {
		fmt.Println("User is logged in by sessionKey", user.Email)

		//redirect to index //TODO: redirect to the path the user was going to
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	//no user in session, need to login user

	// connect to the cluster
	cluster := gocql.NewCluster(constants.IPAddress)
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	//get database hash and user name, by the email
	var databaseHash string
	var previousSessionKey string
	var userID string

	if err := session.Query(`SELECT id, hash, session_key FROM users_by_email WHERE email = ? LIMIT 1`,
		email).Scan(&userID, &databaseHash, &previousSessionKey); err != nil {
		fmt.Println(err)
		log.Fatal(err)
		//TODO: show no user exists by that email if that is the error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// initialize scryptauth on our secret hmac key
	hmacKey := hmac.New(sha256.New, []byte(HmacSecretKey))
	pwHash, err := scryptauth.New(12, hmacKey.Sum(nil))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//decode our db hash to a hash and salt
	pwCost, hash, salt, err := scryptauth.DecodeBase64(databaseHash)
	if err != nil {
		fmt.Print(err)
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// check the hash and salt against the password
	var pass = c.PostForm("pass")
	ok, err := pwHash.Check(pwCost, hash, []byte(pass), salt)
	if !ok {
		fmt.Printf("Error wrong password for user (%s)", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error wrong password for user"})
		return
	}
	fmt.Println("Login OK")

	// create a session key and update the session key for the user
	sessionKey, err := middleware.GenerateSessionID()
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	batch := gocql.NewBatch(gocql.LoggedBatch)

	usersStatement := `INSERT INTO users (session_key, id) VALUES (?, ?)`
	usersEmailStatement := `INSERT INTO users_by_email (session_key, email) VALUES (?, ?)`
	batch.Query(usersStatement, sessionKey, userID)
	batch.Query(usersEmailStatement, sessionKey, email)
	if err := session.ExecuteBatch(batch); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie(constants.RentSessionCookie, sessionKey, 0, "", "", true, true)

	//redirect to index //TODO: redirect to the path the user was going to
	c.Redirect(http.StatusSeeOther, "/")
}

// EditUser route
// requires a post form with the same email as the logged in context
// requires a name to update the profile name to
func EditUser(c *gin.Context) {

	user := middleware.RequestUserFromContext(c)
	fmt.Println("edit user", user)

	// if we alreay have a user from context, it means we have a valid session_key,
	if user.Email != c.PostForm("email") && user.Email != "" || user.Email == "" {
		fmt.Println("Can't edit user from logged out context or differnt user context", user.Email)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't edit user from logged out context or differnt user context"})
		return
	}
	name := c.PostForm("name")
	if name == "" {
		fmt.Println("Can't edit user name to be empty", user.Email)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't edit user name to be empty"})
		return
	}

	// connect to the cluster
	cluster := gocql.NewCluster(constants.IPAddress)
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	usersEmailStatement := `INSERT INTO users_by_email (name, email) VALUES (?, ?)`
	usersStatement := `INSERT INTO users (name, id) VALUES (?, ?)`
	if err := session.Query(usersEmailStatement, name, user.Email).Exec(); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	if err := session.Query(usersStatement, name, user.ID).Exec(); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	//redirect to index //TODO: redirect to the path the user was going to
	c.Redirect(http.StatusSeeOther, "/")
}

// CreateUser route
// requires a postform with a name, password and email
func CreateUser(c *gin.Context) {

	// check the user logging out matches user's session key
	user := middleware.RequestUserFromContext(c)

	fmt.Println(user.Email)
	if user.Email != "" {
		fmt.Println("Can't Create User from a logged in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't Create User from a logged in context"})
		return
	}

	// initialize a scryptauth object
	hmacKey := hmac.New(sha256.New, []byte(HmacSecretKey))
	pwHash, err := scryptauth.New(12, hmacKey.Sum(nil))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.PostForm("pass") == "" {
		fmt.Println("Password is empty!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password is Empty!"})
		return
	}
	// generate a hash and salt from the scrypauth
	hash, salt, err := pwHash.Gen([]byte(c.PostForm("pass")))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//base 64 encoding for easy submitting to db
	hashBase64 := scryptauth.EncodeBase64(pwHash.PwCost, hash, salt)

	// connect to the cluster
	cluster := gocql.NewCluster(constants.IPAddress)
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	sessionKey, err := middleware.GenerateSessionID()
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	uuid, err := gocql.RandomUUID()
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	usersEmailStatement := `INSERT INTO users_by_email (id, name, email, hash, session_key)
	VALUES (?, ?, ?, ?, ?) IF NOT EXISTS`
	name := c.PostForm("name")
	email := c.PostForm("email")

	var uuidCAS, nameCAS, emailCAS, hashCAS, sessionKeyCAS string

	if applied, err :=
		session.Query(usersEmailStatement, uuid, name, email, hashBase64, sessionKey).
			ScanCAS(&nameCAS, &emailCAS, &hashCAS, &uuidCAS, &sessionKeyCAS); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if !applied {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user because email already exists"})
		return
	}

	usersStatement := `INSERT INTO users (id, name, email, hash, session_key) VALUES
	(?, ?, ?, ?, ?)`

	if err := session.Query(usersStatement, uuid, name, email, hashBase64, sessionKey).Exec(); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	// name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie(constants.RentSessionCookie, sessionKey, 0, "", "", true, true)

	c.Redirect(http.StatusSeeOther, "/")
}
