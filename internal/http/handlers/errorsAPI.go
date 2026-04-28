package handlers

import (
	"bernard/internal/domain/entity"
	"net/http"
)

func (h *Handlers) errorResponse(w http.ResponseWriter, statusCode int, message string, details interface{}) {
	h.writeJSON(w, statusCode, entity.ErrorResponse{
		Code:    statusCode,
		Message: message,
		Details: details,
	})
}

func (h *Handlers) serverError(w http.ResponseWriter, r *http.Request, err error) {
	h.log.Error("internal server error", "method", r.Method, "path", r.URL.Path, "error", err)
	h.errorResponse(w, http.StatusInternalServerError, "Internal server error", nil)
}
