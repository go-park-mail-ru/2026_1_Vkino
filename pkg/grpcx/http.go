package grpcx

import (
	"net/http"

	errhttpx "github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap/httpx"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	statusCode, message := errhttpx.DefaultMapper.Map(err)
	httppkg.ErrResponse(w, statusCode, message)
}
