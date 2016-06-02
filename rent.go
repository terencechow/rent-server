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
	router.GET("/posts", routes.PostIndex)
	router.GET("/posts/category/:category", routes.PostIndex)
	router.GET("/posts/user/:user_id", routes.PostsByUser)
	router.GET("/posts/state/:state/:post_id", routes.ShowPost)

	authorized := router.Group("/")
	authorized.Use(middleware.SessionMiddleware())
	{
		/** Authorized routes for Posts **/
		authorized.POST("/posts/create", routes.EditOrCreatePost)
		authorized.POST("/posts/edit", routes.EditOrCreatePost)
		authorized.DELETE("/posts/state/:state/category/:category/user/:user_id/:post_id", routes.DeletePost)

		/** Routes for Authentication **/
		authorized.POST("/register", routes.CreateUser)
		authorized.POST("/login", routes.LoginUser)
		authorized.POST("/logout", routes.LogoutUser)

		/** user routes **/
		authorized.POST("/user/edit", routes.EditUser)
		authorized.DELETE("/user/:user_id", routes.DeleteUser)

		/** Routes for Chats & Messages **/
		router.GET("/chats/user/:user_id", routes.ChatIndex)
		router.GET("/chats/user/:user_id/:chat_id", routes.MessagesIndex)
		router.POST("/chats/create", routes.CreateChat)
		router.POST("/chats/message/create", routes.CreateMessage)

	}
	router.Run(":8080")

}
