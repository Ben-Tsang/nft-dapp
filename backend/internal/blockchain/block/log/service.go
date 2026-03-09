package log

type BlockProcessLogService struct {
	repo *BlockProcessLogRepo
}

func NewBlockProcessLogService(repo *BlockProcessLogRepo) *BlockProcessLogService {
	return &BlockProcessLogService{
		repo: repo,
	}
}

// -------------------------- Service 层补充 --------------------------
func (s *BlockProcessLogService) CreateLog(log *BlockProcessLog) error {
	return s.repo.CreateLog(log)
}

// StatusDesc 返回状态的人性化描述（便于日志/接口展示）
func (b *BlockProcessLog) StatusDesc() string {
	switch b.Status {
	case ProcessStatusFailed:
		return "处理失败"
	case ProcessStatusSuccess:
		return "处理成功"
	case ProcessStatusNoNeed:
		return "无需处理"
	default:
		return "未知状态"
	}
}
