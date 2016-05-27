// TODO:
// 1. double check create user and login user methods. Not entirely sure the methods are proper

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

// ClusterKeyspace - name of keyspace
const ClusterKeyspace string = "rent"

// RentSessionCookie is the constant string name used in cookies
const RentSessionCookie string = "RENT_SESSION_ID"

type key int

const userIDKey key = 0

func requestUserFromContext(c *gin.Context) *User {
	user, _ := c.Get("user")
	if user != nil {
		return user.(*User)
	}
	return &User{}
}

func sessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		cookie, err := c.Cookie(RentSessionCookie)
		if err != nil || cookie == "" {
			// err is not nil then no cookie by that name could be found
			c.Next()
			return
		}

		//a cookie exists, check to ensure it is legit
		// connect to the cluster
		cluster := gocql.NewCluster("127.0.0.1")
		cluster.Keyspace = ClusterKeyspace
		cluster.ProtoVersion = 4
		session, _ := cluster.CreateSession()
		defer session.Close()

		var name, email string
		var id gocql.UUID
		fmt.Println("cookie value is", cookie)
		if err := session.Query(`SELECT id,name,email FROM users_by_session_key WHERE session_key = ? LIMIT 1`,
			cookie).Scan(&id, &name, &email); err != nil {
			fmt.Println("error getting user by session_key", err)
			c.Next()
			return
		}
		user := User{
			ID:    id,
			Name:  name,
			Email: email}
		fmt.Println("setting context to user", user)
		c.Set("user", user)
		c.Next()
	}
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
