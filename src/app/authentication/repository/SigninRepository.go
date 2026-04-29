package repository

import (
	"authentication/commons/constants"
	"context"
	"fmt"
	genericModels "stock_broker_application/src/models"
	"strings"

	"gorm.io/gorm"
)

type SignInRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*genericModels.User, error)
}

type signInRepository struct {
	db *gorm.DB
}

func NewSignInRepository(db *gorm.DB) *signInRepository {
	return &signInRepository{
		db: db,
	}
}

func (repo *signInRepository) GetUserByUsername(ctx context.Context, username string) (*genericModels.User, error) {
	var user genericModels.User

	result := repo.db.WithContext(ctx).Table(constants.UsersTableName).Where(constants.Username, username).First(&user)
	if result.Error != nil {
		fmt.Println("db error: ", result.Error)
		if strings.Contains(result.Error.Error(), gorm.ErrRecordNotFound.Error()) {
			return nil, fmt.Errorf(constants.ErrUserNotFound)
		}
		return nil, fmt.Errorf(constants.DatabaseQueryError)
	}
	return &user, nil
}
