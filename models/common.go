package models

type BaseResponse struct {
	Data   string      `json:"data"`
	Status interface{} `json:"status"`
}
