package httputil

import (
	"encoding/json"
	"net/http"
)

// WriteJSon takes a go object turns it to json and sends to client
// v accepts a response/dto
// json.NewEncoder(w).Encode(v) takes struct v conv to json and streams to network response w
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError returns a standard error message
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// DecodeJson receives data from cient, this converts to a struct
// dst must be a pointer to the dto struct to fill up
// newdecoder reads the raw json and maps the keys to struct
func DecodeJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}
