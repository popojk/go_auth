package services

import (
	"context"

	"go-auth/domain"
)

type UserRepository interface {
	Fetch(ctx context.Context, cursor string, num int64) (res []domain.User, nextCursor string, err error)
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

func (u *UserService) Fetch(ctx context.Context, cursor string, num int64) (res []domain.User, nextCursor string, err error) {
	res, nextCursor, err = u.userRepo.Fetch(ctx, cursor, num)
	if err != nil {
		return nil, "", err
	}

	return
}
