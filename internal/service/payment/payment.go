package payment

import "context"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Unpaid(_ context.Context) map[string]interface{} {
	return map[string]interface{}{
		"total_count": 0,
	}
}
