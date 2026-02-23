package business

import (
	"authentication/models"
	"authentication/repository"
	"context"
	"fmt"
	"stock_broker_application/src/utils"
)

type CreateUserService struct {
	createUserRepository repository.CreateUserRepository // we haev taken a field that is and interface in repo
}

func NewCreateUserService(createUserRepository repository.CreateUserRepository) *CreateUserService {
	return &CreateUserService{
		createUserRepository: createUserRepository, //this is a constructor
	}
}

func (service *CreateUserService) CreateNewUser(ctx context.Context, spanCtx context.Context, bffCreateUserRequest models.BFFCreateUserRequest) error {
	postgresClinet := utils.GetPostgresClient()
	client := postgresClinet.GormDB

	err := service.createUserRepository.CreateNewUser(spanCtx, client, bffCreateUserRequest)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil

}
