package response

import (
	"net/http"

	"github.com/go-chi/render"
)

type APIError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func WriteJSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	render.Status(r, status)
	render.JSON(w, r, payload)
}

func WriteJSONError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	render.Status(r, status)
	render.JSON(w, r, APIError{
		Error: msg,
		Code:  status,
	})
}
