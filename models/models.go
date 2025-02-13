package models

import (
	"time"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:100;unique;not null"`
	CreatedAt time.Time
}

type FriendRequest struct {
	ID         uint   `gorm:"primaryKey"`
	SenderID   uint   `gorm:"not null"`
	ReceiverID uint   `gorm:"not null"`
	Status     string `gorm:"type:enum('pending', 'accepted', 'rejected');default:'pending'"`
	CreatedAt  time.Time
}

type Friend struct {
	ID        uint `gorm:"primaryKey"`
	User1ID   uint `gorm:"not null"`
	User2ID   uint `gorm:"not null"`
	CreatedAt time.Time
}
