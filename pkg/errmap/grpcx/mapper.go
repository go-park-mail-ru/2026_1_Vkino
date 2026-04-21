package grpcx

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Rule struct {
	Code    codes.Code
	Message string
}

type Mapper struct {
	order       []error
	rules       map[error]Rule
	defaultCode codes.Code
	defaultMsg  string
}

func New(order []error, rules map[error]Rule, defaultCode codes.Code, defaultMsg string) *Mapper {
	return &Mapper{
		order:       order,
		rules:       rules,
		defaultCode: defaultCode,
		defaultMsg:  defaultMsg,
	}
}

func (m *Mapper) Map(err error) error {
	if err == nil {
		return nil
	}

	for _, key := range m.order {
		if errors.Is(err, key) {
			rule := m.rules[key]
			return status.Error(rule.Code, rule.Message)
		}
	}

	return status.Error(m.defaultCode, m.defaultMsg)
}