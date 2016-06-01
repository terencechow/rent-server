package routes

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/terencechow/rent/constants"
	"github.com/terencechow/rent/middleware"
	"github.com/terencechow/rent/models"
)

// MessagesIndex route
func MessagesIndex(c *gin.Context) {
	user := middleware.RequestUserFromContext(c)
	if user.ID.String() != c.Param("user_id") || user.ID.String() == "" {
		fmt.Println("Can't view messages from a different user context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't view messages from a different user context"})
		return
	}

	// connect to the cluster
	cluster := gocql.NewCluster(constants.IPAddress)
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	iter := session.Query(`SELECT message_time, owner_id, owner_name,
    borrower_id, borrower_name, message_content, message_sender_id
    FROM messages_by_chat WHERE chat_id = ?`, c.Param("chat_id")).Iter()

	var ownerID, senderID, borrowerID gocql.UUID
	var messageTime time.Time
	var ownerName, borrowerName, messageContent string
	var messages []models.Message

	for iter.Scan(&messageTime, &ownerID, &ownerName, &borrowerID, &borrowerName, &messageContent, &senderID) {
		var senderName string
		if senderID == ownerID {
			senderName = ownerName
		} else {
			senderName = borrowerName
		}
		message := models.Message{
			SenderID:       senderID,
			SenderName:     senderName,
			MessageContent: messageContent,
			MessageTime:    messageTime}
		messages = append(messages, message)
	}

	if err := iter.Close(); err != nil {
		fmt.Println("Error in iter.close", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}

// CreateMessage route
// postform requires user_id, chat_id, message_content, post_id, recipientId
func CreateMessage(c *gin.Context) {
	user := middleware.RequestUserFromContext(c)
	var userID = c.PostForm("user_id")
	if user.ID.String() != userID && user.ID.String() != "" || user.ID.String() == "" {
		fmt.Println("Can't create message chats from a different user context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't create message from a different user context"})
		return
	}

	// connect to the cluster
	cluster := gocql.NewCluster(constants.IPAddress)
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	var chatID = c.PostForm("chat_id")
	var messageTime = time.Now()
	messageID, _ := gocql.RandomUUID()
	messageContent := c.PostForm("message_content")
	var postID = c.PostForm("post_id")
	var recipientID = c.PostForm("recipient_id")

	var messageStatement = `INSERT INTO messages_by_chat
  (chat_id, message_time, message_id, message_content, message_sender_id)
  VALUES (?,?,?,?,?)`

	var chatsStatement = `INSERT INTO chats
    (post_id, user_id, recipient_id, last_message_time)
    VALUES (?, ?, ?, ?)`

	batch := gocql.NewBatch(gocql.LoggedBatch)
	batch.Query(messageStatement, chatID, messageTime, messageID, messageContent, userID)
	batch.Query(chatsStatement, postID, userID, recipientID, messageTime)
	batch.Query(chatsStatement, postID, recipientID, userID, messageTime)

	if err := session.ExecuteBatch(batch); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Success": "Successfully sent message"})
}

// ChatIndex route
func ChatIndex(c *gin.Context) {
	user := middleware.RequestUserFromContext(c)
	if user.ID.String() != c.Param("user_id") || user.ID.String() == "" {
		fmt.Println("Can't view chats from a different user context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't view chats from a different user context"})
		return
	}

	// connect to the cluster
	cluster := gocql.NewCluster(constants.IPAddress)
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	var chatID, postID, userID, recipientID gocql.UUID
	var postName, mainImageURL string
	var lastMessageTime time.Time

	iter := session.Query(`SELECT
    chat_id, post_id, user_id, recipient_id, post_name, main_image_url, last_message_time
    FROM chats WHERE user_id = ?`, user.ID).Iter()

	var chats []models.Chat
	for iter.Scan(&chatID, &postID, &userID, &recipientID, &postName, &mainImageURL, &lastMessageTime) {
		chat := models.Chat{
			ChatID:          chatID,
			PostID:          postID,
			UserID:          userID,
			RecipientID:     recipientID,
			PostName:        postName,
			MainImageURL:    mainImageURL,
			LastMessageTime: lastMessageTime}
		chats = append(chats, chat)
	}

	if err := iter.Close(); err != nil {
		fmt.Println("Error in iter.close", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chats)
}

// CreateChat route
// requires a post form with:
// - user_id, recipient_id, post_id, post_name, main_image_url, recipient_name, message_content
func CreateChat(c *gin.Context) {

	user := middleware.RequestUserFromContext(c)
	var userID = c.PostForm("user_id")
	var recipientID = c.PostForm("recipient_id")
	if user.ID.String() != userID && user.ID.String() != "" || user.ID.String() == "" {
		fmt.Println("Can't create chat from a different user context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't create chat from a different user context"})
		return
	}

	chatID, _ := gocql.RandomUUID()
	var postID = c.PostForm("post_id")
	var postName = c.PostForm("post_name")
	var mainImageURL = c.PostForm("main_image_url")

	lastMessageTime := time.Now()

	// connect to the cluster
	cluster := gocql.NewCluster(constants.IPAddress)
	cluster.Keyspace = constants.ClusterKeyspace
	cluster.ProtoVersion = 4
	session, _ := cluster.CreateSession()
	defer session.Close()

	var chatsStatement = `INSERT INTO chats
    (chat_id, post_id, user_id, recipient_id, post_name, main_image_url, last_message_time)
    VALUES (?,?, ?, ?, ?, ?, ?)`
	if applied, err :=
		session.Query(chatsStatement+" IF NOT EXISTS", chatID, postID, userID, recipientID, postName, mainImageURL, lastMessageTime).
			ScanCAS(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if !applied {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create Chat because a chat between these users already exists for this post"})
		return
	}

	batch := gocql.NewBatch(gocql.LoggedBatch)
	batch.Query(chatsStatement, chatID, postID, recipientID, userID, postName, mainImageURL, lastMessageTime)

	var firstMessageStatement = `INSERT INTO messages_by_chat
  (chat_id, message_time, message_id, owner_id, owner_name, borrower_id, borrower_name,
    message_content, message_sender_id) VALUES (?,?,?,?,?,?,?,?,?)`

	batch.Query(firstMessageStatement, chatID, lastMessageTime,
		recipientID, c.PostForm("recipient_name"), userID, user.Name, c.PostForm("message_content"), userID)

	if err := session.ExecuteBatch(batch); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Success": "Successfully created chat"})
}
