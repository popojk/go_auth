package services

import (
	"context"

	"go-auth/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Fetch(ctx context.Context, page int64, num int64) (res []domain.User, nextPage int64, err error)
	GetById(ctx context.Context, id int64) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
	Store(ctx context.Context, u *domain.User) error
}

// Add repos into service here
type UserService struct {
	userRepo UserRepository
}

// Service constructor
func NewUserService(u UserRepository) *UserService {
	return &UserService{
		userRepo: u,
	}
}

func (u *UserService) Fetch(ctx context.Context, page int64, num int64) (res []domain.User, nextPage int64, err error) {
	res, nextPage, err = u.userRepo.Fetch(ctx, page, num)
	if err != nil {
		return nil, nextPage, err
	}
	return
}

func (u *UserService) GetById(ctx context.Context, id int64) (res domain.User, err error) {
	res, err = u.userRepo.GetById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return
}

func (u *UserService) GetByUsername(ctx context.Context, username string) (res domain.User, err error) {
	res, err = u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return domain.User{}, err
	}
	return
}

func (u *UserService) Store(ctx context.Context, ur *domain.User) (err error) {
	// do check existed user later
	existedUser, _ := u.GetByUsername(ctx, ur.Username)
	if existedUser != (domain.User{}) {
		return domain.ErrConflict
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(ur.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// Stored hashed password in User
	ur.Password = string(hashedPassword)

	err = u.userRepo.Store(ctx, ur)
	return
}
