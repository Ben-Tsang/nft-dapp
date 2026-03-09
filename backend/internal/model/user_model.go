package model

import (
	"time"
)

type User struct {
	ID           string `json:"id" gorm:"primary_key"` // 使用钱包地址
	Username     string `json:"username" gorm:"size:64;uniqueIndex;not null"`
	PasswordHash string `json:"-" gorm:"size:255;not null"`
	Role         string `json:"role" gorm:"size:32;not null;default:'user'"`
	//Address      string    `json:"address" gorm:"size:255;not null;default:''"`
	CreatedAt time.Time `json:"createdAt" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"default:CURRENT_TIMESTAMP"`
}
