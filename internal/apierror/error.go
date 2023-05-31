package apierror

import "net/http"

type HTTPPart struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Error struct {
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	HTTP    HTTPPart               `json:"http"`
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) StatusCode() int {
	return e.HTTP.Code
}

func NewAPIError(msg string, status int) Error {
	return Error{
		Message: msg,
		HTTP: HTTPPart{
			Code:    status,
			Message: http.StatusText(status),
		},
	}
}
