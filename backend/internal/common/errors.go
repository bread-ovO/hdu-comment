package common

import "errors"

var (
	// ErrEmailAlreadyUsed indicates the email has been registered.
	ErrEmailAlreadyUsed = errors.New("email already in use")
	// ErrInvalidCredentials indicates login failure.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrInvalidPhoneNumber indicates an invalid phone format.
	ErrInvalidPhoneNumber = errors.New("invalid phone number")
	// ErrInvalidSMSCode indicates sms verification failure.
	ErrInvalidSMSCode = errors.New("invalid sms code")
	// ErrSMSServiceUnavailable indicates sms auth is disabled or not configured.
	ErrSMSServiceUnavailable = errors.New("sms login unavailable")
	// ErrQQServiceUnavailable indicates QQ OAuth is disabled or not configured.
	ErrQQServiceUnavailable = errors.New("qq login unavailable")
	// ErrInvalidQQState indicates the QQ OAuth state is invalid.
	ErrInvalidQQState = errors.New("invalid qq login state")
	// ErrReviewAlreadyProcessed indicates review status update conflict.
	ErrReviewAlreadyProcessed = errors.New("review already processed")
	// ErrInvalidRefreshToken indicates the provided refresh token is invalid or expired.
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)
