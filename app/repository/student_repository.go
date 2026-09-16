package repository

import (
	"context"
	"fmt"
	"strings"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRepository interface {
	List(
		ctx context.Context,
		search string,
		isActive *bool,
		sortBy string,
		order string,
		page int,
		limit int,
	) ([]model.Student, int, error)

	FindByID(
		ctx context.Context,
		id int,
	) (*model.Student, error)

	Create(
		ctx context.Context,
		req model.CreateStudentRequest,
		ownerID int,
	) (*model.Student, error)

	Replace(
		ctx context.Context,
		id int,
		req model.ReplaceStudentRequest,
	) (*model.Student, error)

	Patch(
		ctx context.Context,
		id int,
		req model.PatchStudentRequest,
	) (*model.Student, error)

	Delete(
		ctx context.Context,
		id int,
	) (bool, error)
}

type studentRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) StudentRepository {
	return &studentRepositoryImpl{
		DB: db,
	}
}

func (r *studentRepositoryImpl) List(
	ctx context.Context,
	search string,
	isActive *bool,
	sortBy string,
	order string,
	page int,
	limit int,
) ([]model.Student, int, error) {

	var conditions []string
	var args []interface{}

	argNumber := 1

	if search != "" {
		conditions = append(
			conditions,
			fmt.Sprintf(
				"LOWER(name) LIKE LOWER($%d)",
				argNumber,
			),
		)

		args = append(args, "%"+search+"%")
		argNumber++
	}

	if isActive != nil {
		conditions = append(
			conditions,
			fmt.Sprintf(
				"is_active = $%d",
				argNumber,
			),
		)

		args = append(args, *isActive)
		argNumber++
	}

	where := ""

	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM students
		%s
	`, where)

	var total int

	if err := r.DB.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	allowedSort := map[string]string{
		"id":        "id",
		"nim":       "nim",
		"name":      "name",
		"grade":     "grade",
		"is_active": "is_active",
		"owner_id":  "owner_id",
	}

	orderBy, ok := allowedSort[sortBy]

	if !ok {
		orderBy = "id"
	}

	orderDirection := "ASC"

	if strings.ToLower(order) == "desc" {
		orderDirection = "DESC"
	}

	offset := (page - 1) * limit

	query := fmt.Sprintf(`
		SELECT
			id,
			nim,
			name,
			grade,
			is_active,
			owner_id
		FROM students
		%s
		ORDER BY %s %s
		LIMIT $%d
		OFFSET $%d
	`, where, orderBy, orderDirection, argNumber, argNumber+1)

	args = append(args, limit, offset)

	rows, err := r.DB.Query(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	result := make([]model.Student, 0)

	for rows.Next() {
		var student model.Student
		var ownerID *int

		if err := rows.Scan(
			&student.ID,
			&student.NIM,
			&student.Name,
			&student.Grade,
			&student.IsActive,
			&ownerID,
		); err != nil {
			return nil, 0, err
		}

		if ownerID != nil {
			student.OwnerID = *ownerID
		}

		result = append(result, student)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (r *studentRepositoryImpl) FindByID(
	ctx context.Context,
	id int,
) (*model.Student, error) {

	var student model.Student
	var ownerID *int

	err := r.DB.QueryRow(
		ctx,
		`
		SELECT
			id,
			nim,
			name,
			grade,
			is_active,
			owner_id
		FROM students
		WHERE id = $1
		`,
		id,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&ownerID,
	)

	if ownerID != nil {
		student.OwnerID = *ownerID
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &student, nil
}

func (r *studentRepositoryImpl) Create(
	ctx context.Context,
	req model.CreateStudentRequest,
	ownerID int,
) (*model.Student, error) {

	var student model.Student
	var retOwnerID *int

	err := r.DB.QueryRow(
		ctx,
		`
		INSERT INTO students (
			nim,
			name,
			grade,
			is_active,
			owner_id
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			nim,
			name,
			grade,
			is_active,
			owner_id
		`,
		req.NIM,
		req.Name,
		req.Grade,
		req.IsActive,
		ownerID,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&retOwnerID,
	)

	if retOwnerID != nil {
		student.OwnerID = *retOwnerID
	}

	if err != nil {
		return nil, err
	}

	return &student, nil
}

func (r *studentRepositoryImpl) Replace(
	ctx context.Context,
	id int,
	req model.ReplaceStudentRequest,
) (*model.Student, error) {

	var student model.Student
	var ownerID *int

	err := r.DB.QueryRow(
		ctx,
		`
		UPDATE students
		SET
			nim = $1,
			name = $2,
			grade = $3,
			is_active = $4
		WHERE id = $5
		RETURNING
			id,
			nim,
			name,
			grade,
			is_active,
			owner_id
		`,
		req.NIM,
		req.Name,
		req.Grade,
		req.IsActive,
		id,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&ownerID,
	)

	if ownerID != nil {
		student.OwnerID = *ownerID
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &student, nil
}

func (r *studentRepositoryImpl) Patch(
	ctx context.Context,
	id int,
	req model.PatchStudentRequest,
) (*model.Student, error) {

	var sets []string
	var args []interface{}

	argNumber := 1

	if req.NIM != nil {
		sets = append(
			sets,
			fmt.Sprintf("nim = $%d", argNumber),
		)

		args = append(args, *req.NIM)
		argNumber++
	}

	if req.Name != nil {
		sets = append(
			sets,
			fmt.Sprintf("name = $%d", argNumber),
		)

		args = append(args, *req.Name)
		argNumber++
	}

	if req.Grade != nil {
		sets = append(
			sets,
			fmt.Sprintf("grade = $%d", argNumber),
		)

		args = append(args, *req.Grade)
		argNumber++
	}

	if req.IsActive != nil {
		sets = append(
			sets,
			fmt.Sprintf("is_active = $%d", argNumber),
		)

		args = append(args, *req.IsActive)
		argNumber++
	}

	if len(sets) == 0 {
		return r.FindByID(ctx, id)
	}

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE students
		SET %s
		WHERE id = $%d
		RETURNING
			id,
			nim,
			name,
			grade,
			is_active,
			owner_id
	`,
		strings.Join(sets, ", "),
		argNumber,
	)

	var student model.Student
	var ownerID *int

	err := r.DB.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&ownerID,
	)

	if ownerID != nil {
		student.OwnerID = *ownerID
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &student, nil
}

func (r *studentRepositoryImpl) Delete(
	ctx context.Context,
	id int,
) (bool, error) {

	commandTag, err := r.DB.Exec(
		ctx,
		`
		DELETE FROM students
		WHERE id = $1
		`,
		id,
	)

	if err != nil {
		return false, err
	}

	return commandTag.RowsAffected() > 0, nil
}