package auth

import (
	"nft_backend/internal/logger"
	"nft_backend/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("change-me-very-strong-secret")

// 用于生成token, 这里相当于实现了jwt的Claims, 同时加上了自己的一些信息
// 注意必须实现了jwt的claims接口才能用于生成token
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(user *model.User) (string, error) {
	now := time.Now()
	expire := now.Add(time.Hour * 24)
	//
	logger.L.Info("生成token使用的id: " + string(user.ID))
	// claims是jwt的payload, 这里设置了一些用户信息, 就是中间的那部分
	// 这里是在registeredClaims基础上添加了一些自己的信息
	// registeredClaims是jwt的标准字段, 这里设置了token的过期时间
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
		},
	}
	//
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenStr string) (*Claims, error) {
	// 解析 JWT 并验证签名
	token, err := jwt.ParseWithClaims(
		tokenStr,  // 传入 token 字符串
		&Claims{}, // 指定解码到的 Claims 类型
		func(token *jwt.Token) (interface{}, error) { // 提供密钥来验证 token 签名
			return jwtSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	// 断言 token.Claims 为 *Claims 类型，并确保 token 是有效的
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims // 如果断言失败或者 token 无效，返回错误
}
