package json_utils

import (
	"encoding/json"
	"net/http"
)

// Вспомогательная функция для парсинга JSON-тела запроса.
func DecodeJSONBody(req *http.Request, dst interface{}) error {
	return json.NewDecoder(req.Body).Decode(dst)
}

// Вспомогательная функция для отправки JSON-ответа.
func SendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}
