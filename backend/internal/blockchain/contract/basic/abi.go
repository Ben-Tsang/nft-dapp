package basic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// LoadABI 加载ABI文件
func LoadABI(abiFilePath string) (abi.ABI, error) {
	abiBytes, err := os.ReadFile(abiFilePath)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("读取ABI文件失败: %v", err)
	}

	var contractData map[string]interface{}
	if err := json.Unmarshal(abiBytes, &contractData); err != nil {
		return abi.ABI{}, fmt.Errorf("解析JSON文件失败: %v", err)
	}

	abiField, exists := contractData["abi"]
	if !exists {
		return abi.ABI{}, fmt.Errorf("ABI字段不存在")
	}

	abiJSON, err := json.Marshal(abiField)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("序列化ABI失败: %v", err)
	}

	contractABI, err := abi.JSON(bytes.NewReader(abiJSON))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("解析ABI失败: %v", err)
	}

	return contractABI, nil
}
