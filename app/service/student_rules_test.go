package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

func TestValidateCreateRequest(t *testing.T) {

	req := model.CreateStudentRequest{
		NIM:      "",
		Name:     "",
		Grade:    120,
		IsActive: true,
	}

	errors := helper.ValidateStruct(req)

	if len(errors) != 3 {
		t.Fatalf(
			"expected 3 errors, got %d",
			len(errors),
		)
	}
}

func TestValidatePatchRequest(t *testing.T) {

	name := ""

	req := model.PatchStudentRequest{
		Name: &name,
	}

	errors := helper.ValidateStruct(req)

	if len(errors) != 1 {
		t.Fatalf(
			"expected 1 error, got %d",
			len(errors),
		)
	}
}

func TestCalculateTotalPages(t *testing.T) {

	tests := []struct {
		total  int
		limit  int
		expect int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{25, 10, 3},
	}

	for _, test := range tests {

		result := CalculateTotalPages(
			test.limit,
			test.total,
		)

		if result != test.expect {

			t.Errorf(
				"expected %d, got %d",
				test.expect,
				result,
			)
		}
	}
}
