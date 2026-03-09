package status

type BlockProcessStatusService struct {
	repo *BlockProcessStatusRepo
}

func NewBlockProcessStatusService(repo *BlockProcessStatusRepo) *BlockProcessStatusService {
	return &BlockProcessStatusService{
		repo: repo,
	}
}

func (b *BlockProcessStatusService) Init(baseNumber int64) error {
	return b.repo.InitBlockStatus(baseNumber)
}

func (b *BlockProcessStatusService) GetLatestBlockNumber(chainID int64) (int64, error) {
	latestBlock, err := b.repo.GetLatestBlock(chainID)
	if err != nil {
		return 0, err
	}
	blockNumber := latestBlock.LatestBlock
	return blockNumber, nil
}

func (b *BlockProcessStatusService) UpdateLatestBlock(chainID int64, newNumber int64) error {
	blockStatus, err := b.repo.GetLatestBlock(chainID)
	if err != nil {
		return err
	}
	blockStatus.LatestBlock = int64(newNumber)
	err = b.repo.Update(chainID, blockStatus)
	if err != nil {
		return err
	}
	return nil
}
