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

func (s *Service) SuccessMessage(_ context.Context) string {
	return "支付成功回调"
}

func (s *Service) FailedMessage(_ context.Context) string {
	return "支付失败回调"
}
