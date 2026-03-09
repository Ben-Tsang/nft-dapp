package person

type Service struct {
	repo *Repository
}

// 构造函数
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// 业务方法, 操作数据库

func (s *Service) CreatePerson(p *Person) error {
	return s.repo.Create(p)
}

func (s *Service) GetPerson(id int, ownerId *int) (*Person, error) {
	return s.repo.ReadSingle(id, ownerId)
}

func (s *Service) GetPersonPage(q PersonPageQuery, ownerId *int) (PersonPageEntity, error) {
	return s.repo.ReadPage(q, ownerId)
}

func (s *Service) UpdatePerson(p *Person, ownerId *int) error {
	return s.repo.Update(p, ownerId)
}

func (s *Service) DeletePerson(id int, ownerId *int) error {
	return s.repo.Delete(id, ownerId)
}
