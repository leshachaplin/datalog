package http

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

func encodeJSONResponse[T any](w http.ResponseWriter, code int, data T) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if code == http.StatusNoContent {
		return nil
	}

	return json.NewEncoder(w).Encode(data)
}

func getClientIP(req *http.Request) string {
	out, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		if xoff := req.Header.Get("X-Original-Forwarded-For"); xoff != "" {
			out = xoff
		} else {
			xff := strings.Split(req.Header.Get("X-Forwarded-For"), ",")
			if len(xff) == 0 {
				return ""
			}

			if xff[0] != req.Header.Get("X-Envoy-External-Address") {
				out = xff[0]
			}
		}
	}

	if ip := net.ParseIP(out); out != "" && ip != nil {
		if ip.IsLoopback() {
			return "127.0.0.1"
		}

		return out
	}

	return "0.0.0.0"
}
