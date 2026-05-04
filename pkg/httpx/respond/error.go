package respond

import (
	"net/http"

	errhttpx "github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap/httpx"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

func Error(w http.ResponseWriter, err error) {
	statusCode, message := errhttpx.DefaultMapper.Map(err)
	httppkg.ErrResponse(w, statusCode, message)
}
