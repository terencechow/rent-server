
package main

import (
  "net/http"
  "github.com/gocql/gocql"
  "github.com/julienschmidt/httprouter"
  "log"
  "fmt"
  // "encoding/json"
)

func User(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
  var id gocql.UUID
  var name string
  var email string
  var post_ids []string
  var chats []string

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "rent"
	session, _ := cluster.CreateSession()
	defer session.Close()

  if err := session.Query(`SELECT id, name, email, post_ids, chats FROM users_by_id WHERE id = ? LIMIT 1`,
     "me").Consistency(gocql.One).Scan(&id, &name, &email, &post_ids, &chats); err != nil {
     log.Fatal(err)
  }
	fmt.Fprintf(w, "Hi thtstere, I love!")
}
