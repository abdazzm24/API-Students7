package middleware

import (
	"github.com/gofiber/fiber/v2"

	"api-students/helper"
)

// RequirePermission memeriksa apakah user memiliki permission tertentu.
//
// Middleware ini harus dipasang setelah RequireAuth.
//
// Urutan:
//
// RequireAuth
//     ↓
// RequirePermission
func RequirePermission(
	permissions *helper.PermissionSet,
	requiredPermission string,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentUser, exists := helper.CurrentUser(c)

		if !exists {
			return helper.Unauthorized("belum terautentikasi")
		}

		allowed := permissions.Can(
			currentUser.Role,
			requiredPermission,
		)

		if !allowed {
			return helper.Forbidden("tidak berhak mengakses resource ini")
		}

		return c.Next()
	}
}

func RequireRole(
	allowedRoles ...string,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentUser, exists := helper.CurrentUser(c)

		if !exists {
			return helper.Unauthorized("belum terautentikasi")
		}

		for _, role := range allowedRoles {
			if currentUser.Role == role {
				return c.Next()
			}
		}

		return helper.Forbidden("role tidak berhak mengakses resource ini")
	}
}