package rest

import (
	"context"
	"net/http"
	"strconv"

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
}

type RestHandler struct {
	UserService UserService
}

const defaultNum = 10

func NewRestHandler(r *gin.Engine, userService UserService) {
	handler := RestHandler{
		UserService: userService,
	}

	r.GET("/users", handler.FetchUser)
	r.GET("/users/detail", handler.GetById)
	r.POST("/users", handler.Store)
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
