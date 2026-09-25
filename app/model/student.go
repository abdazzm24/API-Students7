package model

type Student struct {
	ID       int     `json:"id"`
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
	OwnerID  int     `json:"owner_id"`
}

// Request POST /students
type CreateStudentRequest struct {
	NIM      string  `json:"nim" validate:"required,nim"`
	Name     string  `json:"name" validate:"required,notblank,min=2,max=100"`
	Grade    float64 `json:"grade" validate:"min=0,max=100"`
	IsActive bool    `json:"is_active"`
}

// Request PUT /students/:id
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim" validate:"required,nim"`
	Name     string  `json:"name" validate:"required,notblank,min=2,max=100"`
	Grade    float64 `json:"grade" validate:"min=0,max=100"`
	IsActive bool    `json:"is_active"`
}

// Request PATCH /students/:id
//
// Pointer digunakan agar dapat membedakan:
// nil       = field tidak dikirim
// non-nil   = field dikirim
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty" validate:"omitnil,nim"`
	Name     *string  `json:"name,omitempty" validate:"omitnil,notblank,min=2,max=100"`
	Grade    *float64 `json:"grade,omitempty" validate:"omitnil,min=0,max=100"`
	IsActive *bool    `json:"is_active,omitempty"`
}

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// Pagination lama.
// Masih dipertahankan sampai tahap cursor pagination.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}