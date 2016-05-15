/*Before you execute the program, Launch `cqlsh` and execute commands found in setup.cql */

package main

import (
	"log"
	"net/http"

	"github.com/rs/xhandler"
	"github.com/rs/xmux"
)

func main() {

	c := xhandler.Chain{}
	// Append a context-aware middleware handler
	c.UseC(sessionMiddleware)
	mux := xmux.New()

	/** Routes for Posts **/
	mux.GET("/", xhandler.HandlerFuncC(PostIndex))
	mux.GET("/category/:category", xhandler.HandlerFuncC(PostIndex))
	mux.GET("/category/:category/:post_id", xhandler.HandlerFuncC(ShowPost))
	mux.POST("/user/:user_id/post", xhandler.HandlerFuncC(EditOrCreatePost))
	mux.DELETE("/user/:user_id/category/:category/:post_id", xhandler.HandlerFuncC(DeletePost))

	/** Routes for Authentication **/
	mux.POST("/register", xhandler.HandlerFuncC(CreateUser))
	mux.POST("/login", xhandler.HandlerFuncC(LoginUser))
	mux.POST("/logout", xhandler.HandlerFuncC(LogoutUser))
	mux.DELETE("/user", xhandler.HandlerFuncC(DeleteUser))

	if err := http.ListenAndServe(":8080", c.Handler(mux)); err != nil {
		log.Fatal(err)
	}

}
