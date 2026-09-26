package service

import (
	"strings"

	"api-students/app/model"
)

// ValidateLogin masih digunakan sementara
// sampai auth_service.go diubah menggunakan
// helper.ValidateStruct().
func ValidateLogin(
	req model.LoginRequest,
) map[string]string {

	errs := map[string]string{}

	if strings.TrimSpace(req.Username) == "" {
		errs["username"] = "wajib diisi"
	}

	if req.Password == "" {
		errs["password"] = "wajib diisi"
	}

	return errs
}

// Fungsi ini sengaja tidak lagi menangani Register.
// Validasi Register sekarang berada pada tag
// model.RegisterRequest.
//
// RegisterRequest:
//
// Username -> required,min,max,username
// Email    -> required,email,max
// Password -> required,max,strongpassword
