/*Before you execute the program, Launch `cqlsh` and execute commands found in setup.cql */

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/terencechow/rent/middleware"
	"github.com/terencechow/rent/routes"
)

func main() {

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	/** Routes for Posts **/
	router.GET("/", routes.PostIndex)
	router.GET("/category/:category", routes.PostIndex)
	router.GET("/posts/:state/:post_id", routes.ShowPost)

	authorized := router.Group("/")
	authorized.Use(middleware.SessionMiddleware())
	{
		/** Authorized routes for Posts **/
		authorized.POST("/user/post", routes.EditOrCreatePost)
		authorized.DELETE("/user/:user_id/category/:category/post/:state/:post_id", routes.DeletePost)
		/** Routes for Authentication **/
		authorized.POST("/register", routes.CreateUser)
		authorized.POST("/edit", routes.EditUser)
		authorized.POST("/login", routes.LoginUser)
		authorized.POST("/logout", routes.LogoutUser)
		authorized.DELETE("/user/:user_id", routes.DeleteUser)

	}
	router.Run(":8080")

}
