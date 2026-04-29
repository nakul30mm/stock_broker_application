package repository

import (
	"authentication/commons/constants"
	"context"
	genericModels "stock_broker_application/src/models"

	"gorm.io/gorm"
)

type ValidateUserOtpRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*genericModels.User, error)
}

type validateUserOtpRepository struct {
	db *gorm.DB
}

func NewValidateUserOtpRepository(db *gorm.DB) *validateUserOtpRepository {
	return &validateUserOtpRepository{
		db: db,
	}
}

// this function takes username from request(from service), fetces the user from db and returns the user
func (repo *validateUserOtpRepository) GetUserByUsername(ctx context.Context, username string) (*genericModels.User, error) {
	var user genericModels.User
	result := repo.db.WithContext(ctx).Table(constants.UsersTableName).Where(constants.Username, username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
