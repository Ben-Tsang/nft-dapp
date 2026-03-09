package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"nft_backend/internal/app/service"
	"nft_backend/internal/app/web/response"
	auth2 "nft_backend/internal/auth"
	"nft_backend/internal/model"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	userService *service.UserService
	db          *gorm.DB
	rdb         *redis.Client
}

func NewAuthHandler(userService *service.UserService, rdb *redis.Client) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		rdb:         rdb}
}

type LoginCommonReq struct {
	Username string `json:"username" binding:"required" label:"用户名"`
	Password string `json:"password" binding:"required" label:"密码"`
}

type LoginNonceReq struct {
	Address   string `json:"address" binding:"required" label:"钱包地址"`
	Nonce     string `json:"nonce" binding:"required" label:"随机数"`
	Signature string `json:"signature" binding:"required" label:"签名"`
}

// 使用并发map, key和value是interface{}类型，可以存储任意类型的值。
//var nonceMap sync.Map

func (auth *AuthHandler) Nonce(c *gin.Context) {
	nonce, _ := auth2.GenerateNonce(auth.rdb)
	response.OK(c, "", gin.H{
		"nonce": nonce,
	})
}

// 登录, 前端传入钱包地址, nonce,
func (auth *AuthHandler) Login(c *gin.Context) {

	var loginNonceReq LoginNonceReq

	if err := c.ShouldBindJSON(&loginNonceReq); err != nil {
		response.Fail(c, 400, "参数错误")
	}
	jsonBytes, _ := json.MarshalIndent(loginNonceReq, "", "  ")
	log.Println("loginNonceReq:", string(jsonBytes))
	// 检查nonce是否有效
	nonce := loginNonceReq.Nonce

	nonceValue, err := auth.rdb.Get(context.Background(), nonce).Result()
	if err == redis.Nil {
		response.Fail(c, 400, "nonce不存在")
	} else if err != nil {
		response.Fail(c, 400, "nonce无效: "+err.Error())
	}
	log.Println("登录中的nonce值:", nonceValue)
	auth.rdb.Del(context.Background(), nonce)
	// 检查钱包地址和签名是否匹配, 需要从签名恢复钱包地址
	signature := loginNonceReq.Signature
	addr, err := RecoverAddressFromSignature(nonce, signature)
	if err != nil {
		response.Fail(c, 400, "签名无效: "+err.Error())
	}
	log.Println("传入的钱包地址:", loginNonceReq.Address)
	if addr.Hex() != loginNonceReq.Address {
		response.Fail(c, 400, "签名与钱包地址不符")
	}
	// 生成token, 先检查用户是否存在, 如果用户不存在就用当前钱包地址创用户
	currentUser := model.User{
		ID:           loginNonceReq.Address,
		Username:     loginNonceReq.Address,
		Role:         "0",
		PasswordHash: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	jsonBytes, _ = json.MarshalIndent(currentUser, "", "  ")
	log.Println("组装的用户: " + string(jsonBytes))
	err = auth.userService.CheckAndCreate(&currentUser)
	if err != nil {
		response.Fail(c, 500, "创建用户失败")
	}
	token, err := auth2.GenerateToken(&currentUser)
	if err != nil {
		response.Fail(c, 500, "生成token失败")
	}
	response.OK(c, "", gin.H{
		"token": token,
	})
}

// RecoverAddressFromSignature 从签名恢复钱包地址
// nonce: 签名的原始消息
// signature: 十六进制格式的签名（带或不带0x前缀）
// 返回：恢复的地址和可能的错误
func RecoverAddressFromSignature(nonce, signature string) (common.Address, error) {
	// 1. 清理签名，移除可能的0x前缀
	signature = strings.TrimPrefix(signature, "0x")

	// 2. 解码签名
	sigBytes, err := hexutil.Decode("0x" + signature)
	if err != nil {
		return common.Address{}, fmt.Errorf("签名解码失败: %v", err)
	}

	// 3. 验证签名长度
	if len(sigBytes) != 65 {
		return common.Address{}, fmt.Errorf("签名长度无效: %d, 应为65字节", len(sigBytes))
	}

	// 4. 构造以太坊签名消息格式
	// MetaMask/Ethers使用: \x19Ethereum Signed Message:\n + 消息长度 + 消息
	message := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(nonce), nonce)

	// 5. 计算Keccak256哈希
	hash := crypto.Keccak256Hash([]byte(message))

	// 6. 处理V值（签名中的最后一个字节）
	v := sigBytes[64]

	// 创建签名的副本，以便修改V值
	finalSig := make([]byte, 65)
	copy(finalSig, sigBytes)

	// 7. 调整V值为0或1
	// Ethers v6通常返回0或1，但也可能返回27/28
	if v == 27 || v == 28 {
		finalSig[64] = v - 27
	} else if v == 0 || v == 1 {
		finalSig[64] = v
	} else {
		// 尝试EIP-155格式或直接使用
		finalSig[64] = v % 2 // 取模确保为0或1
	}

	// 8. 从签名恢复公钥
	publicKey, err := crypto.SigToPub(hash.Bytes(), finalSig)
	if err != nil {
		return common.Address{}, fmt.Errorf("恢复公钥失败: %v", err)
	}

	// 9. 从公钥计算地址
	address := crypto.PubkeyToAddress(*publicKey)

	return address, nil
}

// 常规登录, 需要用户名密码
func (auth *AuthHandler) loginCommon(c *gin.Context) {
	var loginCommonReq LoginCommonReq
	// 提取参数
	if err := c.ShouldBindJSON(&loginCommonReq); err != nil {
		response.Fail(c, 400, "参数错误")
	}

	// 查询用户是否存在
	var user model.User
	if err := auth.db.Where("username =?", loginCommonReq.Username).First(&user).Error; err != nil {
		response.Fail(c, 400, "用户名或密码错误")
	}
	fmt.Println(user)
	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginCommonReq.Password)); err != nil {
		response.Fail(c, 400, "用户名或密码错误")
	}

	// 生成token
	token, err := auth2.GenerateToken(&user)
	if err != nil {
		response.Fail(c, 500, "生成token失败")
	}

	response.OK(c, "", gin.H{
		"token":    token,
		"userName": user.Username,
		"role":     user.Role,
	})
}
