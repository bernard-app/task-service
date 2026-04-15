package postgres

import (
	"bernard/internal/domain/entity"
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) CreateUser(ctx context.Context, user entity.User) (*entity.User, error) {
	const op = "usecase.CreateUser"

	query, args, err := sq.
		Insert("users").
		Columns("id", "username", "display_name", "email", "password").
		Values(&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.Password).
		Prefix("RETURNING id, username", "display_name", "email", "password").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: error building query: %s", op, err))
	}

	var result entity.User

	err = s.DB.QueryRow(ctx, query, args...).Scan(&result.ID, &result.Username, &result.DisplayName, &result.Email, &result.Password)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New(fmt.Sprintf("%s: useralready exists: %s", op, err))
		}

		return nil, errors.New(fmt.Sprintf("%s: error creating user: %s", op, err))
	}

	return &result, nil
}

func (s *Storage) UpdateUser(ctx context.Context, user entity.User) (*entity.User, error) {
	const op = "usecase.UpdateUser"

	query, args, err := sq.
		Update("users").
		SetMap(map[string]interface{}{}).
		Where(sq.Eq{"id": user.ID}).
		Prefix("RETURNING id, username, display_name, email, password").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: error building query: %s", op, err))
	}

	var result entity.User
	err = s.DB.QueryRow(ctx, query, args...).Scan(&result.ID, &result.Username, &result.DisplayName, &result.Email, &result.Password)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: error updating user: %s", op, err))
	}

	return &result, nil
}

func (s *Storage) DeleteUser(ctx context.Context, userID uint64) (*entity.User, error) {
	const op = "usecase.DeleteUser"

	query, args, err := sq.
		Delete("users").
		Where(sq.Eq{"id": userID}).
		Prefix("RETURNING id, username", "display_name", "email", "password").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: error building query: %s", op, err))
	}

	var result entity.User

	err = s.DB.QueryRow(ctx, query, args...).Scan(&result.ID, &result.Username, &result.DisplayName, &result.Email, &result.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(fmt.Sprintf("%s: user not found: %s", op, err))
		}

		return nil, errors.New(fmt.Sprintf("%s: error deleting user: %s", op, err))
	}

	return &result, nil
}

func (s *Storage) GetUser(ctx context.Context, userID uint64) (*entity.User, error) {
	const op = "usecase.GetUser"

	query, args, err := sq.
		Select("id", "username", "display_name", "email", "password").
		From("users").
		Where(sq.Eq{"id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: error building query: %s", op, err))
	}

	var user entity.User

	err = s.DB.QueryRow(ctx, query, args...).Scan(&user.ID, user.Username, &user.DisplayName, user.Email, user.Password)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: error getting user: %s", op, err))
	}

	return &user, nil
}

func (s *Storage) GetUserList(ctx context.Context, username string) (*entity.User, error) {}
