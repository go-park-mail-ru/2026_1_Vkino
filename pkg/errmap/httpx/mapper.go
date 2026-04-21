package httpx

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Rule struct {
	Status  int
	Message string
}

type Mapper struct {
	order         []codes.Code
	rules         map[codes.Code]Rule
	defaultStatus int
	defaultMsg    string
}

func New(order []codes.Code, rules map[codes.Code]Rule, defaultStatus int, defaultMsg string) *Mapper {
	return &Mapper{
		order:         order,
		rules:         rules,
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

	for _, code := range m.order {
		if st.Code() == code {
			rule := m.rules[code]
			if rule.Message != "" {
				return rule.Status, rule.Message
			}
			return rule.Status, st.Message()
		}
	}

	return m.defaultStatus, m.defaultMsg
}