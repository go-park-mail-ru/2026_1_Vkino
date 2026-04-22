package httpx

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrResponse struct {
	Status  int
	Message string
}

type Mapper struct {
	mapper        *errmap.Mapper[codes.Code, codes.Code, ErrResponse]
	defaultStatus int
	defaultMsg    string
}

func New(order []codes.Code, rules map[codes.Code]ErrResponse, defaultStatus int, defaultMsg string) *Mapper {
	return &Mapper{
		mapper: errmap.New(order, rules, func(subject codes.Code, key codes.Code) bool {
			return subject == key
		}),
		defaultStatus: defaultStatus,
		defaultMsg:    defaultMsg,
	}
}

func (m *Mapper) Map(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	st, ok := status.FromError(err)
	if !ok {
		return m.defaultStatus, m.defaultMsg
	}

	rule, ok := m.mapper.Resolve(st.Code())
	if !ok {
		return m.defaultStatus, m.defaultMsg
	}

	if rule.Message != "" {
		return rule.Status, rule.Message
	}

	return rule.Status, st.Message()
}
