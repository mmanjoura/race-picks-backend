package models

import "time"

type SignIn struct {
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type SignUp struct {
	ID          int    `json:"id"`
	FullName    string `json:"full_name" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	PhoneNumber string `json:"phone_number"`
}

type User struct {
	ID          int    `json:"id"`
	FullName    string `json:"full_name" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	PhoneNumber string `json:"phone_number"`
	UserType    string `json:"user_type"`
	Profile     string `json:"profile"`
	AvatarUrl  string `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
