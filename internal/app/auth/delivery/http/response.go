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

func WriteError(w http.ResponseWriter, status int, message string) {
	err := WriteJSON(w, status, errorResponse{Error: message})
	if err != nil {
		log.Printf("error writing response: %v", err)
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if _, err := w.Write(append(b, '\n')); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func ReadJSON(r *http.Request, dst any) error {
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
