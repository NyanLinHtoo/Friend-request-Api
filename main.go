package main

import (
	"friend_request_api/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDatabase() *gorm.DB {
	dsn := "root:root@tcp(localhost:3306)/testingdatabase?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the models
	db.AutoMigrate(&models.User{}, &models.FriendRequest{})

	return db
}

func SendFriendRequest(c *gin.Context) {
	type FriendRequestInput struct {
		SenderID   uint `json:"sender_id" binding:"required"`
		ReceiverID uint `json:"receiver_id" binding:"required"`
	}

	var input FriendRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enforce order: smaller ID is always user1_id
	if input.SenderID > input.ReceiverID {
		input.SenderID, input.ReceiverID = input.ReceiverID, input.SenderID
	}

	db := c.MustGet("db").(*gorm.DB)

	// Check if the users are already friends
	var existingFriendship models.Friend
	if err := db.Where("user1_id = ? AND user2_id = ?", input.SenderID, input.ReceiverID).First(&existingFriendship).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Users are already friends"})
		return
	}

	// Check if a friend request already exists
	var existingRequest models.FriendRequest
	if err := db.Where("sender_id = ? AND receiver_id = ?", input.SenderID, input.ReceiverID).First(&existingRequest).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Friend request already exists"})
		return
	}

	// Create a new friend request
	friendRequest := models.FriendRequest{
		SenderID:   input.SenderID,
		ReceiverID: input.ReceiverID,
		Status:     "pending",
	}
	if err := db.Create(&friendRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send friend request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request sent successfully"})
}

func HandleFriendRequest(c *gin.Context) {
	type FriendRequestActionInput struct {
		SenderID   uint   `json:"sender_id" binding:"required"`   // The user who sent the friend request
		ReceiverID uint   `json:"receiver_id" binding:"required"` // The user who received the friend request
		Action     string `json:"action" binding:"required"`      // "accept" or "reject"
	}

	var input FriendRequestActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	// Enforce order: smaller ID is always user1_id
	if input.SenderID > input.ReceiverID {
		input.SenderID, input.ReceiverID = input.ReceiverID, input.SenderID
	}

	// Find the friend request
	var friendRequest models.FriendRequest
	if err := db.Where("sender_id = ? AND receiver_id = ?", input.SenderID, input.ReceiverID).First(&friendRequest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found"})
		return
	}

	if input.Action == "accept" {
		// Create a friendship
		friendship := models.Friend{
			User1ID: friendRequest.SenderID,
			User2ID: friendRequest.ReceiverID,
		}
		if err := db.Create(&friendship).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept friend request"})
			return
		}

		// Delete the friend request
		db.Delete(&friendRequest)

		c.JSON(http.StatusOK, gin.H{"message": "Friend request accepted"})
	} else if input.Action == "reject" {
		// Delete the friend request
		db.Delete(&friendRequest)

		c.JSON(http.StatusOK, gin.H{"message": "Friend request rejected"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
	}
}

func main() {
	r := gin.Default()
	db := ConnectDatabase()

	// Pass the database to the routes
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Define routes
	r.POST("/friend-request", SendFriendRequest)
	r.POST("/friend-request/action", HandleFriendRequest)

	r.Run(":8080") // Start the server
}
