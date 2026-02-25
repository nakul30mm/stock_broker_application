package utils

import (
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashPassword), nil
}

func CompareHashPassword(hashPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}

func CompareUserRequestOTP(OtpFromDB uint64, OtpFromRequest string) bool {
	parsedOTP, err := strconv.ParseUint(OtpFromRequest, 10, 64)
	if err != nil {
		return false
	}
	return OtpFromDB == parsedOTP
}

func CheckOtpExpiry(otpExpiresAt, requestTime time.Time) bool {
	if requestTime.After(otpExpiresAt) {
		return false
	}
	return true
}
