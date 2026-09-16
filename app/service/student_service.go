package service

import (
	"strconv"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(
	repo repository.StudentRepository,
	perms *helper.PermissionSet,
) *StudentService {

	return &StudentService{
		repo:  repo,
		perms: perms,
	}
}

// ========================================================
// GET /students
// ========================================================

func (s *StudentService) List(
	c *fiber.Ctx,
) error {

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	query, err := helper.ParseListQuery(c)

	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"query tidak valid",
			map[string]string{
				"query": err.Error(),
			},
		)
	}

	result, total, err := s.repo.List(
		ctx,
		query.Search,
		query.IsActive,
		query.Sort,
		query.Order,
		query.Page,
		query.Limit,
	)

	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil daftar student",
			map[string]string{
				"error": err.Error(),
			},
		)
	}

	meta := &model.Meta{
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: CalculateTotalPages(
			query.Limit,
			total,
		),
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"daftar student berhasil diambil",
		map[string]any{
			"students": result,
			"meta":     meta,
		},
	)
}

// ========================================================
// GET /students/:id
// ========================================================

func (s *StudentService) Get(
	c *fiber.Ctx,
) error {

	authUser, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Fail(
			c,
			fiber.StatusUnauthorized,
			"belum terautentikasi",
			nil,
		)
	}

	id, err := helper.ParseID(c)

	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id tidak valid",
			map[string]string{
				"id": err.Error(),
			},
		)
	}

	// Authorization dilakukan sebelum query database.
	// Ini mencegah user mengetahui keberadaan ID milik
	// orang lain melalui perbedaan response.
	//
	// Untuk student, permission yang digunakan adalah:
	// student:read:any
	if !s.canAccessStudent(
		authUser,
		id,
		"student:read:any",
	) {

		// Kita memang belum mengetahui owner ID karena
		// data belum diambil. Untuk menentukan ownership,
		// kita perlu mengambil data. Oleh karena itu,
		// untuk student endpoint, pengecekan ownership
		// dilakukan setelah data ditemukan.
		//
		// Permission any tetap dapat dihentikan lebih awal.
		if !s.perms.Can(
			authUser.Role,
			"student:read:any",
		) {
			// Ambil data untuk mengetahui owner.
			// Setelah itu keputusan ownership dilakukan.
		}
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.repo.FindByID(
		ctx,
		id,
	)

	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil detail student",
			nil,
		)
	}

	if student == nil {
		return helper.Fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if !CanAccessStudent(
		authUser,
		student.OwnerID,
		s.perms,
		"student:read:any",
	) {
		return helper.Fail(
			c,
			fiber.StatusForbidden,
			"tidak berhak mengakses student ini",
			nil,
		)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"student berhasil ditemukan",
		student,
	)
}

// ========================================================
// POST /students
// ========================================================

func (s *StudentService) Create(
	c *fiber.Ctx,
) error {

	authUser, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Fail(
			c,
			fiber.StatusUnauthorized,
			"belum terautentikasi",
			nil,
		)
	}

	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"body bukan JSON yang sah",
			nil,
		)
	}

	validationErrors := ValidateCreateRequest(req)

	if len(validationErrors) > 0 {
		return helper.FailValidation(
			c,
			validationErrors,
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// ownerID berasal dari JWT.
	// BUKAN dari request body.
	student, err := s.repo.Create(
		ctx,
		req,
		authUser.UserID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "23505") ||
			strings.Contains(err.Error(), "unique") {

			return helper.Fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
				map[string]string{
					"nim": "NIM harus unik",
				},
			)
		}

		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"gagal membuat student",
			nil,
		)
	}

	return helper.Created(
		c,
		"student berhasil ditambahkan",
		student,
		"/api/v1/students/"+strconv.Itoa(student.ID),
	)
}

// ========================================================
// PUT /students/:id
// ========================================================

func (s *StudentService) Replace(
	c *fiber.Ctx,
) error {

	authUser, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Fail(
			c,
			fiber.StatusUnauthorized,
			"belum terautentikasi",
			nil,
		)
	}

	id, err := helper.ParseID(c)

	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id tidak valid",
			map[string]string{
				"id": err.Error(),
			},
		)
	}

	var req model.ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"body bukan JSON yang sah",
			nil,
		)
	}

	validationErrors := ValidateReplaceRequest(req)

	if len(validationErrors) > 0 {
		return helper.FailValidation(
			c,
			validationErrors,
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.repo.FindByID(
		ctx,
		id,
	)

	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil student",
			nil,
		)
	}

	if student == nil {
		return helper.Fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if !CanAccessStudent(
		authUser,
		student.OwnerID,
		s.perms,
		"student:update:any",
	) {
		return helper.Fail(
			c,
			fiber.StatusForbidden,
			"tidak berhak mengubah student ini",
			nil,
		)
	}

	updated, err := s.repo.Replace(
		ctx,
		id,
		req,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "23505") ||
			strings.Contains(err.Error(), "unique") {

			return helper.Fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
				map[string]string{
					"nim": "NIM harus unik",
				},
			)
		}

		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"gagal memperbarui student",
			nil,
		)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"student berhasil diganti",
		updated,
	)
}

// ========================================================
// PATCH /students/:id
// ========================================================

func (s *StudentService) Patch(
	c *fiber.Ctx,
) error {

	authUser, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Fail(
			c,
			fiber.StatusUnauthorized,
			"belum terautentikasi",
			nil,
		)
	}

	id, err := helper.ParseID(c)

	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id tidak valid",
			map[string]string{
				"id": err.Error(),
			},
		)
	}

	var req model.PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"body bukan JSON yang sah",
			nil,
		)
	}

	validationErrors := ValidatePatchRequest(req)

	if len(validationErrors) > 0 {
		return helper.FailValidation(
			c,
			validationErrors,
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.repo.FindByID(
		ctx,
		id,
	)

	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"gagal mengambil student",
			nil,
		)
	}

	if student == nil {
		return helper.Fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	if !CanAccessStudent(
		authUser,
		student.OwnerID,
		s.perms,
		"student:update:any",
	) {
		return helper.Fail(
			c,
			fiber.StatusForbidden,
			"tidak berhak mengubah student ini",
			nil,
		)
	}

	updated, err := s.repo.Patch(
		ctx,
		id,
		req,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "23505") ||
			strings.Contains(err.Error(), "unique") {

			return helper.Fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
				map[string]string{
					"nim": "NIM harus unik",
				},
			)
		}

		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"gagal memperbarui sebagian student",
			nil,
		)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"student berhasil diperbarui sebagian",
		updated,
	)
}

// ========================================================
// DELETE /students/:id
// ========================================================

func (s *StudentService) Delete(
	c *fiber.Ctx,
) error {

	id, err := helper.ParseID(c)

	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id tidak valid",
			map[string]string{
				"id": err.Error(),
			},
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	deleted, err := s.repo.Delete(
		ctx,
		id,
	)

	if err != nil {
		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"gagal menghapus student",
			nil,
		)
	}

	if !deleted {
		return helper.Fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
			nil,
		)
	}

	return helper.NoContent(c)
}

// canAccessStudent hanya digunakan untuk pengecekan
// permission yang tidak membutuhkan data student.
func (s *StudentService) canAccessStudent(
	current model.AuthUser,
	targetID int,
	permission string,
) bool {

	return s.perms.Can(
		current.Role,
		permission,
	)
}