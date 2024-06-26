package rest

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"go-auth/domain"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	validator "gopkg.in/go-playground/validator.v9"
)

type ResponseError struct {
	Message string `json:"message"`
}

type UserService interface {
	Fetch(ctx context.Context, page int64, num int64) ([]domain.User, int64, error)
	GetById(ctx context.Context, id int64) (domain.User, error)
	Store(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
	Delete(ctx context.Context, id int64) error
}

type AuthService interface {
	Login(ctx context.Context, lu *domain.LoginUser) (string, error)
	VerifyToken(ctx context.Context, jwtToken string) error
	CheckTokenInRedis(ctx context.Context, jwtToken string) (bool, error)
}

type RestHandler struct {
	UserService UserService
	AuthService AuthService
}

const defaultNum = 10

func NewRestHandler(r *gin.Engine, userService UserService, authService AuthService) {
	handler := RestHandler{
		UserService: userService,
		AuthService: authService,
	}

	r.GET("/users", handler.FetchUser)
	r.GET("/users/detail", handler.GetById)
	r.POST("/users", handler.Store)
	r.PUT("/users", handler.Update)
	r.DELETE("/users", handler.Delete)

	r.POST("/login", handler.Login)
	r.GET("/verify", handler.Verify)
}

// User service functions

func (u *RestHandler) FetchUser(c *gin.Context) {

	nums := c.Query("num")
	num, err := strconv.Atoi(nums)
	if err != nil || num == 0 {
		num = defaultNum
	}

	// cursor := c.Query("cursor")
	page := c.Query("page")
	page_num, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	ctx := c.Request.Context()

	listUser, nextPage, err := u.UserService.Fetch(ctx, int64(page_num), int64(num))
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	next_page_num := strconv.FormatInt(nextPage, 10)

	c.Header(`X-Cursor`, next_page_num)
	c.JSON(http.StatusOK, listUser)
}

func (u *RestHandler) GetById(c *gin.Context) {
	idP, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	id := int64(idP)
	ctx := c.Request.Context()

	user, err := u.UserService.GetById(ctx, id)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	c.JSON(http.StatusOK, user)
}

func isRequestValid(m *domain.User) (bool, error) {
	validate := validator.New()
	err := validate.Struct(m)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (u *RestHandler) Store(c *gin.Context) {
	var user domain.User
	if err := c.Bind(&user); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	ok, err := isRequestValid(&user)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ctx := c.Request.Context()
	if err := u.UserService.Store(ctx, &user); err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)

}

func (u *RestHandler) Update(c *gin.Context) {
	var user domain.User
	if err := c.Bind(&user); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	ok, err := isRequestValid(&user)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errpr": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err = u.UserService.Update(ctx, &user); err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (u *RestHandler) Delete(c *gin.Context) {
	idP, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrNotFound.Error())
	}

	id := int64(idP)
	ctx := c.Request.Context()

	err = u.UserService.Delete(ctx, id)
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Done!"})
}

// Auth service functions
func (u *RestHandler) Login(c *gin.Context) {
	var loginUser domain.LoginUser
	if err := c.Bind(&loginUser); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	token, err := u.AuthService.Login(ctx, &loginUser)
	if err != nil {
		c.JSON(http.StatusForbidden, ResponseError{Message: err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (u *RestHandler) Verify(c *gin.Context) {
	// get authorization from request
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		return
	}

	// get bearer token
	bearerToken := strings.TrimPrefix(authHeader, "Bearer ")
	if bearerToken == authHeader {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token is missing"})
		return
	}

	// verify token
	ctx := c.Request.Context()
	// check whether jwt token in redis first
	existed, err := u.AuthService.CheckTokenInRedis(ctx, bearerToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get token from redis"})
		return
	}
	// return result if jwt token existed in redis
	if existed {
		c.JSON(http.StatusOK, gin.H{"message": existed})
		return
	}
	// if jwt token not in redis, call verify function
	err = u.AuthService.VerifyToken(ctx, bearerToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// token verify successfully
	c.JSON(http.StatusOK, gin.H{"message": true})
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	logrus.Error(err)
	switch err {
	case domain.ErrInternalServerError:
		return http.StatusInternalServerError
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
