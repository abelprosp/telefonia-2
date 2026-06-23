package httputil

import (
	"net/http"

	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

type ErrorKind int

const (
	ErrorValidation ErrorKind = iota
	ErrorNotFound
	ErrorBusiness
	ErrorForbidden
	ErrorInternal
)

type AppError struct {
	Kind         ErrorKind
	Notifications []notifications.Notification
}

func ValidationError(n ...notifications.Notification) *AppError {
	return &AppError{Kind: ErrorValidation, Notifications: n}
}

func NotFoundError(n notifications.Notification) *AppError {
	return &AppError{Kind: ErrorNotFound, Notifications: []notifications.Notification{n}}
}

func BusinessError(n ...notifications.Notification) *AppError {
	return &AppError{Kind: ErrorBusiness, Notifications: n}
}

func ForbiddenError(n notifications.Notification) *AppError {
	return &AppError{Kind: ErrorForbidden, Notifications: []notifications.Notification{n}}
}

func InternalError(n notifications.Notification) *AppError {
	return &AppError{Kind: ErrorInternal, Notifications: []notifications.Notification{n}}
}

func (e *AppError) StatusCode() int {
	switch e.Kind {
	case ErrorValidation:
		return http.StatusBadRequest
	case ErrorNotFound:
		return http.StatusNotFound
	case ErrorBusiness:
		return http.StatusConflict
	case ErrorForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func (e *AppError) Error() string {
	if len(e.Notifications) == 0 {
		return "unknown error"
	}
	return e.Notifications[0].Message
}
