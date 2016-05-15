// TODO:
// 1. double check create user and login user methods. Not entirely sure the methods are proper

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/rs/xhandler"
	"golang.org/x/net/context"
)

// RentSessionCookie is the constant string name used in cookies
const RentSessionCookie string = "RENT_SESSION_ID"

type key int

const userIDKey key = 0

func newContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userIDKey, user)
}

func requestUserFromContext(ctx context.Context) *User {
	user := ctx.Value(userIDKey)
	if user != nil {
		return user.(*User)
	}
	return &User{}
}

func sessionMiddleware(next xhandler.HandlerC) xhandler.HandlerC {
	return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("RentSessionCookie")

		if err != nil || cookie.Value == "" {
			// err is not nil when no cookie by that name could be found
			next.ServeHTTPC(ctx, w, r)
			return
		}

		//a cookie exists, check to ensure it is legit
		// connect to the cluster
		cluster := gocql.NewCluster("127.0.0.1")
		cluster.Keyspace = ClusterKeyspace
		cluster.ProtoVersion = 4
		session, _ := cluster.CreateSession()
		defer session.Close()

		var name, email, latlng string
		var id gocql.UUID
		if err := session.Query(`SELECT * FROM users_by_session_key WHERE session_key = ? LIMIT 1`,
			cookie.Value).Scan(&id, &name, &email, &latlng); err != nil {
			fmt.Println("error getting user by session_key", err)
			next.ServeHTTPC(ctx, w, r)
			return
		}
		user := User{
			ID:     id,
			Name:   name,
			Email:  email,
			Latlng: latlng}

		ctx = newContextWithUser(ctx, &user)
		next.ServeHTTPC(ctx, w, r)
	})

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
