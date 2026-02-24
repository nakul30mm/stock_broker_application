package commons

import (
	"authentication/commons/constants"
	"errors"
)

// Add your common functionalities here.

var UserNotFoundError = errors.New(constants.ErrUserNotFound)
var IncorrectPasswordError = errors.New(constants.ErrIncorrectPassword)

// constants for returning keys
const (
	Username = "username"
	Password = "password"
)
