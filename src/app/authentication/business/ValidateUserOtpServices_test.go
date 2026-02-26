package business

import (
	"authentication/commons"
	"authentication/models"
	"authentication/repository"
	"context"
	genericModels "stock_broker_application/src/models"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

// test function for testing user not found error
func TestValidateUserOTP_UserNotFound(t *testing.T) {
	mockRepo := &repository.MockValidateUserOtpRepository{
		User:  nil,
		Error: commons.UserNotFoundError,
	}

	service := NewValidateUserOtpService(mockRepo, nil) //service{repository: mockRepo, db: nil}

	testRequest := models.BFFValidateUserOtpRequest{
		Username: "Aman00",
		Otp:      "1209",
	}

	err := service.ValidateUserOtp(context.Background(), context.Background(), testRequest)
	// if err != commons.UserNotFoundError {
	// t.Errorf("expected %v, but got %v", commons.UserNotFoundError, err)
	// }
	assert.Equal(t, commons.UserNotFoundError, err)
}

// test function for testing incorrect otp error
func TestValidateUserOTP_IncorrectOtp(t *testing.T) {
	mockUser := &genericModels.User{
		Username:     "Aman00",
		OtpSent:      1209,
		OtpExpiresAt: uint64(time.Now().Add(5 * time.Minute).Unix()),
	}

	mockRepo := &repository.MockValidateUserOtpRepository{
		User:  mockUser,
		Error: nil,
	}

	service := NewValidateUserOtpService(mockRepo, nil)

	testRequest := models.BFFValidateUserOtpRequest{
		Username: "Aman00",
		Otp:      "1208",
	}

	err := service.ValidateUserOtp(context.Background(), context.Background(), testRequest)
	// if err != commons.IncorrectOTPError {
	// t.Errorf("expected %v, but got %v", commons.IncorrectOTPError, err)
	// }
	assert.Equal(t, commons.IncorrectOTPError, err)
}

// test function for expired otp
func TestValidateUserOTP_ExpiredOtp(t *testing.T) {
	mockUser := &genericModels.User{
		Username:     "Aman00",
		OtpSent:      1209,
		OtpExpiresAt: uint64(time.Now().Add(-5 * time.Minute).Unix()),
	}

	mockRepo := &repository.MockValidateUserOtpRepository{
		User:  mockUser,
		Error: nil,
	}

	service := NewValidateUserOtpService(mockRepo, nil)

	testRequest := models.BFFValidateUserOtpRequest{
		Username: "Aman00",
		Otp:      "1209",
	}

	err := service.ValidateUserOtp(context.Background(), context.Background(), testRequest)
	assert.Equal(t, commons.OtpExpiredError, err)
}

// test function for happy path request/ successful otp validation
func TestValidateUserOTP_Success(t *testing.T) {
	mockUser := &genericModels.User{
		Username:     "Aman00",
		OtpSent:      1209,
		OtpExpiresAt: uint64(time.Now().Add(5 * time.Minute).Unix()),
	}

	mockRepo := &repository.MockValidateUserOtpRepository{
		User:  mockUser,
		Error: nil,
	}

	service := NewValidateUserOtpService(mockRepo, nil)

	testRequest := models.BFFValidateUserOtpRequest{
		Username: "Aman00",
		Otp:      "1209",
	}

	err := service.ValidateUserOtp(context.Background(), context.Background(), testRequest)
	assert.Equal(t, nil, err)
}
