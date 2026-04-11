package models

type BFFSignInRequest struct {
	Username   string `json:"username" example:"Arijit" validate:"required,min=5,max=32"`
	Password   string `json:"password" example:"Secure@123" validate:"required,min=8,strongPassword,max=20"`
	DeviceType string `json:"device_type" example:"web" validate:"required,oneof=web mobile"`
}

type BFFSignInResponse struct {
	Message string `json:"message"`
}
