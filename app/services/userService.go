package services

import (
	"context"

	"go-auth/domain"
)

type UserRepository interface {
	Fetch(ctx context.Context, page int64, num int64) (res []domain.User, nextPage int64, err error)
	GetById(ctx context.Context, id int64) (res domain.User, err error)
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
