package repository

import (
	"context"
	genericModels "stock_broker_application/src/models"

	"gorm.io/gorm"
)

type MockValidateUserOtpRepository struct {
	Db    *gorm.DB
	User  *genericModels.User
	Error error
}

func (m *MockValidateUserOtpRepository) GetUserByUsername(ctx context.Context, username string) (*genericModels.User, error) {
	return m.User, m.Error
}
