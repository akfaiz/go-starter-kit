package dto

type Response[T any] struct {
	Status  int    `json:"status"            example:"200"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type MessageResponse struct {
	Status  int    `json:"status"            example:"200"`
	Message string `json:"message,omitempty" example:"Operation successful"`
}

type ErrorResponse struct {
	Type     string `json:"type"               example:"about:blank"`
	Title    string `json:"title"              example:"Bad Request"`
	Status   int    `json:"status"             example:"400"`
	Detail   string `json:"detail,omitempty"   example:"Invalid request parameters"`
	Instance string `json:"instance,omitempty" example:"request-id-12345"`
}

type ValidationErrorResponse struct {
	Type     string `json:"type"               example:"about:blank"`
	Title    string `json:"title"              example:"Bad Request"`
	Status   int    `json:"status"             example:"422"`
	Detail   string `json:"detail,omitempty"   example:"Invalid request parameters"`
	Instance string `json:"instance,omitempty" example:"request-id-12345"`
	Errors   []struct {
		Field   string `json:"field" example:"email"`
		Message string `json:"message" example:"Email is required"`
	} `json:"errors,omitempty"`
}

func NewResponse[T any](status int, data T, message ...string) Response[T] {
	var msg string
	if len(message) > 0 {
		msg = message[0]
	}
	return Response[T]{
		Status:  status,
		Message: msg,
		Data:    data,
	}
}

func NewMessage(status int, message string) MessageResponse {
	return MessageResponse{
		Status:  status,
		Message: message,
	}
}
