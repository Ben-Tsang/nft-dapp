package _type

// 1. 定义字符串枚举类型（基于string，约束取值范围）
type NFTEventType string

// 2. 直接定义字符串枚举常量（大写导出，供外部使用）
// 这些值就是最终要存储到数据库的字符串，无需额外转换
const (
	NFTMinted    NFTEventType = "NFTMinted"    // 铸造
	Transfer     NFTEventType = "Transfer"     // 转移
	Burn         NFTEventType = "Burn"         // 销毁
	ItemListed   NFTEventType = "ItemListed"   // 上架
	ItemUnlisted NFTEventType = "ItemUnlisted" // 下架
	Buy          NFTEventType = "Buy"          // 购买
	SetPrice     NFTEventType = "SetPrice"     // 修改价格
)

// 3. 可选但必加：合法性校验（防止传入无效值，比如"MINT"/"mint"）
func (t NFTEventType) IsValid() bool {
	switch t {
	case NFTMinted, Transfer, Burn, ItemListed, ItemUnlisted, Buy, SetPrice:
		return true
	default:
		return false
	}
}

// 4. 可选：简化赋值（等价于直接用常量，语义更友好）
func (t NFTEventType) String() string {
	return string(t) // 直接返回枚举的字符串值
}
