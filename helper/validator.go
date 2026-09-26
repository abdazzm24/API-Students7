package helper

import (
	"errors"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	// Gunakan nama field dari tag JSON sebagai nama field
	// yang dikembalikan pada error validasi.
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]

		if name == "" || name == "-" {
			return field.Name
		}

		return name
	})

	// Tidak boleh mengandung spasi.
	_ = v.RegisterValidation("nospace", func(fl validator.FieldLevel) bool {
		return !strings.ContainsAny(
			fl.Field().String(),
			" \t\n\r",
		)
	})

	// Username hanya boleh berisi:
	// huruf, angka, titik, dan underscore.
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()

		if value == "" {
			return true
		}

		for _, r := range value {
			if !unicode.IsLetter(r) &&
				!unicode.IsDigit(r) &&
				r != '.' &&
				r != '_' {

				return false
			}
		}

		return true
	})

	// Custom validation untuk NIM.
	// NIM mahasiswa pada project ini harus berupa 9 digit angka.
	_ = v.RegisterValidation("nim", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()

		if len(value) != 9 {
			return false
		}

		for _, r := range value {
			if !unicode.IsDigit(r) {
				return false
			}
		}

		return true
	})

	// String tidak boleh hanya berisi whitespace.
	_ = v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		return strings.TrimSpace(fl.Field().String()) != ""
	})

	// Custom validation password.
	_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		return checkPasswordStrength(fl.Field().String()) == ""
	})

	return v
}

// ValidateStruct menjalankan seluruh validation tag pada struct.
func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)

	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError

	if errors.As(err, &invalid) {
		return map[string]string{
			"_": "objek yang divalidasi tidak sah",
		}
	}

	var fieldErrors validator.ValidationErrors

	if !errors.As(err, &fieldErrors) {
		return map[string]string{
			"_": "validasi gagal",
		}
	}

	result := make(map[string]string, len(fieldErrors))

	for _, fe := range fieldErrors {
		// Satu field cukup memiliki satu pesan.
		if _, exists := result[fe.Field()]; exists {
			continue
		}

		result[fe.Field()] = messageFor(fe)
	}

	return result
}

// messageFor menerjemahkan validation tag menjadi
// pesan yang lebih mudah dibaca oleh client.
func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {

	case "required":
		return "wajib diisi"

	case "email":
		return "format email tidak valid"

	case "min":
		if fe.Kind() == reflect.String {
			return "minimal " + fe.Param() + " karakter"
		}

		return "nilai minimal " + fe.Param()

	case "max":
		if fe.Kind() == reflect.String {
			return "maksimal " + fe.Param() + " karakter"
		}

		return "nilai maksimal " + fe.Param()

	case "alphanum":
		return "hanya boleh berisi huruf dan angka"

	case "nospace":
		return "tidak boleh mengandung spasi"

	case "username":
		return "hanya boleh huruf, angka, titik, dan garis bawah"

	case "nim":
		return "NIM harus terdiri dari 9 digit angka"

	case "notblank":
		return "tidak boleh hanya berisi spasi"

	case "strongpassword":
		if value, ok := fe.Value().(string); ok {
			return checkPasswordStrength(value)
		}
		return "password tidak memenuhi syarat"

	case "omitnil":
		return "tidak memenuhi aturan"

	case "oneof":
		return "harus salah satu dari: " +
			strings.ReplaceAll(fe.Param(), " ", ", ")

	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}

// checkPasswordStrength berasal dari aturan password
// yang sebelumnya berada di app/service/auth_rules.go.
//
// Fungsi ini tetap digunakan sebagai fungsi biasa karena
// strongpassword membutuhkan pesan error yang lebih spesifik.
func checkPasswordStrength(password string) string {
	const minPasswordLength = 8

	if len(password) < minPasswordLength {
		return "minimal 8 karakter"
	}

	var hasLetter bool
	var hasDigit bool

	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true

		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}

	weak := map[string]bool{
		"password1":   true,
		"12345678":    true,
		"qwerty123":   true,
		"admin123":    true,
		"password123": true,
	}

	if weak[strings.ToLower(password)] {
		return "password terlalu umum"
	}

	return ""
}
