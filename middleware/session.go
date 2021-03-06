// TODO:
// 1. double check create user and login user methods. Not entirely sure the methods are proper

package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/terencechow/rent/constants"
	"github.com/terencechow/rent/models"
)

type key int

const userIDKey key = 0

//RequestUserFromContext return the current user context or an empty interface
func RequestUserFromContext(c *gin.Context) *models.User {
	user, _ := c.Get("user")
	if user != nil {
		return user.(*models.User)
	}
	return &models.User{}
}

//SessionMiddleware set the current user context
func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		cookie, err := c.Cookie(constants.RentSessionCookie)
		if err != nil || cookie == "" {
			// err is not nil then no cookie by that name could be found
			c.Next()
			return
		}

		//a cookie exists, check to ensure it is legit
		// connect to the cluster
		cluster := gocql.NewCluster(constants.IPAddress)
		cluster.Keyspace = constants.ClusterKeyspace
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
		user := models.User{
			ID:         id,
			Name:       name,
			Email:      email,
			SessionKey: cookie}
		fmt.Println("setting context to user", user)
		c.Set("user", &user)
		c.Next()
	}
}

//GenerateSessionID creates a random session id
func GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
