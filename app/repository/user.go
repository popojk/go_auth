package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"

	"go-auth/domain"
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

func (m *UserRepository) Fetch(ctx context.Context, page int64, num int64) (res []domain.User, nextPage int64, err error) {
	query := `SELECT id, username, password, avatar, updated_at, created_at
	         FROM public.user ORDER BY created_at LIMIT $1 OFFSET ($2 - 1) * $1`
	// decodedCursor, err := tools.DecodeCursor(cursor)
	// if err != nil && cursor != "" {
	// 	fmt.Println("error here")
	// 	return nil, "", domain.ErrBadParamInput
	// }

	res, err = m.fetch(ctx, query, num, page)
	if err != nil {
		return nil, 0, err
	}

	if len(res) == int(num) {
		nextPage = page + 1
	}

	return
}

func (m *UserRepository) GetById(ctx context.Context, id int64) (res domain.User, err error) {
	query := `SELECT id, username, password, avatar, updated_at, created_at
	          FROM public.user WHERE id = $1`
	list, err := m.fetch(ctx, query, id)
	if err != nil {
		return domain.User{}, err
	}

	if len(list) > 0 {
		res = list[0]
	} else {
		return res, domain.ErrNotFound
	}

	return
}

func (m *UserRepository) GetByUsername(ctx context.Context, username string) (res domain.User, err error) {
	query := `SELECT id, username, password, avatar, updated_at, created_at
	          From public.user WHERE username = $1`
	list, err := m.fetch(ctx, query, username)
	if err != nil {
		return domain.User{}, err
	}

	if len(list) > 0 {
		res = list[0]
	} else {
		return res, domain.ErrNotFound
	}
	return
}

func (m *UserRepository) Store(ctx context.Context, u *domain.User) (err error) {
	query := `INSERT INTO public.user (username, password, avatar) VALUES ($1, $2, '')`
	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	_, err = stmt.ExecContext(ctx, u.Username, u.Password)
	if err != nil {
		return
	}
	return
}

func (m *UserRepository) Update(ctx context.Context, u *domain.User) (err error) {
	query := `UPDATE public.user SET`
	args := []interface{}{}
	argsCounter := 1

	if u.Username != "" {
		query += fmt.Sprintf(" username = $%d,", argsCounter)
		args = append(args, u.Username)
		argsCounter++
	}
	if u.Password != "" {
		query += fmt.Sprintf(" password = $%d,", argsCounter)
		args = append(args, u.Password)
		argsCounter++
	}

	query += fmt.Sprintf(" updated_at = $%d WHERE id = $%d", argsCounter, argsCounter+1)
	args = append(args, u.UpdatedAt, u.ID)

	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return
	}

	affect, err := res.RowsAffected()
	if err != nil {
		return
	}

	if affect != 1 {
		err = fmt.Errorf("weird behavior. Total affected: %d", affect)
		return
	}
	return
}

func (u *UserRepository) Delete(ctx context.Context, id int64) (err error) {
	query := `DELETE FROM public.user WHERE id = $1`

	stmt, err := u.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	res, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return
	}

	rowsAffect, err := res.RowsAffected()
	if err != 1 {
		err = fmt.Errorf("wierd behavior, Total Affected: %d", rowsAffect)
		return
	}

	return

}
