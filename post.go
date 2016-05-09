package main

import (
  "net/http"
  "github.com/gocql/gocql"
  "github.com/julienschmidt/httprouter"
  "log"
  "encoding/json"
  "strconv"
  // "fmt"
)

//TODO: authenticate if they are logged in before doing a delete otherwise anyone can just delete everything.
func DeletePost(w http.ResponseWriter, r *http.Request, ps httprouter.Params){
  var post_id_from_form string = ps.ByName("post_id")
  var category string = ps.ByName("category")
  // connect to the cluster
  cluster := gocql.NewCluster("127.0.0.1")
  cluster.Keyspace = "rent"
  cluster.ProtoVersion = 4
  session, _ := cluster.CreateSession()
  defer session.Close()

  if err := session.Query(`DELETE FROM posts_by_category WHERE category = ? and post_id = ?`,
      category, post_id_from_form).Exec(); err != nil {
      log.Fatal(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
  }
  http.Redirect(w, r, "/", http.StatusFound)
}

func ShowPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params){
  var post_id_from_form string = ps.ByName("post_id")
  var category string = ps.ByName("category")
  var user_id, post_id gocql.UUID
  var title, description, latlng string
  var price int
  var available bool
  var images []string

  // connect to the cluster
  cluster := gocql.NewCluster("127.0.0.1")
  cluster.Keyspace = "rent"
  cluster.ProtoVersion = 4
  session, _ := cluster.CreateSession()
  defer session.Close()

  if err := session.Query(`SELECT
   user_id, post_id, title, description, price, available, category, images, latlng
   FROM posts_by_category WHERE category = ? and post_id = ? LIMIT 1`,
      category, post_id_from_form).Scan(&user_id, &post_id, &title, &description, &price, &available, &category, &images, &latlng); err != nil {
      // log.Fatal(err)
      // http.Error(w, err.Error(), http.StatusInternalServerError)
      //TODO: redirect to a 'not found page'
      http.Redirect(w, r, "/", http.StatusFound)
      return
  }

  post := Post{
    User_id: user_id,
    Post_id: post_id,
    Title: title,
    Description: description,
    Price: price,
    Available: available,
    Category: category,
    Images: images,
    Latlng: latlng }

  post_in_json, _ := json.Marshal(post)

  w.Header().Set("Content-Type", "application/json;charset=UTF-8")
  w.WriteHeader(http.StatusOK)
  w.Write(post_in_json)
}

//TODO: authenticate if they are logged in before doing a create / edit
//  otherwise anyone can post and create / edit everything.
func EditOrCreatePost(w http.ResponseWriter, r *http.Request, ps httprouter.Params){

	var title string = r.FormValue("title")
  var description string = r.FormValue("description")
  var category string = r.FormValue("category")
  var latlng string = r.FormValue("latlng")
  var available_from_form string = r.FormValue("available")
	price, _ := strconv.Atoi(r.FormValue("price"))
	var images []string //TODO: handle

  var post_id string = r.FormValue("post_id")
  if (post_id == ""){
    post_id = gocql.TimeUUID().String()
  }

  var available bool
  if (available_from_form == "true") {
    available = true
  } else {
    available = false
  }

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "rent"
  cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	// insert a post TTL 5184000 = 2 months in seconds
  if err := session.Query(`INSERT INTO posts_by_category
		(post_id, title, description, price, available, category, images, latlng)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?) USING TTL 5184000`,
    post_id, title, description, price, available, category, images, latlng).Exec(); err != nil {
      log.Fatal(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
  }
  http.Redirect(w, r, "/", http.StatusFound)
}


//TODO: combine the PostIndex and PostIndexByCategory functions since they basically do the same thing
func PostIndexByCategory(w http.ResponseWriter, r *http.Request, ps httprouter.Params){
  var user_id, post_id gocql.UUID
	var title, description, category, latlng string
	var price int
	var available bool
	var images []string
  // connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "rent"
  cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

  var search_category string = ps.ByName("category")

  iter := session.Query(`SELECT
		user_id, post_id, title, description, price, available, category, images, latlng
		FROM posts_by_category WHERE category = ? LIMIT 10`, search_category).Iter()

	var posts []Post

	for iter.Scan(&user_id, &post_id, &title, &description, &price, &available, &category, &images, &latlng) {
      post := Post{
				User_id: user_id,
				Post_id: post_id,
				Title: title,
				Description: description,
				Price: price,
				Available: available,
				Category: category,
				Images: images,
				Latlng: latlng }
			posts = append(posts,post)
  }
	posts_in_json, _ := json.Marshal(posts)

	if err := iter.Close(); err != nil {
		 log.Fatal(err)
     http.Error(w, err.Error(), http.StatusInternalServerError)
     return
 	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
  w.WriteHeader(http.StatusOK)
 	w.Write(posts_in_json)
}

func PostIndex(w http.ResponseWriter, r *http.Request, ps httprouter.Params){
	var user_id, post_id gocql.UUID
	var title, description, category, latlng string
	var price int
	var available bool
	var images []string

	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "rent"
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	iter := session.Query(`SELECT
		user_id, post_id, title, description, price, available, category, images, latlng
		FROM posts_by_category LIMIT 10`).Iter()

	var posts []Post

	for iter.Scan(&user_id, &post_id, &title, &description, &price, &available, &category, &images, &latlng) {
      post := Post{
				User_id: user_id,
				Post_id: post_id,
				Title: title,
				Description: description,
				Price: price,
				Available: available,
				Category: category,
				Images: images,
				Latlng: latlng }
			posts = append(posts,post)
  }
	posts_in_json, _ := json.Marshal(posts)

	if err := iter.Close(); err != nil {
		 log.Fatal(err)
     http.Error(w, err.Error(), http.StatusInternalServerError)
     return
 	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
  w.WriteHeader(http.StatusOK)
 	w.Write(posts_in_json)
}
