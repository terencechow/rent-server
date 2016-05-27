/*Before you execute the program, Launch `cqlsh` and execute commands found in setup.cql */

package main

import "github.com/gin-gonic/gin"

func main() {

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	/** Routes for Posts **/
	router.GET("/", PostIndex)
	router.GET("/category/:category", PostIndex)
	router.GET("/posts/:state/:post_id", ShowPost)

	authorized := router.Group("/")
	authorized.Use(sessionMiddleware())
	{
		/** Authorized routes for Posts **/
		authorized.POST("/user/post", EditOrCreatePost)
		authorized.DELETE("/user/:user_id/category/:category/post/:state/:post_id", DeletePost)
		/** Routes for Authentication **/
		authorized.POST("/register", CreateUser)
		authorized.POST("/login", LoginUser)
		authorized.POST("/logout", LogoutUser)
		authorized.DELETE("/user/:user_id", DeleteUser)

	}
	router.Run(":8080")

}
