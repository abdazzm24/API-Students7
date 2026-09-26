package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(
		ctx context.Context,
		user model.User,
	) (model.User, error)

	List(
		ctx context.Context,
	) ([]model.User, error)

	FindAfterCursor(
		ctx context.Context,
		q model.CursorQuery,
	) ([]model.User, error)

	FindByID(
		ctx context.Context,
		id int,
	) (model.User, error)

	FindByUsername(
		ctx context.Context,
		username string,
	) (model.User, error)

	Replace(
		ctx context.Context,
		id int,
		req model.ReplaceUserRequest,
	) (model.User, error)

	Patch(
		ctx context.Context,
		id int,
		req model.PatchUserRequest,
	) (model.User, error)

	Delete(
		ctx context.Context,
		id int,
	) error

	UpdateRole(
		ctx context.Context,
		id int,
		role string,
	) (model.User, error)
}

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(
	pool *pgxpool.Pool,
) UserRepository {
	return &userPostgresRepository{
		pool: pool,
	}
}

const userColumns = `
	id,
	username,
	email,
	password,
	role,
	is_active,
	created_at
`

func scanUser(row pgx.Row) (model.User, error) {

	var user model.User

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	return user, err
}

func (r *userPostgresRepository) Create(
	ctx context.Context,
	user model.User,
) (model.User, error) {

	result, err := scanUser(
		r.pool.QueryRow(
			ctx,
			`
			INSERT INTO users (
				username,
				email,
				password,
				role,
				is_active
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING `+userColumns,
			user.Username,
			user.Email,
			user.Password,
			user.Role,
			user.IsActive,
		),
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {
			return model.User{}, ErrDuplicate
		}

		return model.User{}, fmt.Errorf(
			"membuat user: %w",
			err,
		)
	}

	return result, nil
}

func (r *userPostgresRepository) List(
	ctx context.Context,
) ([]model.User, error) {

	rows, err := r.pool.Query(
		ctx,
		`
		SELECT `+userColumns+`
		FROM users
		ORDER BY id ASC
		`,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"mengambil daftar user: %w",
			err,
		)
	}

	defer rows.Close()

	users := make([]model.User, 0)

	for rows.Next() {

		var user model.User

		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan daftar user: %w",
				err,
			)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userPostgresRepository) FindAfterCursor(
	ctx context.Context,
	q model.CursorQuery,
) ([]model.User, error) {
	args := []any{}
	where := " WHERE 1 = 1"

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND username ILIKE $%d", len(args))
	}
	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		where += fmt.Sprintf(" AND is_active = $%d", len(args))
	}
	if q.After != nil {
		args = append(args, q.After.CreatedAt, q.After.ID)
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)",
			len(args)-1, len(args))
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf(
		"SELECT %s FROM users%s ORDER BY created_at DESC, id DESC LIMIT $%d",
		userColumns, where, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar user: %w", err)
	}
	defer rows.Close()

	result := []model.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca row user: %w", err)
		}
		result = append(result, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query: %w", err)
	}
	return result, nil
}

func (r *userPostgresRepository) FindByID(
	ctx context.Context,
	id int,
) (model.User, error) {

	user, err := scanUser(
		r.pool.QueryRow(
			ctx,
			`
			SELECT `+userColumns+`
			FROM users
			WHERE id = $1
			`,
			id,
		),
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}

		return model.User{}, fmt.Errorf(
			"mencari user berdasarkan id: %w",
			err,
		)
	}

	return user, nil
}

func (r *userPostgresRepository) FindByUsername(
	ctx context.Context,
	username string,
) (model.User, error) {

	user, err := scanUser(
		r.pool.QueryRow(
			ctx,
			`
			SELECT `+userColumns+`
			FROM users
			WHERE LOWER(username) = LOWER($1)
			`,
			username,
		),
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}

		return model.User{}, fmt.Errorf(
			"mencari user berdasarkan username: %w",
			err,
		)
	}

	return user, nil
}

func (r *userPostgresRepository) Replace(
	ctx context.Context,
	id int,
	req model.ReplaceUserRequest,
) (model.User, error) {

	user, err := scanUser(
		r.pool.QueryRow(
			ctx,
			`
			UPDATE users
			SET
				username = $1,
				email = $2,
				is_active = $3
			WHERE id = $4
			RETURNING `+userColumns,
			req.Username,
			req.Email,
			req.IsActive,
			id,
		),
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {
			return model.User{}, ErrDuplicate
		}

		return model.User{}, fmt.Errorf(
			"mengganti data user: %w",
			err,
		)
	}

	return user, nil
}

func (r *userPostgresRepository) Patch(
	ctx context.Context,
	id int,
	req model.PatchUserRequest,
) (model.User, error) {

	var sets []string
	var args []interface{}

	argNumber := 1

	if req.Username != nil {
		sets = append(
			sets,
			fmt.Sprintf(
				"username = $%d",
				argNumber,
			),
		)

		args = append(args, *req.Username)
		argNumber++
	}

	if req.Email != nil {
		sets = append(
			sets,
			fmt.Sprintf(
				"email = $%d",
				argNumber,
			),
		)

		args = append(args, *req.Email)
		argNumber++
	}

	if req.IsActive != nil {
		sets = append(
			sets,
			fmt.Sprintf(
				"is_active = $%d",
				argNumber,
			),
		)

		args = append(args, *req.IsActive)
		argNumber++
	}

	if len(sets) == 0 {
		return r.FindByID(ctx, id)
	}

	args = append(args, id)

	user, err := scanUser(
		r.pool.QueryRow(
			ctx,
			fmt.Sprintf(
				`
				UPDATE users
				SET %s
				WHERE id = $%d
				RETURNING `+userColumns,
				strings.Join(sets, ", "),
				argNumber,
			),
			args...,
		),
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {
			return model.User{}, ErrDuplicate
		}

		return model.User{}, fmt.Errorf(
			"memperbarui user: %w",
			err,
		)
	}

	return user, nil
}

func (r *userPostgresRepository) Delete(
	ctx context.Context,
	id int,
) error {

	tag, err := r.pool.Exec(
		ctx,
		`
		DELETE FROM users
		WHERE id = $1
		`,
		id,
	)

	if err != nil {
		return fmt.Errorf(
			"menghapus user: %w",
			err,
		)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *userPostgresRepository) UpdateRole(
	ctx context.Context,
	id int,
	role string,
) (model.User, error) {

	user, err := scanUser(
		r.pool.QueryRow(
			ctx,
			`
			UPDATE users
			SET role = $1
			WHERE id = $2
			RETURNING `+userColumns,
			role,
			id,
		),
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}

		return model.User{}, fmt.Errorf(
			"mengubah role user: %w",
			err,
		)
	}

	return user, nil
}