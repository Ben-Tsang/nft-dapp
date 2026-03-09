package constant

// 定义操作类型枚举结构体，同时存储英文值和中文描述
type OperateType string

// 枚举常量（英文值），通过注释+映射表关联中文
const (
	OperateTypeMint     OperateType = "mint"      // NFT铸造
	OperateTypeListed   OperateType = "listed"    // NFT上架
	OperateTypeUnlisted OperateType = "unlisted"  // NFT下架
	OperateTypeBuy      OperateType = "buy"       // NFT购买
	OperateTypeTransfer OperateType = "transfer"  // NFT转让（预留）
	OperateTypeSetPrice OperateType = "set_price" // NFT改价（预留）
)

// 定义中英文映射表（核心：从常量关联中文，只需维护这一处）
var OperateTypeLabelMap = map[OperateType]string{
	OperateTypeMint:     "NFT铸造",
	OperateTypeListed:   "NFT上架",
	OperateTypeUnlisted: "NFT下架",
	OperateTypeBuy:      "NFT购买",
	OperateTypeTransfer: "NFT转让",
	OperateTypeSetPrice: "NFT改价",
}

// 可选：获取中文标签的便捷方法（封装逻辑，避免重复代码）
func (ot OperateType) GetLabel() string {
	label, ok := OperateTypeLabelMap[ot]
	if !ok {
		return "未知操作" // 兜底默认值
	}
	return label
}

// 可选：获取所有操作类型枚举列表（给前端返回下拉框数据）
func GetAllOperateTypes() []struct {
	Value string `json:"value"`
	Label string `json:"label"`
} {
	var list []struct {
		Value string `json:"value"`
		Label string `json:"label"`
	}
	for typ, label := range OperateTypeLabelMap {
		list = append(list, struct {
			Value string `json:"value"`
			Label string `json:"label"`
		}{
			Value: string(typ),
			Label: label,
		})
	}
	return list
}

// IsValidOperateType 校验操作类型是否合法
func IsValidOperateType(operateType string) bool {
	// 直接遍历映射表，判断字符串是否在合法值中
	for typ := range OperateTypeLabelMap {
		if string(typ) == operateType {
			return true
		}
	}
	return false
}
