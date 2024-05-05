package subscriber

import "sync"

type MemRepository struct {
	data sync.Map
}

func NewMemRepository() *MemRepository {
	return &MemRepository{
		data: sync.Map{},
	}
}

func (s *MemRepository) Create(address string) error {
	s.data.Store(address, struct{}{})
	return nil
}

func (s *MemRepository) IsSubscriber(address string) bool {
	_, ok := s.data.Load(address)
	return ok
}
