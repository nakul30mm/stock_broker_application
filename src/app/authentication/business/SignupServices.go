package business

import (
	"authentication/models"
	"authentication/repository"
	"context"
	"fmt"
	genericErrors "stock_broker_application/src/constants"

	"gorm.io/gorm"
)

type CreateUserService struct {
	createUserRepository repository.CreateUserRepository
	DB                   *gorm.DB
}

func NewCreateUserService(createUserRepository repository.CreateUserRepository, db *gorm.DB) *CreateUserService {
	return &CreateUserService{
		createUserRepository: createUserRepository,
		DB:                   db,
	}
}

func (service *CreateUserService) CreateNewUser(spanCtx context.Context, bffCreateUserRequest models.BFFCreateUserRequest) error {
	tx := service.DB.Begin()

	if tx.Error != nil {
		return fmt.Errorf(genericErrors.ErrBeginTx, tx.Error)
	}

	err := service.createUserRepository.CreateNewUser(spanCtx, tx, bffCreateUserRequest)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf(genericErrors.ErrCommitTx, err)
	}

	return nil
}
