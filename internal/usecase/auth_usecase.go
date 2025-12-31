package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourname/stockbit-appsec/internal/domain"
	"github.com/yourname/stockbit-appsec/internal/repository"
	"github.com/yourname/stockbit-appsec/pkg/utils"
)

type AuthUsecase struct {
	UserRepo *repository.UserRepository
	SecretKey string
}

func NewAuthUsecase(userRepo *repository.UserRepository, secretKey string) *AuthUsecase {
	return &AuthUsecase{
		UserRepo:  userRepo,
		SecretKey: secretKey,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, email, password, role string) error {
	existingUser, err := u.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("email already registered")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	newUser := &domain.User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	return u.UserRepo.CreateUser(ctx, newUser)
}

func (u *AuthUsecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(u.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
