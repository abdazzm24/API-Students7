package service

import (
	"errors"
	"strconv"
	"strings"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(
	repo repository.UserRepository,
	perms *helper.PermissionSet,
) *UserService {

	return &UserService{
		repo:  repo,
		perms: perms,
	}
}

// translateError mengubah error milik repository menjadi AppError.
func translateError(err error, entity string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound(entity + " tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("username atau email sudah digunakan")
	default:
		return helper.Internal(err)
	}
}

// ========================================================
// GET /users
// ========================================================

func (s *UserService) List(
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
		return helper.WriteUsersCSV(c, rows)
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar user berhasil diambil", rows, meta)
}

// ========================================================
// GET /users/:id
// ========================================================

func (s *UserService) Get(
	c *fiber.Ctx,
) error {

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if !CanAccessUser(current, id, s.perms, "user:read:any") {
		return helper.Forbidden("tidak berhak mengakses data user lain")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "user")
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"user ditemukan",
		user,
	)
}

// ========================================================
// POST /users
// ========================================================

func (s *UserService) Create(
	c *fiber.Ctx,
) error {

	var req model.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	passwordHash, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Internal(err)
	}

	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: passwordHash,
		Role:     "user",
		IsActive: true,
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		return translateError(err, "user")
	}

	return helper.Created(
		c,
		"user berhasil dibuat",
		created,
		"/api/v1/users/"+strconv.Itoa(created.ID),
	)
}

// ========================================================
// PUT /users/:id
// ========================================================

func (s *UserService) Replace(
	c *fiber.Ctx,
) error {

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Forbidden("tidak berhak mengubah user lain")
	}

	var req model.ReplaceUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.repo.Replace(ctx, id, req)
	if err != nil {
		return translateError(err, "user")
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"user berhasil diperbarui",
		user,
	)
}

// ========================================================
// PATCH /users/:id
// ========================================================

func (s *UserService) Patch(
	c *fiber.Ctx,
) error {

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Forbidden("tidak berhak mengubah user lain")
	}

	var req model.PatchUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyPatch(req) {
		return helper.BadRequest("minimal satu field harus dikirim")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.repo.Patch(ctx, id, req)
	if err != nil {
		return translateError(err, "user")
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"user berhasil diperbarui",
		user,
	)
}

// ========================================================
// DELETE /users/:id
// ========================================================

func (s *UserService) Delete(
	c *fiber.Ctx,
) error {

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if current.UserID == id {
		return helper.Forbidden("tidak boleh menghapus akun sendiri")
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(err, "user")
	}

	return helper.NoContent(c)
}

// ========================================================
// PATCH /users/:id/role
// ========================================================

func (s *UserService) AssignRole(
	c *fiber.Ctx,
) error {

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := ValidateAssignRole(current, id, req, s.perms); len(errs) > 0 {
		return helper.Validation(errs)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	user, err := s.repo.UpdateRole(ctx, id, strings.TrimSpace(req.Role))
	if err != nil {
		return translateError(err, "user")
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"role user berhasil diubah",
		user,
	)
}
