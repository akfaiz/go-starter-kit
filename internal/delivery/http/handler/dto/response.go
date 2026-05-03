package dto

type Response[T any] struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type GenericResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
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

func NewMessage(status int, message string) GenericResponse {
	return GenericResponse{
		Status:  status,
		Message: message,
	}
}
