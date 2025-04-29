package dto

type UserAddReq struct {
	Id         string `json:"client_id"`
	Capacity   string `json:"capacity"`
	RatePerSec string `json:"rate_per_sec"`
}

type UserDelReq struct {
	Id string `json:"client_id"`
}
