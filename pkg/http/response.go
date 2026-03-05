package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type errorResponse struct {
	Error string
}

func ErrResponse(w http.ResponseWriter, status int, message string) {
	Response(w, status, errorResponse{Error: message})
}

func Response(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	b, err := json.Marshal(v)

	if err != nil {
		log.Printf("marshal json error: %v", err)
		ErrResponse(w, http.StatusInternalServerError, "internal server error")
	}

	w.WriteHeader(status)

	if _, err := w.Write(append(b, '\n')); err != nil {
		log.Printf("write response error: %v", err)
	}
}

func Read(r *http.Request, dst any) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("error in closing body %v\n", err)
		}
	}(r.Body)

	dec := json.NewDecoder(r.Body)

	// если клиент пришлёт лишние поля, которых нет в dst, будет ошибка.
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	var extra any
	if err := dec.Decode(&extra); err != nil {
		if err == io.EOF {
			return nil
		}

		return err
	}

	return fmt.Errorf("request body must contain a single JSON object")
}
