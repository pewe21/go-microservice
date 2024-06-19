package library

type AppResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewResp(message string, data interface{}) *AppResponse {
	return &AppResponse{
		Message: message,
		Data:    data,
	}
}
