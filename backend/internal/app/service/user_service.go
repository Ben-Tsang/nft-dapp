package service

import (
	"encoding/json"
	"fmt"
	"log"
	"nft_backend/internal/app/repository"
	"nft_backend/internal/logger"
	"nft_backend/internal/model"
)

type UserService struct {
	repo *repository.UserRepo
}

func NewUserService(repo *repository.UserRepo) *UserService {
	return &UserService{
		repo: repo,
	}
}

// 新增
func (s *UserService) CheckAndCreate(user *model.User) error {
	jsonBytes, _ := json.MarshalIndent(user, "", "  ")
	fmt.Println("service传入的user: " + string(jsonBytes))
	user2, err := s.repo.GetUser(user.ID)

	log.Println("查询到的user:", user2)
	if err != nil {
		logger.L.Warn("获取用户失败,: " + err.Error())
		return err
	}
	// 用户不存在，创建用户
	if user2 == nil {
		logger.L.Warn("用户不存在, 创建用户...")
		err := s.repo.CreateUser(user)
		if err != nil {
			return err
		}
	}
	return nil
}

//
