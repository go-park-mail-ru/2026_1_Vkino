package grpcx

import (
	"net/http"

	errhttpx "github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap/httpx"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

func HTTPStatusFromGRPC(err error) (int, string) {
	return errhttpx.DefaultMapper.Map(err)
}

func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	statusCode, message := HTTPStatusFromGRPC(err)
	httppkg.ErrResponse(w, statusCode, message)
}
