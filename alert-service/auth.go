package main

import (
	"context"
	"fmt"
	"time"

	database "alert-service/database/sqlc"

	"golang.org/x/crypto/bcrypt"
)

// mein yaha peh hee jwt aur login signup banara hun maa chuday

type Auther interface {
	SignUp(ctx context.Context, req SignUpUserRequest) (SignUpUserResponse, error)
	Login(ctx context.Context, req LoginUserRequest) (LoginUserResponse, error)
}

type auther struct {
	db       database.Querier
	token    Maker
	tokenExp time.Duration
}

func NewAuthSvc(db database.Querier, token Maker, tokenExp time.Duration) Auther {
	return &auther{
		db:       db,
		token:    token,
		tokenExp: tokenExp,
	}
}

func (a *auther) SignUp(ctx context.Context, req SignUpUserRequest) (SignUpUserResponse, error) {
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return SignUpUserResponse{}, err
	}

	createUserParams := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}

	res, err := a.db.CreateUser(ctx, createUserParams)
	if err != nil {
		return SignUpUserResponse{}, ErrUserAlreadyExists
	}

	return SignUpUserResponse{
		UserID:    res.ID,
		CreatedAt: res.CreatedAt,
	}, nil
}

func (a *auther) Login(ctx context.Context, req LoginUserRequest) (LoginUserResponse, error) {
	user, err := a.db.GetUserById(ctx, req.UserID)
	if err != nil {
		return LoginUserResponse{}, err
	}

	err = checkPassword(req.Password, user.HashedPassword)
	if err != nil {
		return LoginUserResponse{}, ErrNotAuthorized
	}

	accessToken, accessPayload, err := a.token.Create(
		user.ID,
		a.tokenExp,
	)
	if err != nil {
		return LoginUserResponse{}, err
	}

	return LoginUserResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
		User:                 SignUpUserResponse{UserID: user.ID, CreatedAt: user.CreatedAt},
	}, nil
}

// hashPassword returns the bcrypt hash of the password
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// checkPassword checks if the provided password is correct or not
func checkPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
