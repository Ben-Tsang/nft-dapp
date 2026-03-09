package constant

type OperateStatus string

// ============================ 操作状态常量（通用） ============================
const (
	OperateStatusSuccess OperateStatus = "success" // 操作成功
	OperateStatusFailed  OperateStatus = "failed"  // 操作失败
)

var OperateStatusLabelMap = map[OperateStatus]string{
	OperateStatusSuccess: "成功",
	OperateStatusFailed:  "失败",
}

// ============================ 辅助方法（可选，提升类型安全性） ============================

// IsValidOperateStatus 校验操作状态是否合法
func IsValidOperateStatus(status string) bool {
	switch status {
	case string(OperateStatusSuccess), string(OperateStatusFailed):
		return true
	default:
		return false
	}
}
