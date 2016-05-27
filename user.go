// TODO:
// 1. double check create user and login user methods. Not entirely sure the methods are proper
// 2. pass the session key through the http header, not as a variable in the param

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"log"
	"os"

	"github.com/gebi/scryptauth"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

// HmacSecretKey is the secret key that hmac is generated on
var HmacSecretKey = os.Getenv("HMAC_SECRET_KEY")

//DeleteUser route to delete user
func DeleteUser(c *gin.Context) {
	user := requestUserFromContext(c)

	if user.ID.String() != c.Param("user_id") {
		fmt.Println("Can't delete user from a different user context")
		c.JSON(500, gin.H{"error": "Can't delete user from a different user context"})
		return
	}

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
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
	usersSessionStatement := `DELETE FROM users_by_session_key WHERE session_key = ?`
	batch.Query(usersStatement, user.ID)
	batch.Query(usersEmailStatement, user.Email)
	batch.Query(usersSessionStatement, user.SessionKey)
	if err := session.ExecuteBatch(batch); err != nil {
		log.Fatal(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	//delete user posts

	// name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie(RentSessionCookie, "", -1, "", "localhost", true, true)
	c.Redirect(200, "/")
}

//LogoutUser route to log out user
func LogoutUser(c *gin.Context) {
	fmt.Println("LogoutUser route")
	// check the user logging out matches user's session key
	user := requestUserFromContext(c)

	if user.Email != c.PostForm("email") && user.Email != "" {
		fmt.Println("Can't logout user from a different user context")
		c.JSON(500, gin.H{"error": "Can't logout user from a different user context"})
		return
	}

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	batch := gocql.NewBatch(gocql.LoggedBatch)
	usersStatement := `DELETE session_key FROM users WHERE id = ?`
	usersEmailStatement := `DELETE session_key FROM users_by_email WHERE email = ?`
	usersSessionStatement := `DELETE session_key FROM users_by_session_key WHERE session_key = ?`
	batch.Query(usersStatement, user.ID)
	batch.Query(usersEmailStatement, user.Email)
	batch.Query(usersSessionStatement, user.SessionKey)
	if err := session.ExecuteBatch(batch); err != nil {
		log.Fatal(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie(RentSessionCookie, "", -1, "", "localhost", true, true)

	c.Redirect(200, "/")
}

//LoginUser route to login a user
func LoginUser(c *gin.Context) {

	var email = c.PostForm("email")
	user := requestUserFromContext(c)

	// if we alreay have a user from context, it means we have a valid session_key,
	if user.Email != "" {
		fmt.Println("User is logged in by sessionKey", user.Email)

		//redirect to index //TODO: redirect to the path the user was going to
		c.Redirect(200, "/")
		return
	}

	//no user in session, need to login user

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	//get database hash and user name, by the email
	var databaseHash string
	var previousSessionKey string

	if err := session.Query(`SELECT hash, session_key FROM users_by_email WHERE email = ? LIMIT 1`,
		email).Scan(&databaseHash, &previousSessionKey); err != nil {
		fmt.Println(err)
		log.Fatal(err)
		//TODO: show no user exists by that email if that is the error
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// initialize scryptauth on our secret hmac key
	hmacKey := hmac.New(sha256.New, []byte(HmacSecretKey))
	pwHash, err := scryptauth.New(12, hmacKey.Sum(nil))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	//decode our db hash to a hash and salt
	pwCost, hash, salt, err := scryptauth.DecodeBase64(databaseHash)
	if err != nil {
		fmt.Print(err)
		log.Fatal(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// check the hash and salt against the password
	var pass = c.PostForm("pass")
	ok, err := pwHash.Check(pwCost, hash, []byte(pass), salt)
	if !ok {
		fmt.Printf("Error wrong password for user (%s)", err)
		c.JSON(500, gin.H{"error": "Error wrong password for user"})
		return
	}
	fmt.Println("Login OK")

	// create a session key and update the session key for the user
	sessionKey, err := generateSessionID()
	if err != nil {
		log.Fatal(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	batch := gocql.NewBatch(gocql.LoggedBatch)

	usersStatement := `INSERT INTO users (session_key) VALUES (?) WHERE id = ?`
	usersEmailStatement := `INSERT INTO users_by_email (session_key) VALUES (?) WHERE email = ?`
	usersSessionStatement := `INSERT INTO users_by_session_key (session_key) VALUES (?)
	 WHERE session_key = ?`
	batch.Query(usersStatement, sessionKey, user.ID)
	batch.Query(usersEmailStatement, sessionKey, email)
	batch.Query(usersSessionStatement, sessionKey, previousSessionKey)
	if err := session.ExecuteBatch(batch); err != nil {
		log.Fatal(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie(RentSessionCookie, sessionKey, 0, "", "localhost", true, true)

	//redirect to index //TODO: redirect to the path the user was going to
	c.Redirect(200, "/")
}

//CreateUser is the route to create a user.
func CreateUser(c *gin.Context) {

	// check the user logging out matches user's session key
	user := requestUserFromContext(c)

	fmt.Println(user.Email)
	if user.Email != "" {
		fmt.Println("Can't Create User from a logged in context")
		c.JSON(500, gin.H{"error": "Can't Create User from a logged in context"})
		return
	}

	// initialize a scryptauth object
	hmacKey := hmac.New(sha256.New, []byte(HmacSecretKey))
	pwHash, err := scryptauth.New(12, hmacKey.Sum(nil))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if c.PostForm("pass") == "" {
		fmt.Println("Password is empty!")
		c.JSON(500, gin.H{"error": "Password is Empty!"})
		return
	}
	// generate a hash and salt from the scrypauth
	hash, salt, err := pwHash.Gen([]byte(c.PostForm("pass")))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	//base 64 encoding for easy submitting to db
	hashBase64 := scryptauth.EncodeBase64(pwHash.PwCost, hash, salt)

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	sessionKey, err := generateSessionID()
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	uuid, err := gocql.RandomUUID()
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	batch := gocql.NewBatch(gocql.LoggedBatch)

	usersStatement := `INSERT INTO users (id, name, email, hash, session_key) VALUES
	(?, ?, ?, ?, ?) IF NOT EXISTS`
	usersEmailStatement := `INSERT INTO users_by_email (id, name, email, hash, session_key)
	VALUES (?, ?, ?, ?, ?) IF NOT EXISTS`
	usersSessionStatement := `INSERT INTO users_by_session_key (id, name, email, hash, session_key)
	VALUES (?, ?, ?, ?, ?) IF NOT EXISTS`
	name := c.PostForm("name")
	email := c.PostForm("email")
	batch.Query(usersStatement, uuid, name, email, hashBase64, sessionKey)
	batch.Query(usersEmailStatement, uuid, name, email, hashBase64, sessionKey)
	batch.Query(usersSessionStatement, uuid, name, email, hashBase64, sessionKey)

	var nameCAS, emailCAS, hashCAS, sessionKeyCAS string
	var uuidCAS gocql.UUID

	if applied, _, err := session.ExecuteBatchCAS(batch, &nameCAS, &emailCAS, &hashCAS, &uuidCAS, &sessionKeyCAS); err != nil {
		log.Fatal(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	} else if !applied {
		fmt.Println("Could not create user because email already exists")
		c.JSON(500, gin.H{"error": "Could not create user because email already exists"})
		return
	}

	// name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie(RentSessionCookie, sessionKey, 0, "", "localhost", true, true)

	c.Redirect(200, "/")
}
