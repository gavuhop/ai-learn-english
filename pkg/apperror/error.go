package apperror

type ErrorResponse struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *ErrorResponse) Error() string { return e.Message }

func New(code, message string) *ErrorResponse {
	return &ErrorResponse{Code: code, Message: message}
}

func (e *ErrorResponse) WithData(data any) *ErrorResponse {
	e.Data = data
	return e
}
