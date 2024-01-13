package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
)

type contextKey string

const (
	Route contextKey = "route"
	Method contextKey = "method"
)


type currency string

const (
	BTC currency = "btcusdt@trade"
	ETH currency = "ethusdt@trade"
	SOL currency = "solusdt@trade"
)

type state string

const (
	Created   state = "created"
	Triggered state = "triggered"
	Deleted   state = "deleted"
	Completed state = "completed"
)


// for auth service
type SignUpUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=7"`
}

type SignUpUserResponse struct {
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginUserRequest struct {
	UserID   int64  `json:"user_id" validate:"required,number,min=1"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=7"`
}

type LoginUserResponse struct {
	AccessToken          string             `json:"access_token"`
	AccessTokenExpiresAt time.Time          `json:"access_token_expires_at"`
	User                 SignUpUserResponse `json:"user"`
}

// for alert service
type CreateAlertRequest struct {
	UserID    int64   `json:"user_id" validate:"required,number,min=1"`
	Currency  string  `json:"currency" validate:"required,oneof=btcusdt@trade ethusdt@trade solusdt@trade"`
	Price     float64 `json:"price" validate:"required,number,min=0"`
	Direction bool    `json:"direction" validate:"required"`
}

type ReadAllAlertsRequest struct {
	UserID int64 `json:"user_id" validate:"required,number,min=1"`
	Limit  int32 `json:"limit" validate:"required,number,min=1,max=100"`
	Offset int32 `json:"offset" validate:"min=0"`
}

type ReadFilerRequest struct {
	UserID int64  `json:"user_id" validate:"required,number,min=1"`
	Status string `json:"status" validate:"required,oneof=created triggered deleted completed"`
	Limit  int32  `json:"limit" validate:"required,number,min=1,max=100"`
	Offset int32  `json:"offset" validate:"min=0"`
}

type UpdateAlertRequest struct {
	AlertID   int64   `json:"alert_id" validate:"required,number,min=1"`
	UserID    int64   `json:"user_id" validate:"required,number,min=1"`
	Currency  string  `json:"currency" validate:"required,oneof=btcusdt@trade ethusdt@trade solusdt@trade"`
	Price     float64 `json:"price" validate:"required,number,min=0"`
	Direction bool    `json:"direction"`
}

type DeleteAlertRequest struct {
	AlertID int64 `json:"alert_id" validate:"required,number,min=1"`
	UserID  int64 `json:"user_id" validate:"required,number,min=1"`
}

var (
	ErrTokenExpired        = errors.New("token has expired")
	ErrInvalidToken        = errors.New("token is invalid")
	ErrNoAuthHeader        = errors.New("no authorization header")
	ErrInvalidAuthHeader   = errors.New("invalid authorization header")
	ErrUnsupportedAuthType = errors.New("unsupported authorization type")
	ErrInvalidKeySize      = fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	ErrBadRequest          = errors.New("bad request")
	ErrNotAuthorized       = errors.New("not authorized")
	ErrSubscriptionFailed  = errors.New("subscription failed")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrDuplicateAlert      = errors.New("duplicate alert")
	ErrAlertNotFound	   = errors.New("alert not found")
)

type ErrValidation struct {
	Err error
}

func NewErrValidation(err error) *ErrValidation {
	return &ErrValidation{Err: err}
}

func (e *ErrValidation) Error() string {
	return e.Err.Error()
}
