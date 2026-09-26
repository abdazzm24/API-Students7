package service

import (
	"strconv"
	"strings"
	"time"

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

	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	rows, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	if format == helper.FormatCSV {
		return helper.WriteStudentsCSV(c, rows)
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(time.Time{}, last.ID)
	}

	return helper.SuccessCursor(c, "daftar student berhasil diambil", rows, meta)
}

// ========================================================
// GET /students/:id
// ========================================================

func (s *StudentService) Get(
	c *fiber.Ctx,
) error {

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return helper.Internal(err)
	}

	if student == nil {
		return helper.NotFound("student tidak ditemukan")
	}

	if !CanAccessStudent(authUser, student.OwnerID, s.perms, "student:read:any") {
		return helper.Forbidden("tidak berhak mengakses student ini")
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
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body bukan JSON yang sah")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.repo.Create(ctx, req, authUser.UserID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "23505") ||
			strings.Contains(err.Error(), "unique") {

			return helper.Conflict("NIM sudah digunakan")
		}

		return helper.Internal(err)
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
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body bukan JSON yang sah")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return helper.Internal(err)
	}

	if student == nil {
		return helper.NotFound("student tidak ditemukan")
	}

	if !CanAccessStudent(authUser, student.OwnerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah student ini")
	}

	updated, err := s.repo.Replace(ctx, id, req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "23505") ||
			strings.Contains(err.Error(), "unique") {

			return helper.Conflict("NIM sudah digunakan")
		}

		return helper.Internal(err)
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
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body bukan JSON yang sah")
	}

	if IsEmptyStudentPatch(req) {
		return helper.BadRequest("minimal satu field harus dikirim")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return helper.Internal(err)
	}

	if student == nil {
		return helper.NotFound("student tidak ditemukan")
	}

	if !CanAccessStudent(authUser, student.OwnerID, s.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah student ini")
	}

	updated, err := s.repo.Patch(ctx, id, req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "23505") ||
			strings.Contains(err.Error(), "unique") {

			return helper.Conflict("NIM sudah digunakan")
		}

		return helper.Internal(err)
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

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return helper.Internal(err)
	}

	if !deleted {
		return helper.NotFound("student tidak ditemukan")
	}

	return helper.NoContent(c)
}

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
