package services

import (
	"context"
	"fmt"
	"time"

	"go-auth/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Fetch(ctx context.Context, page int64, num int64) (res []domain.User, nextPage int64, err error)
	GetById(ctx context.Context, id int64) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
	Store(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
	Delete(ctx context.Context, id int64) error
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

func (u *UserService) Update(ctx context.Context, ur *domain.User) (err error) {
	// Todo. Must check the update user is current user

	// Get current user first
	currentUser, err := u.GetById(ctx, int64(ur.ID))
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	// hash password
	if ur.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(ur.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		ur.Password = string(hashedPassword)
	} else {
		ur.Password = currentUser.Password
	}

	now := time.Now()
	ur.UpdatedAt = &now

	if err := u.userRepo.Update(ctx, ur); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return
}

func (u *UserService) Delete(ctx context.Context, id int64) (err error) {
	// Todo. Must check the update user is current user

	// get current user first
	existedUser, err := u.userRepo.GetById(ctx, id)
	if err != nil {
		return
	}
	if existedUser == (domain.User{}) {
		return domain.ErrNotFound
	}
	if err := u.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return
}
