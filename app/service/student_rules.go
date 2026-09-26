package service

import (
	"strings"

	"api-students/app/model"
)

// ApplyPatch menerapkan perubahan dari PATCH
// ke data Student yang sedang ada.
//
// Validasi bentuk request dilakukan oleh validator.
// Fungsi ini hanya bertugas menggabungkan perubahan.
func ApplyStudentPatch(
	current model.Student,
	req model.PatchStudentRequest,
) model.Student {

	if req.NIM != nil {
		current.NIM = strings.TrimSpace(*req.NIM)
	}

	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}

	if req.Grade != nil {
		current.Grade = *req.Grade
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current
}

// IsEmptyStudentPatch memeriksa apakah PATCH
// tidak mengirim field apa pun.
func IsEmptyStudentPatch(
	req model.PatchStudentRequest,
) bool {

	return req.NIM == nil &&
		req.Name == nil &&
		req.Grade == nil &&
		req.IsActive == nil
}

// CalculateTotalPages masih dipakai oleh pagination lama.
// Akan dihapus/diganti ketika cursor pagination selesai.
func CalculateTotalPages(
	limit int,
	total int,
) int {

	if total == 0 {
		return 0
	}

	return (total + limit - 1) / limit
}