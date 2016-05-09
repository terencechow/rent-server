/*Before you execute the program, Launch `cqlsh` and execute commands found in setup.cql */

package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func main() {
	router := httprouter.New()
  router.GET("/", PostIndex)
	router.GET("/category/:category", PostIndexByCategory)
	router.GET("/category/:category/:post_id", ShowPost)
	router.POST("/post",EditOrCreatePost)
	router.POST("/category/:category/:post_id", DeletePost)
  // router.GET("/hello/:name", handler)
  http.ListenAndServe(":8080", router)

}
