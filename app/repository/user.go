package repository

import (
	"context"
	"database/sql"

	"github.com/sirupsen/logrus"

	"go-auth/domain"
	"go-auth/tools"
)

type UserRepository struct {
	Conn *sql.DB
}

func NewUserRepository(conn *sql.DB) *UserRepository {
	return &UserRepository{conn}
}

func (m *UserRepository) fetch(ctx context.Context, query string, args ...interface{}) (result []domain.User, err error) {
	rows, err := m.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer func() {
		errRow := rows.Close()
		if errRow != nil {
			logrus.Error(errRow)
		}
	}()

	result = make([]domain.User, 0)
	for rows.Next() {
		t := domain.User{}
		err = rows.Scan(
			&t.ID,
			&t.Username,
			&t.Password,
			&t.Avatar,
			&t.UpdatedAt,
			&t.CreatedAt,
		)

		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

func (m *UserRepository) Fetch(ctx context.Context, cursor string, num int64) (res []domain.User, nextCursor string, err error) {
	query := `SELECT id, username, password, avatar, updated_at, created_at
	         FROM user ORDER BY created_at`
	decodedCursor, err := tools.DecodeCursor(cursor)
	if err != nil && cursor != "" {
		return nil, "", domain.ErrBadParamInput
	}

	res, err = m.fetch(ctx, query, decodedCursor, num)
	if err != nil {
		return nil, "", err
	}

	if len(res) == int(num) {
		nextCursor = tools.EncodeCursor(res[len(res)-1].CreatedAt)
	}

	return
}
