package helper

import (
	"api-students/app/model"

	"github.com/gofiber/fiber/v2"
)

func Success(
	c *fiber.Ctx,
	status int,
	message string,
	data any,
) error {

	return c.Status(status).JSON(
		model.WebResponse{
			Success: true,
			Message: message,
			Data:    data,
		},
	)
}

func SuccessCursor(
	c *fiber.Ctx,
	message string,
	data any,
	meta *model.CursorMeta,
) error {

	return c.Status(fiber.StatusOK).JSON(
		model.WebResponse{
			Success: true,
			Message: message,
			Data:    data,
			Meta:    meta,
		},
	)
}

func Created(
	c *fiber.Ctx,
	message string,
	data any,
	location string,
) error {

	c.Set("Location", location)

	return c.Status(fiber.StatusCreated).JSON(
		model.WebResponse{
			Success: true,
			Message: message,
			Data:    data,
		},
	)
}

func Fail(
	c *fiber.Ctx,
	status int,
	message string,
	errors any,
) error {

	return c.Status(status).JSON(
		model.WebResponse{
			Success: false,
			Message: message,
			Errors:  errors,
		},
	)
}

func FailValidation(
	c *fiber.Ctx,
	errors any,
) error {

	return Fail(
		c,
		fiber.StatusUnprocessableEntity,
		"validasi gagal",
		errors,
	)
}

func NoContent(
	c *fiber.Ctx,
) error {

	return c.SendStatus(
		fiber.StatusNoContent,
	)
}
