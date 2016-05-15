// TODO:
// 1. double check create user and login user methods. Not entirely sure the methods are proper
// 2. pass the session key through the http header, not as a variable in the param

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gebi/scryptauth"
	"github.com/gocql/gocql"
	"golang.org/x/net/context"
)

// HmacSecretKey is the secret key that hmac is generated on
var HmacSecretKey = os.Getenv("HMAC_SECRET_KEY")

//DeleteUser route to delete user
func DeleteUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := requestUserFromContext(ctx)

	if user.Email != r.FormValue("email") {
		fmt.Println("Can't delete user from a different user context")
		http.Error(w, "Can't delete user from a different user context", http.StatusInternalServerError)
		return
	}

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	if err := session.Query(`DELETE FROM users_by_email WHERE email = ?`,
		r.FormValue("email")).Exec(); err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//LogoutUser route to log out user
func LogoutUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	// check the user logging out matches user's session key
	user := requestUserFromContext(ctx)

	if user.Email != r.FormValue("email") {
		fmt.Println("Can't logout user from a different user context")
		http.Error(w, "Can't logout user from a different user context", http.StatusInternalServerError)
		return
	}

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	// insert an empty string into the user session_key
	if err := session.Query(`INSERT INTO users_by_email
		session_key VALUES ? WHERE email = ?`,
		"", r.FormValue("email")).Exec(); err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//send back new cookie and expire it immediately
	cookie := http.Cookie{Name: RentSessionCookie, Value: "", Secure: true, HttpOnly: true, MaxAge: -1}
	http.SetCookie(w, &cookie)

	// http.Redirect(w, r, "/", http.StatusFound)
	w.WriteHeader(http.StatusOK)
}

//LoginUser route to login a user
func LoginUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var latlng = r.FormValue("latlng")
	var email = r.FormValue("email")
	user := requestUserFromContext(ctx)

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	// if we alreay have a user from context, it means we have a valid session_key,
	// no need to login the user, simply update the latlng of the user in our db and redirect
	if user != nil {
		// create a session key and update the session key and the latlng for the user

		if err := session.Query(`INSERT INTO users_by_email
  		(latlng) VALUES (? WHERE email = ?`,
			latlng, email).Exec(); err != nil {
			log.Fatal(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//redirect to index //TODO: redirect to the path the user was going to
		http.Redirect(w, r, "/", http.StatusOK)
		return
	}

	//no user in session, need to login user
	//get database hash and user name, by the email
	var databaseHash string
	var name string
	var id gocql.UUID

	if err := session.Query(`SELECT id, name, hash FROM users_by_email WHERE email = ? LIMIT 1`,
		email).Scan(&id, &name, &databaseHash); err != nil {
		fmt.Println(err)
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		//TODO: show no user exists by that email if that is the error
		return
	}

	// initialize scryptauth on our secret hmac key
	hmacKey := hmac.New(sha256.New, []byte(HmacSecretKey))
	pwHash, err := scryptauth.New(12, hmacKey.Sum(nil))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//decode our db hash to a hash and salt
	pwCost, hash, salt, err := scryptauth.DecodeBase64(databaseHash)
	if err != nil {
		fmt.Print(err)
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// check the hash and salt against the password
	var pass = r.FormValue("pass")
	ok, err := pwHash.Check(pwCost, hash, []byte(pass), salt)
	if !ok {
		fmt.Printf("Error wrong password for user (%s)", err)
		//TODO: return an json showing error
		return
	}
	fmt.Println("Login OK")

	// create a session key and update the session key and the latlng for the user
	sessionKey, err := generateSessionID()
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := session.Query(`INSERT INTO users_by_email
		(latlng, session_key) VALUES (?, ?) WHERE email = ?`,
		latlng, sessionKey, email).Exec(); err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//send back new cookie
	cookie := http.Cookie{Name: RentSessionCookie, Value: sessionKey, Secure: true, HttpOnly: true}
	http.SetCookie(w, &cookie)

	//redirect to index //TODO: redirect to the path the user was going to
	http.Redirect(w, r, "/", http.StatusOK)
}

//CreateUser route to create a user
func CreateUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	// check the user logging out matches user's session key
	user := requestUserFromContext(ctx)

	if user != nil {
		fmt.Println("Can't Create User from a logged in context")
		http.Error(w, "Can't Create User from a logged in context", http.StatusInternalServerError)
		return
	}

	// initialize a scryptauth object
	hmacKey := hmac.New(sha256.New, []byte(HmacSecretKey))
	pwHash, err := scryptauth.New(12, hmacKey.Sum(nil))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// generate a hash and salt from the scrypauth
	hash, salt, err := pwHash.Gen([]byte(r.FormValue("pass")))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	uuid, err := gocql.RandomUUID()
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//store new user in db
	if err := session.Query(`INSERT INTO users_by_email
		(id, name, email, hash, session_key) VALUES (?, ?, ?, ?) IF NOT EXISTS`,
		uuid, r.FormValue("name"), r.FormValue("email"), hashBase64, sessionKey).Exec(); err != nil {
		fmt.Println(err)
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{Name: RentSessionCookie, Value: sessionKey, Secure: true, HttpOnly: true, MaxAge: 0}
	http.SetCookie(w, &cookie)

	//redirect to index //TODO: redirect to a specific path
	http.Redirect(w, r, "/", http.StatusOK)
}
