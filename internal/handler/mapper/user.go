package mapper

import (
	"cloud/internal/handler/dto"
	"cloud/internal/model"
	"log/slog"
	"strconv"
)

func AddReqToUser(req dto.UserAddReq) model.User {
	capacity, err := strconv.Atoi(req.Capacity)
	rate, err := strconv.Atoi(req.RatePerSec)
	if err != nil {
		slog.Error("Failed to mapping user struct", "error", err)
		return model.User{}
	}
	return model.User{
		Id:         req.Id,
		Capacity:   capacity,
		RatePerSec: rate,
	}
}

func DelReqToUser(req dto.UserDelReq) model.User {
	return model.User{
		Id: req.Id,
	}
}
