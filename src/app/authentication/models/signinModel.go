package models

type BFFSignInRequest struct {
	Username   string `json:"username" example:"Nakul" validate:"required,min=5,max=32" binding:"required"`
	Password   string `json:"password" example:"Admin@123" validate:"required,min=8,strongPassword,max=20" binding:"required"`
	DeviceType string `json:"device_type" example:"web" validate:"required,oneof=web mobile" binding:"required"`
}

type BFFSignInResponse struct {
	Message string `json:"message"`
}
