package me

import (
	"time"
	"userland/store/postgres"
)

type UserResponse struct {
	Id uint `json:"id"`
	FullName string `json:"fullname"`
	Location string `json:"location"`
	Bio string `json:"bio"`
	Web string `json:"web"`
	Picture string `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUserResponse(u postgres.User) UserResponse {
	return UserResponse{
		Id: u.Id,
		FullName: u.FullName,
		Location: u.Location,
		Bio: u.Bio,
		Web: u.Web,
		Picture: u.Picture,
		CreatedAt: *u.CreatedAt,
	}
}
