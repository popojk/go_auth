package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"go-auth/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    UserRepository
	redisClient *redis.Client
}

func NewAuthService(u UserRepository, rdb *redis.Client) *AuthService {
	return &AuthService{
		userRepo:    u,
		redisClient: rdb,
	}
}

var secretKey = []byte(os.Getenv("SECRET_KEY"))

func (a *AuthService) Login(ctx context.Context, u *domain.LoginUser) (jwt string, err error) {
	// get user by username
	user, err := a.userRepo.GetByUsername(ctx, u.Username)
	if err != nil {
		return "", err
	}

	// authenticate password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		return "", err
	}

	// authenticate success, sign JWT token
	token, err := createToken(u.Username)
	if err != nil {
		return "", err
	}
	// save the token into redis
	err = a.redisClient.Set(ctx, token, "true", 24*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store token in Redis: %v", err)
	}
	// return token
	return token, nil
}

func (a *AuthService) VerifyToken(ctx context.Context, tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	// save JWT token into redis
	err = a.redisClient.Set(ctx, tokenString, "true", 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to store token in Redis: %v", err)
	}
	return nil
}

func (a *AuthService) CheckTokenInRedis(ctx context.Context, tokenString string) (bool, error) {
	val, err := a.redisClient.Get(ctx, tokenString).Result()
	if err == redis.Nil {
		// Key does not exist
		return false, nil
	} else if err != nil {
		// Other error
		return false, err
	}

	// Key exists
	return val == "true", nil
}

func createToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
