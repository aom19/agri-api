package handlers

import (
	"agri-api/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileService *usecase.ProfileService
}

func NewProfileHandler(profileService *usecase.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

// GetProfile returnează profilul utilizatorului autentificat
// @Summary      Get profile
// @Tags         profile
// @Produce      json
// @Success      200 {object} domain.UserProfile
// @Failure      401 {object} object{error=string}
// @Router       /profile [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID := mustUserID(c)
	profile, err := h.profileService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Nu s-a putut încărca profilul"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// UpdateProfile actualizează datele de profil
// @Summary      Update profile
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        body body object{first_name=string,last_name=string,date_of_birth=string} true "Date profil"
// @Success      200 {object} domain.UserProfile
// @Failure      400 {object} object{errors=object}
// @Router       /profile [patch]
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID := mustUserID(c)

	var req struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		DateOfBirth string `json:"date_of_birth"` // format: YYYY-MM-DD
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors(err)})
		return
	}

	var dob *time.Time
	if req.DateOfBirth != "" {
		t, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errors": gin.H{"date_of_birth": "Format invalid, folosește YYYY-MM-DD"}})
			return
		}
		dob = &t
	}

	profile, err := h.profileService.UpdateProfile(userID, req.FirstName, req.LastName, dob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Nu s-a putut actualiza profilul"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// UploadPhoto încarcă poza de profil
// @Summary      Upload profile photo
// @Tags         profile
// @Accept       multipart/form-data
// @Produce      json
// @Param        photo formData file true "Fișier imagine (jpg/png/webp, max 5MB)"
// @Success      200 {object} object{profile_photo=string}
// @Failure      400 {object} object{error=string}
// @Router       /profile/photo [post]
func (h *ProfileHandler) UploadPhoto(c *gin.Context) {
	userID := mustUserID(c)

	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Niciun fișier trimis"})
		return
	}
	defer file.Close()

	url, err := h.profileService.UploadPhoto(userID, file, header)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile_photo": url})
}

// mustUserID extrage userID din contextul JWT (setat de middleware)
func mustUserID(c *gin.Context) int64 {
	sub, _ := c.Get("user_id")
	id, _ := strconv.ParseInt(sub.(string), 10, 64)
	return id
}
