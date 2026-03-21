package constants

import (
	"errors"

	"gorm.io/gorm"
)

// error messages
const (
	ErrInvalidActiontype = "invalid action type"
	ErrScripDoesnotExist = "scrip does not exist"
	ErrUserNotFound      = "user does not exist"
	ErrDatabaseQuery     = "database query error"
)

// Request Validation Errors
const (
	ErrInvalidPayload               = "invalid required payload"
	ErrUnexpectedValue              = "unexpected value for the field."
	ErrFailedToAddScripToWatchlists = "failed to add scrip to watchlists"
	ErrScripNotInWatchlists         = "scrip does not exist in any watchlists of the user"
)

// errors
var (
	InternalServerError       = "internal server error"
	DatabaseQueryError        = errors.New(ErrDatabaseQuery)
	ScripNotFoundError        = errors.New("scrip not found")
	ScripNotInWatchlistsError = errors.New(ErrScripNotInWatchlists)
	UserNotFoundError         = gorm.ErrRecordNotFound
	InvalidActionTypeError    = errors.New(ErrInvalidActiontype)
)

const (
	ErrRequestFailed = "%s request failed"
)
