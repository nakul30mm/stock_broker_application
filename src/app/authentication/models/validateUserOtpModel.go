package models

type BFFValidateUserOtpRequest struct {
	Username   string `json:"username" example:"Arijit" validate:"required,min=5,max=32"`
	Otp        string `json:"otp" validate:"required,otp"`
	DeviceType string `json:"device_type" example:"web" validate:"required,oneof=web mobile"`
}

type BFFValidateUserOtpResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"accessToken"`
}
