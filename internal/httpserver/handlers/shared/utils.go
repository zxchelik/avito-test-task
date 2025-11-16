package shared

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, code ErrorCode, msg string) {
	WriteJSON(w, status, ErrorResponse{
		Error: errorBody{
			Code:    code,
			Message: msg,
		},
	})
}

func WriteInternalError(w http.ResponseWriter, err error, log *slog.Logger) {
	// В проде сюда бы логгер.
	log.Error("Internal Server Error", slog.String("error", err.Error()))
	WriteError(w, http.StatusInternalServerError, ErrorCodeInternal, "internal server error")
}
