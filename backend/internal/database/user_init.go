package database

import (
	"nft_backend/internal/model"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func EnsureDefaultAdmin(db *gorm.DB) error {

	const defaultAdminUsername = "admin"
	const defaultAdminPassword = "123456"

	// 检查默认管理员是否存在
	var admin model.User
	if err := db.Where("username =?", defaultAdminUsername).First(&admin).Error; err != nil {
		// 如果不存在，则创建默认管理员

		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)

		admin = model.User{
			Username:     defaultAdminUsername,
			PasswordHash: string(passwordHash),
			Role:         "admin",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
	}

	return nil
}
