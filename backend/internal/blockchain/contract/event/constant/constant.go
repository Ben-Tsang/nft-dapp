package constant

import "time"

// ============================ 通用事件处理常量 ============================
// EventHandleTimeout 所有事件处理器的默认超时时间
const EventHandleTimeout = 10 * time.Second

// ========== HTTP接口专属常量 ==========
const (
	// HTTPRequestTimeout HTTP接口默认超时时间（查询类接口，如列表/详情）
	HTTPRequestTimeout = 3 * time.Second
	// HTTPLongRequestTimeout 长耗时HTTP接口超时（如区块校正/批量操作）
	HTTPLongRequestTimeout = 10 * time.Second
)

// ============================ 事件类型常量（可选，若需统一管理） ============================
const (
	EventTypeListed   = "ItemListed"   // 上架事件
	EventTypeUnlisted = "ItemUnlisted" // 下架事件
	EventTypeBuy      = "Buy"          // 购买事件
)
