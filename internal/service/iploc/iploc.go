package iploc

import (
	"fmt"
	"strings"

	ipdb "github.com/ipipdotnet/ipdb-go"

	"xj_comp/internal/domain"
)

type Locator interface {
	Find(ip string) ([]string, error)
}

type IPDBLocator struct {
	city *ipdb.City
}

func NewIPDBLocator(path string) (*IPDBLocator, error) {
	city, err := ipdb.NewCity(path)
	if err != nil {
		return nil, fmt.Errorf("open ipdb: %w", err)
	}
	return &IPDBLocator{city: city}, nil
}

func (l *IPDBLocator) Find(ip string) ([]string, error) {
	return l.city.Find(ip, "CN")
}

type Service struct {
	locator Locator
}

func NewService(locator Locator) *Service {
	return &Service{locator: locator}
}

func (s *Service) Find(ip string) domain.IPLocData {
	parts, err := s.locator.Find(ip)
	if err != nil {
		parts = nil
	}
	return domain.IPLocData{Data: strings.TrimSpace(strings.Join(parts, " "))}
}
