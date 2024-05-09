package rest

import (
	"context"
	"net/http"
	"strconv"

	"go-auth/domain"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ResponseError struct {
	Message string `json:"message"`
}

type UserService interface {
	Fetch(ctx context.Context, cursor string, num int64) ([]domain.User, string, error)
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
}

// User service functions

func (u *RestHandler) FetchUser(c *gin.Context) {

	nums := c.Query("num")
	num, err := strconv.Atoi(nums)
	if err != nil || num == 0 {
		num = defaultNum
	}

	cursor := c.Query("cursor")
	ctx := c.Request.Context()

	listUser, nextCursor, err := u.UserService.Fetch(ctx, cursor, int64(num))
	if err != nil {
		c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	c.Header(`X-Cursor`, nextCursor)
	c.JSON(http.StatusOK, listUser)
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
