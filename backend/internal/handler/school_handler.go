package handler

import (
	"bank-sampah-backend/internal/model"
	"bank-sampah-backend/internal/repository"
	"bank-sampah-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SchoolHandler struct {
	schoolRepo *repository.SchoolRepository
}

func NewSchoolHandler(schoolRepo *repository.SchoolRepository) *SchoolHandler {
	return &SchoolHandler{schoolRepo: schoolRepo}
}

// ListSchools returns all registered schools
// GET /api/v1/schools
func (h *SchoolHandler) ListSchools(c *fiber.Ctx) error {
	schools, err := h.schoolRepo.FindAll()
	if err != nil {
		return response.InternalError(c, "Gagal mengambil data sekolah")
	}
	return response.Success(c, "Daftar sekolah", schools)
}

// GetSchool returns a single school by ID
// GET /api/v1/schools/:id
func (h *SchoolHandler) GetSchool(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "ID tidak valid")
	}

	school, err := h.schoolRepo.FindByID(id)
	if err != nil {
		return response.NotFound(c, "Sekolah tidak ditemukan")
	}

	return response.Success(c, "Detail sekolah", school)
}

type CreateSchoolRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	CallbackURL string `json:"callback_url"`
}

// CreateSchool registers a new school and generates API credentials
// POST /api/v1/schools
func (h *SchoolHandler) CreateSchool(c *fiber.Ctx) error {
	var req CreateSchoolRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Format request tidak valid")
	}

	if req.Name == "" || req.Code == "" {
		return response.BadRequest(c, "Nama dan kode sekolah wajib diisi")
	}

	school := &model.School{
		Name:        req.Name,
		Code:        req.Code,
		CallbackURL: req.CallbackURL,
		IsActive:    true,
	}

	if err := h.schoolRepo.Create(school); err != nil {
		return response.InternalError(c, "Gagal mendaftarkan sekolah: "+err.Error())
	}

	// Return with API secret visible (only on creation!)
	return response.Created(c, "Sekolah berhasil didaftarkan. Simpan API Secret dengan aman!", fiber.Map{
		"id":          school.ID,
		"name":        school.Name,
		"code":        school.Code,
		"api_key":     school.APIKey,
		"api_secret":  school.APISecret,
		"callback_url": school.CallbackURL,
	})
}

type UpdateSchoolRequest struct {
	Name        string `json:"name"`
	CallbackURL string `json:"callback_url"`
	IsActive    *bool  `json:"is_active"`
}

// UpdateSchool updates school info
// PUT /api/v1/schools/:id
func (h *SchoolHandler) UpdateSchool(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "ID tidak valid")
	}

	school, err := h.schoolRepo.FindByID(id)
	if err != nil {
		return response.NotFound(c, "Sekolah tidak ditemukan")
	}

	var req UpdateSchoolRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Format request tidak valid")
	}

	if req.Name != "" {
		school.Name = req.Name
	}
	if req.CallbackURL != "" {
		school.CallbackURL = req.CallbackURL
	}
	if req.IsActive != nil {
		school.IsActive = *req.IsActive
	}

	if err := h.schoolRepo.Update(school); err != nil {
		return response.InternalError(c, "Gagal mengupdate sekolah")
	}

	return response.Success(c, "Sekolah berhasil diupdate", school)
}

// RegenerateCredentials creates new API key/secret for a school
// POST /api/v1/schools/:id/regenerate
func (h *SchoolHandler) RegenerateCredentials(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "ID tidak valid")
	}

	school, err := h.schoolRepo.RegenerateCredentials(id)
	if err != nil {
		return response.InternalError(c, "Gagal regenerate kredensial")
	}

	return response.Success(c, "Kredensial API berhasil di-regenerate. Simpan API Secret dengan aman!", fiber.Map{
		"id":         school.ID,
		"api_key":    school.APIKey,
		"api_secret": school.APISecret,
	})
}

// DeleteSchool performs soft delete on a school
// DELETE /api/v1/schools/:id
func (h *SchoolHandler) DeleteSchool(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "ID tidak valid")
	}

	school, err := h.schoolRepo.FindByID(id)
	if err != nil {
		return response.NotFound(c, "Sekolah tidak ditemukan")
	}

	if err := h.schoolRepo.Delete(school.ID); err != nil {
		return response.InternalError(c, "Gagal menghapus sekolah")
	}

	return response.Success(c, "Sekolah berhasil dihapus", nil)
}
