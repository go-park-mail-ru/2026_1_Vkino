package http

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCaseResponse struct {
	Name           string
	StatusCode     int
	Data           interface{}
	ExpectedStatus int
	ExpectedBody   string
}

func TestResponse(t *testing.T) {
	testCases := []TestCaseResponse{
		{
			Name:           "success response",
			StatusCode:     http.StatusOK,
			Data:           struct{ Message string }{Message: "hello"},
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   `{"Message":"hello"}` + "\n",
		},
		{
			Name:           "marshal json error",
			StatusCode:     http.StatusOK,
			Data:           make(chan int),
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedBody:   `{"Error":"internal server error"}` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			w := httptest.NewRecorder()

			Response(w, tc.StatusCode, tc.Data)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.ExpectedStatus, resp.StatusCode)
			assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, tc.ExpectedBody, string(body))
		})
	}
}

type TestCaseRead struct {
	Name        string
	RequestBody string
	ExpectedErr bool
	ExpectedDst interface{}
}

func TestRead(t *testing.T) {
	testCases := []TestCaseRead{
		{
			Name:        "successful read",
			RequestBody: `{"name":"test","age":25}`,
			ExpectedErr: false,
			ExpectedDst: &struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{Name: "test", Age: 25},
		},
		{
			Name:        "invalid json",
			RequestBody: `{"name":"test",age:25}`,
			ExpectedErr: true,
			ExpectedDst: &struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{},
		},
		{
			Name:        "unknown fields",
			RequestBody: `{"name":"test","age":25,"extra":"field"}`,
			ExpectedErr: true,
			ExpectedDst: &struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{},
		},
		{
			Name:        "empty body",
			RequestBody: ``,
			ExpectedErr: true,
			ExpectedDst: &struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tc.RequestBody)))

			dst := &struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{}

			err := Read(req, dst)

			if tc.ExpectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.ExpectedDst, dst)
			}
		})
	}
}
