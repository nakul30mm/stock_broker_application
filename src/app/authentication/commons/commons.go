package commons

import (
	"authentication/commons/constants"
	"errors"
)

// Add your common functionalities here.

var ErrUserNotFound = errors.New(constants.ErrUserNotFound)
var ErrIncorrectPassword = errors.New(constants.ErrIncorrectPassword)

// constants for returning keys
const (
	Username = "username"
	Password = "password"
)
