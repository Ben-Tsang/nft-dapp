package repository

import (
	"errors"
	"log"
	"nft_backend/internal/model"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(user *model.User) error {
	return r.db.Create(&user).Error
}

func (r *UserRepo) GetUser(id string) (*model.User, error) {
	log.Println("获取user传入的id: ", id)
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error // 错误检查
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有找到记录，可以返回 nil 和特定的错误
			log.Println("用户不存在")
			return nil, nil
		}
		// 如果发生其他错误，返回
		log.Println("查询用户失败:", err)
		return nil, err
	}
	// 如果找到了记录，返回用户
	return &user, nil
}
