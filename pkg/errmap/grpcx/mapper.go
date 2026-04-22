package grpcx

import (
	"errors"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrResponse struct {
	Code    codes.Code
	Message string
}

type Mapper struct {
	mapper      *errmap.Mapper[error, error, ErrResponse]
	defaultCode codes.Code
	defaultMsg  string
}

func New(order []error, rules map[error]ErrResponse, defaultCode codes.Code, defaultMsg string) *Mapper {
	return &Mapper{
		mapper: errmap.New(order, rules, func(subject error, key error) bool {
			return errors.Is(subject, key)
		}),
		defaultCode: defaultCode,
		defaultMsg:  defaultMsg,
	}
}

func (m *Mapper) Map(err error) error {
	if err == nil {
		return nil
	}

	rule, ok := m.mapper.Resolve(err)
	if !ok {
		return status.Error(m.defaultCode, m.defaultMsg)
	}

	return status.Error(rule.Code, rule.Message)
}
